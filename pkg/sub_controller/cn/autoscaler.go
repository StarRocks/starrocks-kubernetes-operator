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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	appv1 "k8s.io/api/apps/v1"
)

func (cc *CnController) generateAutoScalerName(src *srapi.StarRocksCluster) string {
	return load.Name(src.Name, src.Spec.StarRocksCnSpec) + "-autoscaler"
}

func (cc *CnController) buildCnAutoscalerParams(scalerInfo srapi.AutoScalingPolicy, target *appv1.StatefulSet, src *srapi.StarRocksCluster) *rutils.PodAutoscalerParams {

	labels := rutils.Labels{}
	labels.AddLabel(target.Labels)
	labels.Add(srapi.ComponentLabelKey, "autoscaler")

	return &rutils.PodAutoscalerParams{
		Namespace:      target.Namespace,
		Name:           cc.generateAutoScalerName(src),
		Labels:         labels,
		AutoscalerType: src.Spec.StarRocksCnSpec.AutoScalingPolicy.Version,
		TargetName:     target.Name,
		// use src as ownerReference for reconciling on autoscaler updated.
		OwnerReferences: target.OwnerReferences,
		ScalerPolicy:    &scalerInfo,
	}
}
