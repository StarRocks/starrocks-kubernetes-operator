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

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
)

var (
	AutoscalerKind  = "HorizontalPodAutoscaler"
	StatefulSetKind = "StatefulSet"
	ServiceKind     = "Service"
)

type PodAutoscalerParams struct {
	AutoscalerType  srapi.AutoScalerVersion
	Namespace       string
	Name            string
	Labels          Labels
	TargetName      string
	OwnerReferences []metav1.OwnerReference
	ScalerPolicy    *srapi.AutoScalingPolicy
}

func BuildHorizontalPodAutoscaler(pap *PodAutoscalerParams, autoScalerVersion srapi.AutoScalerVersion) client.Object {
	if autoScalerVersion == "" {
		autoScalerVersion = pap.AutoscalerType.Complete(k8sutils.KUBE_MAJOR_VERSION, k8sutils.KUBE_MINOR_VERSION)
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

	getObjectMeta := func(pap *PodAutoscalerParams) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name:            pap.Name,
			Namespace:       pap.Namespace,
			Labels:          pap.Labels,
			OwnerReferences: pap.OwnerReferences,
		}
	}

	scaleTargetRef := &v1.CrossVersionObjectReference{
		Name:       pap.TargetName,
		Kind:       StatefulSetKind,
		APIVersion: appv1.SchemeGroupVersion.String(),
	}

	switch autoScalerVersion {
	case srapi.AutoScalerV1:
		return &v1.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(pap),
			Spec: v1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *(scaleTargetRef),
				MaxReplicas:    pap.ScalerPolicy.MaxReplicas,
				MinReplicas:    pap.ScalerPolicy.MinReplicas,
			},
		}
	case srapi.AutoScalerV2:
		hpa := &v2.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(pap),
			Spec: v2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *((*v2.CrossVersionObjectReference)(unsafe.Pointer(scaleTargetRef))),
				MaxReplicas:    pap.ScalerPolicy.MaxReplicas,
				MinReplicas:    pap.ScalerPolicy.MinReplicas,
			},
		}
		// the codes use unsafe.Pointer to convert struct, when audit please notice the correctness about memory assign.
		if pap.ScalerPolicy != nil && pap.ScalerPolicy.HPAPolicy != nil {
			if len(pap.ScalerPolicy.HPAPolicy.Metrics) != 0 {
				metrics := unsafe.Slice((*v2.MetricSpec)(unsafe.Pointer(&pap.ScalerPolicy.HPAPolicy.Metrics[0])),
					len(pap.ScalerPolicy.HPAPolicy.Metrics))
				hpa.Spec.Metrics = metrics
			}
			hpa.Spec.Behavior = (*v2.HorizontalPodAutoscalerBehavior)(unsafe.Pointer(pap.ScalerPolicy.HPAPolicy.Behavior))
		}
		return hpa
	default:
		// case srapi.AutoScalerV2Beta2:
		hpa := &v2beta2.HorizontalPodAutoscaler{
			TypeMeta:   getTypeMeta(autoScalerVersion),
			ObjectMeta: getObjectMeta(pap),
			Spec: v2beta2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: *((*v2beta2.CrossVersionObjectReference)(unsafe.Pointer(scaleTargetRef))),
				MaxReplicas:    pap.ScalerPolicy.MaxReplicas,
				MinReplicas:    pap.ScalerPolicy.MinReplicas,
			},
		}
		if pap.ScalerPolicy != nil && pap.ScalerPolicy.HPAPolicy != nil {
			hpa.Spec.Metrics = pap.ScalerPolicy.HPAPolicy.Metrics
			hpa.Spec.Behavior = pap.ScalerPolicy.HPAPolicy.Behavior
		}
		return hpa
	}
}
