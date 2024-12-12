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
	"fmt"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

const (
	HEALTH_API_PATH = "/api/health"
)

// LifeCycle returns a lifecycle.
func LifeCycle(lifeCycle *corev1.Lifecycle, preStopScriptPath string) *corev1.Lifecycle {
	defaultPreStop := &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{
			Command: []string{preStopScriptPath},
		},
	}

	if lifeCycle == nil {
		return &corev1.Lifecycle{
			PreStop: defaultPreStop,
		}
	}

	preStop := lifeCycle.PreStop
	if preStop == nil {
		preStop = defaultPreStop
	}
	return &corev1.Lifecycle{
		PreStop:   preStop,
		PostStart: lifeCycle.PostStart,
	}
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
	containers := []corev1.Container{container}
	if len(spec.GetSidecars()) > 0 {
		containers = append(containers, spec.GetSidecars()...)
	}
	podSpec := corev1.PodSpec{
		InitContainers:                spec.GetInitContainers(),
		Containers:                    containers,
		Volumes:                       volumes,
		ServiceAccountName:            spec.GetServiceAccount(),
		TerminationGracePeriodSeconds: spec.GetTerminationGracePeriodSeconds(),
		Affinity:                      spec.GetAffinity(),
		TopologySpreadConstraints:     spec.GetTopologySpreadConstraints(),
		Tolerations:                   spec.GetTolerations(),
		ImagePullSecrets:              spec.GetImagePullSecrets(),
		NodeSelector:                  spec.GetNodeSelector(),
		HostAliases:                   spec.GetHostAliases(),
		SchedulerName:                 spec.GetSchedulerName(),
		AutomountServiceAccountToken:  func() *bool { b := false; return &b }(),
		ShareProcessNamespace:         spec.GetShareProcessNamespace(),
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
		ReadOnlyRootFilesystem:   spec.IsReadOnlyRootFilesystem(),
		// set additional Capabilities
		Capabilities: spec.GetCapabilities(),
	}
}

func getDefaultEntrypointScript(spec v1.SpecInterface) string {
	switch v := spec.(type) {
	case *v1.StarRocksFeSpec:
		return fmt.Sprintf("%s/fe_entrypoint.sh", GetStarRocksRootPath(v.FeEnvVars))
	case *v1.StarRocksBeSpec:
		return fmt.Sprintf("%s/be_entrypoint.sh", GetStarRocksRootPath(v.BeEnvVars))
	case *v1.StarRocksCnSpec:
		return fmt.Sprintf("%s/cn_entrypoint.sh", GetStarRocksRootPath(v.CnEnvVars))
	}
	return ""
}

func GetStorageDir(spec v1.SpecInterface) string {
	switch v := spec.(type) {
	case *v1.StarRocksFeSpec:
		return fmt.Sprintf("%s/fe/meta", GetStarRocksRootPath(v.FeEnvVars))
	case *v1.StarRocksBeSpec:
		return fmt.Sprintf("%s/be/storage", GetStarRocksRootPath(v.BeEnvVars))
	case *v1.StarRocksCnSpec:
		return fmt.Sprintf("%s/cn/storage", GetStarRocksRootPath(v.CnEnvVars))
	}
	return ""
}

func GetLogDir(spec v1.SpecInterface) string {
	switch v := spec.(type) {
	case *v1.StarRocksFeSpec:
		return fmt.Sprintf("%s/fe/log", GetStarRocksRootPath(v.FeEnvVars))
	case *v1.StarRocksBeSpec:
		return fmt.Sprintf("%s/be/log", GetStarRocksRootPath(v.BeEnvVars))
	case *v1.StarRocksCnSpec:
		return fmt.Sprintf("%s/cn/log", GetStarRocksRootPath(v.CnEnvVars))
	}
	return ""
}

func GetConfigDir(spec v1.SpecInterface) string {
	switch v := spec.(type) {
	case *v1.StarRocksFeSpec:
		return fmt.Sprintf("%s/fe/conf", GetStarRocksRootPath(v.FeEnvVars))
	case *v1.StarRocksBeSpec:
		return fmt.Sprintf("%s/be/conf", GetStarRocksRootPath(v.BeEnvVars))
	case *v1.StarRocksCnSpec:
		return fmt.Sprintf("%s/cn/conf", GetStarRocksRootPath(v.CnEnvVars))
	}
	return ""
}

func GetPreStopScriptPath(spec v1.SpecInterface) string {
	switch v := spec.(type) {
	case *v1.StarRocksFeSpec:
		return fmt.Sprintf("%s/fe_prestop.sh", GetStarRocksRootPath(v.FeEnvVars))
	case *v1.StarRocksBeSpec:
		return fmt.Sprintf("%s/be_prestop.sh", GetStarRocksRootPath(v.BeEnvVars))
	case *v1.StarRocksCnSpec:
		return fmt.Sprintf("%s/cn_prestop.sh", GetStarRocksRootPath(v.CnEnvVars))
	}
	return ""
}

func GetStarRocksDefaultRootPath() string {
	return "/opt/starrocks"
}

func GetStarRocksRootPath(envVars []corev1.EnvVar) string {
	for _, env := range envVars {
		if common.EqualsIgnoreCase(env.Name, "STARROCKS_ROOT") {
			return env.Value
		}
	}
	return "/opt/starrocks"
}

func ContainerCommand(spec v1.SpecInterface) []string {
	if spec.GetCommand() != nil {
		return spec.GetCommand()
	}

	script := getDefaultEntrypointScript(spec)
	return []string{script}
}

func ContainerArgs(spec v1.SpecInterface) []string {
	if spec.GetArgs() != nil {
		return spec.GetArgs()
	}

	return []string{"$(FE_SERVICE_NAME)"}
}
