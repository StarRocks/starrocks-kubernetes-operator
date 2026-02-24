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

package feobserver

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
	_metaName             = "fe-meta"
	_logName              = "fe-log"
	_feConfigMountPath    = "/etc/starrocks/fe/conf"
	_envFeConfigMountPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy fe observer.
func (fc *FeObserverController) buildPodTemplate(src *srapi.StarRocksCluster,
	config map[string]interface{}) (*corev1.PodTemplateSpec, error) {
	metaName := src.Name + "-" + srapi.DEFAULT_FE_OBSERVER
	observerSpec := src.Spec.StarRocksFeObserverSpec

	vols, volMounts := pod.MountStorageVolumes(observerSpec)
	// add default volume about log, meta if not configure.
	if !k8sutils.HasVolume(vols, _metaName) && !k8sutils.HasMountPath(volMounts, pod.GetStorageDir(observerSpec)) {
		vols, volMounts = pod.MountEmptyDirVolume(vols, volMounts, _metaName, pod.GetStorageDir(observerSpec), "")
	}
	if !k8sutils.HasVolume(vols, _logName) && !k8sutils.HasMountPath(volMounts, pod.GetLogDir(observerSpec)) {
		vols, volMounts = pod.MountEmptyDirVolume(vols, volMounts, _logName, pod.GetLogDir(observerSpec), "")
	}

	// mount configmap, secrets to pod if needed
	vols, volMounts = pod.MountConfigMapInfo(vols, volMounts, observerSpec.ConfigMapInfo, _feConfigMountPath)
	vols, volMounts = pod.MountConfigMaps(observerSpec, vols, volMounts, observerSpec.ConfigMaps)
	vols, volMounts = pod.MountSecrets(vols, volMounts, observerSpec.Secrets)
	if err := k8sutils.CheckVolumes(vols, volMounts); err != nil {
		return nil, err
	}

	feServiceName := service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	envs := pod.Envs(observerSpec, config, feServiceName, src.Namespace, observerSpec.FeEnvVars)
	httpPort := rutils.GetPort(config, rutils.HTTP_PORT)
	feObserverContainer := corev1.Container{
		Name:            srapi.DEFAULT_FE_OBSERVER,
		Image:           observerSpec.Image,
		Command:         pod.ContainerCommand(observerSpec),
		Args:            pod.ContainerArgs(observerSpec),
		Ports:           pod.Ports(observerSpec, config),
		Env:             envs,
		Resources:       observerSpec.ResourceRequirements,
		VolumeMounts:    volMounts,
		ImagePullPolicy: observerSpec.GetImagePullPolicy(),
		StartupProbe:    pod.StartupProbe(observerSpec.GetStartupProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(observerSpec.GetLivenessProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(observerSpec.GetReadinessProbeFailureSeconds(), httpPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle(observerSpec.GetLifecycle(), pod.GetPreStopCommand(observerSpec)),
		SecurityContext: pod.ContainerSecurityContext(observerSpec),
	}
	if pod.GetStarRocksRootPath(observerSpec.FeEnvVars) != pod.GetStarRocksDefaultRootPath() {
		feObserverContainer.WorkingDir = pod.GetStarRocksRootPath(observerSpec.FeEnvVars)
	}

	if observerSpec.ConfigMapInfo.ConfigMapName != "" && observerSpec.ConfigMapInfo.ResolveKey != "" {
		feObserverContainer.Env = append(feObserverContainer.Env, corev1.EnvVar{
			Name:  _envFeConfigMountPath,
			Value: _feConfigMountPath,
		})
	}

	podSpec := pod.Spec(observerSpec, feObserverContainer, vols)
	annotations := pod.Annotations(observerSpec)
	podSpec.SecurityContext = pod.PodSecurityContext(observerSpec)
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaName,
			Namespace:   src.Namespace,
			Annotations: annotations,
			Labels:      pod.Labels(src.Name, observerSpec),
		},
		Spec: podSpec,
	}, nil
}
