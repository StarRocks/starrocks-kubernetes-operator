package fe

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
)

const (
	testFeSTSName       = "kube-starrocks-fe"
	testFeSearchService = "kube-starrocks-fe-search"
	testNamespace       = "default"
)

func feHost(ordinal string) string {
	return testFeSTSName + "-" + ordinal + "." + testFeSearchService + "." + testNamespace + ".svc.cluster.local"
}

// showFrontendsColumns are the SHOW FRONTENDS columns of StarRocks 3.x.
var showFrontendsColumns = []string{
	"Id", "Name", "IP", "EditLogPort", "HttpPort", "QueryPort", "RpcPort", "Role", "ClusterId", "Join", "Alive",
	"ReplayedJournalId", "LastHeartbeat", "IsHelper", "ErrMsg", "StartTime", "Version",
}

func addFrontendRow(rows *sqlmock.Rows, host, role, alive string) {
	rows.AddRow("1", "fe_"+host, host, "9010", "8030", "9030", "9020", role, "123", "true", alive,
		"100", "2026-11-03 10:00:00", "true", "", "2026-11-01 10:00:00", "3.5.0")
}

func TestQueryShowFrontends(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows(showFrontendsColumns)
	addFrontendRow(rows, feHost("0"), "LEADER", "true")
	addFrontendRow(rows, feHost("1"), "FOLLOWER", "true")
	addFrontendRow(rows, feHost("2"), "FOLLOWER", "false")
	mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)

	frontends, err := QueryShowFrontends(context.Background(), db)
	require.NoError(t, err)
	assert.Equal(t, []Frontend{
		{Host: feHost("0"), EditLogPort: "9010", Role: "LEADER", Alive: true},
		{Host: feHost("1"), EditLogPort: "9010", Role: "FOLLOWER", Alive: true},
		{Host: feHost("2"), EditLogPort: "9010", Role: "FOLLOWER", Alive: false},
	}, frontends)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryShowFrontends_ColumnsMatchedByName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	// Older versions have no Id column, and the order of the columns must not matter.
	rows := sqlmock.NewRows([]string{"Alive", "Name", "Role", "EditLogPort", "IP"}).
		AddRow("false", "fe_2", "observer", "9010", feHost("2"))
	mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)

	frontends, err := QueryShowFrontends(context.Background(), db)
	require.NoError(t, err)
	assert.Equal(t, []Frontend{{Host: feHost("2"), EditLogPort: "9010", Role: "OBSERVER", Alive: false}}, frontends)
}

func TestFrontendsToDrop(t *testing.T) {
	dead := func(host, role string) Frontend {
		return Frontend{Host: host, EditLogPort: "9010", Role: role}
	}
	tests := []struct {
		name      string
		frontends []Frontend
		replicas  int32
		want      []string
	}{
		{
			name: "nothing to drop when every frontend is in range",
			frontends: []Frontend{
				{Host: feHost("0"), EditLogPort: "9010", Role: "LEADER", Alive: true},
				{Host: feHost("1"), EditLogPort: "9010", Role: "FOLLOWER", Alive: true},
				{Host: feHost("2"), EditLogPort: "9010", Role: "FOLLOWER", Alive: true},
			},
			replicas: 3,
		},
		{
			name: "drops the dead followers beyond replicas, highest ordinal first",
			frontends: []Frontend{
				{Host: feHost("0"), EditLogPort: "9010", Role: "LEADER", Alive: true},
				dead(feHost("3"), "FOLLOWER"),
				dead(feHost("4"), "OBSERVER"),
				dead(feHost("10"), "FOLLOWER"),
			},
			replicas: 3,
			want:     []string{feHost("10"), feHost("4"), feHost("3")},
		},
		{
			name: "keeps a dead frontend whose pod is still part of the statefulset",
			frontends: []Frontend{
				{Host: feHost("0"), EditLogPort: "9010", Role: "LEADER", Alive: true},
				dead(feHost("1"), "FOLLOWER"),
			},
			replicas: 3,
		},
		{
			name: "keeps a frontend beyond replicas that FE still reports as alive",
			frontends: []Frontend{
				{Host: feHost("2"), EditLogPort: "9010", Role: "FOLLOWER", Alive: true},
			},
			replicas: 2,
		},
		{
			name:      "never drops the leader",
			frontends: []Frontend{dead(feHost("2"), "LEADER")},
			replicas:  2,
		},
		{
			name: "ignores frontends that do not belong to this statefulset",
			frontends: []Frontend{
				dead("other-fe-5."+testFeSearchService+"."+testNamespace+".svc.cluster.local", "FOLLOWER"),
				dead(testFeSTSName+"-5.other-search."+testNamespace+".svc.cluster.local", "FOLLOWER"),
				dead(testFeSTSName+"-5."+testFeSearchService+".other-namespace.svc.cluster.local", "FOLLOWER"),
				dead(testFeSTSName+"-proxy-5."+testFeSearchService+"."+testNamespace+".svc.cluster.local", "FOLLOWER"),
				dead("10.0.0.5", "FOLLOWER"),
				dead(feHost("05"), "FOLLOWER"),
				dead(feHost("-5"), "FOLLOWER"),
				dead(testFeSTSName+"-5", "FOLLOWER"),
			},
			replicas: 3,
		},
		{
			name:      "accepts a host without the cluster domain",
			frontends: []Frontend{dead(testFeSTSName+"-3."+testFeSearchService+"."+testNamespace, "FOLLOWER")},
			replicas:  3,
			want:      []string{testFeSTSName + "-3." + testFeSearchService + "." + testNamespace},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, frontend := range frontendsToDrop(tt.frontends, testFeSTSName, testFeSearchService, testNamespace, tt.replicas) {
				got = append(got, frontend.Host)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDropFrontendStatement(t *testing.T) {
	statement, err := dropFrontendStatement(Frontend{Host: feHost("3"), EditLogPort: "9010", Role: "FOLLOWER"})
	require.NoError(t, err)
	assert.Equal(t, `ALTER SYSTEM DROP FOLLOWER "`+feHost("3")+`:9010"`, statement)

	statement, err = dropFrontendStatement(Frontend{Host: feHost("3"), EditLogPort: "9010", Role: "OBSERVER"})
	require.NoError(t, err)
	assert.Equal(t, `ALTER SYSTEM DROP OBSERVER "`+feHost("3")+`:9010"`, statement)

	_, err = dropFrontendStatement(Frontend{Host: feHost("3"), EditLogPort: "9010", Role: "LEADER"})
	assert.Error(t, err)
}

func TestFeDSN(t *testing.T) {
	assert.Equal(t, "root:pwd@tcp(kube-starrocks-fe-service.default:9030)/",
		feDSN("pwd", "kube-starrocks-fe-service", "default", "9030", SSLModeDisabled))
	assert.Equal(t, "root:@tcp(kube-starrocks-fe-service.default:9030)/?tls=skip-verify",
		feDSN("", "kube-starrocks-fe-service", "default", "9030", SSLModeRequired))
}

const testMetaDir = "/opt/starrocks/fe/meta"

func feStatefulSet(replicas int32, revision string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.StatefulSetKind,
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: testFeSTSName, Namespace: testNamespace},
		Spec: appsv1.StatefulSetSpec{
			Replicas: rutils.GetInt32Pointer(replicas),
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name: "fe",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "fe-meta", MountPath: testMetaDir},
					{Name: "fe-log", MountPath: "/opt/starrocks/fe/log"},
				},
			}}}},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "fe-log"}},
			},
		},
		Status: appsv1.StatefulSetStatus{UpdateRevision: revision},
	}
}

func fePods(revisions []string, ready bool) *corev1.PodList {
	pods := &corev1.PodList{}
	for i, revision := range revisions {
		pods.Items = append(pods.Items, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   testFeSTSName + "-" + strconv.Itoa(i),
				Labels: map[string]string{"controller-revision-hash": revision},
			},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Ready: ready}}},
		})
	}
	return pods
}

func pvc(name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
}

func pvcExists(t *testing.T, k8sClient client.Client, name string) bool {
	err := k8sClient.Get(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: name},
		&corev1.PersistentVolumeClaim{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func TestMetaVolumeClaimTemplates(t *testing.T) {
	assert.Equal(t, []string{"fe-meta"}, metaVolumeClaimTemplates(feStatefulSet(3, ""), testMetaDir))
	assert.Equal(t, []string{"fe-meta"}, metaVolumeClaimTemplates(feStatefulSet(3, ""), testMetaDir+"/"))

	// meta on an emptyDir: there is no PVC to clean up
	sts := feStatefulSet(3, "")
	sts.Spec.VolumeClaimTemplates = sts.Spec.VolumeClaimTemplates[1:]
	assert.Empty(t, metaVolumeClaimTemplates(sts, testMetaDir))
}

func TestFeController_SyncFrontendsInFE(t *testing.T) {
	newDB := func(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })
		return db, mock
	}

	t.Run("scale-in from 3 to 2 drops fe-2 and deletes its meta pvc", func(t *testing.T) {
		k8sClient := fake.NewFakeClient(srapi.Scheme,
			pvc("fe-meta-"+testFeSTSName+"-0"), pvc("fe-meta-"+testFeSTSName+"-1"), pvc("fe-meta-"+testFeSTSName+"-2"),
			pvc("fe-log-"+testFeSTSName+"-2"), pvc("fe-meta-other-fe-2"))
		fc := &FeController{Client: k8sClient}
		db, mock := newDB(t)

		rows := sqlmock.NewRows(showFrontendsColumns)
		addFrontendRow(rows, feHost("0"), "LEADER", "true")
		addFrontendRow(rows, feHost("1"), "FOLLOWER", "true")
		addFrontendRow(rows, feHost("2"), "FOLLOWER", "false")
		mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)
		mock.ExpectExec(`ALTER SYSTEM DROP FOLLOWER "` + feHost("2") + `:9010"`).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(2, ""), feStatefulSet(2, "v2"), fePods([]string{"v2", "v2"}, true), "", db)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())

		assert.False(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-2"))
		// everything else is left alone: in-range meta, the log volume, other StatefulSets
		assert.True(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-0"))
		assert.True(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-1"))
		assert.True(t, pvcExists(t, k8sClient, "fe-log-"+testFeSTSName+"-2"))
		assert.True(t, pvcExists(t, k8sClient, "fe-meta-other-fe-2"))
	})

	t.Run("a missing meta pvc does not block the drop", func(t *testing.T) {
		fc := &FeController{Client: fake.NewFakeClient(srapi.Scheme)}
		db, mock := newDB(t)

		rows := sqlmock.NewRows(showFrontendsColumns)
		addFrontendRow(rows, feHost("0"), "LEADER", "true")
		addFrontendRow(rows, feHost("1"), "OBSERVER", "false")
		mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)
		mock.ExpectExec(`ALTER SYSTEM DROP OBSERVER "` + feHost("1") + `:9010"`).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(1, ""), feStatefulSet(1, "v2"), fePods([]string{"v2"}, true), "", db)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("does nothing when there is no stale frontend", func(t *testing.T) {
		k8sClient := fake.NewFakeClient(srapi.Scheme, pvc("fe-meta-"+testFeSTSName+"-1"))
		fc := &FeController{Client: k8sClient}
		db, mock := newDB(t)

		rows := sqlmock.NewRows(showFrontendsColumns)
		addFrontendRow(rows, feHost("0"), "LEADER", "true")
		addFrontendRow(rows, feHost("1"), "FOLLOWER", "false") // in range: a restarting FE, not a scale-in
		mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)

		err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(2, ""), feStatefulSet(2, "v2"), fePods([]string{"v2", "v2"}, true), "", db)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
		assert.True(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-1"))
	})

	t.Run("does nothing when FE does not report the pods by FQDN", func(t *testing.T) {
		k8sClient := fake.NewFakeClient(srapi.Scheme, pvc("fe-meta-"+testFeSTSName+"-2"))
		fc := &FeController{Client: k8sClient}
		db, mock := newDB(t)

		rows := sqlmock.NewRows(showFrontendsColumns)
		addFrontendRow(rows, "10.0.0.10", "LEADER", "true")
		addFrontendRow(rows, "10.0.0.11", "FOLLOWER", "true")
		addFrontendRow(rows, feHost("2"), "FOLLOWER", "false")
		mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)

		err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(2, ""), feStatefulSet(2, "v2"), fePods([]string{"v2", "v2"}, true), "", db)
		require.ErrorIs(t, err, ErrPodsNotRegistered)
		require.NoError(t, mock.ExpectationsWereMet())
		assert.True(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-2"))
	})

	t.Run("a failed drop is returned and retried by the next reconcile", func(t *testing.T) {
		k8sClient := fake.NewFakeClient(srapi.Scheme, pvc("fe-meta-"+testFeSTSName+"-1"))
		fc := &FeController{Client: k8sClient}
		db, mock := newDB(t)

		for _, result := range []error{errors.New("boom"), nil} {
			rows := sqlmock.NewRows(showFrontendsColumns)
			addFrontendRow(rows, feHost("0"), "LEADER", "true")
			addFrontendRow(rows, feHost("1"), "FOLLOWER", "false")
			mock.ExpectQuery(ShowFrontendsStatement).WillReturnRows(rows)
			exec := mock.ExpectExec(`ALTER SYSTEM DROP FOLLOWER "` + feHost("1") + `:9010"`)
			if result != nil {
				exec.WillReturnError(result)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 0))
			}
		}

		err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(1, ""), feStatefulSet(1, "v2"), fePods([]string{"v2"}, true), "", db)
		require.EqualError(t, err, "boom")
		assert.False(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-1"))

		err = fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
			feStatefulSet(1, ""), feStatefulSet(1, "v2"), fePods([]string{"v2"}, true), "", db)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	notSettled := []struct {
		name      string
		expectSTS *appsv1.StatefulSet
		actualSTS *appsv1.StatefulSet
		pods      *corev1.PodList
		wantErr   error
	}{
		{
			name:      "statefulset replicas not updated yet",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(3, "v2"),
			pods:      fePods([]string{"v2", "v2", "v2"}, true),
			wantErr:   ErrReplicasNotEqual,
		},
		{
			name:      "removed pod is still terminating",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(2, "v2"),
			pods:      fePods([]string{"v2", "v2", "v2"}, true),
			wantErr:   ErrReplicasNotEqual,
		},
		{
			name:      "no pods listed",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(2, "v2"),
			pods:      nil,
			wantErr:   ErrReplicasNotEqual,
		},
		{
			name:      "pods are not the expected ordinals",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(2, "v2"),
			pods: func() *corev1.PodList {
				pods := fePods([]string{"v2", "v2"}, true)
				pods.Items[1].Name = testFeSTSName + "-2"
				return pods
			}(),
			wantErr: ErrReplicasNotEqual,
		},
		{
			name:      "rolling update in progress",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(2, "v2"),
			pods:      fePods([]string{"v2", "v1"}, true),
			wantErr:   ErrHashValueNotEqual,
		},
		{
			name:      "a remaining pod is not ready",
			expectSTS: feStatefulSet(2, ""),
			actualSTS: feStatefulSet(2, "v2"),
			pods:      fePods([]string{"v2", "v2"}, false),
			wantErr:   ErrPodsNotReady,
		},
	}
	for _, tt := range notSettled {
		t.Run("waits: "+tt.name, func(t *testing.T) {
			k8sClient := fake.NewFakeClient(srapi.Scheme, pvc("fe-meta-"+testFeSTSName+"-2"))
			fc := &FeController{Client: k8sClient}
			db, mock := newDB(t)

			err := fc.SyncFrontendsInFE(context.Background(), testFeSearchService, testMetaDir,
				tt.expectSTS, tt.actualSTS, tt.pods, "", db)
			require.ErrorIs(t, err, tt.wantErr)
			// nothing is sent to FE and nothing is deleted before the scale-in has settled
			require.NoError(t, mock.ExpectationsWereMet())
			assert.True(t, pvcExists(t, k8sClient, "fe-meta-"+testFeSTSName+"-2"))
		})
	}
}
