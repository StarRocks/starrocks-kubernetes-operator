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
	"time"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	meta_path          = "/opt/starrocks/fe/meta"
	meta_name          = "fe-meta"
	log_path           = "/opt/starrocks/fe/log"
	log_name           = "fe-log"
	fe_config_path     = "/etc/starrocks/fe/conf"
	env_fe_config_path = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy fe.
func (fc *FeController) buildPodTemplate(src *srapi.StarRocksCluster, config map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_FE
	feSpec := src.Spec.StarRocksFeSpec

	vols, volMounts, vexist := pod.MountStorageVolumes(feSpec)
	// add default volume about log ,meta if not configure.
	if _, ok := vexist[meta_path]; !ok {
		volMounts = append(
			volMounts, corev1.VolumeMount{
				Name:      meta_name,
				MountPath: meta_path,
			})
		vols = append(vols, corev1.Volume{
			Name: meta_name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if _, ok := vexist[log_path]; !ok {
		volMounts = append(volMounts, corev1.VolumeMount{
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

	// mount configmap, secrets to pod if needed
	vols, volMounts = pod.MountConfigMapInfo(vols, volMounts, feSpec.ConfigMapInfo, fe_config_path)
	vols, volMounts = pod.MountConfigMaps(vols, volMounts, feSpec.ConfigMaps)
	vols, volMounts = pod.MountSecrets(vols, volMounts, feSpec.Secrets)

	feExternalServiceName := srapi.GetExternalServiceName(src.Name, feSpec)
	Envs := pod.Envs(src.Spec.StarRocksFeSpec, config, feExternalServiceName, src.Namespace, feSpec.FeEnvVars)
	feContainer := corev1.Container{
		Name:            srapi.DEFAULT_FE,
		Image:           feSpec.Image,
		Command:         []string{"/opt/starrocks/fe_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(feSpec, config),
		Env:             Envs,
		Resources:       feSpec.ResourceRequirements,
		VolumeMounts:    volMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		StartupProbe:    pod.StartupProbe(rutils.GetPort(config, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(rutils.GetPort(config, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(rutils.GetPort(config, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle("/opt/starrocks/fe_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(feSpec),
	}

	if feSpec.ConfigMapInfo.ConfigMapName != "" && feSpec.ConfigMapInfo.ResolveKey != "" {
		feContainer.Env = append(feContainer.Env, corev1.EnvVar{
			Name:  env_fe_config_path,
			Value: fe_config_path,
		})
	}

	podSpec := pod.Spec(feSpec, src.Spec.ServiceAccount, feContainer, vols)
	now := time.Now().Format(time.RFC3339)
	annotations := pod.Annotations(feSpec, src.Annotations, now)
	podSpec.SecurityContext = pod.PodSecurityContext(feSpec)
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Annotations: annotations,
			Labels:      pod.Labels(src.Name, src.Spec.StarRocksFeSpec),
		},
		Spec: podSpec,
	}
}
