/*
 * Copyright 2021-present, StarRocks Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// SpecInterface is a common interface for all starrocks component spec.
// +kubebuilder:object:generate=false
type SpecInterface interface {
	GetReplicas() *int32
	GetServiceName() string
	GetStorageVolumes() []StorageVolume
	GetServiceAccount() string
	GetAffinity() *corev1.Affinity
	GetTolerations() []corev1.Toleration
	GetNodeSelector() map[string]string
	GetImagePullSecrets() []corev1.LocalObjectReference
	GetHostAliases() []corev1.HostAlias
	GetSchedulerName() string
	GetFsGroup() *int64
	GetAnnotations() map[string]string
}

var _ SpecInterface = &StarRocksFeSpec{}
var _ SpecInterface = &StarRocksBeSpec{}
var _ SpecInterface = &StarRocksCnSpec{}

// StarRocksFeSpec defines the desired state of fe.
type StarRocksFeSpec struct {
	StarRocksComponentSpec `json:",inline"`

	//StorageVolumes defines the additional storage for meta storage.
	//+optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	//+optional
	//feEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	FeEnvVars []corev1.EnvVar `json:"feEnvVars,omitempty"`
}

// StarRocksBeSpec defines the desired state of be.
type StarRocksBeSpec struct {
	StarRocksComponentSpec `json:",inline"`

	//StorageVolumes defines the additional storage for meta storage.
	//+optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	//+optional
	//beEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	BeEnvVars []corev1.EnvVar `json:"beEnvVars,omitempty"`
}

// StarRocksCnSpec defines the desired state of cn.
type StarRocksCnSpec struct {
	StarRocksComponentSpec `json:",inline"`

	//+optional
	//cnEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	CnEnvVars []corev1.EnvVar `json:"cnEnvVars,omitempty"`

	//AutoScalingPolicy auto scaling strategy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
}

type StarRocksComponentSpec struct {
	//name of the starrocks be cluster.
	//+optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//Deprecated, not allow set statefulset name.
	Name string `json:"name,omitempty"`

	//annotation for pods. user can config monitor annotation for collect to monitor system.
	Annotations map[string]string `json:"annotations,omitempty"`

	//serviceAccount for access cloud service.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	//A special supplemental group that applies to all containers in a pod.
	// Some volume types allow the Kubelet to change the ownership of that volume
	// to be owned by the pod:
	FsGroup *int64 `json:"fsGroup,omitempty"`

	//Replicas is the number of desired Pod, the number is 1,3,5
	// +kubebuilder:validation:Minimum=0
	//+optional: Defaults to 3
	Replicas *int32 `json:"replicas,omitempty"`

	//Image for a starrocks deployment..
	Image string `json:"image"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	//Service defines the template for the associated Kubernetes Service object.
	//+optional
	Service *StarRocksService `json:"service,omitempty"`

	//defines the specification of resource cpu and mem.
	//+optional
	corev1.ResourceRequirements `json:",inline"`

	//the reference for configMap.
	//+optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	//the reference for secrets.
	//+optional
	Secrets []SecretInfo `json:"secrets,omitempty"`

	//Probe defines the mode probe service in container is alive.
	//+optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	// (Optional) If specified, the pod's nodeSelector，displayName="Map of nodeSelectors to match when scheduling pods on nodes"
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	//+optional
	//If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations for scheduling pods onto some dedicated nodes
	//+optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	//+optional
	//the pod labels for user select or classify pods.
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`

	// SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`
}

func (spec *StarRocksComponentSpec) GetAnnotations() map[string]string {
	return spec.Annotations
}

// StarRocksFeStatus represents the status of starrocks fe.
type StarRocksFeStatus struct {
	StarRocksComponentStatus `json:",inline"`
}

// StarRocksBeStatus represents the status of starrocks be.
type StarRocksBeStatus struct {
	StarRocksComponentStatus `json:",inline"`
}

// StarRocksCnStatus represents the status of starrocks cn.
type StarRocksCnStatus struct {
	StarRocksComponentStatus `json:",inline"`

	//The policy name of autoScale.
	//Deprecated
	HpaName string `json:"hpaName,omitempty"`

	//HorizontalAutoscaler have the autoscaler information.
	HorizontalScaler HorizontalScaler `json:"horizontalScaler,omitempty"`
}

// StarRocksComponentStatus represents the status of a starrocks component.
type StarRocksComponentStatus struct {
	//the name of fe service exposed for user.
	ServiceName string `json:"serviceName,omitempty"`

	//FailedInstances failed pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	//CreatingInstances in creating pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	//RunningInstances in running status pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	//ResourceNames the statefulset names of fe in v1alpha1 version.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of component status. If component have one failed pod phase=failed,
	// also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	//+optional
	//Reason represents the reason of not running.
	Reason string `json:"reason"`
}

type StarRocksService struct {
	//Name assigned to service.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	// +optional
	Name string `json:"name,omitempty"`

	//type of service,the possible value for the service type are : ClusterIP, NodePort, LoadBalancer,ExternalName.
	//More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`

	// Only applies to Service Type: LoadBalancer.
	// This feature depends on whether the underlying cloud-provider supports specifying
	// the loadBalancerIP when a load balancer is created.
	// This field will be ignored if the cloud-provider does not support the feature.
	// This field was under-specified and its meaning varies across implementations,
	// and it cannot support dual-stack.
	// As of Kubernetes v1.24, users are encouraged to use implementation-specific annotations when available.
	// This field may be removed in a future API version.
	// +optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	//Ports the components exposed ports and listen ports in pod.
	// +optional
	Ports []StarRocksServicePort `json:"ports"`

	// Annotations store Kubernetes Service annotations.
	Annotations map[string]string `json:"annotations,omitempty"`
}

type StarRocksServicePort struct {
	//Name of the map about coming port and target port
	Name string `json:"name,omitempty"`

	//Port the pod is exposed on service.
	Port int32 `json:"port"`

	//ContainerPort the service listen in pod.
	ContainerPort int32 `json:"containerPort"`

	//The easiest way to expose fe, cn or be is to use a Service of type `NodePort`.
	NodePort int32 `json:"nodePort,omitempty"`
}

// StarRocksProbe defines the mode for probe be alive.
type StarRocksProbe struct {
	//Type identifies the mode of probe main container
	// +kubebuilder:validation:Enum=tcp;command
	Type string `json:"type"`

	// Number of seconds after the container has started before liveness probes are initiated.
	// Default to 10 seconds.
	// +kubebuilder:validation:Minimum=0
	// +optional
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`

	// How often (in seconds) to perform the probe.
	// Default to Kubernetes default (10 seconds). Minimum value is 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`
}

type ConfigMapInfo struct {
	//the config info for start progress.
	ConfigMapName string `json:"configMapName,omitempty"`

	//the config response key in configmap.
	ResolveKey string `json:"resolveKey,omitempty"`
}

type SecretInfo struct {
	// This must match the Name of a Volume.
	Name string `json:"name,omitempty"`

	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath,omitempty"`
}

func (spec *StarRocksComponentSpec) GetReplicas() *int32 {
	return spec.Replicas
}

func (spec *StarRocksComponentSpec) GetServiceName() string {
	if spec == nil || spec.Service == nil {
		return ""
	}
	return spec.Service.Name
}

func (spec *StarRocksComponentSpec) GetServiceAccount() string {
	return spec.ServiceAccount
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

// GetServiceName returns the service name of starrocks fe.
// If the spec is nil, return empty string.
func (spec *StarRocksFeSpec) GetServiceName() string {
	if spec == nil {
		return ""
	}
	return spec.StarRocksComponentSpec.GetServiceName()
}

func (spec *StarRocksBeSpec) GetServiceName() string {
	if spec == nil {
		return ""
	}
	return spec.StarRocksComponentSpec.GetServiceName()
}

func (spec *StarRocksCnSpec) GetServiceName() string {
	if spec == nil {
		return ""
	}
	return spec.StarRocksComponentSpec.GetServiceName()
}

func (spec *StarRocksFeSpec) GetStorageVolumes() []StorageVolume {
	return spec.StorageVolumes
}

func (spec *StarRocksBeSpec) GetStorageVolumes() []StorageVolume {
	return spec.StorageVolumes
}

func (spec *StarRocksCnSpec) GetStorageVolumes() []StorageVolume {
	return nil
}

func (spec *StarRocksComponentSpec) GetAffinity() *corev1.Affinity {
	return spec.Affinity
}

func (spec *StarRocksComponentSpec) GetTolerations() []corev1.Toleration {
	return spec.Tolerations
}

func (spec *StarRocksComponentSpec) GetNodeSelector() map[string]string {
	return spec.NodeSelector
}

func (spec *StarRocksComponentSpec) GetImagePullSecrets() []corev1.LocalObjectReference {
	return spec.ImagePullSecrets
}

func (spec *StarRocksComponentSpec) GetHostAliases() []corev1.HostAlias {
	return spec.HostAliases
}

func (spec *StarRocksComponentSpec) GetSchedulerName() string {
	return spec.SchedulerName
}

func (spec *StarRocksComponentSpec) GetFsGroup() *int64 {
	return spec.FsGroup
}
