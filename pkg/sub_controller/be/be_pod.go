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

package be

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	_logPath         = "/opt/starrocks/be/log"
	_logName         = "be-log"
	_beConfigPath    = "/etc/starrocks/be/conf"
	_storageName     = "be-storage"
	_storagePath     = "/opt/starrocks/be/storage"
	_envBeConfigPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (be *BeController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) corev1.PodTemplateSpec {
	metaName := src.Name + "-" + srapi.DEFAULT_BE
	beSpec := src.Spec.StarRocksBeSpec

	vols, volumeMounts, vexist := pod.MountStorageVolumes(beSpec)
	// add default volume about log, if meta not configure.
	if _, ok := vexist[_logPath]; !ok {
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _logName, _logPath, "")
	}
	if _, ok := vexist[_storagePath]; !ok {
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _storageName, _storagePath, "")
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, beSpec.ConfigMapInfo, _beConfigPath)
	vols, volumeMounts = pod.MountConfigMaps(vols, volumeMounts, beSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, beSpec.Secrets)

	feExternalServiceName := service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	envs := pod.Envs(src.Spec.StarRocksBeSpec, config, feExternalServiceName, src.Namespace, beSpec.BeEnvVars)
	webServerPort := rutils.GetPort(config, rutils.WEBSERVER_PORT)
	beContainer := corev1.Container{
		Name:            srapi.DEFAULT_BE,
		Image:           beSpec.Image,
		Command:         []string{"/opt/starrocks/be_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(beSpec, config),
		Env:             envs,
		Resources:       beSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(beSpec.GetStartupProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(beSpec.GetLivenessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(beSpec.GetReadinessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle("/opt/starrocks/be_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(beSpec),
	}
	if beSpec.ConfigMapInfo.ConfigMapName != "" && beSpec.ConfigMapInfo.ResolveKey != "" {
		beContainer.Env = append(beContainer.Env, corev1.EnvVar{
			Name:  _envBeConfigPath,
			Value: _beConfigPath,
		})
	}

	podSpec := pod.Spec(beSpec, beContainer, vols)

	annotations := pod.Annotations(beSpec)
	podSpec.SecurityContext = pod.PodSecurityContext(beSpec)
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaName,
			Annotations: annotations,
			Namespace:   src.Namespace,
			Labels:      pod.Labels(src.Name, beSpec),
		},
		Spec: podSpec,
	}
}
