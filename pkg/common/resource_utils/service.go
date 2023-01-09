package resource_utils

import (
	v1alpha12 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
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
	name      string
	namespace string
	ports     []corev1.ServicePort
	selector  map[string]string
	//deal with external access load balancer.
	//serviceType corev1.ServiceType
	labels map[string]string
}

//BuildExternalService build the external service.
func BuildExternalService(src *v1alpha12.StarRocksCluster, name string, serviceType StarRocksServiceType, config map[string]interface{}) corev1.Service {
	var srPorts []v1alpha12.StarRocksServicePort
	//the k8s service type.
	var svcType corev1.ServiceType
	if serviceType == FeService {
		if src.Spec.StarRocksFeSpec.Service != nil && src.Spec.StarRocksFeSpec.Service.Type != "" {
			svcType = src.Spec.StarRocksFeSpec.Service.Type
		} else {
			svcType = corev1.ServiceTypeClusterIP
		}

		httpPort := GetPort(config, HTTP_PORT)
		rpcPort := GetPort(config, RPC_PORT)
		queryPort := GetPort(config, QUERY_PORT)
		editPort := GetPort(config, EDIT_LOG_PORT)
		srPorts = append(srPorts, v1alpha12.StarRocksServicePort{
			Port: httpPort, ContainerPort: httpPort, Name: "http",
		}, v1alpha12.StarRocksServicePort{
			Port: rpcPort, ContainerPort: rpcPort, Name: "rpc",
		}, v1alpha12.StarRocksServicePort{
			Port: queryPort, ContainerPort: queryPort, Name: "query",
		}, v1alpha12.StarRocksServicePort{
			Port: editPort, ContainerPort: editPort, Name: "edit-log"})
	} else if serviceType == BeService {
		if src.Spec.StarRocksBeSpec.Service != nil && src.Spec.StarRocksBeSpec.Service.Type != "" {
			svcType = src.Spec.StarRocksBeSpec.Service.Type
		} else {
			svcType = corev1.ServiceTypeClusterIP
		}
		name = v1alpha12.DEFAULT_BE_SERVICE_NAME
		bePort := GetPort(config, BE_PORT)
		webseverPort := GetPort(config, WEBSERVER_PORT)
		heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
		brpcPort := GetPort(config, BRPC_PORT)
		srPorts = append(srPorts, v1alpha12.StarRocksServicePort{
			Port: bePort, ContainerPort: bePort, Name: "be",
		}, v1alpha12.StarRocksServicePort{
			Port: webseverPort, ContainerPort: webseverPort, Name: "webserver",
		}, v1alpha12.StarRocksServicePort{
			Port: heartPort, ContainerPort: heartPort, Name: "heartbeat",
		}, v1alpha12.StarRocksServicePort{
			Port: brpcPort, ContainerPort: brpcPort, Name: "brpc",
		})

	} else if serviceType == CnService {
		if src.Spec.StarRocksCnSpec.Service != nil && src.Spec.StarRocksCnSpec.Service.Type != "" {
			svcType = src.Spec.StarRocksCnSpec.Service.Type
		} else {
			svcType = corev1.ServiceTypeClusterIP
		}

		name = v1alpha12.DEFAULT_CN_SERVICE_NAME
		thriftPort := GetPort(config, THRIFT_PORT)
		webseverPort := GetPort(config, WEBSERVER_PORT)
		heartPort := GetPort(config, HEARTBEAT_SERVICE_PORT)
		brpcPort := GetPort(config, BRPC_PORT)
		srPorts = append(srPorts, v1alpha12.StarRocksServicePort{
			Port: thriftPort, ContainerPort: thriftPort, Name: "thrift",
		}, v1alpha12.StarRocksServicePort{
			Port: webseverPort, ContainerPort: webseverPort, Name: "webserver",
		}, v1alpha12.StarRocksServicePort{
			Port: heartPort, ContainerPort: heartPort, Name: "heartbeat",
		}, v1alpha12.StarRocksServicePort{
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
			Type:     svcType,
			Selector: labels,
			Ports:    ports,
		},
	}

	hso := serviceHashObject(&svc)
	anno := map[string]string{}
	anno[v1alpha12.ComponentResourceHash] = hash.HashObject(hso)
	anno[v1alpha12.ComponentGeneration] = "1"
	svc.Annotations = anno

	return svc
}

func ServiceDeepEqual(nsvc, oldsvc *corev1.Service) bool {
	var nhsvcValue, ohsvcValue string
	if _, ok := nsvc.Annotations[v1alpha12.ComponentResourceHash]; ok {
		nhsvcValue = nsvc.Annotations[v1alpha12.ComponentResourceHash]
	} else {
		nhsvc := serviceHashObject(nsvc)
		nhsvcValue = hash.HashObject(nhsvc)
	}

	if _, ok := oldsvc.Annotations[v1alpha12.ComponentResourceHash]; ok {
		ohsvcValue = oldsvc.Annotations[v1alpha12.ComponentResourceHash]
	} else {
		ohsvc := serviceHashObject(oldsvc)
		ohsvcValue = hash.HashObject(ohsvc)
	}
	oldGeneration, _ := strconv.ParseInt(oldsvc.Annotations[v1alpha12.ComponentGeneration], 10, 64)

	return nhsvcValue == ohsvcValue && nsvc.Name == oldsvc.Name &&
		nsvc.Namespace == oldsvc.Namespace && oldGeneration == oldsvc.Generation
}

func serviceHashObject(svc *corev1.Service) hashService {
	return hashService{
		name:      svc.Name,
		namespace: svc.Namespace,
		ports:     svc.Spec.Ports,
		selector:  svc.Spec.Selector,
		//serviceType: svc.Spec.Type,
		labels: svc.Labels,
	}
}

//GenerateServiceLabels generate the default labels list starrocks services.
func GenerateServiceLabels(src *v1alpha12.StarRocksCluster, serviceType StarRocksServiceType) Labels {
	labels := Labels{}
	labels.AddLabel(src.Labels)
	if serviceType == FeService {
		labels.Add(v1alpha12.ComponentLabelKey, v1alpha12.DEFAULT_FE)
	} else if serviceType == BeService {
		labels.Add(v1alpha12.ComponentLabelKey, v1alpha12.DEFAULT_BE)
	} else if serviceType == CnService {
		labels.Add(v1alpha12.ComponentLabelKey, v1alpha12.DEFAULT_CN)
	}

	return labels
}
