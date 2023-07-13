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

package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

type StarRocksServiceType string

const (
	FeService StarRocksServiceType = "fe"
	BeService StarRocksServiceType = "be"
	CnService StarRocksServiceType = "cn"
)

// HashService service hash components
type hashService struct {
	name       string
	namespace  string
	finalizers []string
	ports      []corev1.ServicePort
	selector   map[string]string
	// deal with external access load balancer.
	// serviceType corev1.ServiceType
	labels map[string]string
}

// BuildExternalService build the external service. not have selector
func BuildExternalService(src *srapi.StarRocksCluster, name string, serviceType StarRocksServiceType, config map[string]interface{}, selector map[string]string, labels map[string]string) corev1.Service {
	// the k8s service type.
	var srPorts []srapi.StarRocksServicePort
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: src.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
		},
	}

	anno := map[string]string{}
	if serviceType == FeService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_FE
		}
		setServiceType(src.Spec.StarRocksFeSpec.Service, &svc)
		anno = getServiceAnnotations(src.Spec.StarRocksFeSpec.Service)
		srPorts = getFeServicePorts(config)
	} else if serviceType == BeService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_BE
		}
		setServiceType(src.Spec.StarRocksBeSpec.Service, &svc)
		anno = getServiceAnnotations(src.Spec.StarRocksBeSpec.Service)
		srPorts = getBeServicePorts(config)
	} else if serviceType == CnService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_CN
		}
		setServiceType(src.Spec.StarRocksCnSpec.Service, &svc)
		anno = getServiceAnnotations(src.Spec.StarRocksCnSpec.Service)
		srPorts = getCnServicePorts(config)
	}

	ref := metav1.NewControllerRef(src, src.GroupVersionKind())
	svc.OwnerReferences = []metav1.OwnerReference{*ref}

	var ports []corev1.ServicePort
	for _, sp := range srPorts {
		ports = append(ports, corev1.ServicePort{
			Name:       sp.Name,
			Port:       sp.Port,
			NodePort:   sp.NodePort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(int(sp.ContainerPort)),
		})
	}
	// set Ports field before calculate resource hash
	svc.Spec.Ports = ports

	hso := serviceHashObject(&svc)
	anno[srapi.ComponentResourceHash] = hash.HashObject(hso)
	svc.Annotations = anno
	return svc
}

func getFeServicePorts(config map[string]interface{}) (srPorts []srapi.StarRocksServicePort) {
	httpPort := GetPort(config, HTTP_PORT)
	rpcPort := GetPort(config, RPC_PORT)
	queryPort := GetPort(config, QUERY_PORT)
	editPort := GetPort(config, EDIT_LOG_PORT)
	srPorts = append(srPorts, srapi.StarRocksServicePort{
		Port: httpPort, ContainerPort: httpPort, Name: "http",
	}, srapi.StarRocksServicePort{
		Port: rpcPort, ContainerPort: rpcPort, Name: "rpc",
	}, srapi.StarRocksServicePort{
		Port: queryPort, ContainerPort: queryPort, Name: "query",
	}, srapi.StarRocksServicePort{
		Port: editPort, ContainerPort: editPort, Name: "edit-log"})

	return srPorts
}

func getBeServicePorts(config map[string]interface{}) (srPorts []srapi.StarRocksServicePort) {
	bePort := GetPort(config, BE_PORT)
	webserverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)

	srPorts = append(srPorts, srapi.StarRocksServicePort{
		Port: bePort, ContainerPort: bePort, Name: "be",
	}, srapi.StarRocksServicePort{
		Port: webserverPort, ContainerPort: webserverPort, Name: "webserver",
	}, srapi.StarRocksServicePort{
		Port: heartPort, ContainerPort: heartPort, Name: "heartbeat",
	}, srapi.StarRocksServicePort{
		Port: brpcPort, ContainerPort: brpcPort, Name: "brpc",
	})

	return srPorts
}

func getCnServicePorts(config map[string]interface{}) (srPorts []srapi.StarRocksServicePort) {
	thriftPort := GetPort(config, THRIFT_PORT)
	webserverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)
	srPorts = append(srPorts, srapi.StarRocksServicePort{
		Port: thriftPort, ContainerPort: thriftPort, Name: "thrift",
	}, srapi.StarRocksServicePort{
		Port: webserverPort, ContainerPort: webserverPort, Name: "webserver",
	}, srapi.StarRocksServicePort{
		Port: heartPort, ContainerPort: heartPort, Name: "heartbeat",
	}, srapi.StarRocksServicePort{
		Port: brpcPort, ContainerPort: brpcPort, Name: "brpc",
	})

	return srPorts
}

func setServiceType(svc *srapi.StarRocksService, service *corev1.Service) {
	service.Spec.Type = corev1.ServiceTypeClusterIP
	if svc != nil && svc.Type != "" {
		service.Spec.Type = svc.Type
	}

	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && svc.LoadBalancerIP != "" {
		service.Spec.LoadBalancerIP = svc.LoadBalancerIP
	}
}

func getServiceAnnotations(svc *srapi.StarRocksService) map[string]string {
	if svc != nil && svc.Annotations != nil {
		annotations := map[string]string{}
		for key, val := range svc.Annotations {
			annotations[key] = val
		}
		return annotations
	}
	return map[string]string{}
}

func ServiceDeepEqual(nsvc, oldsvc *corev1.Service) bool {
	var nhsvcValue, ohsvcValue string

	nhsvc := serviceHashObject(nsvc)
	klog.V(4).Infof("new service hash object: %+v", nhsvc)
	if _, ok := nsvc.Annotations[srapi.ComponentResourceHash]; ok {
		nhsvcValue = nsvc.Annotations[srapi.ComponentResourceHash]
	} else {
		nhsvcValue = hash.HashObject(nhsvc)
	}

	// calculate the old hash value from the old service, not from annotation.
	ohsvc := serviceHashObject(oldsvc)
	klog.V(4).Infof("old service hash object: %+v", ohsvc)
	ohsvcValue = hash.HashObject(ohsvc)

	return nhsvcValue == ohsvcValue &&
		nsvc.Namespace == oldsvc.Namespace /*&& oldGeneration == oldsvc.Generation*/
}

func serviceHashObject(svc *corev1.Service) hashService {
	return hashService{
		name:       svc.Name,
		namespace:  svc.Namespace,
		finalizers: svc.Finalizers,
		ports:      svc.Spec.Ports,
		selector:   svc.Spec.Selector,
		labels:     svc.Labels,
	}
}

func HaveEqualOwnerReference(svc1 *corev1.Service, svc2 *corev1.Service) bool {
	set := make(map[string]bool)
	for _, o := range svc1.OwnerReferences {
		key := o.Kind + o.Name
		set[key] = true

	}

	for _, o := range svc2.OwnerReferences {
		key := o.Kind + o.Name
		if _, ok := set[key]; ok {
			return true
		}
	}
	return false
}
