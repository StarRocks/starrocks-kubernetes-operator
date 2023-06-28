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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	meta_path          = "/opt/starrocks/fe/meta"
	meta_name          = "fe-meta"
	log_path           = "/opt/starrocks/fe/log"
	log_name           = "fe-log"
	fe_config_path     = "/etc/starrocks/fe/conf"
	env_fe_config_path = "CONFIGMAP_MOUNT_PATH"
)

// fePodLabels generate the fe pod labels and statefulset selector
func (fc *FeController) fePodLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := statefulset.MakeSelector(src.Name, src.Spec.StarRocksFeSpec)
	//podLabels for classify. operator use statefulsetSelector for manage pods.
	if src.Spec.StarRocksFeSpec != nil {
		labels.AddLabel(src.Spec.StarRocksFeSpec.PodLabels)
	}
	return labels
}

// buildPodTemplate construct the podTemplate for deploy fe.
func (fc *FeController) buildPodTemplate(src *srapi.StarRocksCluster, feconfig map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_FE
	feSpec := src.Spec.StarRocksFeSpec

	vexist := make(map[string]bool)
	var volMounts []corev1.VolumeMount
	var vols []corev1.Volume
	for _, sv := range feSpec.StorageVolumes {
		vexist[sv.MountPath] = true
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      sv.Name,
			MountPath: sv.MountPath,
		})

		//TODO: now only support storage class mode.
		vols = append(vols, corev1.Volume{
			Name: sv.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: sv.Name,
				},
			},
		})
	}

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
	vols, volMounts = pod.MountConfigMap(vols, volMounts, feSpec.ConfigMapInfo, fe_config_path)
	vols, volMounts = pod.MountSecrets(vols, volMounts, feSpec.Secrets)

	Envs := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		}, {
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
			},
		}, {
			Name:  srapi.COMPONENT_NAME,
			Value: srapi.DEFAULT_FE,
		}, {
			Name:  srapi.FE_SERVICE_NAME,
			Value: srapi.GetFeExternalServiceName(src) + "." + src.Namespace,
		}, {
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
			},
		}, {
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.hostIP"},
			},
		}, {
			Name:  "HOST_TYPE",
			Value: "FQDN",
		}, {
			Name:  "USER",
			Value: "root",
		},
	}

	Envs = append(Envs, feSpec.FeEnvVars...)
	feContainer := corev1.Container{
		Name:    srapi.DEFAULT_FE,
		Image:   feSpec.Image,
		Command: []string{"/opt/starrocks/fe_entrypoint.sh"},
		Args:    []string{"$(FE_SERVICE_NAME)"},
		Ports: []corev1.ContainerPort{{
			Name:          "http-port",
			ContainerPort: rutils.GetPort(feconfig, rutils.HTTP_PORT),
			Protocol:      corev1.ProtocolTCP,
		}, {
			Name:          "rpc-port",
			ContainerPort: rutils.GetPort(feconfig, rutils.RPC_PORT),
			Protocol:      corev1.ProtocolTCP,
		}, {
			Name:          "query-port",
			ContainerPort: rutils.GetPort(feconfig, rutils.QUERY_PORT),
			Protocol:      corev1.ProtocolTCP,
		}},

		Env:             Envs,
		Resources:       feSpec.ResourceRequirements,
		VolumeMounts:    volMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		StartupProbe:    pod.MakeStartupProbe(rutils.GetPort(feconfig, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		LivenessProbe:   pod.MakeLivenessProbe(rutils.GetPort(feconfig, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.MakeReadinessProbe(rutils.GetPort(feconfig, rutils.HTTP_PORT), pod.HEALTH_API_PATH),
		Lifecycle:       pod.MakeLifeCycle("/opt/starrocks/fe_prestop.sh"),
	}

	if feSpec.ConfigMapInfo.ConfigMapName != "" && feSpec.ConfigMapInfo.ResolveKey != "" {
		feContainer.Env = append(feContainer.Env, corev1.EnvVar{
			Name:  env_fe_config_path,
			Value: fe_config_path,
		})
	}

	sa := feSpec.ServiceAccount
	if feSpec.ServiceAccount != "" {
		sa = feSpec.ServiceAccount
	}

	podSpec := corev1.PodSpec{
		Volumes:                       vols,
		Containers:                    []corev1.Container{feContainer},
		ServiceAccountName:            sa,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
		Affinity:                      feSpec.Affinity,
		ImagePullSecrets:              feSpec.ImagePullSecrets,
		Tolerations:                   feSpec.Tolerations,
		NodeSelector:                  feSpec.NodeSelector,
		HostAliases:                   feSpec.HostAliases,
		SchedulerName:                 feSpec.SchedulerName,
	}

	annos := make(map[string]string)
	//add restart
	if _, ok := src.Annotations[string(srapi.AnnotationFERestartKey)]; ok {
		annos[common.KubectlRestartAnnotationKey] = time.Now().Format(time.RFC3339)
	}
	rutils.Annotations(annos).AddAnnotation(feSpec.Annotations)

	onrootMismatch := corev1.FSGroupChangeOnRootMismatch
	if feSpec.FsGroup == nil {
		sc := &corev1.PodSecurityContext{

			FSGroup:             rutils.GetInt64ptr(common.DefaultFsGroup),
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	} else if *feSpec.FsGroup != 0 {
		sc := &corev1.PodSecurityContext{
			FSGroup:             feSpec.FsGroup,
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Annotations: annos,
			Labels:      fc.fePodLabels(src),
		},
		Spec: podSpec,
	}

}
