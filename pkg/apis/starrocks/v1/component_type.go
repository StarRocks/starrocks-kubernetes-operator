package v1

import corev1 "k8s.io/api/core/v1"

type StarRocksComponentSpec struct {
	StarRocksLoadSpec `json:",inline"`

	// RunAsNonRoot is used to determine whether to run starrocks as a normal user.
	// If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
	// default: nil
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`

	// the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.
	// +optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	// the reference for configMap which allow users to mount any files to container.
	// +optional
	ConfigMaps []ConfigMapReference `json:"configMaps,omitempty"`

	// the reference for secrets.
	// +optional
	Secrets []SecretReference `json:"secrets,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`
}

// StarRocksComponentStatus represents the status of a starrocks component.
type StarRocksComponentStatus struct {
	// the name of fe service exposed for user.
	ServiceName string `json:"serviceName,omitempty"`

	// FailedInstances failed pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	// CreatingInstances in creating pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	// RunningInstances in running status pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	// ResourceNames the statefulset names of fe in v1alpha1 version.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of component status. If component have one failed pod phase=failed,
	// also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.
	Phase MemberPhase `json:"phase"`

	// +optional
	// Reason represents the reason of not running.
	Reason string `json:"reason"`
}

type ConfigMapInfo struct {
	// the config info for start progress.
	ConfigMapName string `json:"configMapName,omitempty"`

	// the config response key in configmap.
	ResolveKey string `json:"resolveKey,omitempty"`
}

type ConfigMapReference MountInfo

type SecretReference MountInfo

type MountInfo struct {
	// This must match the Name of a ConfigMap or Secret in the same namespace, and
	// the length of name must not more than 50 characters.
	Name string `json:"name,omitempty"`

	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath,omitempty"`
}

func (spec *StarRocksComponentSpec) GetHostAliases() []corev1.HostAlias {
	return spec.HostAliases
}

func (spec *StarRocksComponentSpec) GetRunAsNonRoot() (*int64, *int64) {
	runAsNonRoot := spec.RunAsNonRoot
	if runAsNonRoot == nil || *runAsNonRoot == false {
		return nil, nil
	}

	userId := int64(1000)
	groupId := int64(1000)
	return &userId, &groupId
}
