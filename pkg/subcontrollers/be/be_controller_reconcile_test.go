package be_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/controllers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

func newStarRocksClusterController(objects ...runtime.Object) *controllers.StarRocksClusterReconciler {
	srcController := &controllers.StarRocksClusterReconciler{}
	srcController.Recorder = record.NewFakeRecorder(10)
	srcController.Client = fake.NewFakeClient(srapi.Scheme, objects...)
	srcController.Scs = []subcontrollers.ClusterSubController{
		fe.New(srcController.Client),
		be.New(srcController.Client),
		cn.New(srcController.Client),
		feproxy.New(srcController.Client),
	}
	return srcController
}

// TestStarRocksClusterReconciler_BeResourceCreate test the resources created by be controller.
func TestStarRocksClusterReconciler_BeResourceCreate(t *testing.T) {
	// define a StarRocksCluster CR
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
						ConfigMapInfo: srapi.ConfigMapInfo{
							ConfigMapName: "fe-configMap",
							ResolveKey:    "fe.conf",
						},
					},
				},
			},
			StarRocksBeSpec: &srapi.StarRocksBeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "starrocks.com/be:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						ConfigMapInfo: srapi.ConfigMapInfo{
							ConfigMapName: "be-configMap",
							ResolveKey:    "be.conf",
						},
					},
				},
			},
		},
	}

	// mock the fe service is ready
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

	// mock the fe configMap
	feConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fe-configMap",
			Namespace: "default",
		},
		Data: map[string]string{
			"fe.conf": "hello = world",
		},
	}

	// mock the cn configMap
	beConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cn-configMap",
			Namespace: "default",
		},
		Data: map[string]string{
			"be.conf": "hello2 = world2",
		},
	}

	r := newStarRocksClusterController(src, ep, feConfigMap, beConfigMap)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "starrockscluster-sample",
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)

	// check the statefulset is created
	sts := appv1.StatefulSet{}
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample-be",
	}, &sts)
	require.NoError(t, err)

	// check the external service is created
	externalService := corev1.Service{}
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample-be-service",
	}, &externalService)
	require.NoError(t, err)

	// check the internal service is created
	internalService := corev1.Service{}
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample-be-search",
	}, &internalService)
	require.NoError(t, err)
}
