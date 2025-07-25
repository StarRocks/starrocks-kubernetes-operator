// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package load

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func Name(clusterName string, spec v1.SpecInterface) string {
	switch spec.(type) {
	case *v1.StarRocksFeSpec:
		return clusterName + "-" + v1.DEFAULT_FE
	case *v1.StarRocksBeSpec:
		return clusterName + "-" + v1.DEFAULT_BE
	case *v1.StarRocksCnSpec:
		return clusterName + "-" + v1.DEFAULT_CN
	case *v1.StarRocksFeProxySpec:
		return clusterName + "-" + v1.DEFAULT_FE_PROXY
	}
	return ""
}

func Labels(ownerReference string, spec v1.SpecInterface) map[string]string {
	labels := map[string]string{}
	labels[v1.OwnerReference] = ownerReference
	switch spec.(type) {
	case *v1.StarRocksFeSpec:
		labels[v1.ComponentLabelKey] = v1.DEFAULT_FE
	case *v1.StarRocksBeSpec:
		labels[v1.ComponentLabelKey] = v1.DEFAULT_BE
	case *v1.StarRocksCnSpec:
		labels[v1.ComponentLabelKey] = v1.DEFAULT_CN
	case *v1.StarRocksFeProxySpec:
		labels[v1.ComponentLabelKey] = v1.DEFAULT_FE_PROXY
	}
	return labels
}

func Annotations() map[string]string {
	annotations := map[string]string{}
	return annotations
}
