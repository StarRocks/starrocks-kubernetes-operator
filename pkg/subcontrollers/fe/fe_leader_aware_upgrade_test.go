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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

const (
	testNamespace   = "default"
	testCluster     = "sr"
	testFeSts       = "sr-fe"
	oldRevision     = "sr-fe-aaaaaaaaa"
	newRevision     = "sr-fe-bbbbbbbbb"
	testFeQueryPort = int32(9030)
)

type fakeQuerier struct {
	frontends []Frontend
	err       error
	calls     int
}

func (q *fakeQuerier) ShowFrontends(_ context.Context) ([]Frontend, error) {
	q.calls++
	return q.frontends, q.err
}

func fqdn(podName string) string {
	return fmt.Sprintf("%s.sr-fe-search.%s.svc.cluster.local", podName, testNamespace)
}

func feRow(podName, role string, alive bool) Frontend {
	return Frontend{Name: fqdn(podName) + "_9010_1760641496296", Host: fqdn(podName), Role: role, Alive: alive}
}

type testPod struct {
	name     string
	revision string
	ready    bool
	deleting bool
	ip       string
}

func newTestCluster(replicas int32) *srapi.StarRocksCluster {
	return &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{Name: testCluster, Namespace: testNamespace},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(replicas),
						Image:    "starrocks/fe-ubuntu:3.3.10",
					},
				},
			},
		},
	}
}

// newTestSts builds an OnDelete FE StatefulSet whose controller has observed generation 2 and
// wants every pod on newRevision.
func newTestSts(replicas int32, generation int64) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{Kind: "StatefulSet", APIVersion: appsv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: testFeSts, Namespace: testNamespace, Generation: generation},
		Spec: appsv1.StatefulSetSpec{
			Replicas:       rutils.GetInt32Pointer(replicas),
			UpdateStrategy: OnDeleteUpdateStrategy(),
		},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 2,
			UpdateRevision:     newRevision,
			CurrentRevision:    oldRevision,
		},
	}
}

func newTestPods(src *srapi.StarRocksCluster, specs []testPod) []runtime.Object {
	var objs []runtime.Object
	for _, s := range specs {
		labels := load.Selector(src.Name, src.Spec.StarRocksFeSpec)
		labels[appsv1.ControllerRevisionHashLabelKey] = s.revision
		p := &corev1.Pod{
			TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: s.name, Namespace: testNamespace, Labels: labels},
			Status: corev1.PodStatus{
				Phase:             corev1.PodRunning,
				PodIP:             s.ip,
				ContainerStatuses: []corev1.ContainerStatus{{Name: "fe", Ready: s.ready}},
			},
		}
		if s.deleting {
			now := metav1.Now()
			p.DeletionTimestamp = &now
			p.Finalizers = []string{"test/keep"}
		}
		objs = append(objs, p)
	}
	return objs
}

func podNames(t *testing.T, c client.Client) []string {
	t.Helper()
	var list corev1.PodList
	require.NoError(t, c.List(context.Background(), &list, client.InNamespace(testNamespace)))
	var names []string
	for i := range list.Items {
		if list.Items[i].DeletionTimestamp == nil {
			names = append(names, list.Items[i].Name)
		}
	}
	return names
}

func runUpgrade(t *testing.T, src *srapi.StarRocksCluster, sts *appsv1.StatefulSet, pods []testPod,
	querier *fakeQuerier) (client.Client, error) {
	t.Helper()
	objs := append([]runtime.Object{src, sts}, newTestPods(src, pods)...)
	c := fake.NewFakeClient(srapi.Scheme, objs...)
	fc := &FeController{
		Client:   c,
		Recorder: record.NewFakeRecorder(10),
		NewFrontendsQuerier: func(_ context.Context, _ client.Client, _ *srapi.StarRocksCluster,
			_ *appsv1.StatefulSet, _ int32) (FrontendsQuerier, error) {
			return querier, nil
		},
	}
	err := fc.ReconcileLeaderAwareUpgrade(context.Background(), src, sts.DeepCopy(), testFeQueryPort)
	return c, err
}

func TestReconcileLeaderAwareUpgrade(t *testing.T) {
	allReady := func(revisions ...string) []testPod {
		var pods []testPod
		for i, r := range revisions {
			pods = append(pods, testPod{name: fmt.Sprintf("%s-%d", testFeSts, i), revision: r, ready: true})
		}
		return pods
	}
	leaderAt := func(leader string) []Frontend {
		var rows []Frontend
		for _, p := range []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"} {
			role := "FOLLOWER"
			if p == leader {
				role = FrontendRoleLeader
			}
			rows = append(rows, feRow(p, role, true))
		}
		return rows
	}

	tests := []struct {
		name        string
		sts         *appsv1.StatefulSet
		pods        []testPod
		querier     *fakeQuerier
		wantErr     string
		wantRemain  []string
		wantQueried bool
	}{
		{
			name:       "nothing pending: no pod is deleted and FE is not asked",
			sts:        newTestSts(3, 2),
			pods:       allReady(newRevision, newRevision, newRevision),
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
		},
		{
			name:       "statefulset controller has not observed the new spec yet: wait",
			sts:        newTestSts(3, 3),
			pods:       allReady(oldRevision, oldRevision, oldRevision),
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
		},
		{
			name:        "leader at ordinal 0: highest follower goes first",
			sts:         newTestSts(3, 2),
			pods:        allReady(oldRevision, oldRevision, oldRevision),
			querier:     &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain:  []string{"sr-fe-0", "sr-fe-1"},
			wantQueried: true,
		},
		{
			name:        "leader at ordinal 2: it is skipped, ordinal 1 goes first",
			sts:         newTestSts(3, 2),
			pods:        allReady(oldRevision, oldRevision, oldRevision),
			querier:     &fakeQuerier{frontends: leaderAt("sr-fe-2")},
			wantRemain:  []string{"sr-fe-0", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name:        "leader in the middle with the other follower done: remaining follower goes before the leader",
			sts:         newTestSts(3, 2),
			pods:        allReady(oldRevision, oldRevision, newRevision),
			querier:     &fakeQuerier{frontends: leaderAt("sr-fe-1")},
			wantRemain:  []string{"sr-fe-1", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name:        "leader is the only pod left behind: it is replaced last",
			sts:         newTestSts(3, 2),
			pods:        allReady(newRevision, oldRevision, newRevision),
			querier:     &fakeQuerier{frontends: leaderAt("sr-fe-1")},
			wantRemain:  []string{"sr-fe-0", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name:        "leadership moved mid-rollout: the current leader is still deferred",
			sts:         newTestSts(3, 2),
			pods:        allReady(oldRevision, oldRevision, newRevision),
			querier:     &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain:  []string{"sr-fe-0", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "a pod is not ready: wait, nothing is deleted",
			sts:  newTestSts(3, 2),
			pods: []testPod{
				{name: "sr-fe-0", revision: oldRevision, ready: true},
				{name: "sr-fe-1", revision: oldRevision, ready: true},
				{name: "sr-fe-2", revision: newRevision, ready: false},
			},
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
		},
		{
			name: "a pod is terminating: wait, nothing is deleted",
			sts:  newTestSts(3, 2),
			pods: []testPod{
				{name: "sr-fe-0", revision: oldRevision, ready: true},
				{name: "sr-fe-1", revision: oldRevision, ready: true},
				{name: "sr-fe-2", revision: oldRevision, ready: true, deleting: true},
			},
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1"},
		},
		{
			name:       "a pod is missing: wait for the statefulset controller to recreate it",
			sts:        newTestSts(3, 2),
			pods:       allReady(oldRevision, oldRevision),
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1"},
		},
		{
			name:        "SHOW FRONTENDS fails: error, nothing is deleted",
			sts:         newTestSts(3, 2),
			pods:        allReady(oldRevision, oldRevision, oldRevision),
			querier:     &fakeQuerier{err: errors.New("dial tcp: connection refused")},
			wantErr:     "SHOW FRONTENDS failed",
			wantRemain:  []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "an FE is not alive: error, nothing is deleted",
			sts:  newTestSts(3, 2),
			pods: allReady(oldRevision, oldRevision, oldRevision),
			querier: &fakeQuerier{frontends: []Frontend{
				feRow("sr-fe-0", FrontendRoleLeader, true),
				feRow("sr-fe-1", "FOLLOWER", false),
				feRow("sr-fe-2", "FOLLOWER", true),
			}},
			wantErr:     "is not alive",
			wantRemain:  []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "no leader reported: error, nothing is deleted",
			sts:  newTestSts(3, 2),
			pods: allReady(oldRevision, oldRevision, oldRevision),
			querier: &fakeQuerier{frontends: []Frontend{
				feRow("sr-fe-0", "FOLLOWER", true),
				feRow("sr-fe-1", "FOLLOWER", true),
				feRow("sr-fe-2", "FOLLOWER", true),
			}},
			wantErr:     "reports no leader",
			wantRemain:  []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "leader does not match any pod: error, nothing is deleted",
			sts:  newTestSts(3, 2),
			pods: allReady(oldRevision, oldRevision, oldRevision),
			querier: &fakeQuerier{frontends: []Frontend{
				feRow("other-fe-0", FrontendRoleLeader, true),
			}},
			wantErr:     "does not match any fe pod",
			wantRemain:  []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "single replica: the leader is the only pod and is replaced",
			sts:  newTestSts(1, 2),
			pods: allReady(oldRevision),
			querier: &fakeQuerier{frontends: []Frontend{
				feRow("sr-fe-0", FrontendRoleLeader, true),
			}},
			wantRemain:  nil,
			wantQueried: true,
		},
		{
			name: "IP mode: the leader is matched by pod IP",
			sts:  newTestSts(3, 2),
			pods: []testPod{
				{name: "sr-fe-0", revision: oldRevision, ready: true, ip: "10.0.0.1"},
				{name: "sr-fe-1", revision: oldRevision, ready: true, ip: "10.0.0.2"},
				{name: "sr-fe-2", revision: oldRevision, ready: true, ip: "10.0.0.3"},
			},
			querier: &fakeQuerier{frontends: []Frontend{
				{Name: "10.0.0.1_9010_1", Host: "10.0.0.1", Role: "FOLLOWER", Alive: true},
				{Name: "10.0.0.2_9010_1", Host: "10.0.0.2", Role: "FOLLOWER", Alive: true},
				{Name: "10.0.0.3_9010_1", Host: "10.0.0.3", Role: FrontendRoleLeader, Alive: true},
			}},
			wantRemain:  []string{"sr-fe-0", "sr-fe-2"},
			wantQueried: true,
		},
		{
			name: "pods above the replica count are left to the statefulset controller",
			sts:  newTestSts(2, 2),
			pods: []testPod{
				{name: "sr-fe-0", revision: newRevision, ready: true},
				{name: "sr-fe-1", revision: newRevision, ready: true},
				{name: "sr-fe-2", revision: oldRevision, ready: true},
			},
			querier:    &fakeQuerier{frontends: leaderAt("sr-fe-0")},
			wantRemain: []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newTestCluster(*tt.sts.Spec.Replicas)
			c, err := runUpgrade(t, src, tt.sts, tt.pods, tt.querier)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.ElementsMatch(t, tt.wantRemain, podNames(t, c))
			require.Equal(t, tt.wantQueried, tt.querier.calls > 0, "whether SHOW FRONTENDS was run")
		})
	}
}

// TestReconcileLeaderAwareUpgrade_FullRollout walks a whole upgrade with the leader in the middle and
// checks that the leader is replaced exactly once, at the very end.
func TestReconcileLeaderAwareUpgrade_FullRollout(t *testing.T) {
	src := newTestCluster(3)
	sts := newTestSts(3, 2)
	querier := &fakeQuerier{frontends: []Frontend{
		feRow("sr-fe-0", "FOLLOWER", true),
		feRow("sr-fe-1", FrontendRoleLeader, true),
		feRow("sr-fe-2", "FOLLOWER", true),
	}}
	pods := []testPod{
		{name: "sr-fe-0", revision: oldRevision, ready: true},
		{name: "sr-fe-1", revision: oldRevision, ready: true},
		{name: "sr-fe-2", revision: oldRevision, ready: true},
	}
	objs := append([]runtime.Object{src, sts}, newTestPods(src, pods)...)
	c := fake.NewFakeClient(srapi.Scheme, objs...)
	fc := &FeController{
		Client:   c,
		Recorder: record.NewFakeRecorder(10),
		NewFrontendsQuerier: func(_ context.Context, _ client.Client, _ *srapi.StarRocksCluster,
			_ *appsv1.StatefulSet, _ int32) (FrontendsQuerier, error) {
			return querier, nil
		},
	}
	ctx := context.Background()

	// recreate plays the StatefulSet controller: it brings a deleted pod back on the new revision.
	recreate := func(name string) {
		objs := newTestPods(src, []testPod{{name: name, revision: newRevision, ready: true}})
		require.NoError(t, c.Create(ctx, objs[0].(*corev1.Pod)))
	}
	deleted := func() string {
		remaining := map[string]bool{}
		for _, n := range podNames(t, c) {
			remaining[n] = true
		}
		for _, n := range []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"} {
			if !remaining[n] {
				return n
			}
		}
		return ""
	}

	var order []string
	for step := 0; step < 3; step++ {
		require.NoError(t, fc.ReconcileLeaderAwareUpgrade(ctx, src, sts.DeepCopy(), testFeQueryPort))
		victim := deleted()
		require.NotEmpty(t, victim, "step %d should delete a pod", step)
		order = append(order, victim)
		recreate(victim)
	}
	require.Equal(t, []string{"sr-fe-2", "sr-fe-0", "sr-fe-1"}, order)

	// Everything is on the new revision now: a further pass is a no-op.
	before := querier.calls
	require.NoError(t, fc.ReconcileLeaderAwareUpgrade(ctx, src, sts.DeepCopy(), testFeQueryPort))
	require.ElementsMatch(t, []string{"sr-fe-0", "sr-fe-1", "sr-fe-2"}, podNames(t, c))
	require.Equal(t, before, querier.calls)
}

func TestChooseNextPod(t *testing.T) {
	mk := func(names ...string) []corev1.Pod {
		var pods []corev1.Pod
		for _, n := range names {
			pods = append(pods, corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: n}})
		}
		return pods
	}
	require.Equal(t, "sr-fe-10", chooseNextPod(mk("sr-fe-2", "sr-fe-10", "sr-fe-1"), "sr-fe-1"))
	require.Equal(t, "sr-fe-2", chooseNextPod(mk("sr-fe-2", "sr-fe-10", "sr-fe-1"), "sr-fe-10"))
	require.Equal(t, "sr-fe-1", chooseNextPod(mk("sr-fe-1"), "sr-fe-1"))
}

func TestPodOrdinal(t *testing.T) {
	n, ok := podOrdinal("sr-fe", "sr-fe-12")
	require.True(t, ok)
	require.Equal(t, int32(12), n)
	_, ok = podOrdinal("sr-fe", "sr-be-0")
	require.False(t, ok)
	_, ok = podOrdinal("sr-fe", "sr-fe-x")
	require.False(t, ok)
}

func TestLeaderAwareUpgradeFlag(t *testing.T) {
	orig := config.EnableFeLeaderAwareUpgrade
	t.Cleanup(func() { config.EnableFeLeaderAwareUpgrade = orig })

	config.EnableFeLeaderAwareUpgrade = false
	require.False(t, LeaderAwareUpgradeEnabled())
	config.EnableFeLeaderAwareUpgrade = true
	require.True(t, LeaderAwareUpgradeEnabled())

	strategy := OnDeleteUpdateStrategy()
	require.Equal(t, appsv1.OnDeleteStatefulSetStrategyType, strategy.Type)
	require.Nil(t, strategy.RollingUpdate)
}

// TestSyncCluster_LeaderAwareUpgradeForcesOnDelete checks the StatefulSet the FE controller creates
// while the flag is on, and that a first reconcile (no status yet) does not try to delete anything.
func TestSyncCluster_LeaderAwareUpgradeForcesOnDelete(t *testing.T) {
	orig := config.EnableFeLeaderAwareUpgrade
	t.Cleanup(func() { config.EnableFeLeaderAwareUpgrade = orig })
	config.EnableFeLeaderAwareUpgrade = true

	src := newTestCluster(3)
	// A RollingUpdate strategy in the CR is overridden while the flag is on.
	src.Spec.StarRocksFeSpec.UpdateStrategy = &appsv1.StatefulSetUpdateStrategy{
		Type:          appsv1.RollingUpdateStatefulSetStrategyType,
		RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{Partition: rutils.GetInt32Pointer(1)},
	}
	c := fake.NewFakeClient(srapi.Scheme, src)
	querier := &fakeQuerier{}
	fc := &FeController{
		Client:   c,
		Recorder: record.NewFakeRecorder(10),
		NewFrontendsQuerier: func(_ context.Context, _ client.Client, _ *srapi.StarRocksCluster,
			_ *appsv1.StatefulSet, _ int32) (FrontendsQuerier, error) {
			return querier, nil
		},
	}
	require.NoError(t, fc.SyncCluster(context.Background(), src))

	var sts appsv1.StatefulSet
	require.NoError(t, c.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: testFeSts}, &sts))
	require.Equal(t, appsv1.OnDeleteStatefulSetStrategyType, sts.Spec.UpdateStrategy.Type)
	require.Nil(t, sts.Spec.UpdateStrategy.RollingUpdate)
	require.Zero(t, querier.calls, "a brand new statefulset has nothing to roll")

	// With the flag off the CR's own strategy is honored.
	config.EnableFeLeaderAwareUpgrade = false
	c2 := fake.NewFakeClient(srapi.Scheme, src)
	fc2 := &FeController{Client: c2, Recorder: record.NewFakeRecorder(10)}
	require.NoError(t, fc2.SyncCluster(context.Background(), src))
	require.NoError(t, c2.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: testFeSts}, &sts))
	require.Equal(t, appsv1.RollingUpdateStatefulSetStrategyType, sts.Spec.UpdateStrategy.Type)
}

// TestSyncCluster_LeaderAwareUpgradeConvertsExistingStatefulSet covers a StatefulSet created by an
// older operator: it runs RollingUpdate and carries no last-applied annotation, so only the targeted
// patch can clear rollingUpdate.
func TestSyncCluster_LeaderAwareUpgradeConvertsExistingStatefulSet(t *testing.T) {
	orig := config.EnableFeLeaderAwareUpgrade
	t.Cleanup(func() { config.EnableFeLeaderAwareUpgrade = orig })
	config.EnableFeLeaderAwareUpgrade = true

	src := newTestCluster(3)
	existing := &appsv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{Kind: "StatefulSet", APIVersion: appsv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: testFeSts, Namespace: testNamespace},
		Spec: appsv1.StatefulSetSpec{
			Replicas: rutils.GetInt32Pointer(3),
			Selector: &metav1.LabelSelector{MatchLabels: load.Selector(src.Name, src.Spec.StarRocksFeSpec)},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{Partition: rutils.GetInt32Pointer(0)},
			},
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "fe", Image: "old"}}}},
		},
	}
	c := fake.NewFakeClient(srapi.Scheme, src, existing)
	fc := &FeController{
		Client:   c,
		Recorder: record.NewFakeRecorder(10),
		NewFrontendsQuerier: func(_ context.Context, _ client.Client, _ *srapi.StarRocksCluster,
			_ *appsv1.StatefulSet, _ int32) (FrontendsQuerier, error) {
			return &fakeQuerier{}, nil
		},
	}
	require.NoError(t, fc.SyncCluster(context.Background(), src))

	var sts appsv1.StatefulSet
	require.NoError(t, c.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: testFeSts}, &sts))
	require.Equal(t, appsv1.OnDeleteStatefulSetStrategyType, sts.Spec.UpdateStrategy.Type)
	require.Nil(t, sts.Spec.UpdateStrategy.RollingUpdate)
}

func TestReconcileLeaderAwareUpgrade_StatefulSetNotYetVisible(t *testing.T) {
	src := newTestCluster(3)
	c := fake.NewFakeClient(srapi.Scheme, src)
	fc := &FeController{Client: c, Recorder: record.NewFakeRecorder(10)}
	sts := newTestSts(3, 2)
	require.NoError(t, fc.ReconcileLeaderAwareUpgrade(context.Background(), src, sts, testFeQueryPort))
}

func TestEnsureOnDeleteStrategy_MissingStatefulSet(t *testing.T) {
	c := fake.NewFakeClient(srapi.Scheme)
	require.NoError(t, EnsureOnDeleteStrategy(context.Background(), c, testNamespace, testFeSts))
}
