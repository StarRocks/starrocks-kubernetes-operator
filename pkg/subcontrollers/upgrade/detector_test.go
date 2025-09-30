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

func TestDetectUpgrade(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	tests := []struct {
		name             string
		cluster          *srapi.StarRocksCluster
		expectUpgrade    bool
		expectTargetFE   string
	}{
		{
			name: "No upgrade - same versions",
			cluster: &srapi.StarRocksCluster{
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
			},
			expectUpgrade:  false,
			expectTargetFE: "",
		},
		{
			name: "FE upgrade detected",
			cluster: &srapi.StarRocksCluster{
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
			},
			expectUpgrade:  true,
			expectTargetFE: "starrocks/fe-ubuntu:3.3.0",
		},
		{
			name: "No status yet - no upgrade",
			cluster: &srapi.StarRocksCluster{
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
				Status: srapi.StarRocksClusterStatus{},
			},
			expectUpgrade:  false,
			expectTargetFE: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.cluster).Build()
			detector := NewDetector(client)

			upgradeDetected, targetVersions, err := detector.DetectUpgrade(context.Background(), tt.cluster)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectUpgrade, upgradeDetected)

			if tt.expectUpgrade {
				assert.NotNil(t, targetVersions)
				assert.Equal(t, tt.expectTargetFE, targetVersions.FeVersion)
			}
		})
	}
}

func TestIsUpgradeInProgress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	tests := []struct {
		name         string
		cluster      *srapi.StarRocksCluster
		expectResult bool
	}{
		{
			name: "No upgrade state",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{},
			},
			expectResult: false,
		},
		{
			name: "Upgrade detected",
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
			name: "Upgrade preparing",
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
			name: "Upgrade in progress",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhaseInProgress,
					},
				},
			},
			expectResult: true,
		},
		{
			name: "Upgrade completed",
			cluster: &srapi.StarRocksCluster{
				Status: srapi.StarRocksClusterStatus{
					UpgradeState: &srapi.UpgradeState{
						Phase: srapi.UpgradePhaseCompleted,
					},
				},
			},
			expectResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).Build()
			detector := NewDetector(client)

			result := detector.IsUpgradeInProgress(tt.cluster)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestInitializeUpgradeState(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	cluster := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Status: srapi.StarRocksClusterStatus{},
	}

	targetVersions := &srapi.ComponentVersions{
		FeVersion: "starrocks/fe-ubuntu:3.3.0",
		BeVersion: "starrocks/be-ubuntu:3.3.0",
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	detector := NewDetector(client)

	detector.InitializeUpgradeState(cluster, targetVersions)

	assert.NotNil(t, cluster.Status.UpgradeState)
	assert.Equal(t, srapi.UpgradePhaseDetected, cluster.Status.UpgradeState.Phase)
	assert.Equal(t, "starrocks/fe-ubuntu:3.3.0", cluster.Status.UpgradeState.TargetVersion.FeVersion)
	assert.Equal(t, "starrocks/be-ubuntu:3.3.0", cluster.Status.UpgradeState.TargetVersion.BeVersion)
	assert.NotNil(t, cluster.Status.UpgradeState.StartTime)
}

func TestIsFeatureEnabled(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)
	_ = srapi.SchemeBuilder.AddToScheme(scheme)

	tests := []struct {
		name         string
		cluster      *srapi.StarRocksCluster
		expectResult bool
	}{
		{
			name: "Feature enabled - has FE spec",
			cluster: &srapi.StarRocksCluster{
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{},
				},
			},
			expectResult: true,
		},
		{
			name: "Feature disabled - no FE spec",
			cluster: &srapi.StarRocksCluster{
				Spec: srapi.StarRocksClusterSpec{},
			},
			expectResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).Build()
			detector := NewDetector(client)

			result := detector.IsFeatureEnabled(tt.cluster)
			assert.Equal(t, tt.expectResult, result)
		})
	}
}