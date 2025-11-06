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
	Client             client.Client
	Detector           *Detector
	HookExecutor       *HookExecutor
	CustomHookExecutor *CustomHookExecutor
}

// NewManager creates a new upgrade manager
func NewManager(k8sClient client.Client) *Manager {
	return &Manager{
		Client:             k8sClient,
		Detector:           NewDetector(k8sClient),
		HookExecutor:       NewHookExecutor(),
		CustomHookExecutor: NewCustomHookExecutor(k8sClient),
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
	upgradeDetected, upgradeInfoList, err := m.Detector.DetectUpgrade(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to detect upgrade")
		return false, err
	}

	if !upgradeDetected {
		// No upgrade needed
		return false, nil
	}

	// Step 3: Initialize upgrade state for detected components
	logger.Info("New upgrade detected, initializing upgrade state", "componentCount", len(upgradeInfoList))
	m.Detector.InitializeUpgradeState(src, upgradeInfoList)

	// Requeue to handle the upgrade in the next reconciliation
	return true, nil
}

// handleUpgradeInProgress handles an upgrade that is already in progress
func (m *Manager) handleUpgradeInProgress(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	// Check what phase components are in
	hasDetected := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseDetected)) > 0
	hasPreparing := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhasePreparing)) > 0
	hasReady := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseReady)) > 0
	hasInProgress := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseInProgress)) > 0
	hasCompleted := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseCompleted)) > 0
	hasFailed := len(m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseFailed)) > 0

	// Handle based on phase priority
	if hasDetected {
		// Execute pre-upgrade hooks for detected components
		return m.executePreUpgradeHooks(ctx, src)
	}

	if hasPreparing {
		// Still preparing, requeue
		logger.Info("Upgrade preparation in progress")
		return true, nil
	}

	if hasReady {
		// Hooks completed, mark as in progress and let normal reconciliation proceed
		logger.Info("Pre-upgrade hooks completed, proceeding with upgrade")
		m.Detector.MarkUpgradeInProgress(src)
		return false, nil
	}

	if hasInProgress {
		// Check if upgrade is complete
		if m.Detector.CheckUpgradeCompletion(ctx, src) {
			logger.Info("Upgrade completed, executing post-upgrade hooks")
			return m.executePostUpgradeHooks(ctx, src)
		}
		// Still upgrading, let normal reconciliation continue
		return false, nil
	}

	if hasCompleted {
		// Upgrade complete, clear state
		logger.Info("Upgrade fully completed, clearing state")
		m.Detector.ClearUpgradeState(src)
		return false, nil
	}

	if hasFailed {
		// Get failure reason from any failed component
		var reason string
		if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil &&
			src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseFailed {
			reason = src.Status.StarRocksFeStatus.UpgradeState.Reason
		} else if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil &&
			src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseFailed {
			reason = src.Status.StarRocksBeStatus.UpgradeState.Reason
		} else if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil &&
			src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseFailed {
			reason = src.Status.StarRocksCnStatus.UpgradeState.Reason
		}

		// Upgrade failed, leave state for debugging
		logger.Error(nil, "Upgrade failed", "reason", reason)
		return false, fmt.Errorf("upgrade failed: %s", reason)
	}

	return false, nil
}

// executePreUpgradeHooks executes pre-upgrade hooks with security validations
// This executes hooks for ALL components that are in detected/preparing phase
// Supports both predefined (safe, hardcoded) and custom (user-provided) hooks
func (m *Manager) executePreUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")
	logger.Info("Executing pre-upgrade hooks")

	// Update phase to preparing for all detected components
	m.updateComponentsPhaseToPreparing(src)

	// Get list of components that need hooks executed
	componentsToUpgrade := m.Detector.getComponentsInPhase(src, srapi.UpgradePhasePreparing)
	if len(componentsToUpgrade) == 0 {
		logger.Info("No components in preparing phase")
		return false, nil
	}

	logger.Info("Components requiring pre-upgrade hooks", "components", componentsToUpgrade)

	// Execute hooks for each component
	for _, componentType := range componentsToUpgrade {
		logger.Info("Processing pre-upgrade hooks for component", "component", componentType)

		// Check if component has custom hooks configured
		if err := m.executeComponentHooks(ctx, src, componentType, "pre"); err != nil {
			logger.Error(err, "Component hooks failed", "component", componentType)
			m.markComponentsAsFailed(src, fmt.Sprintf("Pre-upgrade hooks failed for %s: %v", componentType, err))
			return false, err
		}
	}

	// Mark preparation as complete for all components
	m.Detector.MarkPreparationComplete(src)
	logger.Info("Pre-upgrade hooks completed successfully for all components")

	// Requeue to proceed with upgrade
	return true, nil
}

// executeComponentHooks executes hooks for a specific component
// This includes both predefined and custom hooks
func (m *Manager) executeComponentHooks(ctx context.Context, src *srapi.StarRocksCluster, componentType ComponentType, phase string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	// Get hook configuration for this component
	hookConfig := m.getComponentHookConfig(src, componentType)
	if hookConfig == nil {
		// No hooks configured, execute default hooks
		logger.Info("No hooks configured, executing default hooks", "component", componentType, "phase", phase)
		return m.executeDefaultHooks(ctx, src, phase)
	}

	// Execute predefined hooks if configured
	if len(hookConfig.Predefined) > 0 {
		logger.Info("Executing predefined hooks", "component", componentType, "phase", phase, "hooks", hookConfig.Predefined)
		if err := m.executePredefinedHooks(ctx, src, hookConfig.Predefined, phase); err != nil {
			return fmt.Errorf("predefined hooks failed: %w", err)
		}
	}

	// Execute custom hooks if configured
	if hookConfig.Custom != nil {
		logger.Info("SECURITY: Executing custom hook from ConfigMap",
			"component", componentType,
			"phase", phase,
			"configMap", hookConfig.Custom.ConfigMapName)

		hookPhase := "pre_upgrade"
		if phase == "post" {
			hookPhase = "post_upgrade"
		}

		if err := m.CustomHookExecutor.ExecuteCustomHook(ctx, src, componentType, hookPhase); err != nil {
			return fmt.Errorf("custom hook failed: %w", err)
		}
	}

	// If no hooks configured at all, execute default hooks
	if len(hookConfig.Predefined) == 0 && hookConfig.Custom == nil {
		logger.Info("No predefined or custom hooks, executing defaults", "component", componentType)
		return m.executeDefaultHooks(ctx, src, phase)
	}

	return nil
}

// getComponentHookConfig gets hook configuration for a component
func (m *Manager) getComponentHookConfig(src *srapi.StarRocksCluster, componentType ComponentType) *srapi.ComponentUpgradeHooks {
	switch componentType {
	case ComponentTypeFE:
		if src.Spec.StarRocksFeSpec != nil {
			return src.Spec.StarRocksFeSpec.UpgradeHooks
		}
	case ComponentTypeBE:
		if src.Spec.StarRocksBeSpec != nil {
			return src.Spec.StarRocksBeSpec.UpgradeHooks
		}
	case ComponentTypeCN:
		if src.Spec.StarRocksCnSpec != nil {
			return src.Spec.StarRocksCnSpec.UpgradeHooks
		}
	}
	return nil
}

// executeDefaultHooks executes the default hardcoded hooks (original behavior)
// These are the safe, pre-reviewed SQL commands
func (m *Manager) executeDefaultHooks(ctx context.Context, src *srapi.StarRocksCluster, phase string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	if phase != "pre" {
		// Default hooks only apply to pre-upgrade phase
		return nil
	}

	logger.Info("Executing default pre-upgrade hooks")

	// Security: Get FE connection with validation
	db, err := m.HookExecutor.GetFEConnection(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to connect to FE for default hooks")
		// If FE is not ready, requeue and try again (don't fail immediately)
		return fmt.Errorf("FE connection failed: %w", err)
	}
	// Security: Ensure database connection is always closed
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close database connection")
		}
	}()

	// Validate FE is ready
	if err := m.HookExecutor.ValidateFEReady(ctx, db); err != nil {
		logger.Error(err, "FE not ready for default hooks")
		return fmt.Errorf("FE not ready: %w", err)
	}

	// Execute default pre-upgrade hooks (original hardcoded behavior)
	if err := m.HookExecutor.ExecutePreUpgradeHooks(ctx, src, db); err != nil {
		logger.Error(err, "Default pre-upgrade hooks failed")
		return fmt.Errorf("default hooks failed: %w", err)
	}

	return nil
}

// executePredefinedHooks executes predefined hooks by name
func (m *Manager) executePredefinedHooks(ctx context.Context, src *srapi.StarRocksCluster, hookNames []string, phase string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")

	// Security: Get FE connection with validation
	db, err := m.HookExecutor.GetFEConnection(ctx, src)
	if err != nil {
		logger.Error(err, "Failed to connect to FE for predefined hooks")
		return fmt.Errorf("FE connection failed: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close database connection")
		}
	}()

	// Validate FE is ready
	if err := m.HookExecutor.ValidateFEReady(ctx, db); err != nil {
		logger.Error(err, "FE not ready for predefined hooks")
		return fmt.Errorf("FE not ready: %w", err)
	}

	// Get hook definitions
	hooks := m.CustomHookExecutor.getPredefinedHooks(hookNames, phase)
	if len(hooks) == 0 {
		logger.Info("No predefined hooks for this phase", "phase", phase)
		return nil
	}

	// Execute each hook
	for _, hook := range hooks {
		logger.Info("Executing predefined hook", "name", hook.Name, "phase", phase)

		if err := m.HookExecutor.ExecuteHookWithRetry(ctx, db, hook); err != nil {
			if hook.Critical {
				return fmt.Errorf("critical predefined hook %s failed: %w", hook.Name, err)
			}
			logger.Error(err, "Non-critical predefined hook failed", "hook", hook.Name)
		}
	}

	return nil
}

// updateComponentsPhaseToPreparing updates the phase to preparing for components in detected phase
func (m *Manager) updateComponentsPhaseToPreparing(src *srapi.StarRocksCluster) {
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksFeStatus.UpgradeState.Phase = srapi.UpgradePhasePreparing
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksBeStatus.UpgradeState.Phase = srapi.UpgradePhasePreparing
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksCnStatus.UpgradeState.Phase = srapi.UpgradePhasePreparing
		}
	}
}

// markComponentsAsFailed marks components as failed
func (m *Manager) markComponentsAsFailed(src *srapi.StarRocksCluster, reason string) {
	if src.Status.StarRocksFeStatus != nil && src.Status.StarRocksFeStatus.UpgradeState != nil {
		if src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksFeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksFeStatus.UpgradeState.Phase = srapi.UpgradePhaseFailed
			src.Status.StarRocksFeStatus.UpgradeState.Reason = reason
		}
	}

	if src.Status.StarRocksBeStatus != nil && src.Status.StarRocksBeStatus.UpgradeState != nil {
		if src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksBeStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksBeStatus.UpgradeState.Phase = srapi.UpgradePhaseFailed
			src.Status.StarRocksBeStatus.UpgradeState.Reason = reason
		}
	}

	if src.Status.StarRocksCnStatus != nil && src.Status.StarRocksCnStatus.UpgradeState != nil {
		if src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhasePreparing ||
			src.Status.StarRocksCnStatus.UpgradeState.Phase == srapi.UpgradePhaseDetected {
			src.Status.StarRocksCnStatus.UpgradeState.Phase = srapi.UpgradePhaseFailed
			src.Status.StarRocksCnStatus.UpgradeState.Reason = reason
		}
	}
}

// executePostUpgradeHooks executes post-upgrade cleanup hooks (non-critical)
func (m *Manager) executePostUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeManager")
	logger.Info("Executing post-upgrade hooks")

	// Get list of components that completed upgrade
	componentsCompleted := m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseInProgress)
	if len(componentsCompleted) == 0 {
		logger.Info("No components in progress phase for post-upgrade hooks")
		return false, nil
	}

	logger.Info("Components requiring post-upgrade hooks", "components", componentsCompleted)

	// Execute post-upgrade hooks for each component (non-critical)
	for _, componentType := range componentsCompleted {
		logger.Info("Processing post-upgrade hooks for component", "component", componentType)

		if err := m.executeComponentHooks(ctx, src, componentType, "post"); err != nil {
			// Post-upgrade hooks are non-critical, just log and continue
			logger.Error(err, "Post-upgrade hooks failed (non-critical)", "component", componentType)
		}
	}

	// Mark upgrade as complete for all components
	m.Detector.MarkUpgradeComplete(src)
	logger.Info("Upgrade process completed successfully for all components")

	return false, nil
}

// ShouldBlockReconciliation returns true if normal reconciliation should be blocked
// This happens when pre-upgrade hooks need to be executed
func (m *Manager) ShouldBlockReconciliation(src *srapi.StarRocksCluster) bool {
	// Check if any component is in detected or preparing phase
	detectedComponents := m.Detector.getComponentsInPhase(src, srapi.UpgradePhaseDetected)
	preparingComponents := m.Detector.getComponentsInPhase(src, srapi.UpgradePhasePreparing)

	// Block reconciliation if any component is in these phases
	return len(detectedComponents) > 0 || len(preparingComponents) > 0
}