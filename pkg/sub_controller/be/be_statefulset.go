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
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildStatefulSetParams generate the params of construct the statefulset.
func (be *BeController) buildStatefulSetParams(src *srapi.StarRocksCluster, beconfig map[string]interface{}, internalServiceName string) rutils.StatefulSetParams {
	beSpec := src.Spec.StarRocksBeSpec

	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := be.buildPodTemplate(src, beconfig)
	pvcs := statefulset.MakePVCList(beSpec.StorageVolumes)
	annos := rutils.Annotations{}
	// add restart annotation on statefulset.
	if _, ok := src.Annotations[string(srapi.AnnotationBERestartKey)]; ok {
		annos.Add(string(srapi.AnnotationBERestartKey), string(srapi.AnnotationRestart))
	}

	return rutils.StatefulSetParams{
		Name:                 statefulset.MakeName(src.Name, beSpec),
		Namespace:            src.Namespace,
		Annotations:          annos,
		Labels:               statefulset.MakeLabels(src.Name, beSpec),
		OwnerReferences:      []metav1.OwnerReference{*or},
		Replicas:             beSpec.Replicas,
		Selector:             statefulset.MakeSelector(src.Name, beSpec),
		PodTemplateSpec:      podTemplateSpec,
		ServiceName:          internalServiceName,
		VolumeClaimTemplates: pvcs,
	}
}
