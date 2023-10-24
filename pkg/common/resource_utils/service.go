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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/constant"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

type StarRocksServiceType string

const (
	FeService      StarRocksServiceType = "fe"
	BeService      StarRocksServiceType = "be"
	CnService      StarRocksServiceType = "cn"
	FeProxyService StarRocksServiceType = "fe-proxy"
)

const (
	FeHTTPPortName    = "http"
	FeRPCPortName     = "rpc"
	FeQueryPortName   = "query"
	FeEditLogPortName = "edit-log"

	BePortName          = "be"
	BeWebserverPortName = "webserver"
	BeHeartbeatPortName = "heartbeat"
	BeBrpcPortName      = "brpc"

	CnThriftPortName    = "thrift"
	CnWebserverPortName = "webserver"
	CnHeartbeatPortName = "heartbeat"
	CnBrpcPortName      = "brpc"
)

// HashService service hash components
type hashService struct {
	name        string
	namespace   string
	finalizers  []string
	ports       []corev1.ServicePort
	selector    map[string]string
	serviceType corev1.ServiceType
	// deal with external access load balancer.
	// serviceType corev1.ServiceType
	labels map[string]string
}

// BuildExternalService build the external service. not have selector
func BuildExternalService(src *srapi.StarRocksCluster, name string, serviceType StarRocksServiceType,
	config map[string]interface{}, selector map[string]string, labels map[string]string) corev1.Service {
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
		spec := src.Spec.StarRocksFeSpec
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_FE
		}
		setServiceType(spec.Service, &svc)
		anno = getServiceAnnotations(spec.Service)
		srPorts = getFeServicePorts(config, spec.Service)
	} else if serviceType == BeService {
		spec := src.Spec.StarRocksBeSpec
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_BE
		}
		setServiceType(spec.Service, &svc)
		anno = getServiceAnnotations(spec.Service)
		srPorts = getBeServicePorts(config, spec.Service)
	} else if serviceType == CnService {
		spec := src.Spec.StarRocksCnSpec
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_CN
		}
		setServiceType(spec.Service, &svc)
		anno = getServiceAnnotations(spec.Service)
		srPorts = getCnServicePorts(config, spec.Service)
	} else if serviceType == FeProxyService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_FE_PROXY
		}
		feproxySpec := src.Spec.StarRocksFeProxySpec
		setServiceType(feproxySpec.Service, &svc)
		anno = getServiceAnnotations(feproxySpec.Service)
		srPorts = []srapi.StarRocksServicePort{
			mergePort(feproxySpec.Service, srapi.StarRocksServicePort{
				Name:          FE_PORXY_HTTP_PORT_NAME,
				Port:          FE_PROXY_HTTP_PORT,
				ContainerPort: FE_PROXY_HTTP_PORT,
			}),
		}
	}

	ref := metav1.NewControllerRef(src, src.GroupVersionKind())
	svc.OwnerReferences = []metav1.OwnerReference{*ref}

	var ports []corev1.ServicePort
	for _, sp := range srPorts {
		servicePort := corev1.ServicePort{
			Name:       sp.Name,
			Port:       sp.Port,
			NodePort:   sp.NodePort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(int(sp.ContainerPort)),
		}
		if servicePort.Name == FeQueryPortName {
			servicePort.AppProtocol = func() *string {
				protocol := "mysql"
				return &protocol
			}()
		}
		ports = append(ports, servicePort)
	}
	// set Ports field before calculate resource hash
	svc.Spec.Ports = ports

	hso := serviceHashObject(&svc)
	anno[srapi.ComponentResourceHash] = hash.HashObject(hso)
	svc.Annotations = anno
	return svc
}

func getFeServicePorts(config map[string]interface{}, service *srapi.StarRocksService) (srPorts []srapi.StarRocksServicePort) {
	httpPort := GetPort(config, HTTP_PORT)
	rpcPort := GetPort(config, RPC_PORT)
	queryPort := GetPort(config, QUERY_PORT)
	editPort := GetPort(config, EDIT_LOG_PORT)
	srPorts = append(srPorts, mergePort(service, srapi.StarRocksServicePort{
		Port: httpPort, ContainerPort: httpPort, Name: FeHTTPPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: rpcPort, ContainerPort: rpcPort, Name: FeRPCPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: queryPort, ContainerPort: queryPort, Name: FeQueryPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: editPort, ContainerPort: editPort, Name: FeEditLogPortName,
	}))

	return srPorts
}

func getBeServicePorts(config map[string]interface{}, service *srapi.StarRocksService) (srPorts []srapi.StarRocksServicePort) {
	bePort := GetPort(config, BE_PORT)
	webserverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)

	srPorts = append(srPorts, mergePort(service, srapi.StarRocksServicePort{
		Port: bePort, ContainerPort: bePort, Name: BePortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: webserverPort, ContainerPort: webserverPort, Name: BeWebserverPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: heartPort, ContainerPort: heartPort, Name: BeHeartbeatPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: brpcPort, ContainerPort: brpcPort, Name: BeBrpcPortName,
	}))

	return srPorts
}

func getCnServicePorts(config map[string]interface{}, service *srapi.StarRocksService) (srPorts []srapi.StarRocksServicePort) {
	thriftPort := GetPort(config, THRIFT_PORT)
	webserverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)
	srPorts = append(srPorts, mergePort(service, srapi.StarRocksServicePort{
		Port: thriftPort, ContainerPort: thriftPort, Name: CnThriftPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: webserverPort, ContainerPort: webserverPort, Name: CnWebserverPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: heartPort, ContainerPort: heartPort, Name: CnHeartbeatPortName,
	}), mergePort(service, srapi.StarRocksServicePort{
		Port: brpcPort, ContainerPort: brpcPort, Name: CnBrpcPortName,
	}))

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

func mergePort(service *srapi.StarRocksService, defaultPort srapi.StarRocksServicePort) srapi.StarRocksServicePort {
	if service == nil || service.Ports == nil {
		return defaultPort
	}
	port := defaultPort
	for _, sp := range service.Ports {
		if sp.Name == defaultPort.Name {
			if sp.Port != 0 {
				port.Port = sp.Port
			}
			if sp.ContainerPort != 0 {
				port.ContainerPort = sp.ContainerPort
			}
			if sp.NodePort != 0 {
				port.NodePort = sp.NodePort
			}
			break
		}
	}
	return port
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
	klog.V(constant.LOG_LEVEL).Infof("new service hash object: %+v", nhsvc)
	if _, ok := nsvc.Annotations[srapi.ComponentResourceHash]; ok {
		nhsvcValue = nsvc.Annotations[srapi.ComponentResourceHash]
	} else {
		nhsvcValue = hash.HashObject(nhsvc)
	}

	// calculate the old hash value from the old service, not from annotation.
	ohsvc := serviceHashObject(oldsvc)
	klog.V(constant.LOG_LEVEL).Infof("old service hash object: %+v", ohsvc)
	ohsvcValue = hash.HashObject(ohsvc)

	return nhsvcValue == ohsvcValue &&
		nsvc.Namespace == oldsvc.Namespace /*&& oldGeneration == oldsvc.Generation*/
}

func serviceHashObject(svc *corev1.Service) hashService {
	return hashService{
		name:        svc.Name,
		namespace:   svc.Namespace,
		finalizers:  svc.Finalizers,
		ports:       svc.Spec.Ports,
		selector:    svc.Spec.Selector,
		labels:      svc.Labels,
		serviceType: svc.Spec.Type,
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
