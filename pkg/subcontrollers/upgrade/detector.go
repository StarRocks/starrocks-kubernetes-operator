/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package upgrade

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// Detector detects upgrades by comparing current and desired component versions
type Detector struct {
	Client client.Client
}

// NewDetector creates a new upgrade detector
func NewDetector(k8sClient client.Client) *Detector {
	return &Detector{
		Client: k8sClient,
	}
}

// DetectUpgrade checks if an upgrade is happening by comparing spec images with status images
func (d *Detector) DetectUpgrade(ctx context.Context, src *srapi.StarRocksCluster) (bool, *srapi.ComponentVersions, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeDetector")

	// Get current versions from running pods/statefulsets
	currentVersions, err := d.getCurrentVersions(ctx, src)
	if err != nil {
		return false, nil, err
	}

	// Get desired versions from spec
	desiredVersions := d.getDesiredVersions(src)

	// Check if there's a version mismatch (upgrade needed)
	upgradeDetected := false

	if currentVersions.FeVersion != "" && desiredVersions.FeVersion != "" {
		if currentVersions.FeVersion != desiredVersions.FeVersion {
			logger.Info("FE upgrade detected",
				"current", currentVersions.FeVersion,
				"desired", desiredVersions.FeVersion)
			upgradeDetected = true
		}
	}

	if currentVersions.BeVersion != "" && desiredVersions.BeVersion != "" {
		if currentVersions.BeVersion != desiredVersions.BeVersion {
			logger.Info("BE upgrade detected",
				"current", currentVersions.BeVersion,
				"desired", desiredVersions.BeVersion)
			upgradeDetected = true
		}
	}

	if currentVersions.CnVersion != "" && desiredVersions.CnVersion != "" {
		if currentVersions.CnVersion != desiredVersions.CnVersion {
			logger.Info("CN upgrade detected",
				"current", currentVersions.CnVersion,
				"desired", desiredVersions.CnVersion)
			upgradeDetected = true
		}
	}

	if upgradeDetected {
		return true, &desiredVersions, nil
	}

	return false, nil, nil
}

// getCurrentVersions retrieves the current running versions from the upgrade state or empty if none
func (d *Detector) getCurrentVersions(ctx context.Context, src *srapi.StarRocksCluster) (srapi.ComponentVersions, error) {
	versions := srapi.ComponentVersions{}

	// If we have an upgrade state, use the current version from there
	if src.Status.UpgradeState != nil {
		return src.Status.UpgradeState.CurrentVersion, nil
	}

	// Otherwise, if components are running, assume current version is what's in spec
	// (this means no upgrade is in progress)
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.Phase == srapi.ComponentRunning {
		if src.Spec.StarRocksFeSpec != nil {
			versions.FeVersion = src.Spec.StarRocksFeSpec.Image
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.Phase == srapi.ComponentRunning {
		if src.Spec.StarRocksBeSpec != nil {
			versions.BeVersion = src.Spec.StarRocksBeSpec.Image
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.Phase == srapi.ComponentRunning {
		if src.Spec.StarRocksCnSpec != nil {
			versions.CnVersion = src.Spec.StarRocksCnSpec.Image
		}
	}

	return versions, nil
}

// getDesiredVersions gets the desired versions from the spec
func (d *Detector) getDesiredVersions(src *srapi.StarRocksCluster) srapi.ComponentVersions {
	versions := srapi.ComponentVersions{}

	if src.Spec.StarRocksFeSpec != nil {
		versions.FeVersion = src.Spec.StarRocksFeSpec.Image
	}

	if src.Spec.StarRocksBeSpec != nil {
		versions.BeVersion = src.Spec.StarRocksBeSpec.Image
	}

	if src.Spec.StarRocksCnSpec != nil {
		versions.CnVersion = src.Spec.StarRocksCnSpec.Image
	}

	return versions
}

// IsUpgradeInProgress checks if an upgrade is currently in progress
func (d *Detector) IsUpgradeInProgress(src *srapi.StarRocksCluster) bool {
	if src.Status.UpgradeState == nil {
		return false
	}

	switch src.Status.UpgradeState.Phase {
	case srapi.UpgradePhaseDetected,
		srapi.UpgradePhasePreparing,
		srapi.UpgradePhaseReady,
		srapi.UpgradePhaseInProgress:
		return true
	default:
		return false
	}
}

// InitializeUpgradeState initializes the upgrade state in the cluster status
func (d *Detector) InitializeUpgradeState(src *srapi.StarRocksCluster, targetVersions *srapi.ComponentVersions) {
	currentVersions, _ := d.getCurrentVersions(context.Background(), src)

	src.Status.UpgradeState = &srapi.UpgradeState{
		Phase:          srapi.UpgradePhaseDetected,
		Reason:         "Upgrade detected, preparing to execute pre-upgrade hooks",
		TargetVersion:  *targetVersions,
		CurrentVersion: currentVersions,
		StartTime:      &metav1.Time{Time: metav1.Now().Time},
	}
}

// ShouldExecutePreUpgradeHooks determines if pre-upgrade hooks should be executed
func (d *Detector) ShouldExecutePreUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) bool {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeDetector")

	// Only execute hooks if upgrade is detected or preparing
	if src.Status.UpgradeState == nil {
		return false
	}

	if src.Status.UpgradeState.Phase == srapi.UpgradePhaseDetected {
		logger.Info("Upgrade detected, pre-upgrade hooks should be executed")
		return true
	}

	return false
}

// MarkPreparationComplete marks the upgrade preparation as complete
func (d *Detector) MarkPreparationComplete(src *srapi.StarRocksCluster) {
	if src.Status.UpgradeState != nil {
		src.Status.UpgradeState.Phase = srapi.UpgradePhaseReady
		src.Status.UpgradeState.Reason = "Pre-upgrade hooks executed successfully, ready to proceed"
	}
}

// MarkUpgradeInProgress marks the upgrade as in progress
func (d *Detector) MarkUpgradeInProgress(src *srapi.StarRocksCluster) {
	if src.Status.UpgradeState != nil {
		src.Status.UpgradeState.Phase = srapi.UpgradePhaseInProgress
		src.Status.UpgradeState.Reason = "Upgrade in progress"
	}
}

// CheckUpgradeCompletion checks if the upgrade has completed
func (d *Detector) CheckUpgradeCompletion(ctx context.Context, src *srapi.StarRocksCluster) bool {
	if src.Status.UpgradeState == nil {
		return false
	}

	// Check if all components have reached their target versions
	currentVersions, _ := d.getCurrentVersions(ctx, src)
	targetVersions := src.Status.UpgradeState.TargetVersion

	feComplete := true
	beComplete := true
	cnComplete := true

	if targetVersions.FeVersion != "" {
		feComplete = currentVersions.FeVersion == targetVersions.FeVersion
	}

	if targetVersions.BeVersion != "" {
		beComplete = currentVersions.BeVersion == targetVersions.BeVersion
	}

	if targetVersions.CnVersion != "" {
		cnComplete = currentVersions.CnVersion == targetVersions.CnVersion
	}

	return feComplete && beComplete && cnComplete
}

// MarkUpgradeComplete marks the upgrade as complete and clears the upgrade state
func (d *Detector) MarkUpgradeComplete(src *srapi.StarRocksCluster) {
	if src.Status.UpgradeState != nil {
		src.Status.UpgradeState.Phase = srapi.UpgradePhaseCompleted
		src.Status.UpgradeState.Reason = "Upgrade completed successfully"
		src.Status.UpgradeState.CompletionTime = &metav1.Time{Time: metav1.Now().Time}
	}
}

// ClearUpgradeState clears the upgrade state after successful completion
func (d *Detector) ClearUpgradeState(src *srapi.StarRocksCluster) {
	// Keep the state for observability but mark as complete
	// Alternatively, you could set it to nil to fully clear it
	// src.Status.UpgradeState = nil
}

// IsFeatureEnabled checks if the automatic upgrade feature is enabled
// For now, we'll enable it by default for all clusters
func (d *Detector) IsFeatureEnabled(src *srapi.StarRocksCluster) bool {
	// This could be controlled by a feature flag or annotation in the future
	// For now, enable for all clusters that have FE component
	return src.Spec.StarRocksFeSpec != nil
}

// GetFEPods returns the list of FE pods for connection
func (d *Detector) GetFEPods(ctx context.Context, src *srapi.StarRocksCluster) (*corev1.PodList, error) {
	podList := &corev1.PodList{}

	labels := map[string]string{
		"app.kubernetes.io/name":      "starrocks",
		"app.kubernetes.io/component": "fe",
		"app.kubernetes.io/instance":  src.Name,
	}

	err := d.Client.List(ctx, podList,
		client.InNamespace(src.Namespace),
		client.MatchingLabels(labels))

	return podList, err
}