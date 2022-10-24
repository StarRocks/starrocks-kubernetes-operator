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

func fePodLabels(src *srapi.StarRocksCluster, stname string) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = stname
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	labels.AddLabel(src.Labels)
	return labels
}

//buildPodTemplate construct the podTemplate for deploy fe.
func (fc *FeController) buildPodTemplate(src *srapi.StarRocksCluster) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_FE
	feSpec := src.Spec.StarRocksFeSpec

	vols := []corev1.Volume{
		//TODO：cancel the configmap for temporary.
		/*{
			Name: srapi.DEFAULT_FE_CONFIG_NAME,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: srapi.DEFAULT_FE_CONFIG_NAME,
					},
				},
			},
		},*/
		{
			Name: srapi.DEFAULT_EMPTDIR_NAME,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	var volMounts []corev1.VolumeMount
	for _, vm := range feSpec.StorageVolumes {
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      vm.Name,
			MountPath: vm.MountPath,
		}, corev1.VolumeMount{
			Name:      srapi.INITIAL_VOLUME_PATH_NAME,
			MountPath: srapi.INITIAL_VOLUME_PATH,
		})
	}

	opContainers := []corev1.Container{
		{
			Name:  srapi.DEFAULT_FE,
			Image: feSpec.Image,
			//TODO: add start command
			Command: []string{"/opt/starrocks/entrypoint-fe.sh"},
			//TODO: add args
			Args: []string{"$(FE_SERVICE_NAME)"},
			Ports: []corev1.ContainerPort{{
				Name:          "http-port",
				ContainerPort: 8030,
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "rpc-port",
				ContainerPort: 9020,
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "query-port",
				ContainerPort: 9030,
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
					Name:  srapi.SERVICE_NAME,
					Value: fc.GetExternalFeServiceName(src) + "." + src.Namespace,
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
				},
			},

			Resources:       feSpec.ResourceRequirements,
			VolumeMounts:    volMounts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			StartupProbe: &corev1.Probe{
				FailureThreshold: 120,
				PeriodSeconds:    5,
				ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(9030)}},
			},
			ReadinessProbe: &corev1.Probe{
				PeriodSeconds:       5,
				InitialDelaySeconds: 5,
				ProbeHandler:        corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(9020)}},
			},
			LivenessProbe: &corev1.Probe{
				FailureThreshold: 5,
				PeriodSeconds:    5,
				ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(9020)}},
			},
		},
	}

	podSpec := corev1.PodSpec{

		Volumes:                       vols,
		Containers:                    opContainers,
		ServiceAccountName:            src.Spec.ServiceAccount,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(30)),
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Labels:      fePodLabels(src, feStatefulSetName(src)),
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
