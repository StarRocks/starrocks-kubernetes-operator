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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	"k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildHorizontalPodAutoscaler(namespace, name, targetName string, labels Labels, references []metav1.OwnerReference, scalepolicy *srapi.AutoScalingPolicy) *v2.HorizontalPodAutoscaler {
	return &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			//TODO: generate cn hpa name
			Name:      name,
			Namespace: namespace,
			//TODO: construct labels
			Labels:          labels,
			OwnerReferences: references,
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				//TODO: generate target's name
				Name:       targetName,
				Kind:       common.CnKind,
				APIVersion: common.CnApiVersionV1ALPHA,
			},
			MaxReplicas: scalepolicy.MaxReplicas,
			MinReplicas: scalepolicy.MinReplicas,
			Metrics:     scalepolicy.HPAPolicy.Metrics,
			Behavior:    scalepolicy.HPAPolicy.Behavior,
		},
	}
}
