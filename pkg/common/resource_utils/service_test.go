package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	svc := BuildExternalService(&src, "test", FeService, make(map[string]interface{}))
	require.Equal(t, corev1.ServiceTypeLoadBalancer, svc.Spec.Type)
}
