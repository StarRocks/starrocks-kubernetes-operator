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
	"encoding/json"
	"reflect"
	"testing"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_getServiceAnnotations(t *testing.T) {
	type args struct {
		svc *srapi.StarRocksService
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "empty service",
			args: args{
				svc: &srapi.StarRocksService{},
			},
			want: map[string]string{},
		},
		{
			name: "service with annotations",
			args: args{
				svc: &srapi.StarRocksService{
					Annotations: map[string]string{
						"test": "test",
					},
				},
			},
			want: map[string]string{"test": "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceAnnotations(tt.args.svc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getServiceAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildExternalService(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Service: &srapi.StarRocksService{
							Type:           corev1.ServiceTypeLoadBalancer,
							LoadBalancerIP: "127.0.0.1",
						},
					},
				},
			},
		},
	}
	type args struct {
		src         *srapi.StarRocksCluster
		name        string
		serviceType StarRocksServiceType
		config      map[string]interface{}
		selector    map[string]string
		labels      map[string]string
	}
	tests := []struct {
		name string
		args args
		want corev1.Service
	}{
		{
			name: "test build external service",
			args: args{
				src:         src,
				name:        "service-name",
				serviceType: FeService,
				config:      map[string]interface{}{},
				selector:    map[string]string{},
				labels:      map[string]string{},
			},
			want: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service-name",
					Namespace: "default",
					Annotations: map[string]string{
						srapi.ComponentResourceHash: "1503664666",
					},
					OwnerReferences: func() []metav1.OwnerReference {
						ref := metav1.NewControllerRef(src, src.GroupVersionKind())
						return []metav1.OwnerReference{*ref}
					}(),
				},
				Spec: corev1.ServiceSpec{
					Type:                     corev1.ServiceTypeLoadBalancer,
					PublishNotReadyAddresses: false,
					LoadBalancerIP:           "127.0.0.1",
					Ports: func() []corev1.ServicePort {
						srPorts := getFeServicePorts(map[string]interface{}{}, nil)
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
								servicePort.AppProtocol = func() *string { v := "mysql"; return &v }()
							}
							ports = append(ports, servicePort)
						}
						return ports
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildExternalService(tt.args.src, tt.args.name, tt.args.serviceType, tt.args.config, tt.args.selector, tt.args.labels)
			gotData, _ := json.Marshal(got)
			wantData, _ := json.Marshal(tt.want)
			if len(gotData) != len(wantData) {
				t.Errorf("BuildExternalService() = %v, want %v", got, tt.want)
				return
			}
			for i := range gotData {
				if gotData[i] != wantData[i] {
					t.Errorf("BuildExternalService() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func Test_getFeServicePorts(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantSrPorts []srapi.StarRocksServicePort
	}{
		{
			name: "test get fe service ports",
			args: args{
				config: map[string]interface{}{},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "http",
					Port:          DefMap[HTTP_PORT],
					ContainerPort: DefMap[HTTP_PORT],
				},
				{
					Name:          "rpc",
					Port:          DefMap[RPC_PORT],
					ContainerPort: DefMap[RPC_PORT],
				},
				{
					Name:          "query",
					Port:          DefMap[QUERY_PORT],
					ContainerPort: DefMap[QUERY_PORT],
				},
				{
					Name:          "edit-log",
					Port:          DefMap[EDIT_LOG_PORT],
					ContainerPort: DefMap[EDIT_LOG_PORT],
				},
			},
		},
		{
			name: "test get fe service ports 2",
			args: args{
				config: map[string]interface{}{
					HTTP_PORT:     "1",
					RPC_PORT:      "2",
					QUERY_PORT:    "3",
					EDIT_LOG_PORT: "4",
				},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "http",
					Port:          1,
					ContainerPort: 1,
				},
				{
					Name:          "rpc",
					Port:          2,
					ContainerPort: 2,
				},
				{
					Name:          "query",
					Port:          3,
					ContainerPort: 3,
				},
				{
					Name:          "edit-log",
					Port:          4,
					ContainerPort: 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSrPorts := getFeServicePorts(tt.args.config, nil); !reflect.DeepEqual(gotSrPorts, tt.wantSrPorts) {
				t.Errorf("getFeServicePorts() = %v, want %v", gotSrPorts, tt.wantSrPorts)
			}
		})
	}
}

func Test_getBeServicePorts(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantSrPorts []srapi.StarRocksServicePort
	}{
		{
			name: "test get be service ports",
			args: args{
				config: map[string]interface{}{},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "be",
					Port:          DefMap[BE_PORT],
					ContainerPort: DefMap[BE_PORT],
				},
				{
					Name:          "webserver",
					Port:          DefMap[WEBSERVER_PORT],
					ContainerPort: DefMap[WEBSERVER_PORT],
				},
				{
					Name:          "heartbeat",
					Port:          DefMap[HEARTBEAT_SERVICE_PORT],
					ContainerPort: DefMap[HEARTBEAT_SERVICE_PORT],
				},
				{
					Name:          "brpc",
					Port:          DefMap[BRPC_PORT],
					ContainerPort: DefMap[BRPC_PORT],
				},
			},
		},
		{
			name: "test get be service ports 2",
			args: args{
				config: map[string]interface{}{
					BE_PORT:                "1",
					WEBSERVER_PORT:         "2",
					HEARTBEAT_SERVICE_PORT: "3",
					BRPC_PORT:              "4",
				},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "be",
					Port:          1,
					ContainerPort: 1,
				},
				{
					Name:          "webserver",
					Port:          2,
					ContainerPort: 2,
				},
				{
					Name:          "heartbeat",
					Port:          3,
					ContainerPort: 3,
				},
				{
					Name:          "brpc",
					Port:          4,
					ContainerPort: 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSrPorts := getBeServicePorts(tt.args.config, nil); !reflect.DeepEqual(gotSrPorts, tt.wantSrPorts) {
				t.Errorf("getBeServicePorts() = %v, want %v", gotSrPorts, tt.wantSrPorts)
			}
		})
	}
}

func Test_getCnServicePorts(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		wantSrPorts []srapi.StarRocksServicePort
	}{
		{
			name: "test get cn service ports",
			args: args{
				config: map[string]interface{}{},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "thrift",
					Port:          DefMap[THRIFT_PORT],
					ContainerPort: DefMap[THRIFT_PORT],
				},
				{
					Name:          "webserver",
					Port:          DefMap[WEBSERVER_PORT],
					ContainerPort: DefMap[WEBSERVER_PORT],
				},
				{
					Name:          "heartbeat",
					Port:          DefMap[HEARTBEAT_SERVICE_PORT],
					ContainerPort: DefMap[HEARTBEAT_SERVICE_PORT],
				},
				{
					Name:          "brpc",
					Port:          DefMap[BRPC_PORT],
					ContainerPort: DefMap[BRPC_PORT],
				},
			},
		},
		{
			name: "test get cn service ports 2",
			args: args{
				config: map[string]interface{}{
					THRIFT_PORT:            "1",
					WEBSERVER_PORT:         "2",
					HEARTBEAT_SERVICE_PORT: "3",
					BRPC_PORT:              "4",
				},
			},
			wantSrPorts: []srapi.StarRocksServicePort{
				{
					Name:          "thrift",
					Port:          1,
					ContainerPort: 1,
				},
				{
					Name:          "webserver",
					Port:          2,
					ContainerPort: 2,
				},
				{
					Name:          "heartbeat",
					Port:          3,
					ContainerPort: 3,
				},
				{
					Name:          "brpc",
					Port:          4,
					ContainerPort: 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSrPorts := getCnServicePorts(tt.args.config, nil); !reflect.DeepEqual(gotSrPorts, tt.wantSrPorts) {
				t.Errorf("getCnServicePorts() = %v, want %v", gotSrPorts, tt.wantSrPorts)
			}
		})
	}
}
