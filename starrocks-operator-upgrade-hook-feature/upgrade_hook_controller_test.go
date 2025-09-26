package subcontrollers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestUpgradeHookController_ShouldExecuteUpgradeHooks(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)

	tests := []struct {
		name     string
		cluster  *srapi.StarRocksCluster
		expected bool
	}{
		{
			name: "should execute with annotation",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationPrepareUpgrade: "true",
					},
				},
			},
			expected: true,
		},
		{
			name: "should execute with spec enabled",
			cluster: &srapi.StarRocksCluster{
				Spec: srapi.StarRocksClusterSpec{
					UpgradePreparation: &srapi.UpgradePreparation{
						Enabled: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "should not execute without triggers",
			cluster: &srapi.StarRocksCluster{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).Build()
			uhc := NewUpgradeHookController(client)

			result := uhc.ShouldExecuteUpgradeHooks(context.TODO(), tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpgradeHookController_ParseHookAnnotation(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	uhc := NewUpgradeHookController(client)

	tests := []struct {
		name        string
		annotation  string
		expectedLen int
	}{
		{
			name:        "single hook",
			annotation:  "disable-tablet-clone",
			expectedLen: 1,
		},
		{
			name:        "multiple hooks",
			annotation:  "disable-tablet-clone,disable-balancer",
			expectedLen: 2,
		},
		{
			name:        "unknown hook",
			annotation:  "unknown-hook",
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := uhc.parseHookAnnotation(tt.annotation)
			assert.Equal(t, tt.expectedLen, len(hooks))
		})
	}
}
