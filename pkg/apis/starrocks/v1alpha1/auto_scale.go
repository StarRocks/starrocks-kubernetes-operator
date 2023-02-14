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

package v1alpha1

import (
	v2 "k8s.io/api/autoscaling/v2"
)

//AutoScalingPolicy defines the auto scale
type AutoScalingPolicy struct {
	//the policy of cn autoscale. operator use autoscaling v2.
	HPAPolicy *HPAPolicy `json:"hpaPolicy,omitempty"`

	//the min numbers of target.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// the max numbers of target.
	//+optional
	MaxReplicas int32 `json:"maxReplicas"`
}

//
type HPAPolicy struct {
	// +optional
	// Metrics specifies how to scale based on a single metric
	Metrics []v2.MetricSpec `json:"metrics,omitempty"`

	// +optional
	// HorizontalPodAutoscalerBehavior configures the scaling behavior of the target
	Behavior *v2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}
