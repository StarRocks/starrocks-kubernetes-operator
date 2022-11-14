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
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cc *CnController) buildCnAutoscalerParams(scalerInfo srapi.AutoScalingPolicy, st *appv1.StatefulSet) *rutils.PodAutoscalerParams {

	labels := rutils.Labels{}
	labels.AddLabel(st.Labels)
	labels.Add(srapi.ComponentLabelKey, "autoscaler")

	ors := []metav1.OwnerReference{*metav1.NewControllerRef(st, st.GroupVersionKind())}
	return &rutils.PodAutoscalerParams{
		Namespace:       st.Namespace,
		Name:            st.Name + "-autoscaler",
		Labels:          labels,
		TargetName:      st.Name,
		OwnerReferences: ors,
		ScalerPolicy:    &scalerInfo,
	}
}
