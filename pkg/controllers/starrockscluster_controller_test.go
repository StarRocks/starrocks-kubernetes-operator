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

	t.Run("empty phase with no statefulsets is initial deployment", func(t *testing.T) {
		cluster := newTestCluster("", "starrocks/fe:3.1.0")
		client := fake.NewFakeClient(srapi.Scheme)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "empty phase should be initial deployment")
	})

	t.Run("reconciling phase is initial deployment even with statefulsets", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterReconciling, "starrocks/fe:3.1.0")
		sts := newTestStatefulSet("test-cluster-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "reconciling phase should be initial deployment")
	})

	t.Run("fe statefulset with image change triggers upgrade", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.2.0")
		sts := newTestStatefulSet("test-cluster-fe")
		sts.Spec.Template.Spec.Containers[0].Image = "starrocks/fe:3.1.0" // Different image
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.True(t, result, "FE StatefulSet with image change should trigger upgrade")
	})

	t.Run("be statefulset without fe treated as initial deployment", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/be:3.2.0")
		sts := newTestStatefulSet("test-cluster-be")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "BE without FE should not trigger upgrade ordering; FE must be reconciled first")
	})

	t.Run("multiple statefulsets with image change triggers upgrade", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.2.0")
		feSts := newTestStatefulSet("test-cluster-fe")
		feSts.Spec.Template.Spec.Containers[0].Image = "starrocks/fe:3.1.0" // Different
		beSts := newTestStatefulSet("test-cluster-be")
		beSts.Spec.Template.Spec.Containers[0].Image = "starrocks/be:3.1.0"
		client := fake.NewFakeClient(srapi.Scheme, feSts, beSts)

		result := isUpgrade(ctx, client, cluster)
		require.True(t, result, "multiple StatefulSets with image change should trigger upgrade")
	})

	t.Run("running cluster without statefulsets is treated as initial deployment", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.1.0")
		client := fake.NewFakeClient(srapi.Scheme) // No StatefulSets

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "running phase without StatefulSets should be treated as initial deployment (status corruption)")
	})

	t.Run("failed phase is initial deployment even with statefulsets", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterFailed, "starrocks/fe:3.1.0")
		sts := newTestStatefulSet("test-cluster-fe")
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "failed phase should be initial deployment")
	})

	t.Run("statefulset with same image does not trigger upgrade", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.1.0")
		sts := newTestStatefulSet("test-cluster-fe")
		sts.Spec.Template.Spec.Containers[0].Image = "starrocks/fe:3.1.0" // Same image
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.False(t, result, "same image version should use initial deployment ordering (no upgrade needed)")
	})

	t.Run("be image change triggers upgrade", func(t *testing.T) {
		cluster := &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksFeSpec: &srapi.StarRocksFeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/fe:3.1.0",
						},
					},
				},
				StarRocksBeSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/be:3.2.0", // New version
						},
					},
				},
			},
		}
		feSts := newTestStatefulSet("test-cluster-fe")
		feSts.Spec.Template.Spec.Containers[0].Image = "starrocks/fe:3.1.0" // Same
		beSts := newTestStatefulSet("test-cluster-be")
		beSts.Spec.Template.Spec.Containers[0].Image = "starrocks/be:3.1.0" // Old version
		client := fake.NewFakeClient(srapi.Scheme, feSts, beSts)

		result := isUpgrade(ctx, client, cluster)
		require.True(t, result, "BE image change should trigger upgrade")
	})

	t.Run("downgrade treated as upgrade for safety", func(t *testing.T) {
		cluster := newTestCluster(srapi.ClusterRunning, "starrocks/fe:3.0.0") // Downgrade to 3.0.0
		sts := newTestStatefulSet("test-cluster-fe")
		sts.Spec.Template.Spec.Containers[0].Image = "starrocks/fe:3.1.0" // Currently on 3.1.0
		client := fake.NewFakeClient(srapi.Scheme, sts)

		result := isUpgrade(ctx, client, cluster)
		require.True(t, result, "downgrade should use upgrade ordering")
	})
}

// TestGetCurrentImageFromStatefulSet tests the image retrieval logic from StatefulSets
func TestGetCurrentImageFromStatefulSet(t *testing.T) {
	ctx := context.Background()

	t.Run("returns empty string when StatefulSet not found", func(t *testing.T) {
		client := fake.NewFakeClient(srapi.Scheme)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "missing-sts")
		require.Equal(t, "", result)
	})

	t.Run("returns empty string when StatefulSet has no containers", func(t *testing.T) {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, sts)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-sts")
		require.Equal(t, "", result)
	})

	t.Run("returns image from fe container by name", func(t *testing.T) {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-fe",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "fe", Image: "starrocks/fe:3.1.0"},
						},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, sts)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-fe")
		require.Equal(t, "starrocks/fe:3.1.0", result)
	})

	t.Run("returns image from be container by name", func(t *testing.T) {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-be",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "be", Image: "starrocks/be:3.1.0"},
						},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, sts)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-be")
		require.Equal(t, "starrocks/be:3.1.0", result)
	})

	t.Run("returns image from cn container by name", func(t *testing.T) {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cn",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "cn", Image: "starrocks/cn:3.1.0"},
						},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, sts)
		result := getCurrentImageFromStatefulSet(ctx, client, "default", "test-cn")
		require.Equal(t, "starrocks/cn:3.1.0", result)
	})
}

// TestCheckForImageChanges tests the image comparison logic
func TestCheckForImageChanges(t *testing.T) {
	ctx := context.Background()

	t.Run("no changes when all images match", func(t *testing.T) {
		cluster := &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksFeSpec: &srapi.StarRocksFeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/fe:3.1.0",
						},
					},
				},
				StarRocksBeSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/be:3.1.0",
						},
					},
				},
				StarRocksCnSpec: &srapi.StarRocksCnSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/cn:3.1.0",
						},
					},
				},
			},
		}
		feSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cluster-fe", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "fe", Image: "starrocks/fe:3.1.0"}},
					},
				},
			},
		}
		beSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cluster-be", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "be", Image: "starrocks/be:3.1.0"}},
					},
				},
			},
		}
		cnSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cluster-cn", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "cn", Image: "starrocks/cn:3.1.0"}},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, feSts, beSts, cnSts)
		result := checkForImageChanges(ctx, client, cluster)
		require.False(t, result, "no changes when all images match")
	})

	t.Run("detects FE image change", func(t *testing.T) {
		cluster := &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksFeSpec: &srapi.StarRocksFeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/fe:3.2.0",
						},
					},
				},
			},
		}
		feSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cluster-fe", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "fe", Image: "starrocks/fe:3.1.0"}},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, feSts)
		result := checkForImageChanges(ctx, client, cluster)
		require.True(t, result, "should detect FE image change")
	})

	t.Run("detects BE image change", func(t *testing.T) {
		cluster := &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksBeSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Image: "starrocks/be:3.2.0",
						},
					},
				},
			},
		}
		beSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cluster-be", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "be", Image: "starrocks/be:3.1.0"}},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, beSts)
		result := checkForImageChanges(ctx, client, cluster)
		require.True(t, result, "should detect BE image change")
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

		isUpgradeScenario := isUpgrade(ctx, client, cluster)
		controllers := getControllersInOrder(isUpgradeScenario, feCtrl, beCtrl, cnCtrl, feProxyCtrl)

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

		isUpgradeScenario := isUpgrade(ctx, client, cluster)
		controllers := getControllersInOrder(isUpgradeScenario, feCtrl, beCtrl, cnCtrl, feProxyCtrl)

		// Check BE is first
		require.Equal(t, "be", controllers[0].GetControllerName())
		// Verify order: BE -> CN -> FE -> FeProxy (StarRocks upgrade procedure)
		require.Equal(t, "cn", controllers[1].GetControllerName())
		require.Equal(t, "fe", controllers[2].GetControllerName())
		require.Equal(t, "feproxy", controllers[3].GetControllerName())
	})
}

// TestIsComponentReady tests component readiness detection
func TestIsComponentReady(t *testing.T) {
	ctx := context.Background()
	cluster := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{},
			StarRocksBeSpec: &srapi.StarRocksBeSpec{},
		},
	}

	t.Run("component not ready when endpoints not found", func(t *testing.T) {
		client := fake.NewFakeClient(srapi.Scheme)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.False(t, result)
	})

	t.Run("component not ready when no addresses", func(t *testing.T) {
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.False(t, result)
	})

	t.Run("component ready when endpoints have addresses", func(t *testing.T) {
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{
						{IP: "10.0.0.1"},
					},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.True(t, result)
	})

	t.Run("component ready when spec is nil", func(t *testing.T) {
		clusterNoCN := &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksFeSpec: &srapi.StarRocksFeSpec{},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme)
		result := isComponentReady(ctx, client, clusterNoCN, "cn")
		require.True(t, result, "CN should be considered ready when not configured")
	})

	t.Run("component not ready when StatefulSet not found", func(t *testing.T) {
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}},
				},
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.False(t, result, "component not ready when StatefulSet doesn't exist")
	})

	t.Run("component not ready when ObservedGeneration lags behind Generation", func(t *testing.T) {
		replicas := int32(3)
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}},
				},
			},
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-cluster-fe",
				Namespace:  "default",
				Generation: 5, // Spec changed
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: 4, // Controller hasn't observed the change yet
				CurrentRevision:    "test-cluster-fe-12345",
				UpdateRevision:     "test-cluster-fe-12345",
				ReadyReplicas:      3,
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints, sts)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.False(t, result, "component not ready when StatefulSet spec change not yet observed")
	})

	t.Run("component not ready when rollout in progress", func(t *testing.T) {
		replicas := int32(3)
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-be-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.2"}},
				},
			},
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-cluster-be",
				Namespace:  "default",
				Generation: 3,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: 3,                       // Observed
				CurrentRevision:    "test-cluster-be-12345", // Old revision
				UpdateRevision:     "test-cluster-be-67890", // New revision - rollout in progress
				ReadyReplicas:      3,
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints, sts)
		result := isComponentReady(ctx, client, cluster, "be")
		require.False(t, result, "component not ready when StatefulSet rollout in progress")
	})

	t.Run("component not ready when replicas not all ready", func(t *testing.T) {
		replicas := int32(3)
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}},
				},
			},
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-cluster-fe",
				Namespace:  "default",
				Generation: 2,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: 2,
				CurrentRevision:    "test-cluster-fe-12345",
				UpdateRevision:     "test-cluster-fe-12345",
				ReadyReplicas:      2, // Only 2 out of 3 ready
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints, sts)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.False(t, result, "component not ready when not all replicas are ready")
	})

	t.Run("component ready when all checks pass", func(t *testing.T) {
		replicas := int32(3)
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-fe-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}, {IP: "10.0.0.2"}, {IP: "10.0.0.3"}},
				},
			},
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-cluster-fe",
				Namespace:  "default",
				Generation: 5,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: 5,                       // Generation matches
				CurrentRevision:    "test-cluster-fe-12345", // Rollout complete
				UpdateRevision:     "test-cluster-fe-12345", // Same revision
				ReadyReplicas:      3,                       // All replicas ready
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints, sts)
		result := isComponentReady(ctx, client, cluster, "fe")
		require.True(t, result, "component ready when all 4 checks pass")
	})

	t.Run("component ready for BE with all checks passing", func(t *testing.T) {
		replicas := int32(2)
		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-be-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "10.0.0.4"}, {IP: "10.0.0.5"}},
				},
			},
		}
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-cluster-be",
				Namespace:  "default",
				Generation: 10,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: 10,
				CurrentRevision:    "test-cluster-be-abc123",
				UpdateRevision:     "test-cluster-be-abc123",
				ReadyReplicas:      2,
			},
		}
		client := fake.NewFakeClient(srapi.Scheme, endpoints, sts)
		result := isComponentReady(ctx, client, cluster, "be")
		require.True(t, result, "BE component ready when all checks pass")
	})
}
