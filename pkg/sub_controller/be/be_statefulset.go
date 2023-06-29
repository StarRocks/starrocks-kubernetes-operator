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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
)

// buildStatefulSetParams generate the params of construct the statefulset.
func (be *BeController) buildStatefulSetParams(src *srapi.StarRocksCluster, beconfig map[string]interface{}, internalServiceName string) statefulset.Params {
	beSpec := src.Spec.StarRocksBeSpec
	podTemplateSpec := be.buildPodTemplate(src, beconfig)
	return statefulset.MakeParams(src, beSpec, podTemplateSpec)
}
