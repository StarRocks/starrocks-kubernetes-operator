/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package upgrade

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// Manager coordinates the entire upgrade process
type Manager struct {
	Client       client.Client
	Detector     *Detector
	HookExecutor *HookExecutor
}

// NewManager creates a new upgrade manager
func NewManager(k8sClient client.Client) *Manager {
	return &Manager{
		Client:       k8sClient,
		Detector:     NewDetector(k8sClient),
		HookExecutor: NewHookExecutor(),
	}
}

// ReconcileUpgrade handles the upgrade reconciliation logic
// Returns: (shouldRequeue, error)
// Security: This function is the main entry point for upgrade management and ensures
// all upgrade operations are performed safely and securely
func (m *Manager) ReconcileUpgrade(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	// Security: Validate input
	if src == nil {
		return false, fmt.Errorf("cluster object cannot be nil")
	}

	// Check if automatic upgrade management is enabled
	if !m.Detector.IsFeatureEnabled(src) {
		return false, nil
	}

	// Step 1: Check if upgrade is already in progress
	if m.Detector.IsUpgradeInProgress(src) {
		return m.handleUpgradeInProgress(ctx, src)
	}

	// Step 2: Detect if a new upgrade is needed
	upgradeDetected, targetVersions, err := m.Detector.DetectUpgrade(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to detect upgrade")
		return false, err
	}

	if !upgradeDetected {
		// No upgrade needed
		return false, nil
	}

	// Step 3: Initialize upgrade state
	logger.Info("New upgrade detected, initializing upgrade state")
	m.Detector.InitializeUpgradeState(src, targetVersions)

	// Requeue to handle the upgrade in the next reconciliation
	return true, nil
}

// handleUpgradeInProgress handles an upgrade that is already in progress
func (m *Manager) handleUpgradeInProgress(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	if src.Status.UpgradeState == nil {
		return false, nil
	}

	switch src.Status.UpgradeState.Phase {
	case srapi.UpgradePhaseDetected:
		// Execute pre-upgrade hooks
		return m.executePreUpgradeHooks(ctx, src)

	case srapi.UpgradePhasePreparing:
		// Still preparing, requeue
		logger.Info("Upgrade preparation in progress")
		return true, nil

	case srapi.UpgradePhaseReady:
		// Hooks completed, mark as in progress and let normal reconciliation proceed
		logger.Info("Pre-upgrade hooks completed, proceeding with upgrade")
		m.Detector.MarkUpgradeInProgress(src)
		return false, nil

	case srapi.UpgradePhaseInProgress:
		// Check if upgrade is complete
		if m.Detector.CheckUpgradeCompletion(ctx, src) {
			logger.Info("Upgrade completed, executing post-upgrade hooks")
			return m.executePostUpgradeHooks(ctx, src)
		}
		// Still upgrading, let normal reconciliation continue
		return false, nil

	case srapi.UpgradePhaseCompleted:
		// Upgrade complete, clear state
		logger.Info("Upgrade fully completed, clearing state")
		m.Detector.ClearUpgradeState(src)
		return false, nil

	case srapi.UpgradePhaseFailed:
		// Upgrade failed, leave state for debugging
		logger.Error(nil, "Upgrade failed", "reason", src.Status.UpgradeState.Reason)
		return false, fmt.Errorf("upgrade failed: %s", src.Status.UpgradeState.Reason)

	default:
		return false, nil
	}
}

// executePreUpgradeHooks executes pre-upgrade hooks with security validations
func (m *Manager) executePreUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")
	logger.Info("Executing pre-upgrade hooks")

	// Update phase to preparing
	src.Status.UpgradeState.Phase = srapi.UpgradePhasePreparing

	// Security: Get FE connection with validation
	db, err := m.HookExecutor.GetFEConnection(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to connect to FE for pre-upgrade hooks")
		// If FE is not ready, requeue and try again (don't fail immediately)
		return true, nil
	}
	// Security: Ensure database connection is always closed
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close database connection")
		}
	}()

	// Validate FE is ready
	if err := m.HookExecutor.ValidateFEReady(ctx, db); err != nil {
		logger.Error(err, "FE not ready for pre-upgrade hooks")
		// Requeue and try again
		return true, nil
	}

	// Execute pre-upgrade hooks
	if err := m.HookExecutor.ExecutePreUpgradeHooks(ctx, src, db); err != nil {
		logger.Error(err, "Pre-upgrade hooks failed")
		src.Status.UpgradeState.Phase = srapi.UpgradePhaseFailed
		src.Status.UpgradeState.Reason = fmt.Sprintf("Pre-upgrade hooks failed: %v", err)
		return false, err
	}

	// Mark preparation as complete
	m.Detector.MarkPreparationComplete(src)
	logger.Info("Pre-upgrade hooks completed successfully")

	// Requeue to proceed with upgrade
	return true, nil
}

// executePostUpgradeHooks executes post-upgrade cleanup hooks (non-critical)
func (m *Manager) executePostUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")
	logger.Info("Executing post-upgrade hooks")

	// Security: Get FE connection with validation
	db, err := m.HookExecutor.GetFEConnection(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to connect to FE for post-upgrade hooks")
		// Post-upgrade hooks are non-critical, mark as complete anyway
		m.Detector.MarkUpgradeComplete(src)
		return false, nil
	}
	// Security: Ensure database connection is always closed
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close database connection")
		}
	}()

	// Execute post-upgrade hooks (non-critical)
	if err := m.HookExecutor.ExecutePostUpgradeHooks(ctx, src, db); err != nil {
		logger.Error(err, "Post-upgrade hooks failed (non-critical)")
		// Continue anyway
	}

	// Mark upgrade as complete
	m.Detector.MarkUpgradeComplete(src)
	logger.Info("Upgrade process completed successfully")

	return false, nil
}

// ShouldBlockReconciliation returns true if normal reconciliation should be blocked
// This happens when pre-upgrade hooks need to be executed
func (m *Manager) ShouldBlockReconciliation(src *srapi.StarRocksCluster) bool {
	if src.Status.UpgradeState == nil {
		return false
	}

	// Block reconciliation during these phases
	switch src.Status.UpgradeState.Phase {
	case srapi.UpgradePhaseDetected, srapi.UpgradePhasePreparing:
		return true
	default:
		return false
	}
}