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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	v2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodAutoscalerParams struct {
	Namespace       string
	Name            string
	Labels          Labels
	TargetName      string
	OwnerReferences []metav1.OwnerReference
	ScalerPolicy    *srapi.AutoScalingPolicy
}

func BuildHorizontalPodAutoscaler(pap *PodAutoscalerParams) *v2.HorizontalPodAutoscaler {
	return &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            pap.Name,
			Namespace:       pap.Namespace,
			Labels:          pap.Labels,
			OwnerReferences: pap.OwnerReferences,
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Name:       pap.TargetName,
				Kind:       "StatefulSet",
				APIVersion: "apps/v1",
			},
			MaxReplicas: pap.ScalerPolicy.MaxReplicas,
			MinReplicas: pap.ScalerPolicy.MinReplicas,
			Metrics:     pap.ScalerPolicy.HPAPolicy.Metrics,
			Behavior:    pap.ScalerPolicy.HPAPolicy.Behavior,
		},
	}
}
