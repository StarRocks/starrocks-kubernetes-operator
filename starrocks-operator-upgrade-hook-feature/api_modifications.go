/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpgradeHook defines pre-upgrade commands to execute
type UpgradeHook struct {
	// Name of the hook for identification
	Name string `json:"name"`

	// SQL command to execute on FE
	Command string `json:"command"`

	// Whether this hook is critical (upgrade fails if hook fails)
	Critical bool `json:"critical,omitempty"`
}

// UpgradePreparation defines upgrade preparation configuration
type UpgradePreparation struct {
	// Enable upgrade preparation hooks
	Enabled bool `json:"enabled,omitempty"`

	// List of pre-upgrade hooks to execute
	Hooks []UpgradeHook `json:"hooks,omitempty"`

	// Timeout for hook execution (default: 300s)
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
}

// Add to StarRocksClusterSpec
type StarRocksClusterSpec struct {
	// ... existing fields ...

	// UpgradePreparation defines pre-upgrade hooks and preparation steps
	// +optional
	UpgradePreparation *UpgradePreparation `json:"upgradePreparation,omitempty"`
}

// Add to StarRocksClusterStatus
type StarRocksClusterStatus struct {
	// ... existing fields ...

	// UpgradePreparationStatus tracks the status of upgrade preparation
	// +optional
	UpgradePreparationStatus *UpgradePreparationStatus `json:"upgradePreparationStatus,omitempty"`
}

// UpgradePreparationStatus tracks upgrade preparation state
type UpgradePreparationStatus struct {
	// Phase of upgrade preparation
	Phase UpgradePreparationPhase `json:"phase,omitempty"`

	// Reason for current phase
	Reason string `json:"reason,omitempty"`

	// Completed hooks
	CompletedHooks []string `json:"completedHooks,omitempty"`

	// Failed hooks
	FailedHooks []string `json:"failedHooks,omitempty"`

	// Timestamp when preparation started
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Timestamp when preparation completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// UpgradePreparationPhase represents the phase of upgrade preparation
type UpgradePreparationPhase string

const (
	// UpgradePreparationPending indicates preparation is pending
	UpgradePreparationPending UpgradePreparationPhase = "Pending"

	// UpgradePreparationRunning indicates preparation is in progress
	UpgradePreparationRunning UpgradePreparationPhase = "Running"

	// UpgradePreparationCompleted indicates preparation completed successfully
	UpgradePreparationCompleted UpgradePreparationPhase = "Completed"

	// UpgradePreparationFailed indicates preparation failed
	UpgradePreparationFailed UpgradePreparationPhase = "Failed"
)
