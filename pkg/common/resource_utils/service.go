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

func BuildExternalService(src *srapi.StarRocksCluster, serviceType StarRocksServiceType) corev1.Service {

	var name string
	var srPorts []srapi.StarRocksServicePort
	if serviceType == FeService {
		name = srapi.DEFAULT_FE_SERVICE_NAME
		if src.Spec.StarRocksFeSpec.Service != nil {
			srPorts = append(srPorts, src.Spec.StarRocksFeSpec.Service.Ports...)
			if src.Spec.StarRocksFeSpec.Service.Name != "" {
				name = src.Spec.StarRocksFeSpec.Service.Name
			}
		}

		srPorts = append(srPorts, srapi.StarRocksServicePort{
			Port: 8030, ContainerPort: 8030, Name: "http-port",
		}, srapi.StarRocksServicePort{
			Port: 9020, ContainerPort: 9020, Name: "rpc-port",
		}, srapi.StarRocksServicePort{
			Port: 9030, ContainerPort: 9030, Name: "query-port",
		}, srapi.StarRocksServicePort{
			Port: 9010, ContainerPort: 9010, Name: "edit-log-port"})
	} else if serviceType == BeService {
		name = srapi.DEFAULT_BE_SERVICE_NAME
		if src.Spec.StarRocksBeSpec.Service != nil {
			srPorts = append(srPorts, src.Spec.StarRocksBeSpec.Service.Ports...)
			if src.Spec.StarRocksBeSpec.Service.Name != "" {
				name = src.Spec.StarRocksBeSpec.Service.Name
			}
		}
		srPorts = append(srPorts, srapi.StarRocksServicePort{
			Port: 9060, ContainerPort: 9060, Name: "be-port",
		}, srapi.StarRocksServicePort{
			Port: 8040, ContainerPort: 8040, Name: "webserver-port",
		}, srapi.StarRocksServicePort{
			Port: 9050, ContainerPort: 9050, Name: "heartbeat-service-port",
		}, srapi.StarRocksServicePort{
			Port: 8060, ContainerPort: 80060, Name: "brpc-port",
		})

	} else if serviceType == CnService {
		name = srapi.DEFAULT_CN_SERVICE_NAME
		if src.Spec.StarRocksCnSpec.Service != nil {
			srPorts = append(srPorts, src.Spec.StarRocksCnSpec.Service.Ports...)
			if src.Spec.StarRocksCnSpec.Service.Name != "" {
				name = src.Spec.StarRocksCnSpec.Service.Name
			}
		}

		srPorts = append(srPorts, srapi.StarRocksServicePort{
			Port: 9060, ContainerPort: 9060, Name: "thrift-port",
		}, srapi.StarRocksServicePort{
			Port: 8040, ContainerPort: 8040, Name: "webserver-port",
		}, srapi.StarRocksServicePort{
			Port: 9050, ContainerPort: 9050, Name: "heartbeat-service-port",
		}, srapi.StarRocksServicePort{
			Port: 8060, ContainerPort: 80060, Name: "brpc-port",
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
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_BE)
	} else if serviceType == BeService {
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_BE)
	} else if serviceType == CnService {
		labels.Add(srapi.ComponentLabelKey, srapi.DEFAULT_CN)
	}

	return labels
}
