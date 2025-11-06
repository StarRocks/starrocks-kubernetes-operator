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

// ComponentType represents the type of component
type ComponentType string

const (
	ComponentTypeFE ComponentType = "FE"
	ComponentTypeBE ComponentType = "BE"
	ComponentTypeCN ComponentType = "CN"
)

// ComponentUpgradeInfo holds upgrade information for a component
type ComponentUpgradeInfo struct {
	ComponentType  ComponentType
	CurrentVersion string
	DesiredVersion string
}

// DetectUpgrade checks if an upgrade is happening by comparing spec images with status images
// Returns: (upgradeDetected, upgradeInfoList, error)
func (d *Detector) DetectUpgrade(ctx context.Context, src *srapi.StarRocksCluster) (bool, []ComponentUpgradeInfo, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeDetector")

	upgradeInfoList := []ComponentUpgradeInfo{}

	// Check FE upgrade
	if src.Spec.StarRocksFeSpec != nil {
		currentVersion := d.getCurrentComponentVersion(src, ComponentTypeFE)
		desiredVersion := src.Spec.StarRocksFeSpec.Image

		if currentVersion != "" && desiredVersion != "" && currentVersion != desiredVersion {
			logger.Info("FE upgrade detected",
				"current", currentVersion,
				"desired", desiredVersion)
			upgradeInfoList = append(upgradeInfoList, ComponentUpgradeInfo{
				ComponentType:  ComponentTypeFE,
				CurrentVersion: currentVersion,
				DesiredVersion: desiredVersion,
			})
		}
	}

	// Check BE upgrade
	if src.Spec.StarRocksBeSpec != nil {
		currentVersion := d.getCurrentComponentVersion(src, ComponentTypeBE)
		desiredVersion := src.Spec.StarRocksBeSpec.Image

		if currentVersion != "" && desiredVersion != "" && currentVersion != desiredVersion {
			logger.Info("BE upgrade detected",
				"current", currentVersion,
				"desired", desiredVersion)
			upgradeInfoList = append(upgradeInfoList, ComponentUpgradeInfo{
				ComponentType:  ComponentTypeBE,
				CurrentVersion: currentVersion,
				DesiredVersion: desiredVersion,
			})
		}
	}

	// Check CN upgrade
	if src.Spec.StarRocksCnSpec != nil {
		currentVersion := d.getCurrentComponentVersion(src, ComponentTypeCN)
		desiredVersion := src.Spec.StarRocksCnSpec.Image

		if currentVersion != "" && desiredVersion != "" && currentVersion != desiredVersion {
			logger.Info("CN upgrade detected",
				"current", currentVersion,
				"desired", desiredVersion)
			upgradeInfoList = append(upgradeInfoList, ComponentUpgradeInfo{
				ComponentType:  ComponentTypeCN,
				CurrentVersion: currentVersion,
				DesiredVersion: desiredVersion,
			})
		}
	}

	return len(upgradeInfoList) > 0, upgradeInfoList, nil
}

// getCurrentComponentVersion retrieves the current running version for a specific component
func (d *Detector) getCurrentComponentVersion(src *srapi.StarRocksCluster, componentType ComponentType) string {
	switch componentType {
	case ComponentTypeFE:
		if src.Status.StarRocksFeStatus != nil {
			if src.Status.StarRocksFeStatus.UpgradeState != nil {
				return src.Status.StarRocksFeStatus.UpgradeState.CurrentVersion
			}
			// If no upgrade state, and component is running, assume current version is what's in spec
			if src.Status.StarRocksFeStatus.Phase == srapi.ComponentRunning && src.Spec.StarRocksFeSpec != nil {
				return src.Spec.StarRocksFeSpec.Image
			}
		}

	case ComponentTypeBE:
		if src.Status.StarRocksBeStatus != nil {
			if src.Status.StarRocksBeStatus.UpgradeState != nil {
				return src.Status.StarRocksBeStatus.UpgradeState.CurrentVersion
			}
			if src.Status.StarRocksBeStatus.Phase == srapi.ComponentRunning && src.Spec.StarRocksBeSpec != nil {
				return src.Spec.StarRocksBeSpec.Image
			}
		}

	case ComponentTypeCN:
		if src.Status.StarRocksCnStatus != nil {
			if src.Status.StarRocksCnStatus.UpgradeState != nil {
				return src.Status.StarRocksCnStatus.UpgradeState.CurrentVersion
			}
			if src.Status.StarRocksCnStatus.Phase == srapi.ComponentRunning && src.Spec.StarRocksCnSpec != nil {
				return src.Spec.StarRocksCnSpec.Image
			}
		}
	}

	return ""
}

// IsUpgradeInProgress checks if an upgrade is currently in progress for any component
func (d *Detector) IsUpgradeInProgress(src *srapi.StarRocksCluster) bool {
	// Check FE
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if d.isUpgradePhaseActive(src.Status.StarRocksFeStatus.UpgradeState.Phase) {
			return true
		}
	}

	// Check BE
	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if d.isUpgradePhaseActive(src.Status.StarRocksBeStatus.UpgradeState.Phase) {
			return true
		}
	}

	// Check CN
	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if d.isUpgradePhaseActive(src.Status.StarRocksCnStatus.UpgradeState.Phase) {
			return true
		}
	}

	return false
}

// isUpgradePhaseActive checks if a phase represents an active upgrade
func (d *Detector) isUpgradePhaseActive(phase srapi.UpgradePhase) bool {
	switch phase {
	case srapi.UpgradePhaseDetected,
		srapi.UpgradePhasePreparing,
		srapi.UpgradePhaseReady,
		srapi.UpgradePhaseInProgress:
		return true
	default:
		return false
	}
}

// InitializeUpgradeState initializes the upgrade state for the specified components
func (d *Detector) InitializeUpgradeState(src *srapi.StarRocksCluster, upgradeInfoList []ComponentUpgradeInfo) {
	for _, info := range upgradeInfoList {
		d.initializeComponentUpgradeState(src, info)
	}
}

// initializeComponentUpgradeState initializes upgrade state for a single component
func (d *Detector) initializeComponentUpgradeState(src *srapi.StarRocksCluster, info ComponentUpgradeInfo) {
	upgradeState := &srapi.UpgradeState{
		Phase:          srapi.UpgradePhaseDetected,
		Reason:         "Upgrade detected, preparing to execute pre-upgrade hooks",
		TargetVersion:  info.DesiredVersion,
		CurrentVersion: info.CurrentVersion,
		StartTime:      &metav1.Time{Time: metav1.Now().Time},
	}

	switch info.ComponentType {
	case ComponentTypeFE:
		if src.Status.StarRocksFeStatus == nil {
			src.Status.StarRocksFeStatus = &srapi.StarRocksFeStatus{}
		}
		src.Status.StarRocksFeStatus.UpgradeState = upgradeState

	case ComponentTypeBE:
		if src.Status.StarRocksBeStatus == nil {
			src.Status.StarRocksBeStatus = &srapi.StarRocksBeStatus{}
		}
		src.Status.StarRocksBeStatus.UpgradeState = upgradeState

	case ComponentTypeCN:
		if src.Status.StarRocksCnStatus == nil {
			src.Status.StarRocksCnStatus = &srapi.StarRocksCnStatus{}
		}
		src.Status.StarRocksCnStatus.UpgradeState = upgradeState
	}
}

// ShouldExecutePreUpgradeHooks determines if pre-upgrade hooks should be executed
func (d *Detector) ShouldExecutePreUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) bool {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeDetector")

	// Check if any component is in detected phase
	components := d.getComponentsInPhase(src, srapi.UpgradePhaseDetected)
	if len(components) > 0 {
		logger.Info("Upgrade detected for components, pre-upgrade hooks should be executed", "components", components)
		return true
	}

	return false
}

// getComponentsInPhase returns a list of components in a specific phase
func (d *Detector) getComponentsInPhase(src *srapi.StarRocksCluster, phase srapi.UpgradePhase) []ComponentType {
	components := []ComponentType{}

	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == phase {
			components = append(components, ComponentTypeFE)
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == phase {
			components = append(components, ComponentTypeBE)
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == phase {
			components = append(components, ComponentTypeCN)
		}
	}

	return components
}

// MarkPreparationComplete marks the upgrade preparation as complete for all components in detected phase
func (d *Detector) MarkPreparationComplete(src *srapi.StarRocksCluster) {
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksFeStatus.UpgradeState.Phase = srapi.UpgradePhaseReady
			src.Status.StarRocksFeStatus.UpgradeState.Reason = "Pre-upgrade hooks executed successfully, ready to proceed"
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksBeStatus.UpgradeState.Phase = srapi.UpgradePhaseReady
			src.Status.StarRocksBeStatus.UpgradeState.Reason = "Pre-upgrade hooks executed successfully, ready to proceed"
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksCnStatus.UpgradeState.Phase = srapi.UpgradePhaseReady
			src.Status.StarRocksCnStatus.UpgradeState.Reason = "Pre-upgrade hooks executed successfully, ready to proceed"
		}
	}
}

// MarkUpgradeInProgress marks the upgrade as in progress for ready components
func (d *Detector) MarkUpgradeInProgress(src *srapi.StarRocksCluster) {
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseReady {
			src.Status.StarRocksFeStatus.UpgradeState.Phase = srapi.UpgradePhaseInProgress
			src.Status.StarRocksFeStatus.UpgradeState.Reason = "Upgrade in progress"
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseReady {
			src.Status.StarRocksBeStatus.UpgradeState.Phase = srapi.UpgradePhaseInProgress
			src.Status.StarRocksBeStatus.UpgradeState.Reason = "Upgrade in progress"
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseReady {
			src.Status.StarRocksCnStatus.UpgradeState.Phase = srapi.UpgradePhaseInProgress
			src.Status.StarRocksCnStatus.UpgradeState.Reason = "Upgrade in progress"
		}
	}
}

// CheckUpgradeCompletion checks if upgrades have completed for all components
func (d *Detector) CheckUpgradeCompletion(ctx context.Context, src *srapi.StarRocksCluster) bool {
	allCompleted := true

	// Check FE
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeFE)
			targetVersion := src.Status.StarRocksFeStatus.UpgradeState.TargetVersion
			if currentVersion != targetVersion {
				allCompleted = false
			}
		}
	}

	// Check BE
	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeBE)
			targetVersion := src.Status.StarRocksBeStatus.UpgradeState.TargetVersion
			if currentVersion != targetVersion {
				allCompleted = false
			}
		}
	}

	// Check CN
	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeCN)
			targetVersion := src.Status.StarRocksCnStatus.UpgradeState.TargetVersion
			if currentVersion != targetVersion {
				allCompleted = false
			}
		}
	}

	return allCompleted
}

// MarkUpgradeComplete marks upgrades as complete for all components that finished
func (d *Detector) MarkUpgradeComplete(src *srapi.StarRocksCluster) {
	now := &metav1.Time{Time: metav1.Now().Time}

	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeFE)
			if currentVersion == src.Status.StarRocksFeStatus.UpgradeState.TargetVersion {
				src.Status.StarRocksFeStatus.UpgradeState.Phase = srapi.UpgradePhaseCompleted
				src.Status.StarRocksFeStatus.UpgradeState.Reason = "Upgrade completed successfully"
				src.Status.StarRocksFeStatus.UpgradeState.CompletionTime = now
			}
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeBE)
			if currentVersion == src.Status.StarRocksBeStatus.UpgradeState.TargetVersion {
				src.Status.StarRocksBeStatus.UpgradeState.Phase = srapi.UpgradePhaseCompleted
				src.Status.StarRocksBeStatus.UpgradeState.Reason = "Upgrade completed successfully"
				src.Status.StarRocksBeStatus.UpgradeState.CompletionTime = now
			}
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseInProgress {
			currentVersion := d.getCurrentComponentVersion(src, ComponentTypeCN)
			if currentVersion == src.Status.StarRocksCnStatus.UpgradeState.TargetVersion {
				src.Status.StarRocksCnStatus.UpgradeState.Phase = srapi.UpgradePhaseCompleted
				src.Status.StarRocksCnStatus.UpgradeState.Reason = "Upgrade completed successfully"
				src.Status.StarRocksCnStatus.UpgradeState.CompletionTime = now
			}
		}
	}
}

// ClearUpgradeState clears the upgrade state after successful completion
func (d *Detector) ClearUpgradeState(src *srapi.StarRocksCluster) {
	// Clear completed upgrade states
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseCompleted {
			src.Status.StarRocksFeStatus.UpgradeState = nil
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseCompleted {
			src.Status.StarRocksBeStatus.UpgradeState = nil
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseCompleted {
			src.Status.StarRocksCnStatus.UpgradeState = nil
		}
	}
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