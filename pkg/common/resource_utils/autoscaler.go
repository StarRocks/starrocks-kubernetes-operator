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

package resource_utils

import (
	"unsafe"

	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
)

var (
	AutoscalerKind         = "HorizontalPodAutoscaler"
	StatefulSetKind        = "StatefulSet"
	StarRocksClusterKind   = "StarRocksCluster"
	StarRocksWarehouseKind = "StarRocksWarehouse"
	ServiceKind            = "Service"
)

// HPAParams defines the parameters for creating a HorizontalPodAutoscaler resource.
type HPAParams struct {
	// Version defines the version of HPA to be used.
	Version srapi.AutoScalerVersion

	// Name is the name of the HorizontalPodAutoscaler resource.
	Name string

	// Namespace is the namespace where the HorizontalPodAutoscaler resource will be created.
	// It will be in the same namespace as the target resource.
	Namespace string

	// Labels are the labels to be applied to the HorizontalPodAutoscaler resource.
	// Now there are two labels: app.kubernetes.io/component: autoscaler and app.starrocks.ownerreference/name: <target-name>.
	Labels map[string]string

	// OwnerReferences includes only one OwnerReference, which is the StarRocksCluster or StarRocksWarehouse object that
	// this HPA belongs to.
	OwnerReferences []metav1.OwnerReference

	// ScalerPolicy defines the scaling policy for the HorizontalPodAutoscaler.
	ScalerPolicy *srapi.AutoScalingPolicy
}

func BuildHPA(hpaParams *HPAParams, autoScalerVersion srapi.AutoScalerVersion) client.Object {
	if autoScalerVersion == "" {
		autoScalerVersion = hpaParams.Version.Complete(k8sutils.KUBE_MAJOR_VERSION, k8sutils.KUBE_MINOR_VERSION)
	}

	getTypeMeta := func(version srapi.AutoScalerVersion) metav1.TypeMeta {
		meta := metav1.TypeMeta{
			Kind: AutoscalerKind,
		}
		switch version {
		case srapi.AutoScalerV1:
			meta.APIVersion = v1.SchemeGroupVersion.String()
		case srapi.AutoScalerV2Beta2:
			meta.APIVersion = v2beta2.SchemeGroupVersion.String()
		case srapi.AutoScalerV2:
			meta.APIVersion = v2.SchemeGroupVersion.String()
		}
		return meta
	}

	getObjectMeta := func(hpaParams *HPAParams) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name:            hpaParams.Name,
			Namespace:       hpaParams.Namespace,
			Labels:          hpaParams.Labels,
			OwnerReferences: hpaParams.OwnerReferences,
		}
	}

	scaleTargetRef := &v1.CrossVersionObjectReference{
		Name:       hpaParams.OwnerReferences[0].Name,
		Kind:       hpaParams.OwnerReferences[0].Kind,
		APIVersion: hpaParams.OwnerReferences[0].APIVersion,
	}

	switch autoScalerVersion {
	case srapi.AutoScalerV1:
		return &v1.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(hpaParams),
			Spec: v1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *(scaleTargetRef),
				MaxReplicas:    hpaParams.ScalerPolicy.MaxReplicas,
				MinReplicas:    hpaParams.ScalerPolicy.MinReplicas,
			},
		}
	case srapi.AutoScalerV2:
		hpa := &v2.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(hpaParams),
			Spec: v2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *((*v2.CrossVersionObjectReference)(unsafe.Pointer(scaleTargetRef))),
				MaxReplicas:    hpaParams.ScalerPolicy.MaxReplicas,
				MinReplicas:    hpaParams.ScalerPolicy.MinReplicas,
			},
		}
		// the codes use unsafe.Pointer to convert struct, when audit please notice the correctness about memory assign.
		if hpaParams.ScalerPolicy != nil && hpaParams.ScalerPolicy.HPAPolicy != nil {
			if len(hpaParams.ScalerPolicy.HPAPolicy.Metrics) != 0 {
				metrics := unsafe.Slice((*v2.MetricSpec)(unsafe.Pointer(&hpaParams.ScalerPolicy.HPAPolicy.Metrics[0])),
					len(hpaParams.ScalerPolicy.HPAPolicy.Metrics))
				hpa.Spec.Metrics = metrics
			}
			hpa.Spec.Behavior = (*v2.HorizontalPodAutoscalerBehavior)(unsafe.Pointer(hpaParams.ScalerPolicy.HPAPolicy.Behavior))
		}
		return hpa
	default:
		// case srapi.AutoScalerV2Beta2:
		hpa := &v2beta2.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(hpaParams),
			Spec: v2beta2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *((*v2beta2.CrossVersionObjectReference)(unsafe.Pointer(scaleTargetRef))),
				MaxReplicas:    hpaParams.ScalerPolicy.MaxReplicas,
				MinReplicas:    hpaParams.ScalerPolicy.MinReplicas,
			},
		}
		if hpaParams.ScalerPolicy != nil && hpaParams.ScalerPolicy.HPAPolicy != nil {
			hpa.Spec.Metrics = hpaParams.ScalerPolicy.HPAPolicy.Metrics
			hpa.Spec.Behavior = hpaParams.ScalerPolicy.HPAPolicy.Behavior
		}
		return hpa
	}
}
