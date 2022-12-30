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

package cn_controller

import (
	v1alpha12 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	log_path           = "/opt/starrocks/cn/log"
	log_name           = "cn-log"
	cn_config_path     = "/etc/starrocks/cn/conf"
	env_cn_config_path = "CONFIGMAP_MOUNT_PATH"
)

//cnPodLabels
func (cc *CnController) cnPodLabels(src *v1alpha12.StarRocksCluster, ownerReferenceName string) rutils.Labels {
	labels := rutils.Labels{}
	labels[v1alpha12.OwnerReference] = ownerReferenceName
	labels[v1alpha12.ComponentLabelKey] = v1alpha12.DEFAULT_CN
	labels.AddLabel(src.Labels)
	return labels
}

//buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(src *v1alpha12.StarRocksCluster, cnconfig map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + v1alpha12.DEFAULT_CN
	cnSpec := src.Spec.StarRocksCnSpec

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

	if cnSpec.ConfigMapInfo.ConfigMapName != "" && cnSpec.ConfigMapInfo.ResolveKey != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      cnSpec.ConfigMapInfo.ConfigMapName,
			MountPath: cn_config_path,
		})

		vols = append(vols, corev1.Volume{
			Name: cnSpec.ConfigMapInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cnSpec.ConfigMapInfo.ConfigMapName,
					},
				},
			},
		})

	}

	cnContainer := corev1.Container{
		Name:    v1alpha12.DEFAULT_CN,
		Image:   cnSpec.Image,
		Command: []string{"/opt/starrocks/cn_entrypoint.sh"},
		Args:    []string{"$(FE_SERVICE_NAME)"},
		Ports: []corev1.ContainerPort{
			{
				Name:          "thrift-port",
				ContainerPort: rutils.GetPort(cnconfig, rutils.THRIFT_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "webserver-port",
				ContainerPort: rutils.GetPort(cnconfig, rutils.WEBSERVER_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "heartbeat-port",
				ContainerPort: rutils.GetPort(cnconfig, rutils.HEARTBEAT_SERVICE_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "brpc-port",
				ContainerPort: rutils.GetPort(cnconfig, rutils.BRPC_PORT),
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
				Name:  v1alpha12.COMPONENT_NAME,
				Value: v1alpha12.DEFAULT_CN,
			}, {
				Name:  v1alpha12.FE_SERVICE_NAME,
				Value: v1alpha12.GetFeExternalServiceName(src),
			}, {
				Name:  "FE_QUERY_PORT",
				Value: strconv.FormatInt(int64(rutils.GetPort(cnconfig, rutils.QUERY_PORT)), 10),
			}, {
				Name:  "HOST_TYPE",
				Value: "FQDN",
			}, {
				Name:  "USER",
				Value: "root",
			},
		},
		Resources:       cnSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe: &corev1.Probe{
			//TODO: default 5min, user can config.
			FailureThreshold: 60,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(cnconfig, rutils.BRPC_PORT),
			}}},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold: 60,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(cnconfig, rutils.HEARTBEAT_SERVICE_PORT),
			}}},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(cnconfig, rutils.THRIFT_PORT),
			}}},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/opt/starrocks/cn_prestop.sh"},
				},
			},
		},
	}

	if cnSpec.ConfigMapInfo.ConfigMapName != "" && cnSpec.ConfigMapInfo.ResolveKey != "" {
		cnContainer.Env = append(cnContainer.Env, corev1.EnvVar{
			Name:  env_cn_config_path,
			Value: cn_config_path,
		})
	}

	podSpec := corev1.PodSpec{
		Containers:                    []corev1.Container{cnContainer},
		Volumes:                       vols,
		ServiceAccountName:            src.Spec.ServiceAccount,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metaname,
			Namespace: src.Namespace,
			Labels:    cc.cnPodLabels(src, cnStatefulSetName(src)),
			//Annotations: src.Annotations,
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
