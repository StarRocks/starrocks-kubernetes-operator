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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//buildStatefulSetParams generate the params of construct the statefulset.
func (be *BeController) buildStatefulSetParams(src *srapi.StarRocksCluster, beconfig map[string]interface{}, internalServiceName string) rutils.StatefulSetParams {
	beSpec := src.Spec.StarRocksBeSpec

	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := be.buildPodTemplate(src, beconfig)
	var pvcs []corev1.PersistentVolumeClaim
	for _, vm := range beSpec.StorageVolumes {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: vm.Name},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: vm.StorageClassName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(vm.StorageSize),
					},
				},
			},
		})
	}

	annos := rutils.Annotations{}
	// add restart annotation on statefulset.
	if _, ok := src.Annotations[string(srapi.AnnotationBERestartKey)]; ok {
		annos.Add(string(srapi.AnnotationBERestartKey), string(srapi.AnnotationRestart))
	}

	return rutils.StatefulSetParams{
		Name:                 srapi.BeStatefulSetName(src),
		Namespace:            src.Namespace,
		Annotations:          annos,
		ServiceName:          internalServiceName,
		PodTemplateSpec:      podTemplateSpec,
		Labels:               be.beStatefulSetsLabels(src),
		Selector:             be.bePodLabels(src),
		OwnerReferences:      []metav1.OwnerReference{*or},
		Replicas:             beSpec.Replicas,
		VolumeClaimTemplates: pvcs,
	}
}

func (be *BeController) beStatefulSetsLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	//once the src labels updated, the statefulset will enter into a can't be modified state.
	//labels.AddLabel(src.Labels)
	return labels
}

func (be *BeController) beStatefulsetSelector(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = srapi.BeStatefulSetName(src)
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	return labels
}
