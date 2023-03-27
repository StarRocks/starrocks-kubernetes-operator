package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type StarRocksServiceType string

const (
	FeService = "fe"
	BeService = "be"
	CnService = "cn"
)

//HashService service hash components
type hashService struct {
	name      string
	namespace string
	ports     []corev1.ServicePort
	selector  map[string]string
	//deal with external access load balancer.
	//serviceType corev1.ServiceType
	labels map[string]string
}

//BuildExternalService build the external service. not have selector
func BuildExternalService(src *srapi.StarRocksCluster, name string, serviceType StarRocksServiceType, config map[string]interface{}, selector map[string]string) corev1.Service {
	//the k8s service type.
	var srPorts []srapi.StarRocksServicePort
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: src.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
		},
	}
	svc.Finalizers = append(svc.Finalizers, srapi.SERVICE_FINALIZER)

	if serviceType == FeService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_FE
		}
		setServiceType(src.Spec.StarRocksFeSpec.Service, &svc)
		srPorts = getFeServicePorts(config)
	} else if serviceType == BeService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_BE
		}

		setServiceType(src.Spec.StarRocksBeSpec.Service, &svc)
		srPorts = getBeServicePorts(config)
	} else if serviceType == CnService {
		if svc.Name == "" {
			svc.Name = src.Name + "-" + srapi.DEFAULT_CN
		}

		setServiceType(src.Spec.StarRocksCnSpec.Service, &svc)
		srPorts = getCnServicePorts(config)
	}

	or := metav1.OwnerReference{
		APIVersion: src.APIVersion,
		Kind:       src.Kind,
		Name:       src.Name,
		UID:        src.UID,
	}
	svc.OwnerReferences = []metav1.OwnerReference{or}

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

	hso := serviceHashObject(&svc)
	anno := map[string]string{}
	anno[srapi.ComponentResourceHash] = hash.HashObject(hso)
	svc.Annotations = anno
	svc.Spec.Ports = ports
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
	webseverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)

	srPorts = append(srPorts, srapi.StarRocksServicePort{
		Port: bePort, ContainerPort: bePort, Name: "be",
	}, srapi.StarRocksServicePort{
		Port: webseverPort, ContainerPort: webseverPort, Name: "webserver",
	}, srapi.StarRocksServicePort{
		Port: heartPort, ContainerPort: heartPort, Name: "heartbeat",
	}, srapi.StarRocksServicePort{
		Port: brpcPort, ContainerPort: brpcPort, Name: "brpc",
	})

	return srPorts
}

func getCnServicePorts(config map[string]interface{}) (srPorts []srapi.StarRocksServicePort) {
	thriftPort := GetPort(config, THRIFT_PORT)
	webseverPort := GetPort(config, WEBSERVER_PORT)
	heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
	brpcPort := GetPort(config, BRPC_PORT)
	srPorts = append(srPorts, srapi.StarRocksServicePort{
		Port: thriftPort, ContainerPort: thriftPort, Name: "thrift",
	}, srapi.StarRocksServicePort{
		Port: webseverPort, ContainerPort: webseverPort, Name: "webserver",
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

func ServiceDeepEqual(nsvc, oldsvc *corev1.Service) bool {
	var nhsvcValue, ohsvcValue string
	if _, ok := nsvc.Annotations[srapi.ComponentResourceHash]; ok {
		nhsvcValue = nsvc.Annotations[srapi.ComponentResourceHash]
	} else {
		nhsvc := serviceHashObject(nsvc)
		nhsvcValue = hash.HashObject(nhsvc)
	}

	if _, ok := oldsvc.Annotations[srapi.ComponentResourceHash]; ok {
		ohsvcValue = oldsvc.Annotations[srapi.ComponentResourceHash]
	} else {
		ohsvc := serviceHashObject(oldsvc)
		ohsvcValue = hash.HashObject(ohsvc)
	}

	return nhsvcValue == ohsvcValue &&
		nsvc.Namespace == oldsvc.Namespace /*&& oldGeneration == oldsvc.Generation*/
}

func serviceHashObject(svc *corev1.Service) hashService {
	return hashService{
		name:      svc.Name,
		namespace: svc.Namespace,
		ports:     svc.Spec.Ports,
		selector:  svc.Spec.Selector,
		labels:    svc.Labels,
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
