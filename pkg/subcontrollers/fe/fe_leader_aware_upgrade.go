/*
Copyright 2021-present, StarRocks Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fe

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/mysqlutil"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

// Leader-aware upgrade.
//
// With the default RollingUpdate strategy the StatefulSet controller replaces FE pods from the highest
// ordinal down and knows nothing about BDBJE roles. If the leader sits in the middle of that order, the
// election it triggers can hand leadership to a pod that is itself replaced a moment later, so one
// upgrade causes two leader failovers.
//
// When config.EnableFeLeaderAwareUpgrade is set, the FE StatefulSet is switched to the OnDelete
// strategy, which makes the StatefulSet controller recreate a pod on the new template only after
// something deletes it. ReconcileLeaderAwareUpgrade is that something: on every reconcile it finds the
// FE pods that still run the old revision, asks FE who the leader is, and deletes exactly one pod,
// preferring followers and leaving the leader for last. Because the leader is re-read on every pass,
// a leadership change mid-rollout still leaves the then-current leader for the end.

const (
	// ShowFrontendsStatement lists every FE known to the cluster together with its BDBJE role.
	ShowFrontendsStatement = "SHOW FRONTENDS"

	// FrontendRoleLeader is the Role column value SHOW FRONTENDS reports for the BDBJE master.
	FrontendRoleLeader = "LEADER"

	// EventReasonLeaderAwareUpgrade is the reason on the events emitted while the operator drives an
	// FE rollout.
	EventReasonLeaderAwareUpgrade = "FeLeaderAwareUpgrade"
)

// Frontend is one row of SHOW FRONTENDS, reduced to what the upgrade needs.
type Frontend struct {
	// Name is the FE's own identifier, e.g. "sr-fe-0.sr-fe-search.ns.svc.cluster.local_9010_1760641496296".
	Name string
	// Host is the IP column: the FQDN of the FE pod in FQDN mode, or its IP address in IP mode.
	Host string
	// Role is LEADER, FOLLOWER or OBSERVER.
	Role string
	// Alive reports whether this FE is reachable from the FE that answered the query.
	Alive bool
}

// IsLeader reports whether this frontend is the BDBJE leader.
func (f Frontend) IsLeader() bool {
	return strings.EqualFold(f.Role, FrontendRoleLeader)
}

// FrontendsQuerier answers SHOW FRONTENDS. The production implementation dials FE over MySQL; tests
// substitute a canned answer.
type FrontendsQuerier interface {
	ShowFrontends(ctx context.Context) ([]Frontend, error)
}

// NewFrontendsQuerierFunc builds the querier for one reconcile. sts is the FE StatefulSet as the
// operator expects it, which is where the root password (MYSQL_PWD) is read from; queryPort is the
// FE query port resolved from the FE config.
type NewFrontendsQuerierFunc func(ctx context.Context, k8sClient client.Client, src *srapi.StarRocksCluster,
	sts *appsv1.StatefulSet, queryPort int32) (FrontendsQuerier, error)

// mysqlFrontendsQuerier runs SHOW FRONTENDS against the FE external service.
type mysqlFrontendsQuerier struct {
	dsn string
}

// newMySQLFrontendsQuerier is the production NewFrontendsQuerierFunc. The root password is taken from
// the MYSQL_PWD environment variable on the FE container, the same convention the CN controller uses;
// an FE cluster without MYSQL_PWD is assumed to have an empty root password.
func newMySQLFrontendsQuerier(ctx context.Context, k8sClient client.Client, src *srapi.StarRocksCluster,
	sts *appsv1.StatefulSet, queryPort int32) (FrontendsQuerier, error) {
	logger := logr.FromContextOrDiscard(ctx)
	rootPassword := ""
	if len(sts.Spec.Template.Spec.Containers) > 0 {
		for _, envVar := range sts.Spec.Template.Spec.Containers[0].Env {
			if envVar.Name != "MYSQL_PWD" {
				continue
			}
			value, err := k8sutils.GetEnvVarValue(ctx, k8sClient, src.Namespace, envVar)
			if err != nil {
				logger.Error(err, "failed to get MYSQL_PWD from fe env vars, use the default password: empty string")
				break
			}
			rootPassword = value
		}
	}
	host := service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec) + "." + src.Namespace
	return &mysqlFrontendsQuerier{
		dsn: mysqlutil.DSN(rootPassword, host, strconv.Itoa(int(queryPort)), config.FeSslMode),
	}, nil
}

func (q *mysqlFrontendsQuerier) ShowFrontends(ctx context.Context) ([]Frontend, error) {
	db, err := sql.Open("mysql", q.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, ShowFrontendsStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ParseShowFrontends(rows)
}

// ParseShowFrontends reads SHOW FRONTENDS rows by column name, so that column order and columns
// added by newer FE versions do not matter.
func ParseShowFrontends(rows *sql.Rows) ([]Frontend, error) {
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
		fe := Frontend{}
		for i, col := range columns {
			switch col {
			case "Name":
				fe.Name = string(values[i])
			case "IP":
				fe.Host = string(values[i])
			case "Role":
				fe.Role = string(values[i])
			case "Alive":
				fe.Alive = strings.EqualFold(string(values[i]), "true")
			}
		}
		frontends = append(frontends, fe)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return frontends, nil
}

// LeaderAwareUpgradeEnabled reports whether the operator was started with --enable-fe-leader-aware-upgrade.
func LeaderAwareUpgradeEnabled() bool {
	return config.EnableFeLeaderAwareUpgrade
}

// OnDeleteUpdateStrategy is the update strategy the FE StatefulSet gets while the operator drives the
// rollout. RollingUpdate must be nil: the API server rejects an OnDelete strategy that carries
// rollingUpdate parameters.
func OnDeleteUpdateStrategy() appsv1.StatefulSetUpdateStrategy {
	return appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}
}

// EnsureOnDeleteStrategy switches an existing FE StatefulSet to OnDelete with a targeted merge patch
// before the regular three-way apply runs. The three-way merge only removes fields that appear in the
// last-applied annotation, and a StatefulSet created by an older operator has no such annotation, so
// on its own it would leave rollingUpdate in place, which the API server rejects next to OnDelete.
// A StatefulSet that does not exist yet or is already OnDelete is left alone.
func EnsureOnDeleteStrategy(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	var actual appsv1.StatefulSet
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &actual); err != nil {
		return client.IgnoreNotFound(err)
	}
	if actual.Spec.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType && actual.Spec.UpdateStrategy.RollingUpdate == nil {
		return nil
	}
	logr.FromContextOrDiscard(ctx).Info("switching fe statefulset to the OnDelete update strategy for the leader-aware upgrade",
		"statefulset", name, "previousType", actual.Spec.UpdateStrategy.Type)
	patch := []byte(`{"spec":{"updateStrategy":{"type":"OnDelete","rollingUpdate":null}}}`)
	return k8sClient.Patch(ctx, &actual, client.RawPatch(types.MergePatchType, patch))
}

// ReconcileLeaderAwareUpgrade advances an FE rollout by at most one pod. It is a no-op when every FE
// pod already runs the StatefulSet's update revision, and it waits (returns nil without acting) while
// the StatefulSet controller has not yet observed the new spec or while any FE pod is not ready or is
// still terminating. Otherwise it deletes one pod that is behind: a follower if any follower is
// behind, and the leader only once it is the last one.
//
// It returns an error, which makes the reconcile retry, when FE cannot be asked who the leader is, when
// any FE reports as not alive, or when no FE reports as leader. Nothing is deleted in those cases.
func (fc *FeController) ReconcileLeaderAwareUpgrade(ctx context.Context, src *srapi.StarRocksCluster,
	expectSts *appsv1.StatefulSet, queryPort int32) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("leaderAwareUpgrade")

	var sts appsv1.StatefulSet
	if err := fc.Client.Get(ctx, types.NamespacedName{Namespace: expectSts.Namespace, Name: expectSts.Name}, &sts); err != nil {
		return err
	}
	if sts.Status.ObservedGeneration < sts.Generation || sts.Status.UpdateRevision == "" {
		logger.Info("waiting for the statefulset controller to observe the new spec",
			"generation", sts.Generation, "observedGeneration", sts.Status.ObservedGeneration)
		return nil
	}

	pods, pending, err := fc.listRolloutPods(ctx, src, &sts)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	if reason := rolloutWaitReason(&sts, pods); reason != "" {
		logger.Info(reason)
		return nil
	}

	leaderPod, err := fc.lookupLeaderPod(ctx, src, expectSts, queryPort, pods)
	if err != nil {
		return err
	}

	victim := chooseNextPod(pending, leaderPod)
	role := "follower"
	if victim == leaderPod {
		role = "leader"
	}
	msg := fmt.Sprintf("replacing fe %s pod %s (%d of %d pods still on the old revision, leader is %s)",
		role, victim, len(pending), len(pods), leaderPod)
	logger.Info(msg, "updateRevision", sts.Status.UpdateRevision)
	fc.Recorder.Event(src, corev1.EventTypeNormal, EventReasonLeaderAwareUpgrade, msg)

	for i := range pending {
		if pending[i].Name == victim {
			return client.IgnoreNotFound(fc.Client.Delete(ctx, &pending[i]))
		}
	}
	return nil
}

// listRolloutPods returns the FE pods that take part in the rollout and, among them, the ones not yet
// on the StatefulSet's update revision. Pods above the replica count are on their way out and are left
// to the StatefulSet controller.
func (fc *FeController) listRolloutPods(ctx context.Context, src *srapi.StarRocksCluster,
	sts *appsv1.StatefulSet) (pods, pending []corev1.Pod, err error) {
	var podList corev1.PodList
	if err = fc.Client.List(ctx, &podList, client.InNamespace(src.Namespace),
		client.MatchingLabels(load.Selector(src.Name, src.Spec.StarRocksFeSpec))); err != nil {
		return nil, nil, err
	}
	replicas := statefulSetReplicas(sts)
	for i := range podList.Items {
		ordinal, ok := podOrdinal(sts.Name, podList.Items[i].Name)
		if !ok || ordinal >= replicas {
			continue
		}
		pods = append(pods, podList.Items[i])
		if podList.Items[i].Labels[appsv1.ControllerRevisionHashLabelKey] != sts.Status.UpdateRevision {
			pending = append(pending, podList.Items[i])
		}
	}
	return pods, pending, nil
}

// rolloutWaitReason explains why no pod may be deleted right now, or returns "" when the rollout may
// proceed: every pod the StatefulSet wants must exist, be running, be ready and not be terminating, so
// that at most one FE is ever down because of the operator.
func rolloutWaitReason(sts *appsv1.StatefulSet, pods []corev1.Pod) string {
	if replicas := statefulSetReplicas(sts); len(pods) < int(replicas) {
		return fmt.Sprintf("waiting for the statefulset controller to create all fe pods (%d of %d exist)", len(pods), replicas)
	}
	for i := range pods {
		p := &pods[i]
		if p.DeletionTimestamp != nil || p.Status.Phase != corev1.PodRunning || !k8sutils.PodIsReady(&p.Status) {
			return fmt.Sprintf("waiting for fe pod %s to be ready before replacing the next one", p.Name)
		}
	}
	return ""
}

// lookupLeaderPod asks FE who the leader is and maps the answer onto one of pods.
func (fc *FeController) lookupLeaderPod(ctx context.Context, src *srapi.StarRocksCluster,
	expectSts *appsv1.StatefulSet, queryPort int32, pods []corev1.Pod) (string, error) {
	newQuerier := fc.NewFrontendsQuerier
	if newQuerier == nil {
		newQuerier = newMySQLFrontendsQuerier
	}
	querier, err := newQuerier(ctx, fc.Client, src, expectSts, queryPort)
	if err != nil {
		return "", fmt.Errorf("leader-aware upgrade: build frontends querier: %w", err)
	}
	frontends, err := querier.ShowFrontends(ctx)
	if err != nil {
		return "", fmt.Errorf("leader-aware upgrade: %s failed, no fe pod will be replaced until it succeeds: %w",
			ShowFrontendsStatement, err)
	}
	leaderPod, err := findLeaderPod(frontends, pods)
	if err != nil {
		return "", fmt.Errorf("leader-aware upgrade: %w", err)
	}
	return leaderPod, nil
}

func statefulSetReplicas(sts *appsv1.StatefulSet) int32 {
	if sts.Spec.Replicas == nil {
		return 1
	}
	return *sts.Spec.Replicas
}

// findLeaderPod maps the LEADER row of SHOW FRONTENDS onto one of the FE pods. In FQDN mode the IP
// column is "<pod>.<search service>.<namespace>.svc...", so the first DNS label is the pod name; in
// IP mode it is the pod IP. Every FE must be alive before a pod is replaced: a dead FE plus one
// deliberately deleted pod could cost BDBJE its quorum.
func findLeaderPod(frontends []Frontend, pods []corev1.Pod) (string, error) {
	if len(frontends) == 0 {
		return "", fmt.Errorf("%s returned no rows", ShowFrontendsStatement)
	}
	var leader *Frontend
	for i := range frontends {
		if !frontends[i].Alive {
			return "", fmt.Errorf("fe %s is not alive, refusing to replace any fe pod", frontends[i].Name)
		}
		if frontends[i].IsLeader() {
			if leader != nil {
				return "", fmt.Errorf("%s reports more than one leader (%s, %s)", ShowFrontendsStatement, leader.Name, frontends[i].Name)
			}
			leader = &frontends[i]
		}
	}
	if leader == nil {
		return "", fmt.Errorf("%s reports no leader, the fe cluster may be electing", ShowFrontendsStatement)
	}
	firstLabel := strings.SplitN(leader.Host, ".", 2)[0]
	for i := range pods {
		if pods[i].Name == firstLabel || (pods[i].Status.PodIP != "" && pods[i].Status.PodIP == leader.Host) {
			return pods[i].Name, nil
		}
	}
	return "", fmt.Errorf("leader %s (host %s) does not match any fe pod", leader.Name, leader.Host)
}

// chooseNextPod returns the name of the pod to delete next: the highest-ordinal pending pod that is not
// the leader, mirroring the StatefulSet controller's own order, or the leader when it is the only pod
// left behind.
func chooseNextPod(pending []corev1.Pod, leaderPod string) string {
	sorted := make([]corev1.Pod, len(pending))
	copy(sorted, pending)
	sort.Slice(sorted, func(i, j int) bool {
		// Names share the StatefulSet prefix, so the ordinal is what differs; a numeric compare keeps
		// pod 10 after pod 2.
		oi, _ := ordinalSuffix(sorted[i].Name)
		oj, _ := ordinalSuffix(sorted[j].Name)
		return oi > oj
	})
	for i := range sorted {
		if sorted[i].Name != leaderPod {
			return sorted[i].Name
		}
	}
	return leaderPod
}

// podOrdinal returns the ordinal of a StatefulSet pod named "<stsName>-<n>".
func podOrdinal(stsName, podName string) (int32, bool) {
	if !strings.HasPrefix(podName, stsName+"-") {
		return 0, false
	}
	return ordinalSuffix(podName)
}

func ordinalSuffix(podName string) (int32, bool) {
	idx := strings.LastIndex(podName, "-")
	if idx < 0 {
		return 0, false
	}
	n, err := strconv.ParseInt(podName[idx+1:], 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
}

// PrepareLeaderAwareUpgrade is the pre-apply half of the SyncCluster wiring: when the flag is on it
// forces the expected StatefulSet to OnDelete and converts an existing StatefulSet the same way.
func PrepareLeaderAwareUpgrade(ctx context.Context, k8sClient client.Client, expectSts *appsv1.StatefulSet, enabled bool) error {
	if !enabled {
		return nil
	}
	expectSts.Spec.UpdateStrategy = OnDeleteUpdateStrategy()
	return EnsureOnDeleteStrategy(ctx, k8sClient, expectSts.Namespace, expectSts.Name)
}

// RunLeaderAwareUpgrade is the post-apply half of the SyncCluster wiring: a no-op unless enabled.
func (fc *FeController) RunLeaderAwareUpgrade(ctx context.Context, src *srapi.StarRocksCluster,
	expectSts *appsv1.StatefulSet, feConfig map[string]interface{}, enabled bool) error {
	if !enabled {
		return nil
	}
	return fc.ReconcileLeaderAwareUpgrade(ctx, src, expectSts, resolveFEQueryPort(feConfig))
}

// resolveFEQueryPort is a small helper so SyncCluster does not need to know how ports are resolved.
func resolveFEQueryPort(feConfig map[string]interface{}) int32 {
	return rutils.GetPort(feConfig, rutils.QUERY_PORT)
}
