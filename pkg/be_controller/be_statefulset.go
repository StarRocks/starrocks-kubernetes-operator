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

package be_controller

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//buildStatefulSetParams generate the params of construct the statefulset.
func (be *BeController) buildStatefulSetParams(src *srapi.StarRocksCluster, beconfig map[string]interface{}) rutils.StatefulSetParams {
	beSpec := src.Spec.StarRocksBeSpec

	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := be.buildPodTemplate(src, beconfig)

	return rutils.StatefulSetParams{
		Name:            beStatefulSetName(src),
		Namespace:       src.Namespace,
		ServiceName:     be.getBeDomainService(),
		PodTemplateSpec: podTemplateSpec,
		Labels:          beStatefulSetsLabels(src),
		Selector:        be.bePodLabels(src, beStatefulSetName(src)),
		OwnerReferences: []metav1.OwnerReference{*or},
		Replicas:        beSpec.Replicas,
	}
}

func beStatefulSetsLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	labels.AddLabel(src.Labels)
	return labels
}

func beStatefulSetName(src *srapi.StarRocksCluster) string {
	stname := src.Name + "-" + srapi.DEFAULT_BE
	if src.Spec.StarRocksBeSpec.Name != "" {
		stname = src.Spec.StarRocksBeSpec.Name
	}
	return stname
}
