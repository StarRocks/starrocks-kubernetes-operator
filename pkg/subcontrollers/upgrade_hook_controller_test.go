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
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: srapi.StarRocksClusterSpec{},
			},
			expected: false,
		},
		{
			name: "should not execute if already completed",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationPrepareUpgrade: "true",
					},
				},
				Status: srapi.StarRocksClusterStatus{
					UpgradePreparationStatus: &srapi.UpgradePreparationStatus{
						Phase: srapi.UpgradePreparationCompleted,
					},
				},
			},
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

func TestUpgradeHookController_GetPredefinedHooks(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	uhc := NewUpgradeHookController(client)

	tests := []struct {
		name        string
		hookNames   string
		expectedLen int
		expectedNames []string
	}{
		{
			name:        "single hook",
			hookNames:   "disable-tablet-clone",
			expectedLen: 1,
			expectedNames: []string{"disable-tablet-clone"},
		},
		{
			name:        "multiple hooks",
			hookNames:   "disable-tablet-clone,disable-balancer",
			expectedLen: 2,
			expectedNames: []string{"disable-tablet-clone", "disable-balancer"},
		},
		{
			name:        "unknown hook",
			hookNames:   "unknown-hook",
			expectedLen: 0,
			expectedNames: []string{},
		},
		{
			name:        "mixed known and unknown",
			hookNames:   "disable-tablet-clone,unknown-hook,disable-balancer",
			expectedLen: 2,
			expectedNames: []string{"disable-tablet-clone", "disable-balancer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := uhc.getPredefinedHooks(tt.hookNames)
			assert.Equal(t, tt.expectedLen, len(hooks))
			
			if tt.expectedLen > 0 {
				for i, expectedName := range tt.expectedNames {
					assert.Equal(t, expectedName, hooks[i].Name)
					assert.NotEmpty(t, hooks[i].Command)
				}
			}
		})
	}
}

func TestUpgradeHookController_GetHooksToExecute(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	uhc := NewUpgradeHookController(client)

	tests := []struct {
		name        string
		cluster     *srapi.StarRocksCluster
		expectedLen int
	}{
		{
			name: "spec-based hooks only",
			cluster: &srapi.StarRocksCluster{
				Spec: srapi.StarRocksClusterSpec{
					UpgradePreparation: &srapi.UpgradePreparation{
						Hooks: []srapi.UpgradeHook{
							{Name: "custom-hook", Command: "SELECT 1"},
						},
					},
				},
			},
			expectedLen: 1,
		},
		{
			name: "annotation-based hooks only",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationUpgradeHooks: "disable-tablet-clone",
					},
				},
			},
			expectedLen: 1,
		},
		{
			name: "both spec and annotation hooks",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationUpgradeHooks: "disable-balancer",
					},
				},
				Spec: srapi.StarRocksClusterSpec{
					UpgradePreparation: &srapi.UpgradePreparation{
						Hooks: []srapi.UpgradeHook{
							{Name: "custom-hook", Command: "SELECT 1"},
						},
					},
				},
			},
			expectedLen: 2,
		},
		{
			name: "no hooks",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       srapi.StarRocksClusterSpec{},
			},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks, err := uhc.getHooksToExecute(tt.cluster)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLen, len(hooks))
		})
	}
}

func TestUpgradeHookController_IsUpgradeReady(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	uhc := NewUpgradeHookController(client)

	tests := []struct {
		name     string
		cluster  *srapi.StarRocksCluster
		expected bool
	}{
		{
			name: "no upgrade preparation needed",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       srapi.StarRocksClusterSpec{},
			},
			expected: true,
		},
		{
			name: "upgrade preparation completed",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationPrepareUpgrade: "true",
					},
				},
				Status: srapi.StarRocksClusterStatus{
					UpgradePreparationStatus: &srapi.UpgradePreparationStatus{
						Phase: srapi.UpgradePreparationCompleted,
					},
				},
			},
			expected: true,
		},
		{
			name: "upgrade preparation running",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationPrepareUpgrade: "true",
					},
				},
				Status: srapi.StarRocksClusterStatus{
					UpgradePreparationStatus: &srapi.UpgradePreparationStatus{
						Phase: srapi.UpgradePreparationRunning,
					},
				},
			},
			expected: false,
		},
		{
			name: "upgrade preparation failed",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						AnnotationPrepareUpgrade: "true",
					},
				},
				Status: srapi.StarRocksClusterStatus{
					UpgradePreparationStatus: &srapi.UpgradePreparationStatus{
						Phase: srapi.UpgradePreparationFailed,
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uhc.IsUpgradeReady(context.TODO(), tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}
