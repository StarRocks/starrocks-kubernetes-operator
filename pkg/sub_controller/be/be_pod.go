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
	"time"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	log_path           = "/opt/starrocks/be/log"
	log_name           = "be-log"
	be_config_path     = "/etc/starrocks/be/conf"
	storage_name       = "be-storage"
	storage_path       = "/opt/starrocks/be/storage"
	env_be_config_path = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (be *BeController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_BE
	beSpec := src.Spec.StarRocksBeSpec

	vols, volumeMounts, vexist := pod.MountStorageVolumes(beSpec)
	// add default volume about log, if meta not configure.
	if _, ok := vexist[log_path]; !ok {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			// use storage volume.
			Name:      log_name,
			MountPath: log_path,
		})
		vols = append(vols, corev1.Volume{
			Name: log_name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
	if _, ok := vexist[storage_path]; !ok {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			// use storage volume.
			Name:      storage_name,
			MountPath: storage_path,
		})

		vols = append(vols, corev1.Volume{
			Name: storage_name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, beSpec.ConfigMapInfo, be_config_path)
	vols, volumeMounts = pod.MountConfigMaps(vols, volumeMounts, beSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, beSpec.Secrets)

	feExternalServiceName := srapi.GetExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	Envs := pod.Envs(src.Spec.StarRocksBeSpec, config, feExternalServiceName, src.Namespace, beSpec.BeEnvVars)
	beContainer := corev1.Container{
		Name:            srapi.DEFAULT_BE,
		Image:           beSpec.Image,
		Command:         []string{"/opt/starrocks/be_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(beSpec, config),
		Env:             Envs,
		Resources:       beSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(rutils.GetPort(config, rutils.WEBSERVER_PORT), pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle("/opt/starrocks/be_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(beSpec),
	}
	if beSpec.ConfigMapInfo.ConfigMapName != "" && beSpec.ConfigMapInfo.ResolveKey != "" {
		beContainer.Env = append(beContainer.Env, corev1.EnvVar{
			Name:  env_be_config_path,
			Value: be_config_path,
		})
	}

	podSpec := pod.Spec(beSpec, src.Spec.ServiceAccount, beContainer, vols)

	now := time.Now().Format(time.RFC3339)
	annotations := pod.Annotations(beSpec, src.Annotations, now)
	podSpec.SecurityContext = pod.PodSecurityContext(beSpec)
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Annotations: annotations,
			Namespace:   src.Namespace,
			Labels:      pod.Labels(src.Name, beSpec),
		},
		Spec: podSpec,
	}
}
