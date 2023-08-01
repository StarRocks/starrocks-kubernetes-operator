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

package cn

import (
	"time"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	_logPath         = "/opt/starrocks/cn/log"
	_logName         = "cn-log"
	_cnConfigPath    = "/etc/starrocks/cn/conf"
	_envCnConfigPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) corev1.PodTemplateSpec {
	metaName := src.Name + "-" + srapi.DEFAULT_CN
	cnSpec := src.Spec.StarRocksCnSpec

	vols, volumeMounts, vexist := pod.MountStorageVolumes(cnSpec)
	// add default volume about log
	if _, ok := vexist[_logPath]; !ok {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      _logName,
			MountPath: _logPath,
		})
		vols = append(vols, corev1.Volume{
			Name: _logName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, cnSpec.ConfigMapInfo, _cnConfigPath)
	vols, volumeMounts = pod.MountConfigMaps(vols, volumeMounts, cnSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, cnSpec.Secrets)

	feExternalServiceName := service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	Envs := pod.Envs(src.Spec.StarRocksCnSpec, config, feExternalServiceName, src.Namespace, cnSpec.CnEnvVars)
	cnContainer := corev1.Container{
		Name:            srapi.DEFAULT_CN,
		Image:           cnSpec.Image,
		Command:         []string{"/opt/starrocks/cn_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(cnSpec, config),
		Env:             Envs,
		Resources:       cnSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle("/opt/starrocks/cn_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(cnSpec),
	}

	if cnSpec.ConfigMapInfo.ConfigMapName != "" && cnSpec.ConfigMapInfo.ResolveKey != "" {
		cnContainer.Env = append(cnContainer.Env, corev1.EnvVar{
			Name:  _envCnConfigPath,
			Value: _cnConfigPath,
		})
	}

	podSpec := pod.Spec(cnSpec, src.Spec.ServiceAccount, cnContainer, vols)
	now := time.Now().Format(time.RFC3339)
	annotations := pod.Annotations(cnSpec, src.Annotations, now)
	podSpec.SecurityContext = pod.PodSecurityContext(cnSpec)
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaName,
			Annotations: annotations,
			Namespace:   src.Namespace,
			Labels:      pod.Labels(src.Name, cnSpec),
		},
		Spec: podSpec,
	}
}
