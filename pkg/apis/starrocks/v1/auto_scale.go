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

package v1

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoScalingPolicy defines the auto scale
type AutoScalingPolicy struct {
	// the policy of autoscaling. operator use autoscaling v2.
	HPAPolicy *HPAPolicy `json:"hpaPolicy,omitempty"`

	// version represents the autoscaler version for cn service. only support v1,v2beta2,v2
	// +optional
	Version AutoScalerVersion `json:"version,omitempty"`

	// MinReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down. It defaults to 1 pod.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// MaxReplicas is the upper limit for the number of pods that can be set by the autoscaler;
	// cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas"`
}

type AutoScalerVersion string

const (
	// AutoScalerV1 the cn service use v1 autoscaler. Reference to https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
	AutoScalerV1 AutoScalerVersion = "v1"

	// AutoScalerV2Beta2 the cn service use v2beta2. Reference to  https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
	AutoScalerV2Beta2 AutoScalerVersion = "v2beta2"

	// AutoScalerV2 the cn service use v2. Reference to  https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
	AutoScalerV2 AutoScalerVersion = "v2"
)

// Complete completes the default value of AutoScalerVersion
func (version AutoScalerVersion) Complete(major, minor string) AutoScalerVersion {
	if version != "" {
		return version
	}
	// operator choose a proper default hpa version by checking ths kubernetes version
	// if kubernetes version > 1.25, use v2 version
	if major == "1" {
		minorNumber, err := strconv.Atoi(minor)
		if err != nil {
			// keep backward compatibility
			return AutoScalerV2Beta2
		}
		if minorNumber > 25 {
			return AutoScalerV2
		} else {
			return AutoScalerV2Beta2
		}
	}
	return AutoScalerV2
}

type HPAPolicy struct {
	// +optional
	// Metrics specifies how to scale based on a single metric
	// the struct copy from k8s.io/api/autoscaling/v2beta2/types.go. the redundancy code will hide the restriction about
	// HorizontalPodAutoscaler version and kubernetes releases matching issue.
	// the splice will have unsafe.Pointer convert, so be careful to edit the struct fields.
	Metrics []MetricSpec `json:"metrics,omitempty"`

	// +optional
	// HorizontalPodAutoscalerBehavior configures the scaling behavior of the target.
	// the struct copy from k8s.io/api/autoscaling/v2beta2/types.go. the redundancy code will hide the restriction about
	// HorizontalPodAutoscaler version and kubernetes releases matching issue.
	// the
	Behavior *HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// MetricSourceType indicates the type of metric.
type MetricSourceType string

const (
	// ObjectMetricSourceType is a metric describing a kubernetes object
	// (for example, hits-per-second on an Ingress object).
	ObjectMetricSourceType MetricSourceType = "Object"
	// PodsMetricSourceType is a metric describing each pod in the current scale
	// target (for example, transactions-processed-per-second).  The values
	// will be averaged together before being compared to the target value.
	PodsMetricSourceType MetricSourceType = "Pods"
	// ResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ResourceMetricSourceType MetricSourceType = "Resource"
	// ContainerResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing a single container in each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ContainerResourceMetricSourceType MetricSourceType = "ContainerResource"
	// ExternalMetricSourceType is a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	ExternalMetricSourceType MetricSourceType = "External"
)

// ObjectMetricSource indicates how to scale on a metric describing a
// kubernetes object (for example, hits-per-second on an Ingress object).
type ObjectMetricSource struct {
	DescribedObject CrossVersionObjectReference `json:"describedObject" protobuf:"bytes,1,name=describedObject"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,3,name=metric"`
}

// MetricTarget defines the target value, average value, or average utilization of a specific metric
type MetricTarget struct {
	// type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type" protobuf:"bytes,1,name=type"`
	// value is the target value of the metric (as a quantity).
	// +optional
	Value *resource.Quantity `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// averageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *resource.Quantity `json:"averageValue,omitempty" protobuf:"bytes,3,opt,name=averageValue"`
	// averageUtilization is the target value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// Currently only valid for Resource metric source type
	// +optional
	AverageUtilization *int32 `json:"averageUtilization,omitempty" protobuf:"bytes,4,opt,name=averageUtilization"`
}

// MetricTargetType specifies the type of metric being targeted, and should be either
// "Value", "AverageValue", or "Utilization"
type MetricTargetType string

const (
	// UtilizationMetricType declares a MetricTarget is an AverageUtilization value
	UtilizationMetricType MetricTargetType = "Utilization"
	// ValueMetricType declares a MetricTarget is a raw value
	ValueMetricType MetricTargetType = "Value"
	// AverageValueMetricType declares a MetricTarget is an
	AverageValueMetricType MetricTargetType = "AverageValue"
)

// MetricIdentifier defines the name and optionally selector for a metric
type MetricIdentifier struct {
	// name is the name of the given metric
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// selector is the string-encoded form of a standard kubernetes label selector for the given metric
	// When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping.
	// When unset, just the metricName will be used to gather metrics.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,name=selector"`
}

// PodsMetricSource indicates how to scale on a metric describing each pod in
// the current scale target (for example, transactions-processed-per-second).
// The values will be averaged together before being compared to the target
// value.
type PodsMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}

// ResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ResourceMetricSource struct {
	// name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// ContainerResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ContainerResourceMetricSource struct {
	// name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// container is the name of the container in the pods of the scaling target
	Container string `json:"container" protobuf:"bytes,3,opt,name=container"`
}

// ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction
type ScalingPolicySelect string

const (
	// MaxPolicySelect selects the policy with the highest possible change.
	MaxPolicySelect ScalingPolicySelect = "Max"
	// MinPolicySelect selects the policy with the lowest possible change.
	MinPolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// HPAScalingPolicyType is the type of the policy which could be used while making scaling decisions.
type HPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy HPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy HPAScalingPolicyType = "Percent"
)

// HPAScalingPolicy is a single policy which must hold true for a specified past interval.
type HPAScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type HPAScalingPolicyType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=HPAScalingPolicyType"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int32 `json:"value" protobuf:"varint,2,opt,name=value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int32 `json:"periodSeconds" protobuf:"varint,3,opt,name=periodSeconds"`
}

// HPAScalingRules configures the scaling behavior for one direction.
// These Rules are applied after calculating DesiredReplicas from metrics for the HPA.
// They can limit the scaling velocity by specifying scaling policies.
// They can prevent flapping by specifying the stabilization window, so that the
// number of replicas is not set instantly, instead, the safest value from the stabilization
// window is chosen.
type HPAScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds *int32 `json:"stabilizationWindowSeconds,omitempty" protobuf:"varint,3,opt,name=stabilizationWindowSeconds"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value MaxPolicySelect is used.
	// +optional
	SelectPolicy *ScalingPolicySelect `json:"selectPolicy,omitempty" protobuf:"bytes,1,opt,name=selectPolicy"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +optional
	Policies []HPAScalingPolicy `json:"policies,omitempty" protobuf:"bytes,2,rep,name=policies"`
}

// HorizontalPodAutoscalerBehavior configures the scaling behavior of the target
// in both Up and Down directions (scaleUp and scaleDown fields respectively).
type HorizontalPodAutoscalerBehavior struct {
	// scaleUp is scaling policy for scaling Up.
	// If not set, the default value is the higher of:
	//   * increase no more than 4 pods per 60 seconds
	//   * double the number of pods per 60 seconds
	// No stabilization is used.
	// +optional
	ScaleUp *HPAScalingRules `json:"scaleUp,omitempty" protobuf:"bytes,1,opt,name=scaleUp"`
	// scaleDown is scaling policy for scaling Down.
	// If not set, the default value is to allow to scale down to minReplicas pods, with a
	// 300 second stabilization window (i.e., the highest recommendation for
	// the last 300sec is used).
	// +optional
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty" protobuf:"bytes,2,opt,name=scaleDown"`
}

// MetricSpec specifies how to scale based on a single metric
// (only `type` and one other matching field should be set at once).
type MetricSpec struct {
	// type is the type of metric source.  It should be one of "ContainerResource", "External",
	// "Object", "Pods" or "Resource", each mapping to a matching field in the object.
	// Note: "ContainerResource" type is available on when the feature-gate
	// HPAContainerMetrics is enabled
	Type MetricSourceType `json:"type" protobuf:"bytes,1,name=type"`

	// object refers to a metric describing a single kubernetes object
	// (for example, hits-per-second on an Ingress object).
	// +optional
	Object *ObjectMetricSource `json:"object,omitempty" protobuf:"bytes,2,opt,name=object"`
	// pods refers to a metric describing each pod in the current scale target
	// (for example, transactions-processed-per-second).  The values will be
	// averaged together before being compared to the target value.
	// +optional
	Pods *PodsMetricSource `json:"pods,omitempty" protobuf:"bytes,3,opt,name=pods"`
	// resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing each pod in the
	// current scale target (e.g. CPU or memory). Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics using the "pods" source.
	// +optional
	Resource *ResourceMetricSource `json:"resource,omitempty" protobuf:"bytes,4,opt,name=resource"`
	// container resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing a single container in
	// each pod of the current scale target (e.g. CPU or memory). Such metrics are
	// built in to Kubernetes, and have special scaling options on top of those
	// available to normal per-pod metrics using the "pods" source.
	// This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.
	// +optional
	ContainerResource *ContainerResourceMetricSource `json:"containerResource,omitempty" protobuf:"bytes,7,opt,name=containerResource"`
	// external refers to a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	// +optional
	External *ExternalMetricSource `json:"external,omitempty" protobuf:"bytes,5,opt,name=external"`
}

// ExternalMetricSource indicates how to scale on a metric not associated with
// any Kubernetes object (for example length of queue in cloud
// messaging service, or QPS from loadbalancer running outside of cluster).
type ExternalMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}
