/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package upgrade

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// HookExecutor executes pre-upgrade and post-upgrade hooks
type HookExecutor struct {
	// Timeout for hook execution
	DefaultTimeout time.Duration
}

// NewHookExecutor creates a new hook executor
func NewHookExecutor() *HookExecutor {
	return &HookExecutor{
		DefaultTimeout: 300 * time.Second,
	}
}

// Hook represents a single upgrade hook
type Hook struct {
	Name     string
	Command  string
	Critical bool
}

// ExecutePreUpgradeHooks executes all pre-upgrade hooks automatically
func (h *HookExecutor) ExecutePreUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")
	logger.Info("Executing pre-upgrade hooks automatically")

	// Define standard pre-upgrade hooks that should run for all upgrades
	hooks := h.getStandardPreUpgradeHooks()

	if src.Status.UpgradeState == nil {
		src.Status.UpgradeState = &srapi.UpgradeState{}
	}

	// Execute each hook
	for _, hook := range hooks {
		logger.Info("Executing pre-upgrade hook", "name", hook.Name, "command", hook.Command)

		err := h.executeHook(ctx, db, hook)
		if err != nil {
			logger.Error(err, "Pre-upgrade hook failed", "hook", hook.Name)

			if hook.Critical {
				src.Status.UpgradeState.Phase = srapi.UpgradePhaseFailed
				src.Status.UpgradeState.Reason = fmt.Sprintf("Critical pre-upgrade hook %s failed: %v", hook.Name, err)
				return fmt.Errorf("critical pre-upgrade hook %s failed: %w", hook.Name, err)
			}

			// Non-critical hook failed, log and continue
			logger.Info("Non-critical hook failed, continuing", "hook", hook.Name)
			continue
		}

		// Track successful hook execution
		src.Status.UpgradeState.HooksExecuted = append(src.Status.UpgradeState.HooksExecuted, hook.Name)
		logger.Info("Pre-upgrade hook completed successfully", "hook", hook.Name)
	}

	logger.Info("All pre-upgrade hooks executed successfully")
	return nil
}

// ExecutePostUpgradeHooks executes post-upgrade cleanup hooks automatically
func (h *HookExecutor) ExecutePostUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")
	logger.Info("Executing post-upgrade hooks automatically")

	// Define standard post-upgrade hooks
	hooks := h.getStandardPostUpgradeHooks()

	// Execute each hook
	for _, hook := range hooks {
		logger.Info("Executing post-upgrade hook", "name", hook.Name, "command", hook.Command)

		err := h.executeHook(ctx, db, hook)
		if err != nil {
			// Post-upgrade hooks are typically non-critical
			logger.Error(err, "Post-upgrade hook failed (non-critical)", "hook", hook.Name)
			continue
		}

		logger.Info("Post-upgrade hook completed successfully", "hook", hook.Name)
	}

	logger.Info("All post-upgrade hooks executed")
	return nil
}

// getStandardPreUpgradeHooks returns the standard pre-upgrade hooks that should run for all upgrades
func (h *HookExecutor) getStandardPreUpgradeHooks() []Hook {
	return []Hook{
		{
			Name:     "disable-tablet-scheduling",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`,
			Critical: true,
		},
		{
			Name:     "disable-load-balancer",
			Command:  `ADMIN SET FRONTEND CONFIG ("disable_balance" = "true")`,
			Critical: true,
		},
	}
}

// getStandardPostUpgradeHooks returns the standard post-upgrade hooks for cleanup
func (h *HookExecutor) getStandardPostUpgradeHooks() []Hook {
	return []Hook{
		{
			Name:     "enable-tablet-scheduling",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")`,
			Critical: false,
		},
		{
			Name:     "enable-load-balancer",
			Command:  `ADMIN SET FRONTEND CONFIG ("disable_balance" = "false")`,
			Critical: false,
		},
	}
}

// executeHook executes a single hook with timeout
func (h *HookExecutor) executeHook(ctx context.Context, db *sql.DB, hook Hook) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, h.DefaultTimeout)
	defer cancel()

	logger.Info("Executing SQL command", "command", hook.Command)

	// Execute the SQL command
	_, err := db.ExecContext(ctx, hook.Command)
	if err != nil {
		return fmt.Errorf("failed to execute hook %s: %w", hook.Name, err)
	}

	return nil
}

// GetFEConnection establishes a connection to the FE for executing hooks
func (h *HookExecutor) GetFEConnection(ctx context.Context, src *srapi.StarRocksCluster) (*sql.DB, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	// Build FE service name following operator convention
	feServiceName := fmt.Sprintf("%s-fe-service", src.Name)

	// Build connection string
	host := fmt.Sprintf("%s.%s.svc.cluster.local", feServiceName, src.Namespace)
	port := "9030" // Default FE query port

	// Check if custom port is configured
	if src.Spec.StarRocksFeSpec != nil && src.Spec.StarRocksFeSpec.Service != nil {
		for _, p := range src.Spec.StarRocksFeSpec.Service.Ports {
			if p.Name == "query" {
				port = fmt.Sprintf("%d", p.Port)
				break
			}
		}
	}

	// Build DSN (Data Source Name)
	// Format: [username[:password]@][protocol[(address)]]/
	dsn := fmt.Sprintf("root@tcp(%s:%s)/", host, port)

	logger.Info("Connecting to FE", "host", host, "port", port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection limits
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping FE: %w", err)
	}

	logger.Info("Successfully connected to FE")
	return db, nil
}

// ValidateFEReady checks if the FE is ready to receive commands
func (h *HookExecutor) ValidateFEReady(ctx context.Context, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	// Simple query to check if FE is ready
	query := "SELECT 1"

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result int
	err := db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return fmt.Errorf("FE not ready: %w", err)
	}

	logger.Info("FE is ready to accept commands")
	return nil
}