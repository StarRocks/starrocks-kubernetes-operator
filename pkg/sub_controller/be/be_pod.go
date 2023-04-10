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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
	"time"
)

const (
	log_path           = "/opt/starrocks/be/log"
	log_name           = "be-log"
	be_config_path     = "/etc/starrocks/be/conf"
	storage_name       = "be-storage"
	storage_path       = "/opt/starrocks/be/storage"
	env_be_config_path = "CONFIGMAP_MOUNT_PATH"
)

//bePodLabels
func (be *BeController) bePodLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := be.beStatefulsetSelector(src)
	//podLables for classify. operator use statefulsetSelector for manage pods.
	if src.Spec.StarRocksBeSpec != nil {
		labels.AddLabel(src.Spec.StarRocksBeSpec.PodLabels)
	}

	return labels
}

//buildPodTemplate construct the podTemplate for deploy cn.
func (be *BeController) buildPodTemplate(src *srapi.StarRocksCluster, beconfig map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_BE
	beSpec := src.Spec.StarRocksBeSpec

	vexist := make(map[string]bool)
	var volumeMounts []corev1.VolumeMount
	var vols []corev1.Volume
	for _, sv := range beSpec.StorageVolumes {
		vexist[sv.MountPath] = true
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      sv.Name,
			MountPath: sv.MountPath,
		})

		vols = append(vols, corev1.Volume{
			Name: sv.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: sv.Name,
				},
			},
		})
	}

	// add default volume about log, if meta not configure.
	if _, ok := vexist[log_path]; !ok {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			//use storage volume.
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
			//use storage volume.
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
			Value: srapi.DEFAULT_CN,
		}, {
			Name:  srapi.FE_SERVICE_NAME,
			Value: srapi.GetFeExternalServiceName(src),
		}, {
			Name:  "FE_QUERY_PORT",
			Value: strconv.FormatInt(int64(rutils.GetPort(beconfig, rutils.QUERY_PORT)), 10),
		}, {
			Name:  "HOST_TYPE",
			Value: "FQDN",
		}, {
			Name:  "USER",
			Value: "root",
		},
	}

	Envs = append(Envs, beSpec.BeEnvVars...)
	beContainer := corev1.Container{
		Name:    srapi.DEFAULT_BE,
		Image:   beSpec.Image,
		Command: []string{"/opt/starrocks/be_entrypoint.sh"},
		Args:    []string{"$(FE_SERVICE_NAME)"},
		Ports: []corev1.ContainerPort{
			{
				Name:          "be-port",
				ContainerPort: rutils.GetPort(beconfig, rutils.BE_PORT),
			}, {
				Name:          "webserver-port",
				ContainerPort: rutils.GetPort(beconfig, rutils.WEBSERVER_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "heartbeat-port",
				ContainerPort: rutils.GetPort(beconfig, rutils.HEARTBEAT_SERVICE_PORT),
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "brpc-port",
				ContainerPort: rutils.GetPort(beconfig, rutils.BRPC_PORT),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env:             Envs,
		Resources:       beSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe: &corev1.Probe{
			FailureThreshold: 60,
			PeriodSeconds:    5,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beconfig, rutils.BRPC_PORT),
			}}},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds:    5,
			FailureThreshold: 3,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beconfig, rutils.HEARTBEAT_SERVICE_PORT),
			}}},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds:    5,
			FailureThreshold: 3,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(beconfig, rutils.THRIFT_PORT),
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

	sa := src.Spec.ServiceAccount
	if beSpec.ServiceAccount != "" {
		sa = beSpec.ServiceAccount
	}

	podSpec := corev1.PodSpec{
		Containers:                    []corev1.Container{beContainer},
		Volumes:                       vols,
		ServiceAccountName:            sa,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
		Affinity:                      beSpec.Affinity,
		Tolerations:                   beSpec.Tolerations,
		ImagePullSecrets:              beSpec.ImagePullSecrets,
		NodeSelector:                  beSpec.NodeSelector,
		HostAliases:                   beSpec.HostAliases,
	}

	annos := make(map[string]string)
	//add restart annotation in podTemplate.
	if _, ok := src.Annotations[string(srapi.AnnotationBERestartKey)]; ok {
		//simulate the kubectl operation.
		annos[common.KubectlRestartAnnotationKey] = time.Now().Format(time.RFC3339)
	}

	rutils.Annotations(annos).AddAnnotation(beSpec.Annotations)

	onrootMismatch := corev1.FSGroupChangeOnRootMismatch
	if beSpec.FsGroup == nil {
		sc := &corev1.PodSecurityContext{
			FSGroup:             rutils.GetInt64ptr(common.DefaultFsGroup),
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	} else if *beSpec.FsGroup != 0 {
		sc := &corev1.PodSecurityContext{
			FSGroup:             beSpec.FsGroup,
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Annotations: annos,
			Namespace:   src.Namespace,
			Labels:      be.bePodLabels(src),
		},
		Spec: podSpec,
	}
}
