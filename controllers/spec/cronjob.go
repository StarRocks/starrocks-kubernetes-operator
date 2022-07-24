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
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// build a cron job base on cn
func MakeCnCronJob(cn *v1alpha1.ComputeNodeGroup) *batchv1beta1.CronJob {
	allLabels := cn.Labels

	if allLabels == nil {
		allLabels = make(map[string]string)
	}
	for k, v := range cn.Spec.PodPolicy.Labels {
		allLabels[k] = v
	}

	selectorLabels := map[string]string{
		"cn":           cn.Name,
		"cn-component": "cronjob",
	}
	injectLabels(allLabels, selectorLabels)

	return &batchv1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1beta1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cn.Namespace,
			Name:      cn.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cn, cn.GroupVersionKind()),
			},
			Labels: allLabels,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          cn.Spec.CronJobPolicy.Schedule,
			ConcurrencyPolicy: "Replace",
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: allLabels,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: allLabels,
						},

						Spec: corev1.PodSpec{
							ServiceAccountName: cn.Name,
							ImagePullSecrets:   cn.Spec.PodPolicy.ImagePullSecrets,
							Containers: []corev1.Container{
								corev1.Container{
									Name:            "job",
									Image:           cn.Spec.Images.ComponentsImage,
									ImagePullPolicy: "Always",
									Command:         []string{"./offline"},
									Env: []corev1.EnvVar{
										// TODO: solve hard-coded
										corev1.EnvVar{
											Name:  common.EnvKeyFeAddrs,
											Value: strings.Join(cn.Spec.FeInfo.Addresses, ","),
										},
										corev1.EnvVar{
											Name: common.EnvKeyFeUsr,
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: cn.Spec.FeInfo.AccountSecret,
													},
													Key: common.EnvKeyFeUsr,
												},
											},
										},
										corev1.EnvVar{
											Name: common.EnvKeyFePwd,
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: cn.Spec.FeInfo.AccountSecret,
													},
													Key: common.EnvKeyFePwd,
												},
											},
										},
										corev1.EnvVar{
											Name:  common.EnvKeyCnNs,
											Value: cn.Namespace,
										},
										corev1.EnvVar{
											Name:  common.EnvKeyCnName,
											Value: cn.Name,
										},
										corev1.EnvVar{
											Name:  common.EnvKeyCnPort,
											Value: common.CnHeartBeatPort,
										},
									},
								},
							},
							RestartPolicy: "OnFailure",
						},
					},
				},
			},
		},
	}
}

// sync changed
func SyncCronJobChanged(current, desired *batchv1beta1.CronJob) {
	current.Spec = desired.Spec
	current.Labels = desired.Labels
	current.Annotations = desired.Annotations
}
