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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputeNodeGroupSpec defines the desired state of ComputeNodeGroup
type ComputeNodeGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	FeInfo   FeInfo `json:"feInfo"`
	CnInfo   CnInfo `json:"cnInfo"`
	Images   Images `json:"images"`
	Replicas int32  `json:"replicas"`
	// TODO: gpa
	// hpa
	// +optional
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
	// pod template
	PodPolicy     PodPolicy     `json:"podPolicy,omitempty"`
	CronJobPolicy CronJobPolicy `json:"cronJobPolicy,omitempty"`
}

// cn node info
type CnInfo struct {
	ConfigMap string `json:"configMap,omitempty"`
}

// fe node info
type FeInfo struct {
	Addresses     []string `json:"addresses"`
	AccountSecret string   `json:"accountSecret"`
}

// offline cronjob running policy
type CronJobPolicy struct {
	Schedule string `json:"schedule,omitempty"`
}

// images
type Images struct {
	CnImage         string `json:"cnImage"`
	ComponentsImage string `json:"componentsImage"`
}

// ServersStatus is the status of the servers of the cluster with both
// ready and not-ready servers
type ServersStatus struct {
	Available    int32 `json:"available"`
	Unavailable  int32 `json:"unavailable"`
	Unregistered int32 `json:"unregistered"`
	Useless      int32 `json:"useless"`
}

// ComputeNodeGroupStatus defines the observed state of ComputeNodeGroup
type ComputeNodeGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions map[CnComponent]ResourceCondition `json:"conditions"`

	// ObservedGeneration is the most recent generation observed for this cluster.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Servers is the list of servers in the cn cluster
	Servers ServersStatus `json:"servers"`
	// UpdatedReplicas is the number of cn servers that has been updated to the latest configuration
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	// Label selector for scaling
	// +optional
	LabelSelector string `json:"labelSelector,omitempty"`

	Replicas int32 `json:"replicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.labelSelector
// ComputeNodeGroup is the Schema for the computenodegroups API
type ComputeNodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeNodeGroupSpec   `json:"spec,omitempty"`
	Status ComputeNodeGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComputeNodeGroupList contains a list of ComputeNodeGroup
type ComputeNodeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeNodeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeNodeGroup{}, &ComputeNodeGroupList{})
}
