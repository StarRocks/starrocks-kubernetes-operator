package v1

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type loadInterface interface {
	GetAnnotations() map[string]string

	GetReplicas() *int32
	GetImagePullSecrets() []corev1.LocalObjectReference
	GetSchedulerName() string
	GetNodeSelector() map[string]string
	GetAffinity() *corev1.Affinity
	GetTopologySpreadConstraints() []corev1.TopologySpreadConstraint
	GetTolerations() []corev1.Toleration
	GetStartupProbeFailureSeconds() *int32
	GetLivenessProbeFailureSeconds() *int32
	GetReadinessProbeFailureSeconds() *int32
	GetLifecycle() *corev1.Lifecycle
	GetService() *StarRocksService
	GetUpdateStrategy() *appv1.StatefulSetUpdateStrategy
	GetShareProcessNamespace() *bool

	GetStorageVolumes() []StorageVolume
	GetServiceAccount() string
	GetImagePullPolicy() corev1.PullPolicy
	GetCommand() []string
	GetArgs() []string
}

type StarRocksLoadSpec struct {
	// defines the specification of resource cpu and mem.
	// +optional
	corev1.ResourceRequirements `json:",inline"`

	// annotation for pods. user can config monitor annotation for collect to monitor system.
	Annotations map[string]string `json:"annotations,omitempty"`

	// +optional
	// the pod labels for user select or classify pods.
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// Replicas is the number of desired Pod.
	// When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
	// applied, the replica count of the StatefulSet object in K8S will be reset to the value
	// specified by the 'Replicas' field, erasing the value previously set by HPA.
	// So operator will set it to nil when HPA policy is enabled.
	//
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Image for a starrocks deployment.
	// +optional
	Image string `json:"image"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to IfNotPresent for compatibility.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the
	// images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for
	// them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"` //nolint:lll

	// SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// (Optional) If specified, the pod's nodeSelector，displayName="Map of nodeSelectors to match when scheduling pods on nodes"
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	// If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations for scheduling pods onto some dedicated nodes
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// (Optional) TopologySpreadConstraints for spreading pods across failure-domains
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// Service defines the template for the associated Kubernetes Service object.
	// +optional
	Service *StarRocksService `json:"service,omitempty"`

	// StorageVolumes defines the additional storage for FE meta or BE/CN storage.
	//
	// For FE component
	//	If the storage volume name is fe-meta or the volume mount path is /opt/starrocks/fe/meta,
	//	then it will be recognized as storing the FE meta.
	//	If the storage volume name is fe-log or the volume mount path is /opt/starrocks/fe/log,
	//	then it will be recognized as storing the FE log.
	//
	// For BE component
	// 	If the storage volume name is be-storage(or be-data) or the volume mount path is /opt/starrocks/be/storage,
	// 	then it will be recognized as storing the BE data.
	// 	If the storage volume name is be-log or the volume mount path is /opt/starrocks/be/log,
	// 	then it will be recognized as storing the BE log.
	//
	// For CN component
	// 	If the storage volume name is cn-log or the volume mount path is /opt/starrocks/cn/log,
	// 	then it will be recognized as storing the cn log.
	//
	// If operator can't find the above storage volume names or the volume mount paths, it will create
	// default storage volumes by using emptyDir.
	//
	// +optional
	StorageVolumes []StorageVolume `json:"storageVolumes,omitempty"`

	// serviceAccount for access cloud service.
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.
	// +optional
	ConfigMapInfo ConfigMapInfo `json:"configMapInfo,omitempty"`

	// StartupProbeFailureSeconds defines the total failure seconds of startup Probe.
	// Default failureThreshold is 60 and the periodSeconds is 5, this means the startup
	// will fail if the pod can't start in 300 seconds. Your StartupProbeFailureSeconds is
	// the total time of seconds before startupProbe give up and fail the container start.
	// If startupProbeFailureSeconds can't be divided by defaultPeriodSeconds, the failureThreshold
	// will be rounded up.
	// Note: you can set it to 0 to disable the startup probe.
	// +optional
	StartupProbeFailureSeconds *int32 `json:"startupProbeFailureSeconds,omitempty"`

	// LivenessProbeFailureSeconds defines the total failure seconds of liveness Probe.
	// Default failureThreshold is 3 and the periodSeconds is 5, this means the liveness
	// will fail if the pod can't respond in 15 seconds. Your LivenessProbeFailureSeconds is
	// the total time of seconds before the container restart. If LivenessProbeFailureSeconds
	// can't be divided by defaultPeriodSeconds, the failureThreshold will be rounded up.
	// Note: you can set it to 0 to disable the liveness probe.
	// +optional
	LivenessProbeFailureSeconds *int32 `json:"livenessProbeFailureSeconds,omitempty"`

	// ReadinessProbeFailureSeconds defines the total failure seconds of readiness Probe.
	// Default failureThreshold is 3 and the periodSeconds is 5, this means the readiness
	// will fail if the pod can't respond in 15 seconds. Your ReadinessProbeFailureSeconds is
	// the total time of seconds before pods becomes not ready. If ReadinessProbeFailureSeconds
	// can't be divided by defaultPeriodSeconds, the failureThreshold will be rounded up.
	// Note: you can set it to 0 to disable the readiness probe.
	// +optional
	ReadinessProbeFailureSeconds *int32 `json:"readinessProbeFailureSeconds,omitempty"`

	// Lifecycle describes actions that the management system should take in response to container lifecycle events.
	// By default, Operator will add corresponding preStop hooks for different components. For example, the preStop
	// script for the FE Component is /opt/starrocks/fe_prestop.sh, for the BE Component is /opt/starrocks/be_prestop.sh,
	// and for the CN Component is /opt/starrocks/cn_prestop.sh.
	// +optional
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`

	// Share a single process namespace between all of the containers in a pod.
	// When this is set containers will be able to view and signal processes from other containers
	// in the same pod, and the first process in each container will not be assigned PID 1.
	// Optional: Default to false.
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty"`
}

// StarRocksService defines external service for starrocks component.
type StarRocksService struct {
	// Annotations store Kubernetes Service annotations.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

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

	// Ports are the ports that are exposed by this service.
	// You can override the default port information by specifying the same StarRocksServicePort.Name in the ports list.
	// e.g. if you want to use a dedicated node port, you can just specify the StarRocksServicePort.Name and
	// StarRocksServicePort.NodePort field.
	// +optional
	Ports []StarRocksServicePort `json:"ports,omitempty"`

	// If specified and supported by the platform, this will restrict traffic through the cloud-provider
	// load-balancer will be restricted to the specified client IPs. This field will be ignored if the
	// cloud-provider does not support the feature.
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/
	// +optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty"`
}

// StarRocksServicePort defines the port that will be exposed by this service.
// To assign a specific port or nodePort to a service, you should specify them by the corresponding name or
// containerPort in the service configuration. If both containerPort and name are specified, containerPort takes precedence.
type StarRocksServicePort struct {
	// Name of this port within the service.
	// For fe, port name can be http, query, rpc, edit-log, and their default container port is 8030, 9030, 9020, 9010.
	// For be, port name can be webserver, heartbeat, brpc, be, and their default container port is 8040, 9050, 8060, 9060.
	// For cn, port name can be webserver, heartbeat, brpc, thrift, and their default container port is 8040, 9050, 8060, 9060.
	// For fe proxy, port name can be http-port, and its default container port is 8080.
	// +optional
	Name string `json:"name,omitempty"`

	// Port that will be exposed by this service.
	// +optional
	Port int32 `json:"port,omitempty"`

	// ContainerPort of the service port.
	// For fe, port name can be http, query, rpc, edit-log, and their default container port is 8030, 9030, 9020, 9010.
	// For be, port name can be webserver, heartbeat, brpc, be, and their default container port is 8040, 9050, 8060, 9060.
	// For cn, port name can be webserver, heartbeat, brpc, thrift, and their default container port is 8040, 9050, 8060, 9060.
	// For fe proxy, port name can be http-port, and its default container port is 8080.
	// +optional
	ContainerPort int32 `json:"containerPort,omitempty"`

	// NodePort is used to specify the port on each node on which the service is exposed.
	// The range of valid ports is 30000-32767
	// +optional
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

func (spec *StarRocksLoadSpec) GetReplicas() *int32 {
	return spec.Replicas
}

func (spec *StarRocksLoadSpec) GetStorageVolumes() []StorageVolume {
	return spec.StorageVolumes
}

func (spec *StarRocksLoadSpec) GetServiceAccount() string {
	return spec.ServiceAccount
}

func (spec *StarRocksLoadSpec) GetAffinity() *corev1.Affinity {
	return spec.Affinity
}

func (spec *StarRocksLoadSpec) GetTopologySpreadConstraints() []corev1.TopologySpreadConstraint {
	return spec.TopologySpreadConstraints
}

func (spec *StarRocksLoadSpec) GetTolerations() []corev1.Toleration {
	return spec.Tolerations
}

func (spec *StarRocksLoadSpec) GetService() *StarRocksService {
	return spec.Service
}

func (spec *StarRocksLoadSpec) GetNodeSelector() map[string]string {
	return spec.NodeSelector
}

func (spec *StarRocksLoadSpec) GetImagePullSecrets() []corev1.LocalObjectReference {
	return spec.ImagePullSecrets
}

func (spec *StarRocksLoadSpec) GetAnnotations() map[string]string {
	return spec.Annotations
}

func (spec *StarRocksLoadSpec) GetSchedulerName() string {
	return spec.SchedulerName
}

func (spec *StarRocksLoadSpec) GetStartupProbeFailureSeconds() *int32 {
	return spec.StartupProbeFailureSeconds
}

func (spec *StarRocksLoadSpec) GetLivenessProbeFailureSeconds() *int32 {
	return spec.LivenessProbeFailureSeconds
}

func (spec *StarRocksLoadSpec) GetReadinessProbeFailureSeconds() *int32 {
	return spec.ReadinessProbeFailureSeconds
}

func (spec *StarRocksLoadSpec) GetLifecycle() *corev1.Lifecycle {
	return spec.Lifecycle
}

func (spec *StarRocksLoadSpec) GetShareProcessNamespace() *bool {
	return spec.ShareProcessNamespace
}

func (spec *StarRocksLoadSpec) GetImagePullPolicy() corev1.PullPolicy {
	if spec.ImagePullPolicy == "" {
		return corev1.PullIfNotPresent
	}
	return spec.ImagePullPolicy
}
