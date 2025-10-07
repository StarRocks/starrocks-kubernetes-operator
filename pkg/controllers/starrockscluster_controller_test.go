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

package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

func newStarRocksClusterController(objects ...runtime.Object) *StarRocksClusterReconciler {
	srcController := &StarRocksClusterReconciler{}
	srcController.Recorder = record.NewFakeRecorder(10)
	srcController.Client = fake.NewFakeClient(srapi.Scheme, objects...)
	srcController.FeController = fe.New(srcController.Client, fake.GetEventRecorderFor(nil))
	srcController.BeController = be.New(srcController.Client, fake.GetEventRecorderFor(srcController.Recorder))
	srcController.CnController = cn.New(srcController.Client, fake.GetEventRecorderFor(nil))
	srcController.FeProxyController = feproxy.New(srcController.Client, fake.GetEventRecorderFor(nil))
	return srcController
}

func TestReconcileConstructFeResource(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "starrocks.com/fe:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						StorageVolumes: []srapi.StorageVolume{
							{
								Name:             "fe-meta",
								StorageClassName: rutils.GetStringPointer("shard-data"),
								MountPath:        "/data/fe/meta",
								StorageSize:      "10Gi",
							},
						},
					},
				},
			},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "test",
					},
				},
			},
			StarRocksBeSpec: &srapi.StarRocksBeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "test",
					},
				},
			},
		},
	}

	r := newStarRocksClusterController(src)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "starrockscluster-sample"}})
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}

// Helper to create a basic cluster for upgrade tests
func newTestCluster(phase srapi.Phase, image string) *srapi.StarRocksCluster {
	return &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image: image,
					},
				},
			},
		},
		Status: srapi.StarRocksClusterStatus{
			Phase: phase,
		},
	}
}

// Helper to create a StatefulSet with a specific image for upgrade tests
func newTestStatefulSet(name string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "starrocks/fe:3.1.0"},
					},
				},
			},
		},
	}
}

// TestIsUpgrade tests the main upgrade detection logic
func TestIsUpgrade(t *testing.T) {
	ctx := context.Background()

	t.Run("not running cluster returns false", func(t *testing.T) {
		cluster := newTestCluster("", "starrocks/fe:3.1.0")
		client := fake.NewFakeClient(srapi.Scheme)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result)
	})

	t.Run("running cluster with no statefulset returns false", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.1.0")
		client := fake.NewFakeClient(srapi.Scheme)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result)
	})

	t.Run("running cluster with same image returns false", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.1.0")
		sts := newTestStatefulSet("test-cluster-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result)
	})

	t.Run("running cluster with different image returns true", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.2.0")
		sts := newTestStatefulSet("test-cluster-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.True(t, result)
	})
}

// TestGetCurrentImageFromStatefulSet tests image retrieval from StatefulSets
func TestGetCurrentImageFromStatefulSet(t *testing.T) {
	ctx := context.Background()

	t.Run("missing statefulset returns empty", func(t *testing.T) {
		client := fake.NewFakeClient(srapi.Scheme)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-fe")
		require.Equal(t, "", result)
	})

	t.Run("existing statefulset returns image", func(t *testing.T) {
		sts := newTestStatefulSet("test-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-fe")
		require.Equal(t, "starrocks/fe:3.1.0", result)
	})
}

// TestGetControllersInOrder tests controller ordering based on deployment scenario
func TestGetControllersInOrder(t *testing.T) {
	ctx := context.Background()
	feCtrl := fe.New(nil, nil)
	beCtrl := be.New(nil, nil)
	cnCtrl := cn.New(nil, nil)
	feProxyCtrl := feproxy.New(nil, nil)

	t.Run("initial deployment uses FE-first order", func(t *testing.T) {
		cluster := newTestCluster("", "starrocks/fe:3.1.0")
		client := fake.NewFakeClient(srapi.Scheme)

		controllers := getControllersInOrder(ctx, client, cluster, feCtrl, beCtrl, cnCtrl, feProxyCtrl)

		// Check FE is first
		require.Equal(t, "fe", controllers[0].GetControllerName())
		// Verify order: FE -> BE -> CN -> FeProxy
		require.Equal(t, "be", controllers[1].GetControllerName())
		require.Equal(t, "cn", controllers[2].GetControllerName())
		require.Equal(t, "feproxy", controllers[3].GetControllerName())
	})

	t.Run("upgrade uses BE-first order", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.2.0")
		sts := newTestStatefulSet("test-cluster-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		controllers := getControllersInOrder(ctx, client, cluster, feCtrl, beCtrl, cnCtrl, feProxyCtrl)

		// Check BE is first
		require.Equal(t, "be", controllers[0].GetControllerName())
		// Verify order: BE -> CN -> FE -> FeProxy (StarRocks upgrade procedure)
		require.Equal(t, "cn", controllers[1].GetControllerName())
		require.Equal(t, "fe", controllers[2].GetControllerName())
		require.Equal(t, "feproxy", controllers[3].GetControllerName())
	})
}
