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

func BuildHorizontalPodAutoscaler(src *srapi.StarRocksCluster) *v2.HorizontalPodAutoscaler {
	if src.Spec.StarRocksCnSpec == nil || src.Spec.StarRocksCnSpec.AutoScalingPolicy == nil ||
		src.Spec.StarRocksCnSpec.AutoScalingPolicy.HPAPolicy == nil {
		return nil
	}

	cnspec := src.Spec.StarRocksCnSpec
	return &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			//TODO: 构造cn hpa的名字
			Name:      cnspec,
			Namespace: src.Namespace,
			//TODO: 构造labels
			Labels: cn.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(src, src.GroupVersionKind()),
			},
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				//TODO: 构建target的name
				Name:       cn.Name,
				Kind:       common.CnKind,
				APIVersion: common.CnApiVersionV1ALPHA,
			},
			MaxReplicas: cnspec.AutoScalingPolicy.MaxReplicas,
			MinReplicas: cnspec.AutoScalingPolicy.MinReplicas,
			Metrics:     cnspec.AutoScalingPolicy.HPAPolicy.Metrics,
			Behavior:    cnspec.AutoScalingPolicy.HPAPolicy.Behavior,
		},
	}
}
