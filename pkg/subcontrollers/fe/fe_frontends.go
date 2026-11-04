package fe

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

const (
	ShowFrontendsStatement = "SHOW FRONTENDS"

	frontendRoleLeader   = "LEADER"
	frontendRoleFollower = "FOLLOWER"
	frontendRoleObserver = "OBSERVER"
)

var (
	ErrReplicasNotEqual  = errors.New("replicas not equal")
	ErrHashValueNotEqual = errors.New("hash value not equal")
	ErrPodsNotReady      = errors.New("not all fe pods are ready")
	ErrPodsNotRegistered = errors.New("not all fe pods are registered in FE by their FQDN")
)

// Frontend is one row of SHOW FRONTENDS.
type Frontend struct {
	// Host is the IP column. With HOST_TYPE=FQDN it is the FQDN of the FE pod, e.g.
	// kube-starrocks-fe-2.kube-starrocks-fe-search.default.svc.cluster.local.
	Host        string
	EditLogPort string
	// Role is LEADER, FOLLOWER or OBSERVER.
	Role  string
	Alive bool

	index int // the pod ordinal parsed from Host, set by frontendsToDrop
}

// QueryShowFrontends runs SHOW FRONTENDS and returns one Frontend per row. Columns are matched by
// name, so it works across StarRocks versions that add or reorder columns.
func QueryShowFrontends(ctx context.Context, db *sql.DB) ([]Frontend, error) {
	rows, err := db.QueryContext(ctx, ShowFrontendsStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var frontends []Frontend
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		frontend := Frontend{}
		for i, col := range columns {
			switch col {
			case "IP":
				frontend.Host = string(values[i])
			case "EditLogPort":
				frontend.EditLogPort = string(values[i])
			case "Role":
				frontend.Role = strings.ToUpper(string(values[i]))
			case "Alive":
				frontend.Alive = strings.EqualFold(string(values[i]), "true")
			}
		}
		frontends = append(frontends, frontend)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return frontends, nil
}

// dropFrontendStatement returns the statement that removes frontend from the FE cluster.
func dropFrontendStatement(frontend Frontend) (string, error) {
	switch frontend.Role {
	case frontendRoleFollower, frontendRoleObserver:
		return fmt.Sprintf("ALTER SYSTEM DROP %s \"%s:%s\"", frontend.Role, frontend.Host, frontend.EditLogPort), nil
	default:
		return "", fmt.Errorf("can not drop frontend %s:%s with role %q", frontend.Host, frontend.EditLogPort, frontend.Role)
	}
}

// frontendOrdinal returns the ordinal of the StatefulSet pod that host belongs to. host must be the FQDN of a pod
// of this StatefulSet, i.e. <stsName>-<ordinal>.<searchServiceName>.<namespace>[.svc.<cluster domain>].
func frontendOrdinal(host, stsName, searchServiceName, namespace string) (int, bool) {
	podName, rest, found := strings.Cut(host, ".")
	if !found || !strings.HasPrefix("."+rest+".", "."+searchServiceName+"."+namespace+".") {
		return 0, false
	}
	return podOrdinal(podName, stsName)
}

// podOrdinal returns the ordinal of podName, which must be <stsName>-<ordinal>.
func podOrdinal(podName, stsName string) (int, bool) {
	ordinal, found := strings.CutPrefix(podName, stsName+"-")
	if !found {
		return 0, false
	}
	index, err := strconv.Atoi(ordinal)
	if err != nil || index < 0 || strconv.Itoa(index) != ordinal {
		return 0, false
	}
	return index, true
}

// frontendsToDrop returns the frontends that were removed from the FE StatefulSet by a scale-in but are
// still registered in the FE cluster, ordered from the highest pod ordinal to the lowest.
//
// A frontend is only returned when all of these hold:
//   - its host is the FQDN of a pod of this StatefulSet (see frontendOrdinal), so FEs added by hand or belonging
//     to another StatefulSet are never touched;
//   - the ordinal is >= replicas, so the StatefulSet no longer runs that pod;
//   - FE reports it as not alive;
//   - it is not the leader.
func frontendsToDrop(frontends []Frontend, stsName, searchServiceName, namespace string, replicas int32) []Frontend {
	var result []Frontend
	for _, frontend := range frontends {
		index, ok := frontendOrdinal(frontend.Host, stsName, searchServiceName, namespace)
		if !ok || index < int(replicas) || frontend.Alive || frontend.Role == frontendRoleLeader {
			continue
		}
		frontend.index = index
		result = append(result, frontend)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].index > result[j].index
	})
	return result
}

// allPodsRegistered reports whether every pod the StatefulSet runs shows up in SHOW FRONTENDS by its FQDN. If not,
// FE does not identify its nodes the way frontendOrdinal expects (e.g. it does not run with HOST_TYPE=FQDN), and
// the operator must not guess which frontends are stale.
func allPodsRegistered(frontends []Frontend, stsName, searchServiceName, namespace string, replicas int32) bool {
	registered := make(map[int]bool)
	for _, frontend := range frontends {
		if index, ok := frontendOrdinal(frontend.Host, stsName, searchServiceName, namespace); ok {
			registered[index] = true
		}
	}
	for i := 0; i < int(replicas); i++ {
		if !registered[i] {
			return false
		}
	}
	return true
}

// metaVolumeClaimTemplates returns the names of the volume claim templates of sts that are mounted at metaDir.
func metaVolumeClaimTemplates(sts *appsv1.StatefulSet, metaDir string) []string {
	if len(sts.Spec.Template.Spec.Containers) == 0 {
		return nil
	}
	mounted := make(map[string]bool)
	for _, mount := range sts.Spec.Template.Spec.Containers[0].VolumeMounts {
		if path.Clean(mount.MountPath) == path.Clean(metaDir) {
			mounted[mount.Name] = true
		}
	}
	var names []string
	for i := range sts.Spec.VolumeClaimTemplates {
		if name := sts.Spec.VolumeClaimTemplates[i].Name; mounted[name] {
			names = append(names, name)
		}
	}
	return names
}

// feDSN builds the go-sql-driver DSN used to connect to FE through its external service.
func feDSN(rootPassword, feServiceName, namespace, queryPort, sslMode string) string {
	return fmt.Sprintf("root:%s@tcp(%s.%s:%s)/%s",
		rootPassword, feServiceName, namespace, queryPort, SSLModeDSNSuffix(sslMode))
}

// syncFrontendsInFE looks up the actual FE StatefulSet and pods, and drops the FE nodes that a scale-in removed.
func (fc *FeController) syncFrontendsInFE(ctx context.Context, src *srapi.StarRocksCluster,
	feConfig map[string]interface{}, expectSTS *appsv1.StatefulSet) error {
	feSpec := src.Spec.StarRocksFeSpec

	var actualSTS appsv1.StatefulSet
	if err := fc.Client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: expectSTS.Name}, &actualSTS); err != nil {
		return err
	}

	var actualPods corev1.PodList
	if err := fc.Client.List(ctx, &actualPods, client.InNamespace(src.Namespace),
		client.MatchingLabels(load.Selector(src.Name, feSpec))); err != nil {
		return err
	}

	rootPassword := ""
	if len(expectSTS.Spec.Template.Spec.Containers) > 0 {
		for _, envVar := range expectSTS.Spec.Template.Spec.Containers[0].Env {
			if envVar.Name == "MYSQL_PWD" {
				var err error
				rootPassword, err = k8sutils.GetEnvVarValue(ctx, fc.Client, src.Namespace, envVar)
				if err != nil {
					return err
				}
			}
		}
	}
	dsn := feDSN(rootPassword, service.ExternalServiceName(src.Name, feSpec), src.Namespace,
		strconv.Itoa(int(rutils.GetPort(feConfig, rutils.QUERY_PORT))), config.FeSslMode)

	return fc.SyncFrontendsInFE(ctx, service.SearchServiceName(src.Name, feSpec), pod.GetStorageDir(feSpec),
		expectSTS, &actualSTS, &actualPods, dsn, nil)
}

// SyncFrontendsInFE drops the FE nodes that are still registered in the FE cluster after a scale-in removed
// their pods.
//
// Reducing the FE replicas only shrinks the StatefulSet. The removed FE stays in SHOW FRONTENDS as a dead member
// of the BDBJE replication group, and BDBJE can not clean its old journal files (meta/bdb/*.jdb) until every
// registered member has caught up, so fe-meta keeps growing until FE runs out of disk.
//
// Like the compute node cleanup, it only acts once the scale-in has settled: the StatefulSet matches the expected
// one, it runs exactly the expected number of pods, and every pod is ready and on the current revision. Then, for
// each frontend selected by frontendsToDrop, it deletes the PVCs of that pod's FE meta volume and runs ALTER SYSTEM
// DROP FOLLOWER/OBSERVER.
//
// The meta PVC has to go too: StatefulSets keep PVCs on scale-in, and a dropped FE can not rejoin the cluster with
// its old meta (StarRocks requires an empty meta dir and --helper). Without the PVC, a later scale-out starts that
// pod with no meta, and the FE entrypoint adds it back with ALTER SYSTEM ADD FOLLOWER. The PVC is deleted before the
// DROP so that both steps are retried by the next reconcile if either fails. If db is nil, a connection is opened
// with dsn.
func (fc *FeController) SyncFrontendsInFE(ctx context.Context, searchServiceName, metaDir string,
	expectSTS *appsv1.StatefulSet, actualSTS *appsv1.StatefulSet, actualPods *corev1.PodList,
	dsn string, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx)

	expectReplicas, err := waitForScaleInSettled(ctx, expectSTS, actualSTS, actualPods)
	if err != nil {
		return err
	}

	if db == nil {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
	}

	frontends, err := QueryShowFrontends(ctx, db)
	if err != nil {
		logger.Error(err, "query SHOW FRONTENDS failed", "sql", ShowFrontendsStatement)
		return err
	}

	if !allPodsRegistered(frontends, actualSTS.Name, searchServiceName, actualSTS.Namespace, expectReplicas) {
		logger.Info("fe pods are not registered by their FQDN in SHOW FRONTENDS, skip dropping frontends")
		return ErrPodsNotRegistered
	}

	metaClaims := metaVolumeClaimTemplates(actualSTS, metaDir)
	for _, frontend := range frontendsToDrop(frontends, actualSTS.Name, searchServiceName, actualSTS.Namespace, expectReplicas) {
		statement, err := dropFrontendStatement(frontend)
		if err != nil {
			return err
		}
		for _, claim := range metaClaims {
			pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Namespace: actualSTS.Namespace,
				Name:      fmt.Sprintf("%s-%s-%d", claim, actualSTS.Name, frontend.index),
			}}
			if err = fc.Client.Delete(ctx, pvc); err != nil && !apierrors.IsNotFound(err) {
				logger.Error(err, "delete fe meta pvc failed", "pvc", pvc.Name)
				return err
			}
			logger.Info("delete fe meta pvc of the removed fe pod", "pvc", pvc.Name)
		}
		if _, err = db.ExecContext(ctx, statement); err != nil {
			logger.Error(err, "drop frontend failed", "sql", statement)
			return err
		}
		logger.Info("drop frontend success", "host", frontend.Host, "role", frontend.Role)
	}
	return nil
}

// waitForScaleInSettled returns the expected FE replicas once the StatefulSet matches the expected one and runs
// exactly that many pods, all ready and on the current revision. Otherwise it returns the reason to wait.
func waitForScaleInSettled(ctx context.Context, expectSTS *appsv1.StatefulSet, actualSTS *appsv1.StatefulSet,
	actualPods *corev1.PodList) (int32, error) {
	logger := logr.FromContextOrDiscard(ctx)

	expectReplicas := int32(1)
	if expectSTS.Spec.Replicas != nil {
		expectReplicas = *expectSTS.Spec.Replicas
	}
	stsReplicas := int32(1)
	if actualSTS.Spec.Replicas != nil {
		stsReplicas = *actualSTS.Spec.Replicas
	}
	if expectReplicas != stsReplicas {
		logger.Info("expect replicas is not equal to statefulset replicas", "expectReplicas", expectReplicas, "stsReplicas", stsReplicas)
		return 0, ErrReplicasNotEqual
	}
	if expectHashValue, equal := rutils.StatefulSetDeepEqual(expectSTS, actualSTS); !equal {
		logger.Info("the actual statefulset is not operator expected", "expectHashValue", expectHashValue)
		return 0, ErrHashValueNotEqual
	}
	if actualPods == nil || len(actualPods.Items) != int(expectReplicas) {
		actualReplicas := 0
		if actualPods != nil {
			actualReplicas = len(actualPods.Items)
		}
		logger.Info("the expected number of fe pods is not equal to the actual number",
			"expectReplicas", expectReplicas, "actualReplicas", actualReplicas)
		return 0, ErrReplicasNotEqual
	}
	const controllerRevisionHashKey = "controller-revision-hash"
	for i := range actualPods.Items {
		pod := &actualPods.Items[i]
		if pod.Labels[controllerRevisionHashKey] == "" ||
			pod.Labels[controllerRevisionHashKey] != actualSTS.Status.UpdateRevision {
			logger.Info("there is an old fe pod, continue to wait for the statefulset to be updated", "pod", pod.Name)
			return 0, ErrHashValueNotEqual
		}
		if index, ok := podOrdinal(pod.Name, actualSTS.Name); !ok || index >= int(expectReplicas) {
			logger.Info("fe pod is not one of the expected ordinals, continue to wait", "pod", pod.Name)
			return 0, ErrReplicasNotEqual
		}
		if !k8sutils.PodIsReady(&pod.Status) {
			logger.Info("fe pod is not ready, continue to wait", "pod", pod.Name)
			return 0, ErrPodsNotReady
		}
	}
	return expectReplicas, nil
}
