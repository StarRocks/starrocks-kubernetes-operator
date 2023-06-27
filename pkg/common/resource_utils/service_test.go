package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func Test_BuildExternalService(t *testing.T) {
	src := srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				Service: &srapi.StarRocksService{
					Type:           corev1.ServiceTypeLoadBalancer,
					LoadBalancerIP: "127.0.0.1",
				},
			},
		},
	}

	svc := BuildExternalService(&src, "test", FeService, make(map[string]interface{}), make(map[string]string), make(map[string]string))
	require.Equal(t, corev1.ServiceTypeLoadBalancer, svc.Spec.Type)
}

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
