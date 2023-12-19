package cn_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
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

// TestStarRocksClusterReconciler_CnResourceCreate test the resources created by cn controller.
func TestStarRocksClusterReconciler_CnResourceCreate(t *testing.T) {
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
							ResolveKey:    "cn.conf",
						},
					},
				},
			},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						// After uncommenting Replicas: rutils.GetInt32Pointer(3), due to the conflict between Replicas
						// and AutoScalingPolicy, the Cn Controller will delete the Replicas field. This ultimately leads
						// to the execution of ctrl.Result{Requeue: true}, r.PatchStarRocksCluster(ctx, src).
						//Replicas: rutils.GetInt32Pointer(3),
						Image: "starrocks.com/cn:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						ConfigMapInfo: srapi.ConfigMapInfo{
							ConfigMapName: "cn-configMap",
							ResolveKey:    "cn.conf",
						},
					},
				},
				AutoScalingPolicy: &srapi.AutoScalingPolicy{
					HPAPolicy: &srapi.HPAPolicy{
						Metrics: []v2beta2.MetricSpec{{
							Type: v2beta2.PodsMetricSourceType,
							Object: &v2beta2.ObjectMetricSource{
								DescribedObject: v2beta2.CrossVersionObjectReference{
									Kind:       "statefulset",
									Name:       "starrockscluster-sample-cn",
									APIVersion: "apps/v2beta2",
								},
								Target: v2beta2.MetricTarget{
									Type:               v2beta2.ValueMetricType,
									Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
									AverageUtilization: rutils.GetInt32Pointer(1),
								},
								Metric: v2beta2.MetricIdentifier{
									Name: "test",
									Selector: &metav1.LabelSelector{
										MatchLabels: make(map[string]string),
									},
								},
							},
							Pods: &v2beta2.PodsMetricSource{
								Metric: v2beta2.MetricIdentifier{
									Name: "test",
									Selector: &metav1.LabelSelector{
										MatchLabels: make(map[string]string),
									},
								},
								Target: v2beta2.MetricTarget{
									Type:               v2beta2.ValueMetricType,
									Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
									AverageUtilization: rutils.GetInt32Pointer(1),
								},
							},
							Resource: &v2beta2.ResourceMetricSource{
								Name: "test",
								Target: v2beta2.MetricTarget{
									Type:               v2beta2.ValueMetricType,
									Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
									AverageUtilization: rutils.GetInt32Pointer(1),
								},
							},
							ContainerResource: &v2beta2.ContainerResourceMetricSource{
								Name: "test",
								Target: v2beta2.MetricTarget{
									Type:               v2beta2.ValueMetricType,
									Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
									AverageUtilization: rutils.GetInt32Pointer(1),
								},
								Container: "test",
							},
							External: &v2beta2.ExternalMetricSource{
								Metric: v2beta2.MetricIdentifier{
									Name: "test",
									Selector: &metav1.LabelSelector{
										MatchLabels: make(map[string]string),
									},
								},
								Target: v2beta2.MetricTarget{
									Type:               v2beta2.ValueMetricType,
									Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
									AverageUtilization: rutils.GetInt32Pointer(1),
								},
							},
						}},
					},
					MaxReplicas: 1,
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
	cnConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cn-configMap",
			Namespace: "default",
		},
		Data: map[string]string{
			"cn.conf": "hello2 = world2",
		},
	}

	r := newStarRocksClusterController(src, ep, feConfigMap, cnConfigMap)
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
		Name:      "starrockscluster-sample-cn",
	}, &sts)
	require.NoError(t, err)

	// check the external service is created
	externalService := corev1.Service{}
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample-cn-service",
	}, &externalService)
	require.NoError(t, err)

	// check the internal service is created
	internalService := corev1.Service{}
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "starrockscluster-sample-cn-search",
	}, &internalService)
	require.NoError(t, err)
}
