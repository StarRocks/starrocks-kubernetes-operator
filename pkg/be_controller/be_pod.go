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

package be_controller

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	log_path           = "/opt/starrocks/be/log"
	log_name           = "be-log"
	be_config_path     = "/etc/starrocks/be/conf"
	env_be_config_path = "CONFIGMAP_MOUNT_PATH"
)

//bePodLabels
func (cc *BeController) bePodLabels(src *srapi.StarRocksCluster, ownerReferenceName string) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = ownerReferenceName
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	labels.AddLabel(src.Labels)
	return labels
}

//buildPodTemplate construct the podTemplate for deploy be.
func (cc *BeController) buildPodTemplate(src *srapi.StarRocksCluster, beConfig map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_BE
	beSpec := src.Spec.StarRocksBeSpec

	//generate the default emptydir for log.
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      log_name,
			MountPath: log_path,
		},
	}

	vols := []corev1.Volume{
		{
			Name: log_name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	for _, sv := range beSpec.StorageVolumes {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
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

	if beSpec.ConfigMapInfo.ConfigMapName != "" && beSpec.ConfigMapInfo.ResolveKey != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      beSpec.ConfigMapInfo.ConfigMapName,
			MountPath: be_config_path,
		})

		vols = append(vols, corev1.Volume{
			Name: beSpec.ConfigMapInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: beSpec.ConfigMapInfo.ConfigMapName,
					},
				},
			},
		})

	}

	beContainer := corev1.Container{
		Name:    srapi.DEFAULT_BE,
		Image:   beSpec.Image,
		Command: []string{"/opt/starrocks/be_entrypoint.sh"},
		Args:    []string{"$(FE_SERVICE_NAME)"},
		Ports: []corev1.ContainerPort{
			{
				Name:          "thrift-port",
				ContainerPort: rutils.GetPort(beConfig, rutils.THRIFT_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "webserver-port",
				ContainerPort: rutils.GetPort(beConfig, rutils.WEBSERVER_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "heartbeat-port",
				ContainerPort: rutils.GetPort(beConfig, rutils.HEARTBEAT_SERVICE_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "brpc-port",
				ContainerPort: rutils.GetPort(beConfig, rutils.BRPC_PORT),
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
				Value: srapi.GetFeExternalServiceName(src),
			}, {
				Name:  "FE_QUERY_PORT",
				Value: strconv.FormatInt(int64(rutils.GetPort(beConfig, rutils.QUERY_PORT)), 10),
			}, {
				Name:  "HOST_TYPE",
				Value: "FQDN",
			}, {
				Name:  "USER",
				Value: "root",
			},
		},
		Resources:       beSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe: &corev1.Probe{
			FailureThreshold: 30,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beConfig, rutils.BRPC_PORT),
			}}},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold: 30,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beConfig, rutils.HEARTBEAT_SERVICE_PORT),
			}}},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beConfig, rutils.THRIFT_PORT),
			}}},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/opt/starrocks/be_prestop.sh"},
				},
			},
		},
	}

	if beSpec.ConfigMapInfo.ConfigMapName != "" && beSpec.ConfigMapInfo.ResolveKey != "" {
		beContainer.Env = append(beContainer.Env, corev1.EnvVar{
			Name:  env_be_config_path,
			Value: be_config_path,
		})
	}

	podSpec := corev1.PodSpec{
		Containers:                    []corev1.Container{beContainer},
		Volumes:                       vols,
		ServiceAccountName:            src.Spec.ServiceAccount,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Labels:      cc.bePodLabels(src, beStatefulSetName(src)),
			Annotations: src.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: src.APIVersion,
					Kind:       src.Kind,
					Name:       src.Name,
				},
			},
		},
		Spec: podSpec,
	}
}
