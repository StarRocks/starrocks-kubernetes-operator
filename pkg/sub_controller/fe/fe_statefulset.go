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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildStatefulSetParams generate the params of construct the statefulset.
func (fc *FeController) buildStatefulSetParams(src *srapi.StarRocksCluster, feconfig map[string]interface{}, internalServiceName string) rutils.StatefulSetParams {
	feSpec := src.Spec.StarRocksFeSpec
	pvcs := statefulset.MakePVCList(feSpec.StorageVolumes)
	statefulSetName := statefulset.MakeName(src.Name, src.Spec.StarRocksFeSpec)
	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := fc.buildPodTemplate(src, feconfig)
	// add restart annotation on statefulset.
	annotations := rutils.Annotations{}
	if _, ok := src.Annotations[string(srapi.AnnotationFERestartKey)]; ok {
		annotations.Add(string(srapi.AnnotationFERestartKey), string(srapi.AnnotationRestart))
	}

	return rutils.StatefulSetParams{
		Name:                 statefulSetName,
		Namespace:            src.Namespace,
		Annotations:          annotations,
		Labels:               statefulset.MakeLabels(src.Name, src.Spec.StarRocksFeSpec),
		OwnerReferences:      []metav1.OwnerReference{*or},
		Replicas:             feSpec.Replicas,
		Selector:             statefulset.MakeSelector(src.Name, src.Spec.StarRocksFeSpec),
		PodTemplateSpec:      podTemplateSpec,
		ServiceName:          internalServiceName,
		VolumeClaimTemplates: pvcs,
	}
}
