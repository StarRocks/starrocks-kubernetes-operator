/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package upgrade

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestReconcileUpgrade_NoUpgrade(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	cluster := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image: "starrocks/fe-ubuntu:3.2.0",
					},
				},
			},
		},
		Status: srapi.StarRocksClusterStatus{
			StarRocksFeStatus: &srapi.StarRocksFeStatus{
				StarRocksComponentStatus: srapi.StarRocksComponentStatus{
					Phase: srapi.ComponentRunning,
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cluster).Build()
	manager := NewManager(client)

	shouldRequeue, err := manager.ReconcileUpgrade(context.Background(), cluster)

	assert.NoError(t, err)
	assert.False(t, shouldRequeue)
	assert.Nil(t, cluster.Status.UpgradeState)
}

func TestReconcileUpgrade_NewUpgradeDetected(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	cluster := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image: "starrocks/fe-ubuntu:3.3.0",
					},
				},
			},
		},
		Status: srapi.StarRocksClusterStatus{
			StarRocksFeStatus: &srapi.StarRocksFeStatus{
				StarRocksComponentStatus: srapi.StarRocksComponentStatus{
					Phase: srapi.ComponentRunning,
				},
			},
			UpgradeState: &srapi.UpgradeState{
				CurrentVersion: srapi.ComponentVersions{
					FeVersion: "starrocks/fe-ubuntu:3.2.0",
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cluster).Build()
	manager := NewManager(client)

	shouldRequeue, err := manager.ReconcileUpgrade(context.Background(), cluster)

	assert.NoError(t, err)
	assert.True(t, shouldRequeue)
	assert.NotNil(t, cluster.Status.UpgradeState)
	assert.Equal(t, srapi.UpgradePhaseDetected, cluster.Status.UpgradeState.Phase)
}

func TestShouldBlockReconciliation(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	tests := []struct {
		name         string
		cluster      *srapi.StarRocksCluster
		expectResult bool
	}{
		{
			name: "No upgrade state - don't block",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{},
			},
			expectResult: false,
		},
		{
			name: "Upgrade detected - block",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhaseDetected,
					},
				},
			},
			expectResult: true,
		},
		{
			name: "Upgrade preparing - block",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhasePreparing,
					},
				},
			},
			expectResult: true,
		},
		{
			name: "Upgrade ready - don't block",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhaseReady,
					},
				},
			},
			expectResult: false,
		},
		{
			name: "Upgrade in progress - don't block",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhaseInProgress,
					},
				},
			},
			expectResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).Build()
			manager := NewManager(client)

			result := manager.ShouldBlockReconciliation(tt.cluster)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}