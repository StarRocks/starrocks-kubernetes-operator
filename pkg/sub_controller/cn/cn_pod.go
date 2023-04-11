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
	log_path           = "/opt/starrocks/cn/log"
	log_name           = "cn-log"
	cn_config_path     = "/etc/starrocks/cn/conf"
	env_cn_config_path = "CONFIGMAP_MOUNT_PATH"
)

//buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(src *srapi.StarRocksCluster, cnconfig map[string]interface{}) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_CN
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
			Value: strconv.FormatInt(int64(rutils.GetPort(cnconfig, rutils.QUERY_PORT)), 10),
		}, {
			Name:  "HOST_TYPE",
			Value: "FQDN",
		}, {
			Name:  "USER",
			Value: "root",
		},
	}

	Envs = append(Envs, cnSpec.CnEnvVars...)
	cnContainer := corev1.Container{
		Name:    srapi.DEFAULT_CN,
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
		Env:             Envs,
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
			PeriodSeconds:    5,
			FailureThreshold: 3,
			ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: rutils.GetPort(cnconfig, rutils.HEARTBEAT_SERVICE_PORT),
			}}},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds:    5,
			FailureThreshold: 3,
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

	sa := src.Spec.ServiceAccount
	if cnSpec.ServiceAccount != "" {
		sa = cnSpec.ServiceAccount
	}

	podSpec := corev1.PodSpec{
		Containers:                    []corev1.Container{cnContainer},
		Volumes:                       vols,
		ServiceAccountName:            sa,
		TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
		Affinity:                      cnSpec.Affinity,
		Tolerations:                   cnSpec.Tolerations,
		ImagePullSecrets:              cnSpec.ImagePullSecrets,
		NodeSelector:                  cnSpec.NodeSelector,
		HostAliases:                   cnSpec.HostAliases,
	}
	annos := make(map[string]string)
	//add restart
	if _, ok := src.Annotations[string(srapi.AnnotationCNRestartKey)]; ok {
		annos[common.KubectlRestartAnnotationKey] = time.Now().Format(time.RFC3339)
	}
	//add annotations for cn pods.
	rutils.Annotations(annos).AddAnnotation(cnSpec.Annotations)

	onrootMismatch := corev1.FSGroupChangeOnRootMismatch
	if cnSpec.FsGroup == nil {
		sc := &corev1.PodSecurityContext{
			FSGroup:             rutils.GetInt64ptr(common.DefaultFsGroup),
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	} else if *cnSpec.FsGroup != 0 {
		sc := &corev1.PodSecurityContext{
			RunAsUser:           cnSpec.FsGroup,
			FSGroupChangePolicy: &onrootMismatch,
		}
		podSpec.SecurityContext = sc
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Annotations: annos,
			Namespace:   src.Namespace,
			Labels:      cc.cnPodLabels(src),
		},
		Spec: podSpec,
	}
}

func (cc *CnController) cnPodLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := cc.cnStatefulsetSelector(src)
	//podLables for classify. operator use statefulsetSelector for manage pods.
	if src.Spec.StarRocksCnSpec != nil {
		labels.AddLabel(src.Spec.StarRocksCnSpec.PodLabels)
	}
	return labels
}
