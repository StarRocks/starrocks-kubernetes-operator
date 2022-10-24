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
		klog.Info("CreateOrUpdateService service Name, Ports, Selector, ServiceType, Labels have not change ", "namespace ", svc.Namespace, " name ", svc.Name)
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
	klog.Info("Creating resource service ", "namespace ", object.GetNamespace(), " name ", object.GetName(), " kind ", object.GetObjectKind().GroupVersionKind().Kind)
	if err := k8sclient.Create(ctx, object); err != nil {
		return err
	}
	return nil
}

func UpdateClientObject(ctx context.Context, k8sclient client.Client, object client.Object) error {
	klog.Info("Updating resource service ", "namespace ", object.GetNamespace(), " name ", object.GetName(), " kind l", object.GetObjectKind())
	if err := k8sclient.Update(ctx, object); err != nil {
		return err
	}
	return nil
}

func DeleteClientObject(ctx context.Context, k8sclient client.Client, namespace, name string) (bool, error) {
	var ob client.Object
	err := k8sclient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, ob)
	if err != nil && apierrors.IsNotFound(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	if err := k8sclient.Delete(ctx, ob); err != nil {
		return true, nil
	}
	return true, nil
}

func PodIsReady(status *corev1.PodStatus) bool {
	for _, cs := range status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}

	return true
}
