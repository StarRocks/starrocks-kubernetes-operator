package feproxy_test

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
	cdapi.Register()
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

// TestCelerDataClusterReconciler_FeProxyResourceCreate test the resources created by fe proxy controller.
func TestCelerDataClusterReconciler_FeProxyResourceCreate(t *testing.T) {
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
					},
				},
			},
			CelerDataFeProxySpec: &cdapi.CelerDataFeProxySpec{
				CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
					Replicas: rutils.GetInt32Pointer(3),
					Image:    "celerdata.com/cn:2.40",
					ResourceRequirements: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
							corev1.ResourceMemory: resource.MustParse("16G"),
						},
					},
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

	r := newCelerDataClusterController(src, ep)
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
	sts := appsv1.Deployment{}
	client := r.Client
	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample-fe-proxy",
	}, &sts)
	require.NoError(t, err)

	// check the external service is created
	externalService := corev1.Service{}
	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample-fe-proxy-service",
	}, &externalService)
	require.NoError(t, err)
}
