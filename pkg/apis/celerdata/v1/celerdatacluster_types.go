package v1

import (
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SpecInterface defines the common interface that must be implemented by all CelerData component specs
// (FE, BE, CN, FE Proxy). It provides methods to configure pod and container settings like security context,
// lifecycle hooks, networking, and storage.
// All components including CelerDataFeSpec, CelerDataBeSpec, CelerDataCnSpec, CelerDataFeProxySpec have implemented
// the SpecInterface. If a method has the same implementation, we will implement in CelerDataLoadSpec which implements
// the loadInterface interface.
// +kubebuilder:object:generate=false
type SpecInterface interface {
	loadInterface

	// GetHostAliases returns the host aliases configuration for the pod. Host aliases allow adding entries to
	// the pod's /etc/hosts file to configure custom host-to-IP mappings.
	GetHostAliases() []corev1.HostAlias

	// GetRunAsNonRoot returns the user ID and group ID to run the container as non-root.
	// Returns (*uid, *gid) where nil means use container defaults.
	GetRunAsNonRoot() (*int64, *int64)

	// GetTerminationGracePeriodSeconds returns the grace period in seconds for pod termination.
	// This is how long to wait for the pod to terminate gracefully before forcefully killing it.
	GetTerminationGracePeriodSeconds() *int64

	// GetCapabilities returns the Linux capabilities configuration for the container.
	// Capabilities allow granting specific privileges to processes.
	GetCapabilities() *corev1.Capabilities

	// GetSidecars returns the list of sidecar containers to add to the pod.
	// Sidecars run alongside the main container to provide additional functionality.
	GetSidecars() []corev1.Container

	// GetInitContainers returns the list of init containers to run before the main containers.
	// Init containers run to completion in order, before the main containers start.
	GetInitContainers() []corev1.Container

	// IsReadOnlyRootFilesystem returns whether the container's root filesystem should be read-only.
	// A read-only root filesystem prevents modifications to improve security.
	IsReadOnlyRootFilesystem() *bool

	// GetSysctls returns the sysctl parameters to set for the container.
	// Sysctls allow configuring kernel parameters at runtime.
	GetSysctls() []corev1.Sysctl

	// GetCommand returns the command to run in the container.
	// This overrides the container image's default entrypoint.
	GetCommand() []string

	// GetArgs returns the arguments to pass to the container command.
	// These are the arguments passed to either the default entrypoint or GetCommand().
	GetArgs() []string

	// GetUpdateStrategy returns the update strategy for the StatefulSet.
	// This controls how pods are replaced during updates.
	GetUpdateStrategy() *appsv1.StatefulSetUpdateStrategy

	// GetMinReadySeconds returns the minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	GetMinReadySeconds() *int32

	GetPodManagementPolicy() appsv1.PodManagementPolicyType
}

var _ SpecInterface = &CelerDataFeSpec{}

var _ SpecInterface = &CelerDataBeSpec{}

var _ SpecInterface = &CelerDataCnSpec{}

var _ SpecInterface = &CelerDataFeProxySpec{}

// CelerDataClusterSpec defines the desired state of CelerDataCluster
type CelerDataClusterSpec struct {
	// Specify a Service Account for CelerDataCluster use k8s cluster.
	// +optional
	// Deprecated: component use serviceAccount in own's field.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// CelerDataFeSpec define fe configuration for start fe service.
	CelerDataFeSpec *CelerDataFeSpec `json:"celerDataFeSpec,omitempty"`

	// CelerDataBeSpec define be configuration for start be service.
	CelerDataBeSpec *CelerDataBeSpec `json:"celerDataBeSpec,omitempty"`

	// CelerDataCnSpec define cn configuration for start cn service.
	CelerDataCnSpec *CelerDataCnSpec `json:"celerDataCnSpec,omitempty"`

	// CelerDataLoadSpec define a proxy for fe.
	CelerDataFeProxySpec *CelerDataFeProxySpec `json:"celerDataFeProxySpec,omitempty"`

	// +optional
	// DisasterRecovery is used to determine whether to enter disaster recovery mode.
	DisasterRecovery *DisasterRecovery `json:"disasterRecovery,omitempty"`

	// +optional
	// WaitForFullRollout controls rolling upgrade behavior. When set to true, the operator
	// will wait for FE StatefulSet to be fully rolled out (all replicas ready and at the
	// same revision) before updating BE/CN components. This prevents cascading failures
	// if a bad FE version is deployed.
	// When false (default), BE/CN updates can proceed as soon as any FE pod is ready.
	// Defaults to false for backward compatibility.
	WaitForFullRollout bool `json:"waitForFullRollout,omitempty"`
}

// CelerDataClusterStatus defines the observed state of CelerDataCluster.
type CelerDataClusterStatus struct {
	// Represents the state of cluster. the possible value are: running, failed, pending
	Phase Phase `json:"phase"`

	// Reason represents the errors when calling sub-controllers
	Reason string `json:"reason,omitempty"`

	// Represents the status of fe. the status have running, failed and creating pods.
	CelerDataFeStatus *CelerDataFeStatus `json:"celerDataFeStatus,omitempty"`

	// Represents the status of be. the status have running, failed and creating pods.
	CelerDataBeStatus *CelerDataBeStatus `json:"celerDataBeStatus,omitempty"`

	// Represents the status of cn. the status have running, failed and creating pods.
	CelerDataCnStatus *CelerDataCnStatus `json:"celerDataCnStatus,omitempty"`

	// Represents the status of fe proxy. the status have running, failed and creating pods.
	CelerDataFeProxyStatus *CelerDataFeProxyStatus `json:"celerDataFeProxyStatus,omitempty"`

	// +optional
	// DisasterRecoveryStatus represents the status of disaster recovery.
	DisasterRecoveryStatus *DisasterRecoveryStatus `json:"disasterRecoveryStatus,omitempty"`
}

// CelerDataFeSpec defines the desired state of fe.
type CelerDataFeSpec struct {
	CelerDataComponentSpec `json:",inline"`

	// +optional
	// feEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	FeEnvVars []corev1.EnvVar `json:"feEnvVars,omitempty"`
}

// CelerDataBeSpec defines the desired state of be.
type CelerDataBeSpec struct {
	CelerDataComponentSpec `json:",inline"`

	// +optional
	// beEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	BeEnvVars []corev1.EnvVar `json:"beEnvVars,omitempty"`
}

// CelerDataCnSpec defines the desired state of cn.
type CelerDataCnSpec struct {
	CelerDataComponentSpec `json:",inline"`

	// +optional
	// cnEnvVars is a slice of environment variables that are added to the pods, the default is empty.
	CnEnvVars []corev1.EnvVar `json:"cnEnvVars,omitempty"`

	// AutoScalingPolicy auto scaling strategy
	AutoScalingPolicy *AutoScalingPolicy `json:"autoScalingPolicy,omitempty"`
}

// CelerDataFeProxySpec defines the specification for FE Proxy
// Note: it includes CelerDataLoadSpec, not CelerDataComponentSpec
type CelerDataFeProxySpec struct {
	CelerDataLoadSpec `json:",inline"`

	Resolver string `json:"resolver,omitempty"`
}

// CelerDataFeStatus represents the status of CelerData fe.
type CelerDataFeStatus struct {
	CelerDataComponentStatus `json:",inline"`
}

// CelerDataBeStatus represents the status of CelerData be.
type CelerDataBeStatus struct {
	CelerDataComponentStatus `json:",inline"`
}

type CelerDataFeProxyStatus struct {
	CelerDataComponentStatus `json:",inline"`
}

// CelerDataCnStatus represents the status of CelerData cn.
type CelerDataCnStatus struct {
	CelerDataComponentStatus `json:",inline"`

	// The policy name of autoScale.
	// Deprecated
	HpaName string `json:"hpaName,omitempty"`

	// HorizontalAutoscaler have the autoscaler information.
	HorizontalScaler HorizontalScaler `json:"horizontalScaler,omitempty"`

	// Replicas is the total number of non-terminated CN pods targeted.
	Replicas int32 `json:"replicas,omitempty"`

	// Selector for CN pods. The HPA will use this selector to know which pods to monitor.
	Selector string `json:"selector,omitempty"`
}

func (spec *CelerDataFeSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.CelerDataComponentSpec.GetReplicas()
}

func (spec *CelerDataBeSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.CelerDataComponentSpec.GetReplicas()
}

func (spec *CelerDataCnSpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.CelerDataComponentSpec.GetReplicas()
}

func (spec *CelerDataFeProxySpec) GetReplicas() *int32 {
	if spec == nil {
		return nil
	}
	return spec.CelerDataLoadSpec.GetReplicas()
}

// GetHostAliases
// fe proxy does not have field HostAliases, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetHostAliases() []corev1.HostAlias {
	// fe proxy do not support host alias
	return nil
}

// GetRunAsNonRoot
// fe proxy does not have field RunAsNonRoot, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetRunAsNonRoot() (*int64, *int64) {
	// fe proxy will set run as nginx user by default, and can not be changed by crd
	return nil, nil
}

// GetTerminationGracePeriodSeconds
// fe proxy does not have field TerminationGracePeriodSeconds, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetTerminationGracePeriodSeconds() *int64 {
	return nil
}

// GetCapabilities
// fe proxy does not have field Capabilities, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetCapabilities() *corev1.Capabilities { return nil }

// GetSidecars
// fe proxy does not have field Sidecars, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetSidecars() []corev1.Container {
	return nil
}

// GetInitContainers
// fe proxy does not have field InitContainers, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetInitContainers() []corev1.Container {
	return nil
}

// GetCommand
// fe proxy does not have field command, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetCommand() []string {
	return nil
}

// GetArgs
// fe proxy does not have field args
func (spec *CelerDataFeProxySpec) GetArgs() []string {
	return nil
}

// GetUpdateStrategy
// fe proxy is deployed by deployment, and it does not have field UpdateStrategy
func (spec *CelerDataFeProxySpec) GetUpdateStrategy() *appsv1.StatefulSetUpdateStrategy {
	return nil
}

// IsReadOnlyRootFilesystem
// fe proxy does not have field ReadOnlyRootFilesystem, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) IsReadOnlyRootFilesystem() *bool {
	return nil
}

// GetSysctls
// fe proxy does not have field Sysctls, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetSysctls() []corev1.Sysctl { return nil }

// GetMinReadySeconds
// fe proxy does not have field MinReadySeconds, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetMinReadySeconds() *int32 {
	return nil
}

// GetPodManagementPolicy
// fe proxy does not have field PodManagementPolicy, the reason why implementing this method is
// that CelerDataFeProxySpec needs to implement SpecInterface interface
func (spec *CelerDataFeProxySpec) GetPodManagementPolicy() appsv1.PodManagementPolicyType {
	return ""
}

// Phase is defined under status, e.g.
// 1. CelerDataClusterStatus.Phase represents the phase of CelerData cluster.
// 2. CelerDataWarehouseStatus.Phase represents the phase of CelerData warehouse.
// The possible value for cluster phase are: running, failed, pending, deleting.
type Phase string

// ComponentPhase represent the component phase. e.g.
// 1. CelerDataCluster contains three components: FE, CN, BE.
// 2. CelerDataWarehouse reuse the CN component.
// The possible value for component phase are: reconciling, failed, running.
type ComponentPhase string

const (
	// ClusterRunning represents CelerData cluster is running.
	ClusterRunning Phase = "running"

	// ClusterFailed represents CelerData cluster failed.
	ClusterFailed Phase = "failed"

	// ClusterReconciling represents some component is reconciling
	ClusterReconciling Phase = "reconciling"
)

const (
	// ComponentReconciling the CelerData component is reconciling
	ComponentReconciling ComponentPhase = "reconciling"

	// ComponentFailed the pod of component is failed
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

// CelerDataCluster defines a CelerData cluster deployment.
// +kubebuilder:object:root=true
// +kubebuilder:metadata:annotations="version=v1.11.4"
// +kubebuilder:resource:shortName=cdc
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.celerDataCnSpec.replicas,statuspath=.status.celerDataCnStatus.replicas,selectorpath=.status.celerDataCnStatus.selector
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="FeStatus",type=string,JSONPath=`.status.celerDataFeStatus.phase`
// +kubebuilder:printcolumn:name="BeStatus",type=string,JSONPath=`.status.celerDataBeStatus.phase`
// +kubebuilder:printcolumn:name="CnStatus",type=string,JSONPath=`.status.celerDataCnStatus.phase`
// +kubebuilder:printcolumn:name="FeProxyStatus",type=string,JSONPath=`.status.celerDataFeProxyStatus.phase`
// +kubebuilder:storageversion
// +k8s:openapi-gen=true
// +genclient
//
//nolint:lll
type CelerDataCluster struct {
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired state of the CelerData cluster.
	Spec CelerDataClusterSpec `json:"spec,omitempty"`

	// Most recent observed status of the CelerData cluster
	Status CelerDataClusterStatus `json:"status,omitempty"`
}

const (
	EmptyDir = "emptyDir"
	HostPath = "hostPath"
)

// StorageVolume defines additional PVC template for StatefulSets and volumeMount for pods that mount this PVC.
type StorageVolume struct {
	// name of a storage volume.
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name"`

	// storageClassName is the name of the StorageClass required by the claim.
	// If storageClassName is not set, the default StorageClass of kubernetes will be used.
	// there are some special storageClassName: emptyDir, hostPath. In this case, It will use emptyDir or hostPath, not PVC.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// StorageSize is a valid memory size type based on powers-of-2, so 1Mi is 1024Ki.
	// Supported units:Mi, Gi, GiB, Ti, Ti, Pi, Ei, Ex: `512Mi`.
	// It will take effect only when storageClassName is real storage class, not emptyDir or hostPath.
	// +kubebuilder:validation:Pattern:="(^0|([0-9]*l[.])?[0-9]+((M|G|T|E|P)i))$"
	// +optional
	StorageSize string `json:"storageSize,omitempty"`

	// HostPath Represents a host path mapped into a pod.
	// If StorageClassName is hostPath, HostPath is required.
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// MountPath specify the path of volume mount.
	MountPath string `json:"mountPath"`

	// SubPath within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	SubPath string `json:"subPath,omitempty"`
}

var ErrHostPathRequired = errors.New("if storageClassName is hostPath, hostPath and hostPath.path is required")

func (storageVolume *StorageVolume) Validate() error {
	if storageVolume.StorageClassName != nil {
		if *storageVolume.StorageClassName == HostPath {
			if storageVolume.HostPath == nil {
				return ErrHostPathRequired
			}
			if storageVolume.HostPath.Path == "" {
				return ErrHostPathRequired
			}
		}
	} else if storageVolume.HostPath != nil {
		if storageVolume.HostPath.Path == "" {
			return ErrHostPathRequired
		}
	}
	return nil
}

// CelerDataClusterList contains a list of CelerDataCluster
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type CelerDataClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CelerDataCluster `json:"items"`
}
