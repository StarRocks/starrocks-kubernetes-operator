package service

import (
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	corev1 "k8s.io/api/core/v1"
)

func MakeSearchService(serviceName string, spec interface{}, externalService *corev1.Service, ports []corev1.ServicePort) *corev1.Service {
	searchSvc := &corev1.Service{}
	externalService.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	searchSvc.Name = serviceName
	searchSvc.Spec = corev1.ServiceSpec{
		ClusterIP: "None",
		Ports:     ports,
		Selector:  externalService.Spec.Selector,
		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	return searchSvc
}

// SearchServiceName get the domain service name, the domain service for statefulset.
// domain service have PublishNotReadyAddresses. while used PublishNotReadyAddresses, the fe start need all instance domain can resolve.
func SearchServiceName(clusterName string, spec interface{}) string {
	switch spec.(type) {
	case *v1.StarRocksBeSpec:
		return clusterName + "-be-search"
	case *v1.StarRocksCnSpec:
		return clusterName + "-cn-search"
	case *v1.StarRocksFeSpec:
		return clusterName + "-fe-search"
	default:
		return ""
	}
}
