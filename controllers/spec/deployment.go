/*
Copyright 2022 StarRocks.

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
package spec

import (
	"strings"

	"github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const CnVolumeName = "cn-volume"

// build a depolyment base on cn
func MakeCnDeployment(cn *v1alpha1.ComputeNodeGroup) *appsv1.Deployment {
	annotation := cn.Annotations
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cn.Name,
			Namespace: cn.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cn, cn.GroupVersionKind()),
			},
			Annotations: annotation,
			Labels:      cn.Labels,
		},
		Spec: makeCnDeploymentSpec(cn),
	}
}

func makeCnDeploymentSpec(cn *v1alpha1.ComputeNodeGroup) appsv1.DeploymentSpec {
	allLabels := cn.Labels

	if allLabels == nil {
		allLabels = make(map[string]string)
	}
	for k, v := range cn.Spec.PodPolicy.Labels {
		allLabels[k] = v
	}

	selectorLabels := map[string]string{
		"cn":           cn.Name,
		"cn-component": "deployment",
	}

	injectLabels(allLabels, selectorLabels)

	return appsv1.DeploymentSpec{
		Replicas: &cn.Spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: selectorLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: cn.Annotations,
				Labels:      allLabels,
			},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					v1.Volume{
						Name: CnVolumeName,
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: cn.Spec.CnInfo.ConfigMap,
								},
							},
						},
					},
				},
				Containers: []v1.Container{
					makeCnContainer(cn),
					makeCnResisterContainer(cn),
				},
				RestartPolicy:    "Always",
				NodeSelector:     cn.Spec.PodPolicy.NodeSelector,
				ImagePullSecrets: cn.Spec.PodPolicy.ImagePullSecrets,
				Affinity:         cn.Spec.PodPolicy.Affinity,
				SecurityContext:  cn.Spec.PodPolicy.SecurityContext,
			},
		},
		Strategy:        appsv1.DeploymentStrategy{},
		MinReadySeconds: 0,
	}
}

func makeCnContainer(cn *v1alpha1.ComputeNodeGroup) v1.Container {
	cmd := []string{"be/bin/start_cn.sh"}
	if cn.Spec.PodPolicy.Command != nil {
		cmd = cn.Spec.PodPolicy.Command
	}
	return v1.Container{
		Name:            "cn-container",
		Image:           cn.Spec.Images.CnImage,
		ImagePullPolicy: "IfNotPresent",
		Resources:       cn.Spec.PodPolicy.Resources,
		Command:         cmd,
		Args:            cn.Spec.PodPolicy.Args,
		StartupProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/api/health",
					Port: intstr.IntOrString{
						IntVal: 8040,
					},
				},
			},
			PeriodSeconds:    10,
			FailureThreshold: 30,
		},
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/health",
					Port: intstr.IntOrString{
						IntVal: 8060,
					},
				},
			},
			PeriodSeconds:    20,
			FailureThreshold: 3,
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      CnVolumeName,
				ReadOnly:  false,
				MountPath: "/data/starrocks/be/conf/cn.conf",
				SubPath:   "cn.conf",
			},
		},
	}
}

func makeCnResisterContainer(cn *v1alpha1.ComputeNodeGroup) v1.Container {
	return v1.Container{
		Name:            "register",
		Image:           cn.Spec.Images.ComponentsImage,
		ImagePullPolicy: "Always",
		Command:         []string{"./register"},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1000m"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1000m"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
			},
		},
		Env: []v1.EnvVar{
			// TODO: solve hard-coded
			v1.EnvVar{
				Name:  common.EnvKeyFeAddrs,
				Value: strings.Join(cn.Spec.FeInfo.Addresses, ","),
			},
			v1.EnvVar{
				Name: common.EnvKeyFeUsr,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: cn.Spec.FeInfo.AccountSecret,
						},
						Key: common.EnvKeyFeUsr,
					},
				},
			},
			v1.EnvVar{
				Name: common.EnvKeyFePwd,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: cn.Spec.FeInfo.AccountSecret,
						},
						Key: common.EnvKeyFePwd,
					},
				},
			},
			v1.EnvVar{
				Name:  common.EnvKeyCnPort,
				Value: common.CnHeartBeatPort,
			},
		},
	}
}

// sync changed
func SyncDeploymentChanged(current, desired *appsv1.Deployment) {
	// k8s would assign a default annotation for deployment
	if deployRevision, ok := current.Annotations["deployment.kubernetes.io/revision"]; ok {
		desired.Annotations["deployment.kubernetes.io/revision"] = deployRevision
	}

	current.Spec.Replicas = desired.Spec.Replicas
	current.Spec.Template.Spec = desired.Spec.Template.Spec
	current.Spec.Template.Labels = desired.Spec.Template.Labels
	current.Labels = desired.Labels
	current.Annotations = desired.Annotations
}

func injectLabels(src, fixed map[string]string) {
	if src == nil {
		src = make(map[string]string)
	}
	for k, v := range fixed {
		src[k] = v
	}
}
