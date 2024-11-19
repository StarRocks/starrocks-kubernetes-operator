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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

const (
	_logName         = "be-log"
	_beConfigPath    = "/etc/starrocks/be/conf"
	_storageName     = "be-storage"
	_storageName2    = "be-data" // helm chart use this format
	_envBeConfigPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (be *BeController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) (*corev1.PodTemplateSpec, error) {
	metaName := src.Name + "-" + srapi.DEFAULT_BE
	beSpec := src.Spec.StarRocksBeSpec

	vols, volumeMounts := pod.MountStorageVolumes(beSpec)

	if !k8sutils.HasVolume(vols, _storageName) && !k8sutils.HasVolume(vols, _storageName2) &&
		!k8sutils.HasMountPath(volumeMounts, pod.GetStorageDir(beSpec)) {
		// Changing the volume name to _storageName2 is fine, it will only affect users who did not persist data.
		// The reason why we need to change the volume name is that the helm chart uses the format _storageName2
		// Keeping the same suffix will make user easy to use feature, like init-containers and sidecars.
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _storageName2, pod.GetStorageDir(beSpec), "")
	}
	if !k8sutils.HasVolume(vols, _logName) && !k8sutils.HasMountPath(volumeMounts, pod.GetLogDir(beSpec)) {
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _logName, pod.GetLogDir(beSpec), "")
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, beSpec.ConfigMapInfo, _beConfigPath)
	vols, volumeMounts = pod.MountConfigMaps(beSpec, vols, volumeMounts, beSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, beSpec.Secrets)
	if err := k8sutils.CheckVolumes(vols, volumeMounts); err != nil {
		return nil, err
	}

	feExternalServiceName := service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	envs := pod.Envs(src.Spec.StarRocksBeSpec, config, feExternalServiceName, src.Namespace, beSpec.BeEnvVars)
	webServerPort := rutils.GetPort(config, rutils.WEBSERVER_PORT)
	beContainer := corev1.Container{
		Name:            srapi.DEFAULT_BE,
		Image:           beSpec.Image,
		Command:         pod.ContainerCommand(beSpec),
		Args:            pod.ContainerArgs(beSpec),
		Ports:           pod.Ports(beSpec, config),
		Env:             envs,
		Resources:       beSpec.ResourceRequirements,
		ImagePullPolicy: beSpec.GetImagePullPolicy(),
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(beSpec.GetStartupProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(beSpec.GetLivenessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(beSpec.GetReadinessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle(beSpec.GetLifecycle(), pod.GetPreStopScriptPath(beSpec)),
		SecurityContext: pod.ContainerSecurityContext(beSpec),
	}
	if pod.GetStarRocksRootPath(beSpec.BeEnvVars) != pod.GetStarRocksDefaultRootPath() {
		beContainer.WorkingDir = pod.GetStarRocksRootPath(beSpec.BeEnvVars)
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
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaName,
			Annotations: annotations,
			Namespace:   src.Namespace,
			Labels:      pod.Labels(src.Name, beSpec),
		},
		Spec: podSpec,
	}, nil
}
