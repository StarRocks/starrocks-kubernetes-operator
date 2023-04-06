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
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//buildStatefulSetParams generate the params of construct the statefulset.
func (fc *FeController) buildStatefulSetParams(src *srapi.StarRocksCluster, feconfig map[string]interface{}, internalServiceName string) rutils.StatefulSetParams {
	feSpec := src.Spec.StarRocksFeSpec
	var pvcs []corev1.PersistentVolumeClaim
	for _, vm := range feSpec.StorageVolumes {
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

	stname := srapi.FeStatefulSetName(src)
	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := fc.buildPodTemplate(src, feconfig)
	annos := rutils.Annotations{}
	// add restart annotation on statefulset.
	if _, ok := src.Annotations[string(srapi.AnnotationFERestartKey)]; ok {
		annos.Add(string(srapi.AnnotationFERestartKey), string(srapi.AnnotationRestart))
	}

	return rutils.StatefulSetParams{
		Name:                 stname,
		Namespace:            src.Namespace,
		Replicas:             feSpec.Replicas,
		Annotations:          annos,
		VolumeClaimTemplates: pvcs,
		ServiceName:          internalServiceName,
		Labels:               fc.feStatefulSetsLabels(src),
		PodTemplateSpec:      podTemplateSpec,
		Selector:             fc.feStatefulsetSelector(src),
		OwnerReferences:      []metav1.OwnerReference{*or},
	}
}

//try not to modify the labels, as the statefulset don't alloy do it.
func (fc *FeController) feStatefulSetsLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	//once the labels updated, the statefulset will enter into a not reconcile state.
	//labels.AddLabel(src.Labels)
	return labels
}

func (fc *FeController) feStatefulsetSelector(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = srapi.FeStatefulSetName(src)
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	return labels
}
