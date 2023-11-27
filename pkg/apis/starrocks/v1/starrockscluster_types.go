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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StarRocksClusterSpec defines the desired state of StarRocksCluster
type StarRocksClusterSpec struct {
	// Specify a Service Account for starRocksCluster use k8s cluster.
	// +optional
	// Deprecated: component use serviceAccount in own's field.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// StarRocksFeSpec define fe configuration for start fe service.
	StarRocksFeSpec *StarRocksFeSpec `json:"starRocksFeSpec,omitempty"`

	// StarRocksBeSpec define be configuration for start be service.
	StarRocksBeSpec *StarRocksBeSpec `json:"starRocksBeSpec,omitempty"`

	// StarRocksCnSpec define cn configuration for start cn service.
	StarRocksCnSpec *StarRocksCnSpec `json:"starRocksCnSpec,omitempty"`

	// StarRocksLoadSpec define a proxy for fe.
	StarRocksFeProxySpec *StarRocksFeProxySpec `json:"starRocksFeProxySpec,omitempty"`
}

// StarRocksClusterStatus defines the observed state of StarRocksCluster.
type StarRocksClusterStatus struct {
	// Represents the state of cluster. the possible value are: running, failed, pending
	Phase Phase `json:"phase"`

	// Represents the status of fe. the status have running, failed and creating pods.
	StarRocksFeStatus *StarRocksFeStatus `json:"starRocksFeStatus,omitempty"`

	// Represents the status of be. the status have running, failed and creating pods.
	StarRocksBeStatus *StarRocksBeStatus `json:"starRocksBeStatus,omitempty"`

	// Represents the status of cn. the status have running, failed and creating pods.
	StarRocksCnStatus *StarRocksCnStatus `json:"starRocksCnStatus,omitempty"`

	// Represents the status of fe proxy. the status have running, failed and creating pods.
	StarRocksFeProxyStatus *StarRocksFeProxyStatus `json:"starRocksFeProxyStatus,omitempty"`
}

// SpecInterface is a common interface for all starrocks component spec.
// +kubebuilder:object:generate=false
type SpecInterface interface {
	loadInterface
	GetHostAliases() []corev1.HostAlias
	GetRunAsNonRoot() (*int64, *int64)
	GetTerminationGracePeriodSeconds() *int64
}

var _ SpecInterface = &StarRocksFeSpec{}
var _ SpecInterface = &StarRocksBeSpec{}
var _ SpecInterface = &StarRocksCnSpec{}
var _ SpecInterface = &StarRocksFeProxySpec{}

// StarRocksFeSpec defines the desired state of fe.
type StarRocksFeSpec struct {
	StarRocksComponentSpec `json:",inline"`

	// +optional
	// feEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	FeEnvVars []corev1.EnvVar `json:"feEnvVars,omitempty"`
}

// StarRocksBeSpec defines the desired state of be.
type StarRocksBeSpec struct {
	StarRocksComponentSpec `json:",inline"`

	// +optional
	// beEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	BeEnvVars []corev1.EnvVar `json:"beEnvVars,omitempty"`
}

// StarRocksCnSpec defines the desired state of cn.
type StarRocksCnSpec struct {
	StarRocksComponentSpec `json:",inline"`

	// +optional
	// cnEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	CnEnvVars []corev1.EnvVar `json:"cnEnvVars,omitempty"`

	// AutoScalingPolicy auto scaling strategy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
}

type StarRocksFeProxySpec struct {
	StarRocksLoadSpec `json:",inline"`

	Resolver string `json:"resolver,omitempty"`
}

// StarRocksFeStatus represents the status of starrocks fe.
type StarRocksFeStatus struct {
	StarRocksComponentStatus `json:",inline"`
}

// StarRocksBeStatus represents the status of starrocks be.
type StarRocksBeStatus struct {
	StarRocksComponentStatus `json:",inline"`
}

type StarRocksFeProxyStatus struct {
	StarRocksComponentStatus `json:",inline"`
}

// StarRocksCnStatus represents the status of starrocks cn.
type StarRocksCnStatus struct {
	StarRocksComponentStatus `json:",inline"`

	// The policy name of autoScale.
	// Deprecated
	HpaName string `json:"hpaName,omitempty"`

	// HorizontalAutoscaler have the autoscaler information.
	HorizontalScaler HorizontalScaler `json:"horizontalScaler,omitempty"`
}

func (spec *StarRocksFeSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.StarRocksComponentSpec.GetReplicas()
}

func (spec *StarRocksBeSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.StarRocksComponentSpec.GetReplicas()
}

func (spec *StarRocksCnSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.StarRocksComponentSpec.GetReplicas()
}

func (spec *StarRocksFeProxySpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.StarRocksLoadSpec.GetReplicas()
}

// GetHostAliases
// fe proxy does not have field HostAliases, the reason why implementing this method is
// that StarRocksFeProxySpec needs to implement SpecInterface interface
func (spec *StarRocksFeProxySpec) GetHostAliases() []corev1.HostAlias {
	// fe proxy do not support host alias
	return nil
}

// GetRunAsNonRoot
// fe proxy does not have field RunAsNonRoot, the reason why implementing this method is
// that StarRocksFeProxySpec needs to implement SpecInterface interface
func (spec *StarRocksFeProxySpec) GetRunAsNonRoot() (*int64, *int64) {
	// fe proxy will set run as nginx user by default, and can not be changed by crd
	return nil, nil
}

// GetTerminationGracePeriodSeconds
// fe proxy does not have field TerminationGracePeriodSeconds, the reason why implementing this method is
// that StarRocksFeProxySpec needs to implement SpecInterface interface
func (spec *StarRocksFeProxySpec) GetTerminationGracePeriodSeconds() *int64 {
	return nil
}

// Phase is defined under status, e.g.
// 1. StarRocksClusterStatus.Phase represents the phase of starrocks cluster.
// 2. StarRocksWarehouseStatus.Phase represents the phase of starrocks warehouse.
// The possible value for cluster phase are: running, failed, pending, deleting.
type Phase string

// ComponentPhase represent the component phase. e.g.
// 1. StarRocksCluster contains three components: FE, CN, BE.
// 2. StarRocksWarehouse reuse the CN component.
// The possible value for component phase are: reconciling, failed, running.
type ComponentPhase string

const (
	// ClusterRunning represents starrocks cluster is running.
	ClusterRunning Phase = "running"

	// ClusterFailed represents starrocks cluster failed.
	ClusterFailed Phase = "failed"

	// ClusterPending represents the starrocks cluster is creating
	ClusterPending Phase = "pending"

	// ClusterDeleting waiting all resource deleted
	ClusterDeleting Phase = "deleting"
)

const (
	// ComponentReconciling the starrocks have component in starting.
	ComponentReconciling ComponentPhase = "reconciling"

	// ComponentFailed have at least one service failed.
	ComponentFailed ComponentPhase = "failed"

	// ComponentRunning all components runs available.
	ComponentRunning ComponentPhase = "running"
)

// AnnotationOperationValue present the operation for fe, cn, be.
type AnnotationOperationValue string

type HorizontalScaler struct {
	// the horizontal scaler name
	Name string `json:"name,omitempty"`

	// the horizontal version.
	Version AutoScalerVersion `json:"version,omitempty"`
}

// StarRocksCluster defines a starrocks cluster deployment.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=src
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="FeStatus",type=string,JSONPath=`.status.starRocksFeStatus.phase`
// +kubebuilder:printcolumn:name="CnStatus",type=string,JSONPath=`.status.starRocksCnStatus.phase`
// +kubebuilder:printcolumn:name="BeStatus",type=string,JSONPath=`.status.starRocksBeStatus.phase`
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +genclient
type StarRocksCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired state of the starrocks cluster.
	Spec StarRocksClusterSpec `json:"spec,omitempty"`

	// Most recent observed status of the starrocks cluster
	Status StarRocksClusterStatus `json:"status,omitempty"`
}

// StorageVolume defines additional PVC template for StatefulSets and volumeMount for pods that mount this PVC
type StorageVolume struct {
	// name of a storage volume.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name"`

	// storageClassName is the name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// StorageSize is a valid memory size type based on powers-of-2, so 1Mi is 1024Ki.
	// Supported units:Mi, Gi, GiB, Ti, Ti, Pi, Ei, Ex: `512Mi`.
	// +kubebuilder:validation:Pattern:="(^0|([0-9]*l[.])?[0-9]+((M|G|T|E|P)i))$"
	// +optional
	StorageSize string `json:"storageSize,omitempty"`

	// MountPath specify the path of volume mount.
	MountPath string `json:"mountPath"`

	// SubPath within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	SubPath string `json:"subPath,omitempty"`
}

// StarRocksClusterList contains a list of StarRocksCluster
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type StarRocksClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StarRocksCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StarRocksCluster{}, &StarRocksClusterList{})
}
