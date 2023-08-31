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

package pkg

import (
	"context"
	"testing"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newStarRocksClusterController(objects ...runtime.Object) *StarRocksClusterReconciler {
	srcController := &StarRocksClusterReconciler{}
	srcController.Recorder = record.NewFakeRecorder(10)
	srcController.Client = k8sutils.NewFakeClient(Scheme, objects...)
	fc := fe.New(srcController.Client)
	cc := cn.New(srcController.Client)
	srcController.Scs = make(map[string]sub_controller.ClusterSubController)
	srcController.Scs[feControllerName] = fc
	srcController.Scs[cnControllerName] = cc
	return srcController
}

func TestReconcileConstructFeResource(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			ServiceAccount: "starrocksAccount",
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
								Name:             "fe-storage",
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

func TestStarRocksClusterReconciler_FeReconcileSuccess(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			ServiceAccount: "starrocksAccount",
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
								Name:             "fe-storage",
								StorageClassName: rutils.GetStringPointer("shard-data"),
								StorageSize:      "10Gi",
								MountPath:        "/data/fe/meta",
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

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "default",
					Labels: map[string]string{
						srapi.OwnerReference:    src.Name + "-" + srapi.DEFAULT_FE,
						srapi.ComponentLabelKey: srapi.DEFAULT_FE,
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{
						Ready: true,
					}},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "default",
					Labels: map[string]string{
						srapi.OwnerReference:    src.Name + "-" + srapi.DEFAULT_FE,
						srapi.ComponentLabelKey: srapi.DEFAULT_FE,
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{
						Ready: true,
					}},
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod3",
					Namespace: "default",
					Labels: map[string]string{
						srapi.OwnerReference:    src.Name + "-" + srapi.DEFAULT_FE,
						srapi.ComponentLabelKey: srapi.DEFAULT_FE,
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{
						Ready: true,
					}},
				},
			},
		},
	}

	st := &appv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{
		Name:      src.Name + "-" + srapi.DEFAULT_FE,
		Namespace: "default",
	}}

	r := newStarRocksClusterController(src, podList, st)
	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample",
	}})

	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}

func TestStarRocksClusterReconciler_CnResourceCreate(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			ServiceAccount: "starrocksAccount",
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
					},
				},
			},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "starrocks.com/cn:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
					},
				},
			},
		},
	}

	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample-fe-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{
				IP: "127.0.0.1",
			}},
		}},
	}
	r := newStarRocksClusterController(src, ep)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "starrockscluster-sample"}})
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}

func TestStarRocksClusterReconciler_CnStatus(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			ServiceAccount: "starrocksAccount",
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
					},
				},
			},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "starrocks.com/cn:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
					},
				},
			},
		},
	}

	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample-fe-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{
				IP: "127.0.0.1",
			}},
		}},
	}

	st := &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "starrockscluster-sample-cn",
			Namespace: "default",
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: rutils.GetInt32Pointer(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					srapi.OwnerReference:    "starrockscluster-sample-cn",
					srapi.ComponentLabelKey: srapi.DEFAULT_CN,
				},
			},
		},
	}

	r := newStarRocksClusterController(src, ep, st)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "starrockscluster-sample"}})
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}
