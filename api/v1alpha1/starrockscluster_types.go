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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specify a Service Account for starRocksCluster
	//+optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	//StarRocksFeSpec is the fe specification.
	StarRocksFeSpec *StarRocksFeSpec `json:"starRocksFeSpec,omitempty"`

	//StarRocksBeSpec is the be specification.
	StarRocksBeSpec *StarRocksBeSpec `json:"starRocksBeSpec,omitempty"`

	//StarRocksCnSpec is the cn specification.
	StarRocksCnSpec *StarRocksCnSpec `json:"starRocksCnSpec,omitempty"`
}

// StarRocksClusterStatus defines the observed state of StarRocksCluster
type StarRocksClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase ClusterPhase `json:"phase"`
	//used pointer for the component
	StarRocksFeStatus *StarRocksFeStatus `json:"starRocksFeStatus,omitempty"`

	StarRocksBeStatus *StarRocksBeStatus `json:"starRocksBeStatus,omitempty"`

	StarRocksCnStatus *StarRocksCnStatus `json:"starRocksCnStatus,omitempty"`
}

type ClusterPhase string
type MemberPhase string

const (
	//ClusterRunning starrocks cluster is running
	ClusterRunning = "running"

	//ClusterFailed starrocks cluster have failed for some reason, the status explain what happened.
	ClusterFailed = "failed"

	//ClusterPending starrocks cluster is creating
	ClusterPending = "pending"

	//ClusterWaiting waiting cluster running
	//ClusterWaiting = "waiting"
)

const (
	//ComponentReconciling the component of starrocks cluster is dynamic adjustment.
	ComponentReconciling = "reconciling"

	ComponentFailed = "failed"
	//
	ComponentRunning = "running"

	ComponentWaiting = "waiting"
)

//StarRocksFeStatus the status of starrocksfe.
type StarRocksFeStatus struct {
	ServiceName       string      `json:"serviceName,omitempty"`
	FailedInstances   []string    `json:"failedInstances,omitempty"`
	CreatingInstances []string    `json:"creatingInstances,omitempty"`
	RunningInstances  []string    `json:"runningInstances,omitempty"`
	ResourceNames     []string    `json:"resourceNames,omitempty"`
	Phase             MemberPhase `json:"phase"`
	Reason            string      `json:"reason"`
}

type StarRocksBeStatus struct {
	ServiceName       string      `json:"serviceName,omitempty"`
	FailedInstances   []string    `json:"failedInstances,omitempty"`
	CreatingInstances []string    `json:"creatingInstances,omitempty"`
	RunningInstances  []string    `json:"runningInstances,omitempty"`
	ResourceNames     []string    `json:"resourceNames,omitempty"`
	Phase             MemberPhase `json:"phase"`
	Reason            string      `json:"reason"`
}

type StarRocksCnStatus struct {
	ServiceName         string      `json:"serviceName,omitempty"`
	FailedInstanceNames []string    `json:"failedInstanceNames,omitempty"`
	CreatingInstances   []string    `json:"creatingInstances,omitempty"`
	RunningInstances    []string    `json:"runningInstances,omitempty"`
	ResourceNames       []string    `json:"resourceNames,omitempty"`
	Phase               MemberPhase `json:"phase"`
	Reason              string      `json:"reason"`
}

// StarRocksCluster is the Schema for the starrocksclusters API
//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=src
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
// +k8s:openapi-gen=true
type StarRocksCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StarRocksClusterSpec   `json:"spec,omitempty"`
	Status StarRocksClusterStatus `json:"status,omitempty"`
}

const (
	// TCPProbeType represents the readiness prob method with TCP
	TCPProbeType string = "tcp"
	// CommandProbeType represents the readiness prob method with arbitrary unix `exec` call format commands
	CommandProbeType string = "command"
)

//StarRocksFeSpec defines the desired state of fe.
// +k8s:openapi-gen=true
type StarRocksFeSpec struct {
	//name of the starrocks be cluster.
	//+optional
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name,omitempty"`

	//Replicas is the number of desired fe Pod, the number is 1,3,5
	//+optional: Defaults to 3
	Replicas *int32 `json:"replicas,omitempty"`

	//Image is the container image for the fe.
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
// +k8s:openapi-gen=true
type StarRocksBeSpec struct {
	//Replicas is the number of desired be Pod
	// Optional: Defaults to 3
	Replicas int32 `json:"replicas,omitempty"`

	//Image is the container image for the be.
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

	//Probe defines the mode probe service in container is alive.
	//+optional
	Probe *StarRocksProbe `json:"probe,omitempty"`

	//StorageVolumes defines the additional storage for fe.
	//+optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	//ReplicaInstance is the names of replica starrocksbe cluster.
	//+optional
	ReplicaInstances []string `json:"ReplicaInstances,omitempty"`
}

//StarRocksCnSpec defines the desired state of cn.
// +k8s:openapi-gen=true
type StarRocksCnSpec struct {
	//name of the starrocks be cluster.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//+optional
	Name string `json:"name,omitempty"`

	//Replicas is the number of desired cn Pod.
	// +kubebuilder:validation:Minimum=0
	//+optional
	Replicas *int32 `json:"replicas,omitempty"`

	//Image is the container image for the cn.
	Image string `json:"image"`

	//Service defines the template for the associated Kubernetes Service object.
	//the service for user access cn.
	Service *StarRocksService `json:"service,omitempty"`

	//+optional
	//set the fe service for register cn, when not set, will use the fe config to find.
	FeServiceName string `json:"feServiceName,omitempty"`

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

	// StorageSize is a valid memory size type based on powers-of-2, so 1MB is 1024KB.
	// Supported units:MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB Ex: `512MB`.
	// +kubebuilder:validation:Pattern:="(^0|([0-9]*[.])?[0-9]+((M|G|T|E|P)i?)?B)$"
	StorageSize string `json:"storageSize"`

	//MountPath specify the path of volume mount.
	MountPath string `json:"mountPath,omitempty"`
}

type StarRocksService struct {
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name,omitempty"`

	Ports []StarRocksServicePort `json:"ports"`
}

type StarRocksServicePort struct {
	Name          string `json:"name,omitempty"`
	Port          int32  `json:"port"`
	ContainerPort int32  `json:"containerPort"`
	NodePort      int32  `json:"nodePort,omitempty"`
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

//+kubebuilder:object:root=true
// StarRocksClusterList contains a list of StarRocksCluster
// +k8s:openapi-gen=true
type StarRocksClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StarRocksCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StarRocksCluster{}, &StarRocksClusterList{})
}
