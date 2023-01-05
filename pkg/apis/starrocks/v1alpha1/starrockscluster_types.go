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
	//+optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	//StarRocksFeSpec define fe configuration for start fe service.
	StarRocksFeSpec *StarRocksFeSpec `json:"starRocksFeSpec,omitempty"`

	//StarRocksBeSpec define be configuration for start be service.
	StarRocksBeSpec *StarRocksBeSpec `json:"starRocksBeSpec,omitempty"`

	//StarRocksCnSpec define cn configuration for start cn service.
	StarRocksCnSpec *StarRocksCnSpec `json:"starRocksCnSpec,omitempty"`
}

// StarRocksClusterStatus defines the observed state of StarRocksCluster.
type StarRocksClusterStatus struct {
	//Represents the state of cluster. the possible value are: running, failed, pending
	Phase ClusterPhase `json:"phase"`
	//Represents the status of fe. the status have running, failed and creating pods.
	StarRocksFeStatus *StarRocksFeStatus `json:"starRocksFeStatus,omitempty"`

	//Represents the status of be. the status have running, failed and creating pods.
	StarRocksBeStatus *StarRocksBeStatus `json:"starRocksBeStatus,omitempty"`

	//Represents the status of cn. the status have running, failed and creating pods.
	StarRocksCnStatus *StarRocksCnStatus `json:"starRocksCnStatus,omitempty"`
}

type ClusterPhase string
type MemberPhase string

const (
	//ClusterRunning represents starrocks cluster is running.
	ClusterRunning = "running"

	//ClusterFailed represents starrocks cluster failed.
	ClusterFailed = "failed"

	//ClusterPending represents the starrocks cluster is creating
	ClusterPending = "pending"

	//ClusterWaiting waiting cluster running
	//ClusterWaiting = "waiting"
)

const (
	//ComponentReconciling the starrocks have component in starting.
	ComponentReconciling = "reconciling"
	//ComponentFailed have at least one service failed.
	ComponentFailed = "failed"
	//ComponentRunning all components runs available.
	ComponentRunning = "running"
	//ComponentWaiting service wait for reconciling.
	ComponentWaiting = "waiting"
)

//StarRocksFeStatus represents the status of starrocks fe.
type StarRocksFeStatus struct {
	//the name of fe service exposed for user.
	ServiceName string `json:"serviceName,omitempty"`

	//FailedInstances failed fe pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	//CreatingInstances in creating pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	//RunningInstances in running status pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	//ResourceNames the statefulset names of fe in v1alpha1 version.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of fe status. If fe have one failed pod phase=failed,
	// also if fe have one creating pod phase=creating, also if fe all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	//+optional
	//Reason represents the reason of not running.
	Reason string `json:"reason"`
}

// StarRocksBeStatus represents the status of starrocks be.
type StarRocksBeStatus struct {
	//the name of be service for fe find be instance.
	ServiceName string `json:"serviceName,omitempty"`

	//FailedInstances deploy failed instance of be.
	FailedInstances []string `json:"failedInstances,omitempty"`

	//CreatingInstances represents status in creating pods of be.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	//RunningInstances represents status in running pods of be.
	RunningInstances []string `json:"runningInstances,omitempty"`

	//The statefulset names of be.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of be status. If be have one failed pod phase=failed,
	// also if be have one creating pod phase=creating, also if be all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// the reason for the phase.
	//+optional
	Reason string `json:"reason"`
}

type StarRocksCnStatus struct {
	//the name of cn service for fe find cn instance.
	ServiceName string `json:"serviceName,omitempty"`

	//FailedInstances deploy failed cn pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	//CreatingInstances in creating status cn pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	//RunningInstances in running status be pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	//The statefulset names of be.
	ResourceNames []string `json:"resourceNames,omitempty"`

	//The policy name of autoScale.
	HpaName string `json:"HpaName,omitempty"`

	// Phase the value from all pods of cn status. If cn have one failed pod phase=failed,
	// also if cn have one creating pod phase=creating, also if cn all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// the reason for the phase.
	//+optional
	Reason string `json:"reason"`
}

// StarRocksCluster defines a starrocks cluster deployment.
//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=src
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="FeStatus",type=string,JSONPath=`.status.starRocksFeStatus.phase`
//+kubebuilder:printcolumn:name="CnStatus",type=string,JSONPath=`.status.starRocksCnStatus.phase`
//+kubebuilder:printcolumn:name="BeStatus",type=string,JSONPath=`.status.starRocksBeStatus.phase`
//+kubebuilder:storageversion
// +k8s:openapi-gen=true
// +genclient
type StarRocksCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//Specification of the desired state of the starrocks cluster.
	Spec StarRocksClusterSpec `json:"spec,omitempty"`

	//Most recent observed status of the starrocks cluster
	Status StarRocksClusterStatus `json:"status,omitempty"`
}

const (
	// TCPProbeType represents the readiness prob method with TCP
	TCPProbeType string = "tcp"
	// CommandProbeType represents the readiness prob method with arbitrary unix `exec` call format commands
	CommandProbeType string = "command"
)

//StarRocksFeSpec defines the desired state of fe.
type StarRocksFeSpec struct {
	//name of the starrocks be cluster.
	//+optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name,omitempty"`

	//Replicas is the number of desired fe Pod, the number is 1,3,5
	//+optional: Defaults to 3
	Replicas *int32 `json:"replicas,omitempty"`

	//Image for a starrocks fe deployment..
	Image string `json:"image"`

	//Service defines the template for the associated Kubernetes Service object.
	//+optional
	Service *StarRocksService `json:"service,omitempty"`

	//defines the specification of resource cpu and mem.
	//+optional
	corev1.ResourceRequirements `json:",inline"`

	//the reference for fe configMap.
	//+optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	//Probe defines the mode probe service in container is alive.
	//+optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	//StorageVolumes defines the additional storage for fe
	//+optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`
}

//StarRocksBeSpec defines the desired state of be.
type StarRocksBeSpec struct {
	//Replicas is the number of desired be Pod. the default value=3
	// Optional
	Replicas *int32 `json:"replicas,omitempty"`

	//Image for a starrocks be deployment.
	Image string `json:"image"`

	//name of the starrocks be cluster.
	//+optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name,omitempty"`

	//Service defines the template for the associated Kubernetes Service object.
	//the service for user access be.
	//+optional
	Service *StarRocksService `json:"service,omitempty"`

	//defines the specification of resource cpu and mem.
	//+optional
	corev1.ResourceRequirements `json:",inline"`

	//the reference for be configMap.
	//+optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	//Probe defines the mode probe service in container is alive.
	//+optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	//StorageVolumes defines the additional storage for fe.
	//+optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	//ReplicaInstance is the names of replica starrocksbe cluster.
	//+optional
	//+deprecated, temp deprecated.
	ReplicaInstances []string `json:"ReplicaInstances,omitempty"`
}

//StarRocksCnSpec defines the desired state of cn.
type StarRocksCnSpec struct {
	//name of the starrocks be cluster.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//+optional
	Name string `json:"name,omitempty"`

	//Replicas is the number of desired cn Pod.
	// +kubebuilder:validation:Minimum=0
	//+optional
	Replicas *int32 `json:"replicas,omitempty"`

	//Image for a starrocks cn deployment.
	Image string `json:"image"`

	//Service defines the template for the associated Kubernetes Service object.
	//the service for user access cn.
	Service *StarRocksService `json:"service,omitempty"`

	//+optional
	//+deprecated,
	//set the fe service for register cn, when not set, will use the fe config to find.
	//FeServiceName string `json:"feServiceName,omitempty"`

	//the reference for cn configMap.
	//+optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	//defines the specification of resource cpu and mem.
	corev1.ResourceRequirements `json:",inline"`

	//Probe defines the mode probe service in container is alive.
	Probe *StarRocksProbe `json:"probe,omitempty"`

	//AutoScalingPolicy auto scaling strategy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
}

type ConfigMapInfo struct {
	//the config info for start progress.
	ConfigMapName string `json:"configMapName,omitempty"`

	//the config response key in configmap.
	ResolveKey string `json:"resolveKey,omitempty"`
}

//StorageVolume defines additional PVC template for StatefulSets and volumeMount for pods that mount this PVC
type StorageVolume struct {
	//name of a storage volume.
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

	//MountPath specify the path of volume mount.
	MountPath string `json:"mountPath,omitempty"`
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

	//Ports the components exposed ports and listen ports in pod.
	// +optional
	Ports []StarRocksServicePort `json:"ports"`
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

//StarRocksProbe defines the mode for probe be alive.
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

// StarRocksClusterList contains a list of StarRocksCluster
//+kubebuilder:object:root=true
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
