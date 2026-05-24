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

package resource_utils

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

// FeIngressName returns the name of the Ingress for the FE web UI.
func FeIngressName(clusterName string) string {
	return clusterName + "-fe-ingress"
}

// BuildFeIngress builds an Ingress that routes external HTTP traffic to the FE web UI.
// The backend always targets the FE external service's "http" port by name, so the
// generated Ingress can never point at the MySQL query port (an L4 protocol that a
// standard Ingress cannot route).
func BuildFeIngress(object object.StarRocksObject, feIngress *srapi.FeIngress,
	labels map[string]string) networkingv1.Ingress {
	feServiceName := service.ExternalServiceName(object.SubResourcePrefixName, (*srapi.StarRocksFeSpec)(nil))
	pathType := networkingv1.PathTypePrefix

	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        FeIngressName(object.SubResourcePrefixName),
			Namespace:   object.GetNamespace(),
			Labels:      labels,
			Annotations: feIngress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: feIngress.IngressClassName,
			TLS:              feIngress.TLS,
			Rules: []networkingv1.IngressRule{
				{
					Host: feIngress.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: feServiceName,
											Port: networkingv1.ServiceBackendPort{
												Name: FeHTTPPortName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ref := metav1.NewControllerRef(object, object.GroupVersionKind())
	ingress.OwnerReferences = []metav1.OwnerReference{*ref}
	return ingress
}
