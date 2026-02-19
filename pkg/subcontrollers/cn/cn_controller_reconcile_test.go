package cn_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
	rutils "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/common/resource_utils"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/controllers"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/k8sutils/fake"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/be"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/cn"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/fe"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/feproxy"
)

func newCelerDataClusterController(objects ...runtime.Object) *controllers.CelerDataClusterReconciler {
	srcController := &controllers.CelerDataClusterReconciler{}
	srcController.Recorder = record.NewFakeRecorder(10)
	srcController.Client = fake.NewFakeClient(cdapi.Scheme, objects...)
	srcController.Scs = []subcontrollers.ClusterSubController{
		fe.New(srcController.Client, fake.GetEventRecorderFor(nil)),
		be.New(srcController.Client, fake.GetEventRecorderFor(srcController.Recorder)),
		cn.New(srcController.Client, fake.GetEventRecorderFor(nil)),
		feproxy.New(srcController.Client, fake.GetEventRecorderFor(nil)),
	}
	return srcController
}

// TestCelerDataClusterReconciler_CnResourceCreate test the resources created by cn controller.
func TestCelerDataClusterReconciler_CnResourceCreate(t *testing.T) {
	// define a CelerDataCluster CR
	src := &cdapi.CelerDataCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celerdatacluster-sample",
			Namespace: "default",
		},
		Spec: cdapi.CelerDataClusterSpec{
			CelerDataFeSpec: &cdapi.CelerDataFeSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "celerdata.com/fe:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						ConfigMapInfo: cdapi.ConfigMapInfo{
							ConfigMapName: "fe-configMap",
							ResolveKey:    "cn.conf",
						},
					},
				},
			},
			CelerDataCnSpec: &cdapi.CelerDataCnSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						// After uncommenting Replicas: rutils.GetInt32Pointer(3), due to the conflict between Replicas
						// and AutoScalingPolicy, the Cn Controller will delete the Replicas field. This ultimately leads
						// to the execution of ctrl.Result{Requeue: true}, r.PatchCelerDataCluster(ctx, src).
						// Replicas: rutils.GetInt32Pointer(3),
						Image: "celerdata.com/cn:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						ConfigMapInfo: cdapi.ConfigMapInfo{
							ConfigMapName: "cn-configMap",
							ResolveKey:    "cn.conf",
						},
					},
				},
				AutoScalingPolicy: &cdapi.AutoScalingPolicy{
					HPAPolicy: &cdapi.HPAPolicy{
						Metrics: []v2beta2.MetricSpec{{
							Type: v2beta2.PodsMetricSourceType,
							Object: &v2beta2.ObjectMetricSource{
								DescribedObject: v2beta2.CrossVersionObjectReference{
									Kind:       "statefulset",
									Name:       "celerdatacluster-sample-cn",
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
			Name:      "celerdatacluster-sample-fe-service",
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

	r := newCelerDataClusterController(src, ep, feConfigMap, cnConfigMap)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "celerdatacluster-sample",
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)

	// check the statefulset is created
	sts := appsv1.StatefulSet{}
	client := r.Client
	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample-cn",
	}, &sts)
	require.NoError(t, err)

	// check the external service is created
	externalService := corev1.Service{}
	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample-cn-service",
	}, &externalService)
	require.NoError(t, err)

	// check the internal service is created
	internalService := corev1.Service{}
	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample-cn-search",
	}, &internalService)
	require.NoError(t, err)
}
