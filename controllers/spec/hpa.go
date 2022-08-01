/*
Copyright 2022 StarRocks.

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
package spec

import (
	"github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// build a hpa base on cn
func MakeCnHPA(cn *v1alpha1.ComputeNodeGroup) *autoscalingv2beta2.HorizontalPodAutoscaler {
	if cn.Spec.AutoScalingPolicy == nil {
		return nil
	}
	if cn.Spec.AutoScalingPolicy.HPAPolicy == nil {
		return nil
	}
	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cn.Name,
			Namespace: cn.Namespace,
			Labels:    cn.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cn, cn.GroupVersionKind()),
			},
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Name:       cn.Name,
				Kind:       common.CnKind,
				APIVersion: common.CnApiVersionV1ALPHA,
			},
			MaxReplicas: cn.Spec.AutoScalingPolicy.MaxReplicas,
			MinReplicas: cn.Spec.AutoScalingPolicy.MinReplicas,
			Metrics:     cn.Spec.AutoScalingPolicy.HPAPolicy.Metrics,
			Behavior:    cn.Spec.AutoScalingPolicy.HPAPolicy.Behavior,
		},
	}
}

// sync changed
// only some fields would be synced
func SyncHPAChanged(current, desired *autoscalingv2beta2.HorizontalPodAutoscaler) {
	current.Labels = makeAnnotationsOrLabels(desired.Labels, current.Labels)
	current.Spec = desired.Spec
}
