package fe_test

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

func TestCelerDataClusterReconciler_FeReconcileSuccess(t *testing.T) {
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
						StorageVolumes: []cdapi.StorageVolume{
							{
								Name:             "fe-meta",
								StorageClassName: rutils.GetStringPointer("shard-data"),
								StorageSize:      "10Gi",
								MountPath:        "/data/fe/meta",
							},
						},
					},
				},
			},
			CelerDataCnSpec: &cdapi.CelerDataCnSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "test",
					},
				},
			},
			CelerDataBeSpec: &cdapi.CelerDataBeSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
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
						cdapi.OwnerReference:    src.Name + "-" + cdapi.DEFAULT_FE,
						cdapi.ComponentLabelKey: cdapi.DEFAULT_FE,
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
						cdapi.OwnerReference:    src.Name + "-" + cdapi.DEFAULT_FE,
						cdapi.ComponentLabelKey: cdapi.DEFAULT_FE,
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
						cdapi.OwnerReference:    src.Name + "-" + cdapi.DEFAULT_FE,
						cdapi.ComponentLabelKey: cdapi.DEFAULT_FE,
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

	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{
		Name:      src.Name + "-" + cdapi.DEFAULT_FE,
		Namespace: "default",
	}}

	r := newCelerDataClusterController(src, podList, sts)
	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: "default",
		Name:      "celerdatacluster-sample",
	}})

	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}
