package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CelerDataWarehouseSpec defines the desired state of CelerDataWarehouse
type CelerDataWarehouseSpec struct {
	// CelerDataCluster is the name of a CelerDataCluster which the warehouse belongs to.
	CelerDataCluster string `json:"celerDataCluster"`

	// Template define component configuration.
	Template *WarehouseComponentSpec `json:"template"`
}

// WarehouseComponentSpec defines the desired state of component.
type WarehouseComponentSpec struct {
	CelerDataComponentSpec `json:",inline"`

	// +optional
	// envVars is a slice of environment variables that are added to the pods, the default is empty.
	EnvVars []corev1.EnvVar `json:"envVars,omitempty"`

	// AutoScalingPolicy defines auto scaling policy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
}

func (componentSpec *WarehouseComponentSpec) ToCnSpec() *CelerDataCnSpec {
	return &CelerDataCnSpec{
		CelerDataComponentSpec: componentSpec.CelerDataComponentSpec,
		CnEnvVars:              componentSpec.EnvVars,
		AutoScalingPolicy:      componentSpec.AutoScalingPolicy,
	}
}

// WarehouseComponentStatus represents the status of component.
// +kubebuilder:object:generate=false
type WarehouseComponentStatus = CelerDataCnStatus

// CelerDataWarehouseStatus defines the observed state of CelerDataWarehouse.
type CelerDataWarehouseStatus struct {
	*WarehouseComponentStatus `json:",inline"`
}

// CelerDataWarehouse defines a CelerData warehouse.
// +kubebuilder:object:root=true
// +kubebuilder:metadata:annotations="version=v1.11.5"
// +kubebuilder:resource:shortName=cdw
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.template.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="reason",type=string,JSONPath=`.status.reason`
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +genclient
type CelerDataWarehouse struct {
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec represents the specification of desired state of a CelerData warehouse.
	Spec CelerDataWarehouseSpec `json:"spec,omitempty"`

	// Status represents the recent observed status of the CelerData warehouse.
	Status CelerDataWarehouseStatus `json:"status,omitempty"`
}

// CelerDataWarehouseList contains a list of CelerDataWarehouse
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type CelerDataWarehouseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CelerDataWarehouse `json:"items"`
}
