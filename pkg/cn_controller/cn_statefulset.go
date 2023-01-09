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

package cn_controller

import (
	v1alpha12 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cnStatefulSetsLabels(src *v1alpha12.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[v1alpha12.OwnerReference] = src.Name
	labels[v1alpha12.ComponentLabelKey] = v1alpha12.DEFAULT_CN
	labels.AddLabel(src.Labels)
	return labels
}

func cnStatefulSetName(src *v1alpha12.StarRocksCluster) string {
	stname := src.Name + "-" + v1alpha12.DEFAULT_CN
	if src.Spec.StarRocksCnSpec.Name != "" {
		stname = src.Spec.StarRocksCnSpec.Name
	}
	return stname
}

//buildStatefulSetParams generate the params of construct the statefulset.
func (cc *CnController) buildStatefulSetParams(src *v1alpha12.StarRocksCluster, cnconfig map[string]interface{}) rutils.StatefulSetParams {
	cnSpec := src.Spec.StarRocksCnSpec
	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	podTemplateSpec := cc.buildPodTemplate(src, cnconfig)

	return rutils.StatefulSetParams{
		Name:            cnStatefulSetName(src),
		Namespace:       src.Namespace,
		ServiceName:     cc.getCnSearchService(),
		PodTemplateSpec: podTemplateSpec,
		Labels:          cnStatefulSetsLabels(src),
		Selector:        cc.cnPodLabels(src, cnStatefulSetName(src)),
		OwnerReferences: []metav1.OwnerReference{*or},
		Replicas:        cnSpec.Replicas,
	}
}
