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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WarehouseComponentSpec
// +kubebuilder:object:generate=false
type WarehouseComponentSpec = StarRocksCnSpec

// StarRocksWarehouseSpec defines the desired state of StarRocksWarehouse
type StarRocksWarehouseSpec struct {
	// StarRocksCluster is the name of a StarRocksCluster which the warehouse belongs to.
	StarRocksCluster string `json:"starRocksCluster"`

	// Template define component configuration.
	Template *StarRocksCnSpec `json:"template"`
}

// WarehouseComponentStatus
// +kubebuilder:object:generate=false
type WarehouseComponentStatus = StarRocksCnStatus

// StarRocksWarehouseStatus defines the observed state of StarRocksWarehouse.
type StarRocksWarehouseStatus struct {
	// Phase represents the state of a warehouse. The possible value are: running, failed, pending and deleting.
	Phase Phase `json:"phase"`

	// WarehouseStatus represents the status of cn service. The status has reconciling, failed and running.
	*WarehouseComponentStatus `json:",inline"`
}

// StarRocksWarehouse defines a starrocks warehouse.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=warehouse
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CnStatus",type=string,JSONPath=`.status.starRocksCnStatus.phase`
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +genclient
type StarRocksWarehouse struct {
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec represents the specification of desired state of a starrocks warehouse.
	Spec StarRocksWarehouseSpec `json:"spec,omitempty"`

	// Status represents the recent observed status of the starrocks warehouse.
	Status StarRocksWarehouseStatus `json:"status,omitempty"`
}

// StarRocksWarehouseList contains a list of StarRocksWarehouse
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type StarRocksWarehouseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StarRocksWarehouse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StarRocksWarehouse{}, &StarRocksWarehouseList{})
}
