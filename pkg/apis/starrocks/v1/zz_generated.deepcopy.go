//go:build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoScalingPolicy) DeepCopyInto(out *AutoScalingPolicy) {
	*out = *in
	if in.HPAPolicy != nil {
		in, out := &in.HPAPolicy, &out.HPAPolicy
		*out = new(HPAPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoScalingPolicy.
func (in *AutoScalingPolicy) DeepCopy() *AutoScalingPolicy {
	if in == nil {
		return nil
	}
	out := new(AutoScalingPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapInfo) DeepCopyInto(out *ConfigMapInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapInfo.
func (in *ConfigMapInfo) DeepCopy() *ConfigMapInfo {
	if in == nil {
		return nil
	}
	out := new(ConfigMapInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapReference) DeepCopyInto(out *ConfigMapReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapReference.
func (in *ConfigMapReference) DeepCopy() *ConfigMapReference {
	if in == nil {
		return nil
	}
	out := new(ConfigMapReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DisasterRecovery) DeepCopyInto(out *DisasterRecovery) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DisasterRecovery.
func (in *DisasterRecovery) DeepCopy() *DisasterRecovery {
	if in == nil {
		return nil
	}
	out := new(DisasterRecovery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DisasterRecoveryStatus) DeepCopyInto(out *DisasterRecoveryStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DisasterRecoveryStatus.
func (in *DisasterRecoveryStatus) DeepCopy() *DisasterRecoveryStatus {
	if in == nil {
		return nil
	}
	out := new(DisasterRecoveryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HPAPolicy) DeepCopyInto(out *HPAPolicy) {
	*out = *in
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]v2beta2.MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Behavior != nil {
		in, out := &in.Behavior, &out.Behavior
		*out = new(v2beta2.HorizontalPodAutoscalerBehavior)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HPAPolicy.
func (in *HPAPolicy) DeepCopy() *HPAPolicy {
	if in == nil {
		return nil
	}
	out := new(HPAPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalScaler) DeepCopyInto(out *HorizontalScaler) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalScaler.
func (in *HorizontalScaler) DeepCopy() *HorizontalScaler {
	if in == nil {
		return nil
	}
	out := new(HorizontalScaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MountInfo) DeepCopyInto(out *MountInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MountInfo.
func (in *MountInfo) DeepCopy() *MountInfo {
	if in == nil {
		return nil
	}
	out := new(MountInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretReference) DeepCopyInto(out *SecretReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretReference.
func (in *SecretReference) DeepCopy() *SecretReference {
	if in == nil {
		return nil
	}
	out := new(SecretReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksBeSpec) DeepCopyInto(out *StarRocksBeSpec) {
	*out = *in
	in.StarRocksComponentSpec.DeepCopyInto(&out.StarRocksComponentSpec)
	if in.BeEnvVars != nil {
		in, out := &in.BeEnvVars, &out.BeEnvVars
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksBeSpec.
func (in *StarRocksBeSpec) DeepCopy() *StarRocksBeSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksBeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksBeStatus) DeepCopyInto(out *StarRocksBeStatus) {
	*out = *in
	in.StarRocksComponentStatus.DeepCopyInto(&out.StarRocksComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksBeStatus.
func (in *StarRocksBeStatus) DeepCopy() *StarRocksBeStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksBeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksCluster) DeepCopyInto(out *StarRocksCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksCluster.
func (in *StarRocksCluster) DeepCopy() *StarRocksCluster {
	if in == nil {
		return nil
	}
	out := new(StarRocksCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StarRocksCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksClusterList) DeepCopyInto(out *StarRocksClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StarRocksCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksClusterList.
func (in *StarRocksClusterList) DeepCopy() *StarRocksClusterList {
	if in == nil {
		return nil
	}
	out := new(StarRocksClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StarRocksClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksClusterSpec) DeepCopyInto(out *StarRocksClusterSpec) {
	*out = *in
	if in.StarRocksFeSpec != nil {
		in, out := &in.StarRocksFeSpec, &out.StarRocksFeSpec
		*out = new(StarRocksFeSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksBeSpec != nil {
		in, out := &in.StarRocksBeSpec, &out.StarRocksBeSpec
		*out = new(StarRocksBeSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksCnSpec != nil {
		in, out := &in.StarRocksCnSpec, &out.StarRocksCnSpec
		*out = new(StarRocksCnSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksFeProxySpec != nil {
		in, out := &in.StarRocksFeProxySpec, &out.StarRocksFeProxySpec
		*out = new(StarRocksFeProxySpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DisasterRecovery != nil {
		in, out := &in.DisasterRecovery, &out.DisasterRecovery
		*out = new(DisasterRecovery)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksClusterSpec.
func (in *StarRocksClusterSpec) DeepCopy() *StarRocksClusterSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksClusterStatus) DeepCopyInto(out *StarRocksClusterStatus) {
	*out = *in
	if in.StarRocksFeStatus != nil {
		in, out := &in.StarRocksFeStatus, &out.StarRocksFeStatus
		*out = new(StarRocksFeStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksBeStatus != nil {
		in, out := &in.StarRocksBeStatus, &out.StarRocksBeStatus
		*out = new(StarRocksBeStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksCnStatus != nil {
		in, out := &in.StarRocksCnStatus, &out.StarRocksCnStatus
		*out = new(StarRocksCnStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.StarRocksFeProxyStatus != nil {
		in, out := &in.StarRocksFeProxyStatus, &out.StarRocksFeProxyStatus
		*out = new(StarRocksFeProxyStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.DisasterRecoveryStatus != nil {
		in, out := &in.DisasterRecoveryStatus, &out.DisasterRecoveryStatus
		*out = new(DisasterRecoveryStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksClusterStatus.
func (in *StarRocksClusterStatus) DeepCopy() *StarRocksClusterStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksCnSpec) DeepCopyInto(out *StarRocksCnSpec) {
	*out = *in
	in.StarRocksComponentSpec.DeepCopyInto(&out.StarRocksComponentSpec)
	if in.CnEnvVars != nil {
		in, out := &in.CnEnvVars, &out.CnEnvVars
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AutoScalingPolicy != nil {
		in, out := &in.AutoScalingPolicy, &out.AutoScalingPolicy
		*out = new(AutoScalingPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksCnSpec.
func (in *StarRocksCnSpec) DeepCopy() *StarRocksCnSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksCnSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksCnStatus) DeepCopyInto(out *StarRocksCnStatus) {
	*out = *in
	in.StarRocksComponentStatus.DeepCopyInto(&out.StarRocksComponentStatus)
	out.HorizontalScaler = in.HorizontalScaler
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksCnStatus.
func (in *StarRocksCnStatus) DeepCopy() *StarRocksCnStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksCnStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksComponentSpec) DeepCopyInto(out *StarRocksComponentSpec) {
	*out = *in
	in.StarRocksLoadSpec.DeepCopyInto(&out.StarRocksLoadSpec)
	if in.RunAsNonRoot != nil {
		in, out := &in.RunAsNonRoot, &out.RunAsNonRoot
		*out = new(bool)
		**out = **in
	}
	if in.Capabilities != nil {
		in, out := &in.Capabilities, &out.Capabilities
		*out = new(corev1.Capabilities)
		(*in).DeepCopyInto(*out)
	}
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]ConfigMapReference, len(*in))
		copy(*out, *in)
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]SecretReference, len(*in))
		copy(*out, *in)
	}
	if in.HostAliases != nil {
		in, out := &in.HostAliases, &out.HostAliases
		*out = make([]corev1.HostAlias, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TerminationGracePeriodSeconds != nil {
		in, out := &in.TerminationGracePeriodSeconds, &out.TerminationGracePeriodSeconds
		*out = new(int64)
		**out = **in
	}
	if in.Sidecars != nil {
		in, out := &in.Sidecars, &out.Sidecars
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.UpdateStrategy != nil {
		in, out := &in.UpdateStrategy, &out.UpdateStrategy
		*out = new(appsv1.StatefulSetUpdateStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadOnlyRootFilesystem != nil {
		in, out := &in.ReadOnlyRootFilesystem, &out.ReadOnlyRootFilesystem
		*out = new(bool)
		**out = **in
	}
	if in.Sysctls != nil {
		in, out := &in.Sysctls, &out.Sysctls
		*out = make([]corev1.Sysctl, len(*in))
		copy(*out, *in)
	}
	if in.PersistentVolumeClaimRetentionPolicy != nil {
		in, out := &in.PersistentVolumeClaimRetentionPolicy, &out.PersistentVolumeClaimRetentionPolicy
		*out = new(appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksComponentSpec.
func (in *StarRocksComponentSpec) DeepCopy() *StarRocksComponentSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksComponentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksComponentStatus) DeepCopyInto(out *StarRocksComponentStatus) {
	*out = *in
	if in.FailedInstances != nil {
		in, out := &in.FailedInstances, &out.FailedInstances
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CreatingInstances != nil {
		in, out := &in.CreatingInstances, &out.CreatingInstances
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RunningInstances != nil {
		in, out := &in.RunningInstances, &out.RunningInstances
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ResourceNames != nil {
		in, out := &in.ResourceNames, &out.ResourceNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksComponentStatus.
func (in *StarRocksComponentStatus) DeepCopy() *StarRocksComponentStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksComponentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksFeProxySpec) DeepCopyInto(out *StarRocksFeProxySpec) {
	*out = *in
	in.StarRocksLoadSpec.DeepCopyInto(&out.StarRocksLoadSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksFeProxySpec.
func (in *StarRocksFeProxySpec) DeepCopy() *StarRocksFeProxySpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksFeProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksFeProxyStatus) DeepCopyInto(out *StarRocksFeProxyStatus) {
	*out = *in
	in.StarRocksComponentStatus.DeepCopyInto(&out.StarRocksComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksFeProxyStatus.
func (in *StarRocksFeProxyStatus) DeepCopy() *StarRocksFeProxyStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksFeProxyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksFeSpec) DeepCopyInto(out *StarRocksFeSpec) {
	*out = *in
	in.StarRocksComponentSpec.DeepCopyInto(&out.StarRocksComponentSpec)
	if in.FeEnvVars != nil {
		in, out := &in.FeEnvVars, &out.FeEnvVars
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksFeSpec.
func (in *StarRocksFeSpec) DeepCopy() *StarRocksFeSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksFeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksFeStatus) DeepCopyInto(out *StarRocksFeStatus) {
	*out = *in
	in.StarRocksComponentStatus.DeepCopyInto(&out.StarRocksComponentStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksFeStatus.
func (in *StarRocksFeStatus) DeepCopy() *StarRocksFeStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksFeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksLoadSpec) DeepCopyInto(out *StarRocksLoadSpec) {
	*out = *in
	in.ResourceRequirements.DeepCopyInto(&out.ResourceRequirements)
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodLabels != nil {
		in, out := &in.PodLabels, &out.PodLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TopologySpreadConstraints != nil {
		in, out := &in.TopologySpreadConstraints, &out.TopologySpreadConstraints
		*out = make([]corev1.TopologySpreadConstraint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(StarRocksService)
		(*in).DeepCopyInto(*out)
	}
	if in.StorageVolumes != nil {
		in, out := &in.StorageVolumes, &out.StorageVolumes
		*out = make([]StorageVolume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.ConfigMapInfo = in.ConfigMapInfo
	if in.StartupProbeFailureSeconds != nil {
		in, out := &in.StartupProbeFailureSeconds, &out.StartupProbeFailureSeconds
		*out = new(int32)
		**out = **in
	}
	if in.LivenessProbeFailureSeconds != nil {
		in, out := &in.LivenessProbeFailureSeconds, &out.LivenessProbeFailureSeconds
		*out = new(int32)
		**out = **in
	}
	if in.ReadinessProbeFailureSeconds != nil {
		in, out := &in.ReadinessProbeFailureSeconds, &out.ReadinessProbeFailureSeconds
		*out = new(int32)
		**out = **in
	}
	if in.Lifecycle != nil {
		in, out := &in.Lifecycle, &out.Lifecycle
		*out = new(corev1.Lifecycle)
		(*in).DeepCopyInto(*out)
	}
	if in.ShareProcessNamespace != nil {
		in, out := &in.ShareProcessNamespace, &out.ShareProcessNamespace
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksLoadSpec.
func (in *StarRocksLoadSpec) DeepCopy() *StarRocksLoadSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksLoadSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksProbe) DeepCopyInto(out *StarRocksProbe) {
	*out = *in
	if in.InitialDelaySeconds != nil {
		in, out := &in.InitialDelaySeconds, &out.InitialDelaySeconds
		*out = new(int32)
		**out = **in
	}
	if in.PeriodSeconds != nil {
		in, out := &in.PeriodSeconds, &out.PeriodSeconds
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksProbe.
func (in *StarRocksProbe) DeepCopy() *StarRocksProbe {
	if in == nil {
		return nil
	}
	out := new(StarRocksProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksService) DeepCopyInto(out *StarRocksService) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]StarRocksServicePort, len(*in))
		copy(*out, *in)
	}
	if in.LoadBalancerSourceRanges != nil {
		in, out := &in.LoadBalancerSourceRanges, &out.LoadBalancerSourceRanges
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksService.
func (in *StarRocksService) DeepCopy() *StarRocksService {
	if in == nil {
		return nil
	}
	out := new(StarRocksService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksServicePort) DeepCopyInto(out *StarRocksServicePort) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksServicePort.
func (in *StarRocksServicePort) DeepCopy() *StarRocksServicePort {
	if in == nil {
		return nil
	}
	out := new(StarRocksServicePort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksWarehouse) DeepCopyInto(out *StarRocksWarehouse) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksWarehouse.
func (in *StarRocksWarehouse) DeepCopy() *StarRocksWarehouse {
	if in == nil {
		return nil
	}
	out := new(StarRocksWarehouse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StarRocksWarehouse) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksWarehouseList) DeepCopyInto(out *StarRocksWarehouseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StarRocksWarehouse, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksWarehouseList.
func (in *StarRocksWarehouseList) DeepCopy() *StarRocksWarehouseList {
	if in == nil {
		return nil
	}
	out := new(StarRocksWarehouseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StarRocksWarehouseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksWarehouseSpec) DeepCopyInto(out *StarRocksWarehouseSpec) {
	*out = *in
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(WarehouseComponentSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksWarehouseSpec.
func (in *StarRocksWarehouseSpec) DeepCopy() *StarRocksWarehouseSpec {
	if in == nil {
		return nil
	}
	out := new(StarRocksWarehouseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StarRocksWarehouseStatus) DeepCopyInto(out *StarRocksWarehouseStatus) {
	*out = *in
	if in.WarehouseComponentStatus != nil {
		in, out := &in.WarehouseComponentStatus, &out.WarehouseComponentStatus
		*out = new(StarRocksCnStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StarRocksWarehouseStatus.
func (in *StarRocksWarehouseStatus) DeepCopy() *StarRocksWarehouseStatus {
	if in == nil {
		return nil
	}
	out := new(StarRocksWarehouseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageVolume) DeepCopyInto(out *StorageVolume) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	if in.HostPath != nil {
		in, out := &in.HostPath, &out.HostPath
		*out = new(corev1.HostPathVolumeSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageVolume.
func (in *StorageVolume) DeepCopy() *StorageVolume {
	if in == nil {
		return nil
	}
	out := new(StorageVolume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WarehouseComponentSpec) DeepCopyInto(out *WarehouseComponentSpec) {
	*out = *in
	in.StarRocksComponentSpec.DeepCopyInto(&out.StarRocksComponentSpec)
	if in.EnvVars != nil {
		in, out := &in.EnvVars, &out.EnvVars
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AutoScalingPolicy != nil {
		in, out := &in.AutoScalingPolicy, &out.AutoScalingPolicy
		*out = new(AutoScalingPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WarehouseComponentSpec.
func (in *WarehouseComponentSpec) DeepCopy() *WarehouseComponentSpec {
	if in == nil {
		return nil
	}
	out := new(WarehouseComponentSpec)
	in.DeepCopyInto(out)
	return out
}
