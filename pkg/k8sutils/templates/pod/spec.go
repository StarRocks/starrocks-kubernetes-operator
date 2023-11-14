// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pod

import (
	"os"
	"strconv"
	"strings"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	HEALTH_API_PATH = "/api/health"
)

// StartupProbe returns a startup probe.
func StartupProbe(startupProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 60
	var defaultPeriodSeconds int32 = 5
	return completeProbe(startupProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

// LivenessProbe returns a liveness probe.
func LivenessProbe(livenessProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 3
	var defaultPeriodSeconds int32 = 5
	return completeProbe(livenessProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

// ReadinessProbe returns a readiness probe.
func ReadinessProbe(readinessProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 3
	var defaultPeriodSeconds int32 = 5
	return completeProbe(readinessProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

// LifeCycle returns a lifecycle.
func LifeCycle(preStopScriptPath string) *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{preStopScriptPath},
			},
		},
	}
}

func getProbe(port int32, path string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: path,
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: port,
			},
		},
	}
}

func completeProbe(failureSeconds *int32, defaultFailureThreshold int32, defaultPeriodSeconds int32,
	probeHandler corev1.ProbeHandler) *corev1.Probe {
	probe := &corev1.Probe{}
	if failureSeconds != nil && *failureSeconds > 0 {
		probe.FailureThreshold = (*failureSeconds + defaultPeriodSeconds - 1) / defaultPeriodSeconds
	} else {
		probe.FailureThreshold = defaultFailureThreshold
	}
	probe.PeriodSeconds = defaultPeriodSeconds
	probe.ProbeHandler = probeHandler
	return probe
}

func getVolumeName(mountInfo v1.MountInfo) string {
	suffixLen := 4
	suffix := hash.HashObject(mountInfo)
	if len(suffix) > suffixLen {
		suffix = suffix[:suffixLen]
	}
	return mountInfo.Name + "-" + suffix
}

func MountConfigMaps(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	references []v1.ConfigMapReference) ([]corev1.Volume, []corev1.VolumeMount) {
	for _, reference := range references {
		volumeName := getVolumeName(v1.MountInfo(reference))
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: reference.Name,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: reference.MountPath,
			SubPath:   reference.SubPath,
		})
	}
	return volumes, volumeMounts
}

func MountSecrets(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	references []v1.SecretReference) ([]corev1.Volume, []corev1.VolumeMount) {
	for _, reference := range references {
		volumeName := getVolumeName(v1.MountInfo(reference))
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: reference.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: reference.MountPath,
			SubPath:   reference.SubPath,
		})
	}
	return volumes, volumeMounts
}

func MountStorageVolumes(spec v1.SpecInterface) ([]corev1.Volume, []corev1.VolumeMount, map[string]bool) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	vexist := make(map[string]bool)
	for _, sv := range spec.GetStorageVolumes() {
		// do not use getVolumeName for backward compatibility
		vexist[sv.MountPath] = true
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      sv.Name,
			MountPath: sv.MountPath,
			SubPath:   sv.SubPath,
		})

		volumes = append(volumes, corev1.Volume{
			Name: sv.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: sv.Name,
				},
			},
		})
	}
	return volumes, volumeMounts, vexist
}

func MountConfigMapInfo(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	cmInfo v1.ConfigMapInfo, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	if cmInfo.ConfigMapName != "" && cmInfo.ResolveKey != "" {
		// do not use getVolumeName for backward compatibility
		volumes = append(volumes, corev1.Volume{
			Name: cmInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cmInfo.ConfigMapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      cmInfo.ConfigMapName,
			MountPath: mountPath,
		})
	}
	return volumes, volumeMounts
}

func Labels(clusterName string, spec v1.SpecInterface) map[string]string {
	labels := load.Selector(clusterName, spec)
	switch v := spec.(type) {
	case *v1.StarRocksBeSpec:
		if v != nil {
			labels.AddLabel(v.PodLabels)
		}
	case *v1.StarRocksCnSpec:
		if v != nil {
			labels.AddLabel(v.PodLabels)
		}
	case *v1.StarRocksFeSpec:
		if v != nil {
			labels.AddLabel(v.PodLabels)
		}
	}
	return labels
}

func Envs(spec v1.SpecInterface, config map[string]interface{},
	feExternalServiceName string, namespace string, envs []corev1.EnvVar) []corev1.EnvVar {
	// copy envs
	envs = append([]corev1.EnvVar(nil), envs...)

	keys := make(map[string]bool)
	for _, env := range envs {
		keys[env.Name] = true
	}

	unsupportedEnvironments := make(map[string]bool)
	if unsupportedEnvs := os.Getenv("KUBE_STARROCKS_UNSUPPORTED_ENVS"); unsupportedEnvs != "" {
		for _, name := range strings.Split(unsupportedEnvs, ",") {
			unsupportedEnvironments[name] = true
		}
	}

	addEnv := func(envVar corev1.EnvVar) {
		if !keys[envVar.Name] && !unsupportedEnvironments[envVar.Name] {
			keys[envVar.Name] = true
			envs = append(envs, envVar)
		}
	}

	for _, envVar := range []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.hostIP"},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
			},
		},
		{
			Name:  "HOST_TYPE",
			Value: "FQDN",
		},
	} {
		addEnv(envVar)
	}

	switch spec.(type) {
	case *v1.StarRocksFeSpec:
		for _, envVar := range []corev1.EnvVar{
			{
				Name:  v1.COMPONENT_NAME,
				Value: v1.DEFAULT_FE,
			},
			{
				Name:  v1.FE_SERVICE_NAME,
				Value: feExternalServiceName + "." + namespace,
			},
		} {
			addEnv(envVar)
		}
	case *v1.StarRocksBeSpec:
		for _, envVar := range []corev1.EnvVar{
			{
				Name:  v1.COMPONENT_NAME,
				Value: v1.DEFAULT_BE,
			},
			{
				Name:  v1.FE_SERVICE_NAME,
				Value: feExternalServiceName,
			},
			{
				Name:  "FE_QUERY_PORT",
				Value: strconv.FormatInt(int64(rutils.GetPort(config, rutils.QUERY_PORT)), 10),
			},
		} {
			addEnv(envVar)
		}
	case *v1.StarRocksCnSpec:
		for _, envVar := range []corev1.EnvVar{
			{
				Name:  v1.COMPONENT_NAME,
				Value: v1.DEFAULT_CN,
			},
			{
				Name:  v1.FE_SERVICE_NAME,
				Value: feExternalServiceName,
			},
			{
				Name:  "FE_QUERY_PORT",
				Value: strconv.FormatInt(int64(rutils.GetPort(config, rutils.QUERY_PORT)), 10),
			},
		} {
			addEnv(envVar)
		}
	}

	return envs
}

func Ports(spec v1.SpecInterface, config map[string]interface{}) []corev1.ContainerPort {
	var ports []corev1.ContainerPort
	switch spec.(type) {
	case *v1.StarRocksFeSpec:
		ports = append(ports, []corev1.ContainerPort{
			{
				Name:          "http-port",
				ContainerPort: rutils.GetPort(config, rutils.HTTP_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "rpc-port",
				ContainerPort: rutils.GetPort(config, rutils.RPC_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "query-port",
				ContainerPort: rutils.GetPort(config, rutils.QUERY_PORT),
				Protocol:      corev1.ProtocolTCP,
			},
		}...)
	case *v1.StarRocksBeSpec:
		ports = append(ports, []corev1.ContainerPort{
			{
				Name:          "be-port",
				ContainerPort: rutils.GetPort(config, rutils.BE_PORT),
			}, {
				Name:          "webserver-port",
				ContainerPort: rutils.GetPort(config, rutils.WEBSERVER_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "heartbeat-port",
				ContainerPort: rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "brpc-port",
				ContainerPort: rutils.GetPort(config, rutils.BRPC_PORT),
				Protocol:      corev1.ProtocolTCP,
			},
		}...)
	case *v1.StarRocksCnSpec:
		ports = append(ports, []corev1.ContainerPort{
			{
				Name:          "thrift-port",
				ContainerPort: rutils.GetPort(config, rutils.THRIFT_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "webserver-port",
				ContainerPort: rutils.GetPort(config, rutils.WEBSERVER_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "heartbeat-port",
				ContainerPort: rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "brpc-port",
				ContainerPort: rutils.GetPort(config, rutils.BRPC_PORT),
				Protocol:      corev1.ProtocolTCP,
			},
		}...)
	case *v1.StarRocksFeProxySpec:
		ports = append(ports, []corev1.ContainerPort{
			{
				Name:          rutils.FE_PORXY_HTTP_PORT_NAME,
				ContainerPort: rutils.FE_PROXY_HTTP_PORT,
				Protocol:      corev1.ProtocolTCP,
			},
		}...)
	}
	return ports
}

func Spec(spec v1.SpecInterface, container corev1.Container, volumes []corev1.Volume) corev1.PodSpec {
	podSpec := corev1.PodSpec{
		Containers:                    []corev1.Container{container},
		Volumes:                       volumes,
		ServiceAccountName:            spec.GetServiceAccount(),
		TerminationGracePeriodSeconds: spec.GetTerminationGracePeriodSeconds(),
		Affinity:                      spec.GetAffinity(),
		Tolerations:                   spec.GetTolerations(),
		ImagePullSecrets:              spec.GetImagePullSecrets(),
		NodeSelector:                  spec.GetNodeSelector(),
		HostAliases:                   spec.GetHostAliases(),
		SchedulerName:                 spec.GetSchedulerName(),
		AutomountServiceAccountToken:  func() *bool { b := false; return &b }(),
	}
	return podSpec
}

func Annotations(spec v1.SpecInterface) map[string]string {
	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}
	return annotations
}

func PodSecurityContext(spec v1.SpecInterface) *corev1.PodSecurityContext {
	_, groupID := spec.GetRunAsNonRoot()
	fsGroup := (*int64)(nil)
	if groupID != nil {
		fsGroup = groupID
	}
	onRootMismatch := corev1.FSGroupChangeOnRootMismatch
	sc := &corev1.PodSecurityContext{
		FSGroupChangePolicy: &onRootMismatch,
		FSGroup:             fsGroup,
	}
	return sc
}

func ContainerSecurityContext(spec v1.SpecInterface) *corev1.SecurityContext {
	userID, groupID := spec.GetRunAsNonRoot()

	var runAsNonRoot *bool
	if userID != nil && *userID != 0 {
		b := true
		runAsNonRoot = &b
	}
	return &corev1.SecurityContext{
		RunAsUser:                userID,
		RunAsGroup:               groupID,
		RunAsNonRoot:             runAsNonRoot,
		AllowPrivilegeEscalation: func() *bool { b := false; return &b }(),
		// starrocks will create pid file, eg.g /opt/starrocks/fe/bin/fe.pid, so set it to false
		ReadOnlyRootFilesystem: func() *bool { b := false; return &b }(),
	}
}
