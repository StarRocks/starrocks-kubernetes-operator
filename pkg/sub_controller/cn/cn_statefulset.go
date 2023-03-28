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
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cc *CnController) cnStatefulSetsLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_CN
	//once src labels updated, the statefulset will enter into a not reconcile state.
	//labels.AddLabel(src.Labels)
	return labels
}

//buildStatefulSetParams generate the params of construct the statefulset.
func (cc *CnController) buildStatefulSetParams(src *srapi.StarRocksCluster, cnconfig map[string]interface{}, internalServiceName string) rutils.StatefulSetParams {
	cnSpec := src.Spec.StarRocksCnSpec
	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := cc.buildPodTemplate(src, cnconfig)

	annos := rutils.Annotations{}
	// add restart annotation on statefulset.
	if _, ok := src.Annotations[string(srapi.AnnotationCNRestartKey)]; ok {
		annos.Add(string(srapi.AnnotationCNRestartKey), string(srapi.AnnotationRestart))
	}

	return rutils.StatefulSetParams{
		Name:            srapi.CnStatefulSetName(src),
		Annotations:     annos,
		Namespace:       src.Namespace,
		ServiceName:     internalServiceName,
		PodTemplateSpec: podTemplateSpec,
		Labels:          cc.cnStatefulSetsLabels(src),
		Selector:        cc.cnPodLabels(src),
		OwnerReferences: []metav1.OwnerReference{*or},
		Replicas:        cnSpec.Replicas,
	}
}

func (cc *CnController) cnStatefulsetSelector(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = srapi.CnStatefulSetName(src)
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_CN
	return labels
}
