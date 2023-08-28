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

package v1alpha1

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
}

// StarRocksClusterStatus defines the observed state of StarRocksCluster.
type StarRocksClusterStatus struct {
	// Represents the state of cluster. the possible value are: running, failed, pending
	Phase ClusterPhase `json:"phase"`
	// Represents the status of fe. the status have running, failed and creating pods.
	StarRocksFeStatus *StarRocksFeStatus `json:"starRocksFeStatus,omitempty"`

	// Represents the status of be. the status have running, failed and creating pods.
	StarRocksBeStatus *StarRocksBeStatus `json:"starRocksBeStatus,omitempty"`

	// Represents the status of cn. the status have running, failed and creating pods.
	StarRocksCnStatus *StarRocksCnStatus `json:"starRocksCnStatus,omitempty"`
}

// represent the cluster phase. the possible value for cluster phase are: running, failed, pending.
type ClusterPhase string

// represent the component phase about be, cn, be. the possible value for component phase are: reconciliing, failed, running, waitting.
type MemberPhase string

const (
	// ClusterRunning represents starrocks cluster is running.
	ClusterRunning ClusterPhase = "running"

	// ClusterFailed represents starrocks cluster failed.
	ClusterFailed ClusterPhase = "failed"

	// ClusterPending represents the starrocks cluster is creating
	ClusterPending ClusterPhase = "pending"

	// ClusterDeleting waiting all resource deleted
	ClusterDeleting = "deleting"
)

const (
	// ComponentReconciling the starrocks have component in starting.
	ComponentReconciling MemberPhase = "reconciling"
	// ComponentFailed have at least one service failed.
	ComponentFailed MemberPhase = "failed"
	// ComponentRunning all components runs available.
	ComponentRunning MemberPhase = "running"
)

// AnnotationOperationValue present the operation for fe, cn, be.
type AnnotationOperationValue string

const (
	// represent the user want to restart all fe pods.
	AnnotationRestart AnnotationOperationValue = "restart"
	// represent all fe pods have restarted.
	AnnotationRestartFinished AnnotationOperationValue = "finished"
	// represent at least one pod on restarting
	AnnotationRestarting AnnotationOperationValue = "restarting"
)

// Operation response key in annnotation, the annotation key be associated with annotation value represent the process status of sr operation.
type AnnotationOperationKey string

const (
	// the fe annotation key for restart
	AnnotationFERestartKey AnnotationOperationKey = "app.starrocks.fe.io/restart"

	// the be annotation key for restart be
	AnnotationBERestartKey AnnotationOperationKey = "app.starrocks.be.io/restart"

	// the cn annotation key for restart cn
	AnnotationCNRestartKey AnnotationOperationKey = "app.starrocks.cn.io/restart"
)

// StarRocksFeStatus represents the status of starrocks fe.
type StarRocksFeStatus struct {
	// the name of fe service exposed for user.
	ServiceName string `json:"serviceName,omitempty"`

	// FailedInstances failed fe pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	// CreatingInstances in creating pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	// RunningInstances in running status pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	// ResourceNames the statefulset names of fe in v1alpha1 version.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of fe status. If fe have one failed pod phase=failed,
	// also if fe have one creating pod phase=creating, also if fe all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// +optional
	// Reason represents the reason of not running.
	Reason string `json:"reason"`
}

// StarRocksBeStatus represents the status of starrocks be.
type StarRocksBeStatus struct {
	// the name of be service for fe find be instance.
	ServiceName string `json:"serviceName,omitempty"`

	// FailedInstances deploy failed instance of be.
	FailedInstances []string `json:"failedInstances,omitempty"`

	// CreatingInstances represents status in creating pods of be.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	// RunningInstances represents status in running pods of be.
	RunningInstances []string `json:"runningInstances,omitempty"`

	// The statefulset names of be.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of be status. If be have one failed pod phase=failed,
	// also if be have one creating pod phase=creating, also if be all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// the reason for the phase.
	// +optional
	Reason string `json:"reason"`
}

type HorizontalScaler struct {
	// the deploy horizontal scaler name
	Name string `json:"name,omitempty"`

	// the deploy horizontal version.
	Version AutoScalerVersion `json:"version,omitempty"`
}

type StarRocksCnStatus struct {
	// the name of cn service for fe find cn instance.
	ServiceName string `json:"serviceName,omitempty"`

	// FailedInstances deploy failed cn pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	// CreatingInstances in creating status cn pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	// RunningInstances in running status be pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	// The statefulset names of be.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// The policy name of autoScale.
	// Deprecated
	HpaName string `json:"hpaName,omitempty"`

	// HorizontalAutoscaler have the autoscaler information.
	HorizontalScaler HorizontalScaler `json:"horizontalScaler,omitempty"`

	// Phase the value from all pods of cn status. If cn have one failed pod phase=failed,
	// also if cn have one creating pod phase=creating, also if cn all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// the reason for the phase.
	// +optional
	Reason string `json:"reason"`
}

// StarRocksCluster defines a starrocks cluster deployment.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=src
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="FeStatus",type=string,JSONPath=`.status.starRocksFeStatus.phase`
// +kubebuilder:printcolumn:name="CnStatus",type=string,JSONPath=`.status.starRocksCnStatus.phase`
// +kubebuilder:printcolumn:name="BeStatus",type=string,JSONPath=`.status.starRocksBeStatus.phase`
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

const (
	// TCPProbeType represents the readiness prob method with TCP
	TCPProbeType string = "tcp"
	// CommandProbeType represents the readiness prob method with arbitrary unix `exec` call format commands
	CommandProbeType string = "command"
)

// StarRocksFeSpec defines the desired state of fe.
type StarRocksFeSpec struct {
	// name of the starrocks be cluster.
	// +optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	// Deprecated, not allow set statefulset name.
	Name string `json:"name,omitempty"`

	// annotation for fe pods. user can config monitor annotation for collect to monitor system.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// serviceAccount for fe access cloud service.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Replicas is the number of desired fe Pod, the number is 1,3,5
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Image for a starrocks fe deployment..
	Image string `json:"image"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	// Service defines the template for the associated Kubernetes Service object.
	// +optional
	Service *StarRocksService `json:"service,omitempty"`

	// defines the specification of resource cpu and mem.
	// +optional
	corev1.ResourceRequirements `json:",inline"`

	// the reference for fe configMap.
	// +optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	// Probe defines the mode probe service in container is alive.
	// +optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	// StorageVolumes defines the additional storage for fe meta storage.
	// +optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	// (Optional) If specified, the pod's nodeSelector，displayName="Map of nodeSelectors to match when scheduling pods on nodes"
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	// feEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	FeEnvVars []corev1.EnvVar `json:"feEnvVars,omitempty"`

	// +optional
	// If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations for scheduling pods onto some dedicated nodes
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// +optional
	// the pod labels for user select or classify pods.
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`
}

// StarRocksBeSpec defines the desired state of be.
type StarRocksBeSpec struct {
	// Replicas is the number of desired be Pod. the default value=3
	// Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Image for a starrocks be deployment.
	Image string `json:"image"`

	// annotation for be pods. user can config monitor annotation for collect to monitor system.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	// serviceAccount for be access cloud service.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// name of the starrocks be cluster.
	// +optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	// Deprecated
	Name string `json:"name,omitempty"`

	// Service defines the template for the associated Kubernetes Service object.
	// the service for user access be.
	// +optional
	Service *StarRocksService `json:"service,omitempty"`

	// defines the specification of resource cpu and mem.
	// +optional
	corev1.ResourceRequirements `json:",inline"`

	// the reference for be configMap.
	// +optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	// Probe defines the mode probe service in container is alive.
	// +optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	// StorageVolumes defines the additional storage for be storage data and log.
	// +optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	// ReplicaInstance is the names of replica starrocksbe cluster.
	// +optional
	// Deprecated, temp deprecated.
	ReplicaInstances []string `json:"replicaInstances,omitempty"`

	// (Optional) If specified, the pod's nodeSelector，displayName="Map of nodeSelectors to match when scheduling pods on nodes"
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	// beEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	BeEnvVars []corev1.EnvVar `json:"beEnvVars,omitempty"`

	// +optional
	// If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations for scheduling pods onto some dedicated nodes
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// podLabels for user selector or classify pods.
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`
}

// StarRocksCnSpec defines the desired state of cn.
type StarRocksCnSpec struct {
	// name of the starrocks cn cluster.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	// +optional
	// Deprecated: , the statefulset name don't allow set, prevent accidental modification.
	Name string `json:"name,omitempty"`

	// annotation for cn pods. user can config monitor annotation for collect to monitor system.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// serviceAccount for cn access cloud service.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Replicas is the number of desired cn Pod.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Image for a starrocks cn deployment.
	Image string `json:"image"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	// Service defines the template for the associated Kubernetes Service object.
	// the service for user access cn.
	Service *StarRocksService `json:"service,omitempty"`

	// +optional
	// set the fe service for register cn, when not set, will use the fe config to find.
	// Deprecated,
	// FeServiceName string `json:"feServiceName,omitempty"`

	// the reference for cn configMap.
	// +optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	// defines the specification of resource cpu and mem.
	corev1.ResourceRequirements `json:",inline"`

	// Probe defines the mode probe service in container is alive.
	Probe *StarRocksProbe `json:"probe,omitempty"`

	// AutoScalingPolicy auto scaling strategy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`

	// (Optional) If specified, the pod's nodeSelector，displayName="Map of nodeSelectors to match when scheduling pods on nodes"
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	// cnEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	CnEnvVars []corev1.EnvVar `json:"cnEnvVars,omitempty"`

	// +optional
	// If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations for scheduling pods onto some dedicated nodes
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// +optional

	// podLabels for user selector or classify pods.
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`
}

type ConfigMapInfo struct {
	// the config info for start progress.
	ConfigMapName string `json:"configMapName,omitempty"`

	// the config response key in configmap.
	ResolveKey string `json:"resolveKey,omitempty"`
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
	StorageSize string `json:"storageSize"`

	// MountPath specify the path of volume mount.
	MountPath string `json:"mountPath,omitempty"`
}

type StarRocksService struct {
	// Name assigned to service.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	// +optional
	Name string `json:"name,omitempty"`

	// type of service,the possible value for the service type are : ClusterIP, NodePort, LoadBalancer,ExternalName.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
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

	// Ports the components exposed ports and listen ports in pod.
	// +optional
	Ports []StarRocksServicePort `json:"ports"`
}

type StarRocksServicePort struct {
	// Name of the map about coming port and target port
	Name string `json:"name,omitempty"`

	// Port the pod is exposed on service.
	Port int32 `json:"port"`

	// ContainerPort the service listen in pod.
	ContainerPort int32 `json:"containerPort"`

	// The easiest way to expose fe, cn or be is to use a Service of type `NodePort`.
	NodePort int32 `json:"nodePort,omitempty"`
}

// StarRocksProbe defines the mode for probe be alive.
type StarRocksProbe struct {
	// Type identifies the mode of probe main container
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
