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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
)

func TestMakeSearchService(t *testing.T) {
	type args struct {
		serviceName     string
		externalService *corev1.Service
		ports           []corev1.ServicePort
		labels          map[string]string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Service
	}{
		{
			name: "test MakeSearchService",
			args: args{
				serviceName: "test",
				externalService: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label_to_be_discarded": "test",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"test": "test",
						},
					},
				},
				ports: []corev1.ServicePort{
					{
						Name: "test",
						Port: 18030,
					},
				},
				labels: map[string]string{
					"test": "test",
				},
			},
			want: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"test": "test",
					},
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "None",
					Ports: []corev1.ServicePort{
						{
							Name: "test",
							Port: 18030,
						},
					},
					Selector: map[string]string{
						"test": "test",
					},
					PublishNotReadyAddresses: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeSearchService(tt.args.serviceName,
				tt.args.externalService, tt.args.ports, tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeSearchService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeSearchServiceForSpec(t *testing.T) {
	cluster := &v1.StarRocksCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.GroupVersion.String(),
			Kind:       "StarRocksCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	observerSpec := &v1.StarRocksFeObserverSpec{}
	labels := load.Labels(cluster.Name, observerSpec)
	ports := []corev1.ServicePort{
		{
			Name: "query-port",
			Port: 9030,
		},
	}

	got := MakeSearchServiceForSpec(object.NewFromCluster(cluster), observerSpec, ports, labels)

	if got.Name != "test-fe-observer-search" {
		t.Errorf("MakeSearchServiceForSpec() name = %v, want %v", got.Name, "test-fe-observer-search")
	}
	if got.Namespace != "default" {
		t.Errorf("MakeSearchServiceForSpec() namespace = %v, want %v", got.Namespace, "default")
	}
	if !reflect.DeepEqual(got.Labels, labels) {
		t.Errorf("MakeSearchServiceForSpec() labels = %v, want %v", got.Labels, labels)
	}
	if !reflect.DeepEqual(got.Spec.Selector, load.Selector(cluster.Name, observerSpec)) {
		t.Errorf("MakeSearchServiceForSpec() selector = %v, want %v", got.Spec.Selector, load.Selector(cluster.Name, observerSpec))
	}
	if got.Spec.ClusterIP != "None" {
		t.Errorf("MakeSearchServiceForSpec() ClusterIP = %v, want %v", got.Spec.ClusterIP, "None")
	}
	if !got.Spec.PublishNotReadyAddresses {
		t.Errorf("MakeSearchServiceForSpec() PublishNotReadyAddresses = %v, want %v", got.Spec.PublishNotReadyAddresses, true)
	}
	if !reflect.DeepEqual(got.Spec.Ports, ports) {
		t.Errorf("MakeSearchServiceForSpec() ports = %v, want %v", got.Spec.Ports, ports)
	}
	if len(got.OwnerReferences) != 1 || got.OwnerReferences[0].Name != "test" || got.OwnerReferences[0].Kind != "StarRocksCluster" {
		t.Errorf("MakeSearchServiceForSpec() OwnerReferences = %v, want one StarRocksCluster owner named test", got.OwnerReferences)
	}
}

func TestSearchServiceName(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test SearchServiceName for be",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksBeSpec{},
			},
			want: "test-be-search",
		},
		{
			name: "test SearchServiceName for cn",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksCnSpec{},
			},
			want: "test-cn-search",
		},
		{
			name: "test SearchServiceName for fe",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: "test-fe-search",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchServiceName(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("SearchServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchServiceName_WithNil(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test SearchServiceName for be",
			args: args{
				clusterName: "test",
				spec:        (*v1.StarRocksBeSpec)(nil),
			},
			want: "test-be-search",
		},
		{
			name: "test SearchServiceName for cn",
			args: args{
				clusterName: "test",
				spec:        (*v1.StarRocksCnSpec)(nil),
			},
			want: "test-cn-search",
		},
		{
			name: "test SearchServiceName for fe",
			args: args{
				clusterName: "test",
				spec:        (*v1.StarRocksFeSpec)(nil),
			},
			want: "test-fe-search",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchServiceName(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("SearchServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFeExternalServiceName(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// NOTE: we must not input a nil value for spec, otherwise the following error will occur:
		// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
		// {
		//	name: "test1",
		//	args: args{
		//		clusterName: "test",
		//		spec:        (*StarRocksFeSpec)(nil),
		//	},
		//	want: "test-fe-service",
		// },
		{
			name: "test2",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: "test-fe-service",
		},
		{
			name: "fe observer does not have external service",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeObserverSpec{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExternalServiceName(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("GetExternalServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
