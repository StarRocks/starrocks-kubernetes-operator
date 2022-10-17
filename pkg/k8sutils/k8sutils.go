package k8sutils

import (
	"context"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdateService(ctx context.Context, k8sclient client.Client, svc *corev1.Service) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	var esvc corev1.Service
	err := k8sclient.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, &esvc)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sclient, svc)
	} else if err != nil {
		return err
	}

	if rutils.
		ServiceDeepEqual(svc, &esvc) {
		klog.Info("CreateOrUpdateService service Name, Ports, Selector, ServiceType, Labels have not change", "namespace", svc.Namespace, "name", svc.Name)
		return nil
	}

	// Apply immutable fields from the existing service.
	svc.Spec.IPFamilies = esvc.Spec.IPFamilies
	svc.Spec.IPFamilyPolicy = esvc.Spec.IPFamilyPolicy
	svc.Spec.ClusterIP = esvc.Spec.ClusterIP
	svc.Spec.ClusterIPs = esvc.Spec.ClusterIPs

	rutils.MergeMetadata(&svc.ObjectMeta, esvc.ObjectMeta)
	return UpdateClientObject(ctx, k8sclient, svc)
}

func CreateClientObject(ctx context.Context, k8sclient client.Client, object client.Object) error {
	klog.Info("Creating resource service", "namespace", object.GetNamespace(), "name", object.GetName(), "kind", object.GetObjectKind())
	if err := k8sclient.Create(ctx, object); err != nil {
		return err
	}
	return nil
}

func UpdateClientObject(ctx context.Context, k8sclient client.Client, object client.Object) error {
	klog.Info("Updating resource service", "namespace", object.GetNamespace(), "name", object.GetName(), "kind", object.GetObjectKind())
	if err := k8sclient.Update(ctx, object); err != nil {
		return err
	}
	return nil
}
