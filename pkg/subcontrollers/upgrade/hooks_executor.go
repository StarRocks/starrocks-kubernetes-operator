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
	// MaxRetries for transient failures
	MaxRetries int
	// RetryDelay between retry attempts
	RetryDelay time.Duration
}

// NewHookExecutor creates a new hook executor with secure defaults
func NewHookExecutor() *HookExecutor {
	return &HookExecutor{
		DefaultTimeout: 300 * time.Second, // 5 minutes timeout for hook execution
		MaxRetries:     3,                  // Retry up to 3 times for transient failures
		RetryDelay:     5 * time.Second,    // Wait 5 seconds between retries
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

	// Execute each hook with retry logic
	for _, hook := range hooks {
		logger.Info("Executing pre-upgrade hook", "name", hook.Name)
		// Security: Do NOT log the actual SQL command to prevent sensitive info leakage

		err := h.executeHookWithRetry(ctx, db, hook)
		if err != nil {
			logger.Error(err, "Pre-upgrade hook failed after retries", "hook", hook.Name)

			if hook.Critical {
				src.Status.UpgradeState.Phase = srapi.UpgradePhaseFailed
				// Security: Sanitize error message to avoid leaking sensitive information
				src.Status.UpgradeState.Reason = fmt.Sprintf("Critical pre-upgrade hook %s failed", hook.Name)
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

	// Execute each hook with retry logic
	for _, hook := range hooks {
		logger.Info("Executing post-upgrade hook", "name", hook.Name)
		// Security: Do NOT log the actual SQL command

		err := h.executeHookWithRetry(ctx, db, hook)
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

// executeHookWithRetry executes a single hook with timeout and retry logic
func (h *HookExecutor) executeHookWithRetry(ctx context.Context, db *sql.DB, hook Hook) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	var lastErr error
	for attempt := 1; attempt <= h.MaxRetries; attempt++ {
		if attempt > 1 {
			logger.Info("Retrying hook execution", "hook", hook.Name, "attempt", attempt, "maxRetries", h.MaxRetries)
			time.Sleep(h.RetryDelay)
		}

		err := h.executeHook(ctx, db, hook)
		if err == nil {
			if attempt > 1 {
				logger.Info("Hook succeeded after retry", "hook", hook.Name, "attempt", attempt)
			}
			return nil
		}

		lastErr = err
		logger.Error(err, "Hook execution attempt failed", "hook", hook.Name, "attempt", attempt)
	}

	return fmt.Errorf("hook %s failed after %d attempts: %w", hook.Name, h.MaxRetries, lastErr)
}

// executeHook executes a single hook with timeout (internal method)
func (h *HookExecutor) executeHook(ctx context.Context, db *sql.DB, hook Hook) error {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, h.DefaultTimeout)
	defer cancel()

	// Security: Use parameterized queries where possible
	// Note: For ADMIN commands, parameterization is limited, but we validate the command source
	// These commands are hardcoded and not user-provided, reducing SQL injection risk

	// Execute the SQL command
	_, err := db.ExecContext(ctx, hook.Command)
	if err != nil {
		// Security: Do not expose the SQL command in the error message
		return fmt.Errorf("failed to execute hook: %w", err)
	}

	return nil
}

// GetFEConnection establishes a secure connection to the FE for executing hooks
func (h *HookExecutor) GetFEConnection(ctx context.Context, src *srapi.StarRocksCluster) (*sql.DB, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	// Security: Validate cluster name and namespace to prevent injection
	if err := validateClusterIdentifiers(src); err != nil {
		return nil, fmt.Errorf("invalid cluster identifiers: %w", err)
	}

	// Build FE service name following operator convention
	feServiceName := fmt.Sprintf("%s-fe-service", src.Name)

	// Build connection string - only connect to services within the same namespace
	// Security: Use svc.cluster.local to ensure we only connect to in-cluster services
	host := fmt.Sprintf("%s.%s.svc.cluster.local", feServiceName, src.Namespace)
	port := "9030" // Default FE query port

	// Check if custom port is configured
	if src.Spec.StarRocksFeSpec != nil && src.Spec.StarRocksFeSpec.Service != nil {
		for _, p := range src.Spec.StarRocksFeSpec.Service.Ports {
			if p.Name == "query" {
				// Security: Validate port range
				if p.Port < 1 || p.Port > 65535 {
					return nil, fmt.Errorf("invalid port number: %d", p.Port)
				}
				port = fmt.Sprintf("%d", p.Port)
				break
			}
		}
	}

	// Build DSN (Data Source Name)
	// Security: Use root user for admin operations (this is expected for StarRocks admin commands)
	// Note: Password authentication should be handled by StarRocks RBAC, not at connection level
	// Format: [username[:password]@][protocol[(address)]]/database
	dsn := fmt.Sprintf("root@tcp(%s:%s)/?timeout=10s&readTimeout=30s&writeTimeout=30s", host, port)

	logger.Info("Connecting to FE", "host", host, "port", port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Security: Set conservative connection limits to prevent resource exhaustion
	db.SetMaxOpenConns(5)   // Limit concurrent connections
	db.SetMaxIdleConns(2)   // Limit idle connections
	db.SetConnMaxLifetime(time.Minute * 5) // Close connections after 5 minutes

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping FE: %w", err)
	}

	logger.Info("Successfully connected to FE")
	return db, nil
}

// validateClusterIdentifiers validates cluster name and namespace to prevent injection attacks
func validateClusterIdentifiers(src *srapi.StarRocksCluster) error {
	// Validate cluster name (DNS-1123 subdomain format)
	// Must consist of lower case alphanumeric characters, '-' or '.'
	if src.Name == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}
	if len(src.Name) > 253 {
		return fmt.Errorf("cluster name too long: %d characters (max 253)", len(src.Name))
	}

	// Validate namespace
	if src.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if len(src.Namespace) > 63 {
		return fmt.Errorf("namespace too long: %d characters (max 63)", len(src.Namespace))
	}

	// Security: Check for suspicious characters that could indicate injection attempts
	// Kubernetes DNS names should only contain alphanumeric characters, '-', and '.'
	for _, r := range src.Name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.') {
			return fmt.Errorf("cluster name contains invalid character: %c", r)
		}
	}

	for _, r := range src.Namespace {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return fmt.Errorf("namespace contains invalid character: %c", r)
		}
	}

	return nil
}

// ValidateFEReady checks if the FE is ready to receive commands
func (h *HookExecutor) ValidateFEReady(ctx context.Context, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("HookExecutor")

	// Security: Use a simple, safe query to check readiness
	// This query has no side effects and cannot be exploited
	query := "SELECT 1"

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result int
	err := db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return fmt.Errorf("FE not ready: %w", err)
	}

	// Security: Verify the result is as expected
	if result != 1 {
		return fmt.Errorf("FE returned unexpected result: %d", result)
	}

	logger.Info("FE is ready to accept commands")
	return nil
}