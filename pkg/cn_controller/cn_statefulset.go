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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cnStatefulSetsLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_CN
	labels.AddLabel(src.Labels)
	return labels
}

func cnStatefulSetName(src *srapi.StarRocksCluster) string {
	stname := src.Name + "-" + srapi.DEFAULT_CN
	if src.Spec.StarRocksCnSpec.Name != "" {
		stname = src.Spec.StarRocksCnSpec.Name
	}
	return stname
}

//buildStatefulSetParams generate the params of construct the statefulset.
func (cc *CnController) buildStatefulSetParams(src *srapi.StarRocksCluster, cnconfig map[string]interface{}) rutils.StatefulSetParams {
	cnSpec := src.Spec.StarRocksCnSpec
	or := metav1.OwnerReference{
		UID:        src.UID,
		Kind:       src.Kind,
		APIVersion: src.APIVersion,
		Name:       src.Name,
	}

	return rutils.StatefulSetParams{
		Name:            cnStatefulSetName(src),
		Namespace:       src.Namespace,
		ServiceName:     cc.getCnDomainService(),
		PodTemplateSpec: cc.buildPodTemplate(src, cnconfig),
		Labels:          cnStatefulSetsLabels(src),
		Selector:        cc.cnPodLabels(src, cnStatefulSetName(src)),
		OwnerReferences: []metav1.OwnerReference{or},
		Replicas:        cnSpec.Replicas,
	}
}
