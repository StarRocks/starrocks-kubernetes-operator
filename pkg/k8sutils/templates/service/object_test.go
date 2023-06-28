package service

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestMakeSearchService(t *testing.T) {
	type args struct {
		serviceName     string
		spec            interface{}
		externalService *corev1.Service
		ports           []corev1.ServicePort
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
				spec:        &v1.StarRocksBeSpec{},
				externalService: &corev1.Service{
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
			},
			want: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
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
			if got := MakeSearchService(tt.args.serviceName, tt.args.spec, tt.args.externalService, tt.args.ports); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeSearchService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchServiceName(t *testing.T) {
	type args struct {
		clusterName string
		spec        interface{}
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
		spec        interface{}
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
