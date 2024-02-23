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

package fe

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

const (
	_metaPath        = "/opt/starrocks/fe/meta"
	_metaName        = "fe-meta"
	_logPath         = "/opt/starrocks/fe/log"
	_logName         = "fe-log"
	_feConfigPath    = "/etc/starrocks/fe/conf"
	_envFeConfigPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy fe.
func (fc *FeController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) (*corev1.PodTemplateSpec, error) {
	metaName := src.Name + "-" + srapi.DEFAULT_FE
	feSpec := src.Spec.StarRocksFeSpec

	vols, volMounts, vexist := pod.MountStorageVolumes(feSpec)
	// add default volume about log, meta if not configure.
	if _, ok := vexist[_metaPath]; !ok {
		vols, volMounts = pod.MountEmptyDirVolume(vols, volMounts, _metaName, _metaPath, "")
	}
	if _, ok := vexist[_logPath]; !ok {
		vols, volMounts = pod.MountEmptyDirVolume(vols, volMounts, _logName, _logPath, "")
	}

	// mount configmap, secrets to pod if needed
	vols, volMounts = pod.MountConfigMapInfo(vols, volMounts, feSpec.ConfigMapInfo, _feConfigPath)
	vols, volMounts = pod.MountConfigMaps(vols, volMounts, feSpec.ConfigMaps)
	vols, volMounts = pod.MountSecrets(vols, volMounts, feSpec.Secrets)
	if err := k8sutils.CheckVolumes(vols, volMounts); err != nil {
		return nil, err
	}

	feExternalServiceName := service.ExternalServiceName(src.Name, feSpec)
	envs := pod.Envs(src.Spec.StarRocksFeSpec, config, feExternalServiceName, src.Namespace, feSpec.FeEnvVars)
	httpPort := rutils.GetPort(config, rutils.HTTP_PORT)
	feContainer := corev1.Container{
		Name:            srapi.DEFAULT_FE,
		Image:           feSpec.Image,
		Command:         []string{"/opt/starrocks/fe_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(feSpec, config),
		Env:             envs,
		Resources:       feSpec.ResourceRequirements,
		VolumeMounts:    volMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		StartupProbe:    pod.StartupProbe(feSpec.GetStartupProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(feSpec.GetLivenessProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(feSpec.GetReadinessProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle(feSpec.GetLifecycle(), "/opt/starrocks/fe_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(feSpec),
	}

	if feSpec.ConfigMapInfo.ConfigMapName != "" && feSpec.ConfigMapInfo.ResolveKey != "" {
		feContainer.Env = append(feContainer.Env, corev1.EnvVar{
			Name:  _envFeConfigPath,
			Value: _feConfigPath,
		})
	}

	podSpec := pod.Spec(feSpec, feContainer, vols)
	annotations := pod.Annotations(feSpec)
	podSpec.SecurityContext = pod.PodSecurityContext(feSpec)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaName,
			Namespace:   src.Namespace,
			Annotations: annotations,
			Labels:      pod.Labels(src.Name, src.Spec.StarRocksFeSpec),
		},
		Spec: podSpec,
	}, nil
}
