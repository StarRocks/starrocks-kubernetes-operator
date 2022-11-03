package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

type StarRocksServiceType string

const (
	FeService = "fe"
	BeService = "be"
	CnService = "cn"
)

//HashService service hash components
type hashService struct {
	name        string
	namespace   string
	ports       []corev1.ServicePort
	selector    map[string]string
	serviceType corev1.ServiceType
	labels      map[string]string
}

//BuildExternalService build the external service.
func BuildExternalService(src *srapi.StarRocksCluster, name string, serviceType StarRocksServiceType, config map[string]interface{}) corev1.Service {
	var srPorts []srapi.StarRocksServicePort
	if serviceType == FeService {
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
	} else if serviceType == BeService {
		name = srapi.DEFAULT_BE_SERVICE_NAME
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

	} else if serviceType == CnService {
		name = srapi.DEFAULT_CN_SERVICE_NAME
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
	}

	labels := GenerateServiceLabels(src, serviceType)
	or := metav1.OwnerReference{
		APIVersion: src.APIVersion,
		Kind:       src.Kind,
		Name:       src.Name,
		UID:        src.UID,
	}

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

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       src.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{or},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
		},
	}

	hso := serviceHashObject(&svc)
	anno := map[string]string{}
	anno[srapi.ComponentResourceHash] = hash.HashObject(hso)
	anno[srapi.ComponentGeneration] = "1"
	svc.Annotations = anno

	return svc
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
	oldGeneration, _ := strconv.ParseInt(oldsvc.Annotations[srapi.ComponentGeneration], 10, 64)

	return nhsvcValue == ohsvcValue && nsvc.Name == oldsvc.Name &&
		nsvc.Namespace == oldsvc.Namespace && oldGeneration == oldsvc.Generation
}

func serviceHashObject(svc *corev1.Service) hashService {
	return hashService{
		name:        svc.Name,
		namespace:   svc.Namespace,
		ports:       svc.Spec.Ports,
		selector:    svc.Spec.Selector,
		serviceType: svc.Spec.Type,
		labels:      svc.Labels,
	}
}

//GenerateServiceLabels generate the default labels list starrocks services.
func GenerateServiceLabels(src *srapi.StarRocksCluster, serviceType StarRocksServiceType) Labels {
	labels := Labels{}
	labels.AddLabel(src.Labels)
	if serviceType == FeService {
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_FE)
	} else if serviceType == BeService {
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_BE)
	} else if serviceType == CnService {
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_CN)
	}

	return labels
}
