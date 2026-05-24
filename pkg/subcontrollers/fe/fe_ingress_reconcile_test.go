package fe_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

func feClusterWithIngress(ingress *srapi.FeIngress) *srapi.StarRocksCluster {
	return &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "starrockscluster-sample", Namespace: "default"},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "starrocks.com/fe:latest",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("1G"),
							},
						},
						StorageVolumes: []srapi.StorageVolume{{
							Name:             "fe-meta",
							StorageClassName: rutils.GetStringPointer("shard-data"),
							StorageSize:      "10Gi",
							MountPath:        "/data/fe/meta",
						}},
					},
				},
				Ingress: ingress,
			},
		},
	}
}

// TestFeController_SyncCluster_Ingress verifies the declarative reconcile behavior:
// an Ingress is created when feSpec.Ingress is set, and removed when it is cleared.
func TestFeController_SyncCluster_Ingress(t *testing.T) {
	className := "nginx"
	src := feClusterWithIngress(&srapi.FeIngress{
		IngressClassName: &className,
		Host:             "sr.example.com",
		Annotations:      map[string]string{"a": "b"},
	})
	client := fake.NewFakeClient(srapi.Scheme, src)
	controller := fe.New(client, fake.GetEventRecorderFor(nil))

	require.NoError(t, controller.SyncCluster(context.Background(), src))

	key := types.NamespacedName{Name: rutils.FeIngressName(src.Name), Namespace: src.Namespace}
	var ingress networkingv1.Ingress
	require.NoError(t, client.Get(context.Background(), key, &ingress),
		"ingress should be created when feSpec.Ingress is set")
	require.Equal(t, "sr.example.com", ingress.Spec.Rules[0].Host)
	backend := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service
	require.Equal(t, src.Name+"-fe-service", backend.Name, "backend must target the FE external service")
	require.Equal(t, rutils.FeHTTPPortName, backend.Port.Name, "backend must target the http (web UI) port")

	// Clearing the field must remove the previously created Ingress (declarative convergence).
	src.Spec.StarRocksFeSpec.Ingress = nil
	require.NoError(t, controller.SyncCluster(context.Background(), src))
	err := client.Get(context.Background(), key, &networkingv1.Ingress{})
	require.True(t, apierrors.IsNotFound(err), "ingress should be deleted when feSpec.Ingress is cleared")
}

// TestFeController_ClearCluster_DeletesIngress verifies the Ingress is removed when the
// StarRocksCluster is being deleted.
func TestFeController_ClearCluster_DeletesIngress(t *testing.T) {
	className := "nginx"
	src := feClusterWithIngress(&srapi.FeIngress{IngressClassName: &className, Host: "sr.example.com"})
	client := fake.NewFakeClient(srapi.Scheme, src)
	controller := fe.New(client, fake.GetEventRecorderFor(nil))

	require.NoError(t, controller.SyncCluster(context.Background(), src))

	// Simulate cluster deletion: ClearCluster only acts when the CR is being deleted and
	// the FE status exists.
	now := metav1.Now()
	src.DeletionTimestamp = &now
	src.Status.StarRocksFeStatus = &srapi.StarRocksFeStatus{}

	require.NoError(t, controller.ClearCluster(context.Background(), src))

	key := types.NamespacedName{Name: rutils.FeIngressName(src.Name), Namespace: src.Namespace}
	err := client.Get(context.Background(), key, &networkingv1.Ingress{})
	require.True(t, apierrors.IsNotFound(err), "ingress should be deleted on cluster deletion")
}

// TestFeController_SyncCluster_Ingress_Update verifies that changing feSpec.Ingress patches
// the existing Ingress, and that an annotation written out-of-band by an ingress controller
// survives the update. This is the reason ApplyIngress uses a three-way merge patch instead
// of a full replace: the operator must not clobber controller-managed fields.
func TestFeController_SyncCluster_Ingress_Update(t *testing.T) {
	src := feClusterWithIngress(&srapi.FeIngress{
		Host:        "old.example.com",
		Annotations: map[string]string{"nginx.ingress.kubernetes.io/rewrite-target": "/"},
	})
	client := fake.NewFakeClient(srapi.Scheme, src)
	controller := fe.New(client, fake.GetEventRecorderFor(nil))

	require.NoError(t, controller.SyncCluster(context.Background(), src))

	key := types.NamespacedName{Name: rutils.FeIngressName(src.Name), Namespace: src.Namespace}

	// Simulate an ingress controller writing its own annotation onto the live object.
	var live networkingv1.Ingress
	require.NoError(t, client.Get(context.Background(), key, &live))
	live.Annotations["ingress.kubernetes.io/backends"] = "HEALTHY"
	require.NoError(t, client.Update(context.Background(), &live))

	// Change the host and re-sync.
	src.Spec.StarRocksFeSpec.Ingress.Host = "new.example.com"
	require.NoError(t, controller.SyncCluster(context.Background(), src))

	var updated networkingv1.Ingress
	require.NoError(t, client.Get(context.Background(), key, &updated))
	require.Equal(t, "new.example.com", updated.Spec.Rules[0].Host, "host change must be applied")
	require.Equal(t, "/", updated.Annotations["nginx.ingress.kubernetes.io/rewrite-target"],
		"operator-managed annotation must be retained")
	require.Equal(t, "HEALTHY", updated.Annotations["ingress.kubernetes.io/backends"],
		"controller-added annotation must be preserved, not clobbered by the operator")
}

// TestFeController_SyncCluster_Ingress_Idempotent verifies repeated syncs with an unchanged
// spec do not error and converge to a single Ingress carrying the resource-hash annotation
// that makes subsequent syncs no-ops.
func TestFeController_SyncCluster_Ingress_Idempotent(t *testing.T) {
	src := feClusterWithIngress(&srapi.FeIngress{Host: "sr.example.com"})
	client := fake.NewFakeClient(srapi.Scheme, src)
	controller := fe.New(client, fake.GetEventRecorderFor(nil))

	require.NoError(t, controller.SyncCluster(context.Background(), src))
	require.NoError(t, controller.SyncCluster(context.Background(), src))

	key := types.NamespacedName{Name: rutils.FeIngressName(src.Name), Namespace: src.Namespace}
	var ingress networkingv1.Ingress
	require.NoError(t, client.Get(context.Background(), key, &ingress))
	require.NotEmpty(t, ingress.Annotations[srapi.ComponentResourceHash],
		"resource hash annotation should be set so subsequent syncs are no-ops")
}
