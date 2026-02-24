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

package service

import (
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func MakeSearchService(serviceName string, externalService *corev1.Service, ports []corev1.ServicePort,
	defaultLabels map[string]string) *corev1.Service {
	searchSvc := &corev1.Service{}
	externalService.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	// Set annotations to nil so external service annotations aren't copied. Interal / search service annotations aren't
	// needed, and adding them can be an issue since some service annotations can only be used when `type` is
	// 'LoadBalancer', e.g. service.beta.kubernetes.io/load-balancer-source-ranges.
	searchSvc.Annotations = nil
	searchSvc.Name = serviceName
	// Set labels to the default labels so external service labels aren't copied. Since labels are used to select objects,
	// adding them to the internal / search service can be detrimental.
	searchSvc.Labels = defaultLabels
	searchSvc.Spec = corev1.ServiceSpec{
		ClusterIP: "None",
		Ports:     ports,
		Selector:  externalService.Spec.Selector,
		// value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	return searchSvc
}

// SearchServiceName get the domain service name, the domain service for statefulset.
// domain service have PublishNotReadyAddresses. while used PublishNotReadyAddresses, the fe start need all instance domain can resolve.
func SearchServiceName(clusterName string, spec v1.SpecInterface) string {
	switch spec.(type) {
	case *v1.StarRocksBeSpec:
		return clusterName + "-be-search"
	case *v1.StarRocksCnSpec:
		return clusterName + "-cn-search"
	case *v1.StarRocksFeSpec:
		return clusterName + "-fe-search"
	case *v1.StarRocksFeObserverSpec:
		return clusterName + "-fe-observer-search"
	default:
		return ""
	}
}

// ExternalServiceName generate the name of external service.
func ExternalServiceName(clusterName string, spec v1.SpecInterface) string {
	switch spec.(type) {
	case *v1.StarRocksFeSpec:
		return clusterName + "-" + v1.DEFAULT_FE + "-service"
	case *v1.StarRocksFeObserverSpec:
		return clusterName + "-" + v1.DEFAULT_FE_OBSERVER + "-service"
	case *v1.StarRocksBeSpec:
		return clusterName + "-" + v1.DEFAULT_BE + "-service"
	case *v1.StarRocksCnSpec:
		return clusterName + "-" + v1.DEFAULT_CN + "-service"
	case *v1.StarRocksFeProxySpec:
		return clusterName + "-" + v1.DEFAULT_FE_PROXY + "-service"
	}
	return ""
}
