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

package fe_controller

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	meta_path          = "/opt/starrocks/fe/meta"
	meta_name          = "fe-meta"
	log_path           = "/opt/starrocks/fe/log"
	log_name           = "fe-log"
	fe_config_path     = "/etc/starrocks/fe/conf"
	env_fe_config_path = "CONFIGMAP_MOUNT_PATH"
)

//fePodLabels generate the fe pod labels and statefulset selector
func fePodLabels(src *srapi.StarRocksCluster, ownerReferenceName string) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = ownerReferenceName
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	labels.AddLabel(src.Labels)
	return labels
}

//buildPodTemplate construct the podTemplate for deploy fe.
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

	if feSpec.ConfigMapInfo.ConfigMapName != "" && feSpec.ConfigMapInfo.ResolveKey != "" {
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      feSpec.ConfigMapInfo.ConfigMapName,
			MountPath: fe_config_path,
		})
		vols = append(vols, corev1.Volume{
			Name: feSpec.ConfigMapInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: feSpec.ConfigMapInfo.ConfigMapName,
					},
				},
			},
		})
	}

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
		},
		},
		Env: []corev1.EnvVar{
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
		},

		Resources:       feSpec.ResourceRequirements,
		VolumeMounts:    volMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		StartupProbe: &corev1.Probe{
			FailureThreshold: 60,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(feconfig, rutils.HTTP_PORT),
			}}},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(feconfig, rutils.QUERY_PORT),
			}}},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold: 60,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(feconfig, rutils.RPC_PORT),
			}}},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/opt/starrocks/fe_prestop.sh"},
				},
			},
		},
	}

	if feSpec.ConfigMapInfo.ConfigMapName != "" && feSpec.ConfigMapInfo.ResolveKey != "" {
		feContainer.Env = append(feContainer.Env, corev1.EnvVar{
			Name:  env_fe_config_path,
			Value: fe_config_path,
		})
	}

	podSpec := corev1.PodSpec{
		Volumes:                       vols,
		Containers:                    []corev1.Container{feContainer},
		ServiceAccountName:            src.Spec.ServiceAccount,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Labels:      fePodLabels(src, feStatefulSetName(src)),
			Annotations: src.Annotations,
			/*OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: src.APIVersion,
					Kind:       src.Kind,
					Name:       src.Name,
					UID:        src.UID,
				},
			},*/
		},
		Spec: podSpec,
	}

}
