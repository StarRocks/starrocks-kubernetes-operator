/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package subcontrollers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

const (
	AnnotationPrepareUpgrade = "starrocks.com/prepare-upgrade"
	AnnotationUpgradeHooks   = "starrocks.com/upgrade-hooks"
)

// UpgradeHookController manages pre-upgrade hooks
type UpgradeHookController struct {
	Client client.Client
}

// NewUpgradeHookController creates a new upgrade hook controller
func NewUpgradeHookController(k8sClient client.Client) *UpgradeHookController {
	return &UpgradeHookController{
		Client: k8sClient,
	}
}

// ShouldExecuteUpgradeHooks checks if upgrade hooks should be executed
func (uhc *UpgradeHookController) ShouldExecuteUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) bool {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController")

	// Check for prepare-upgrade annotation
	if annotations := src.GetAnnotations(); annotations != nil {
		if prepareUpgrade, exists := annotations[AnnotationPrepareUpgrade]; exists && prepareUpgrade == "true" {
			// Check if we already completed preparation for this generation
			if src.Status.UpgradePreparationStatus != nil &&
				src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationCompleted {
				logger.Info("Upgrade preparation already completed")
				return false
			}
			return true
		}
	}

	// Check for spec-based upgrade preparation
	if src.Spec.UpgradePreparation != nil && src.Spec.UpgradePreparation.Enabled {
		if src.Status.UpgradePreparationStatus != nil &&
			src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationCompleted {
			logger.Info("Upgrade preparation already completed")
			return false
		}
		return true
	}

	return false
}

// ExecuteUpgradeHooks executes pre-upgrade hooks
func (uhc *UpgradeHookController) ExecuteUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController")
	logger.Info("Starting upgrade preparation hooks")

	// Initialize status if not exists
	if src.Status.UpgradePreparationStatus == nil {
		src.Status.UpgradePreparationStatus = &srapi.UpgradePreparationStatus{
			Phase:     srapi.UpgradePreparationPending,
			StartTime: &metav1.Time{Time: time.Now()},
		}
	}

	// Update status to running
	src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationRunning

	hooks, err := uhc.getHooksToExecute(src)
	if err != nil {
		src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationFailed
		src.Status.UpgradePreparationStatus.Reason = fmt.Sprintf("Failed to get hooks: %v", err)
		return err
	}

	if len(hooks) == 0 {
		logger.Info("No hooks to execute, marking as completed")
		src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationCompleted
		src.Status.UpgradePreparationStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		return nil
	}

	// Get FE connection
	db, err := uhc.getFEConnection(ctx, src)
	if err != nil {
		src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationFailed
		src.Status.UpgradePreparationStatus.Reason = fmt.Sprintf("Failed to connect to FE: %v", err)
		return err
	}
	defer db.Close()

	// Execute hooks
	for _, hook := range hooks {
		logger.Info("Executing hook", "name", hook.Name, "command", hook.Command)
		
		err := uhc.executeHook(ctx, db, hook)
		if err != nil {
			logger.Error(err, "Hook execution failed", "hook", hook.Name)
			src.Status.UpgradePreparationStatus.FailedHooks = append(
				src.Status.UpgradePreparationStatus.FailedHooks, hook.Name)
			
			if hook.Critical {
				src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationFailed
				src.Status.UpgradePreparationStatus.Reason = fmt.Sprintf("Critical hook %s failed: %v", hook.Name, err)
				return err
			}
			continue
		}
		
		src.Status.UpgradePreparationStatus.CompletedHooks = append(
			src.Status.UpgradePreparationStatus.CompletedHooks, hook.Name)
		logger.Info("Hook completed successfully", "hook", hook.Name)
	}

	// Mark as completed
	src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationCompleted
	src.Status.UpgradePreparationStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	src.Status.UpgradePreparationStatus.Reason = "All hooks executed successfully"

	logger.Info("Upgrade preparation completed successfully")
	return nil
}

// getHooksToExecute determines which hooks to execute
func (uhc *UpgradeHookController) getHooksToExecute(src *srapi.StarRocksCluster) ([]srapi.UpgradeHook, error) {
	var hooks []srapi.UpgradeHook

	// Check spec-based hooks first
	if src.Spec.UpgradePreparation != nil && len(src.Spec.UpgradePreparation.Hooks) > 0 {
		hooks = append(hooks, src.Spec.UpgradePreparation.Hooks...)
	}

	// Check annotation-based hooks
	if annotations := src.GetAnnotations(); annotations != nil {
		if hookNames, exists := annotations[AnnotationUpgradeHooks]; exists {
			predefinedHooks := uhc.getPredefinedHooks(hookNames)
			hooks = append(hooks, predefinedHooks...)
		}
	}

	return hooks, nil
}

// getPredefinedHooks returns predefined hooks based on names
func (uhc *UpgradeHookController) getPredefinedHooks(hookNames string) []srapi.UpgradeHook {
	var hooks []srapi.UpgradeHook
	
	names := strings.Split(hookNames, ",")
	for _, name := range names {
		name = strings.TrimSpace(name)
		switch name {
		case "disable-tablet-clone":
			hooks = append(hooks, srapi.UpgradeHook{
				Name:     "disable-tablet-clone",
				Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`,
				Critical: true,
			})
		case "disable-balancer":
			hooks = append(hooks, srapi.UpgradeHook{
				Name:     "disable-balancer",
				Command:  `ADMIN SET FRONTEND CONFIG ("disable_balance"="true")`,
				Critical: true,
			})
		case "enable-tablet-clone":
			hooks = append(hooks, srapi.UpgradeHook{
				Name:     "enable-tablet-clone",
				Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")`,
				Critical: false,
			})
		case "enable-balancer":
			hooks = append(hooks, srapi.UpgradeHook{
				Name:     "enable-balancer",
				Command:  `ADMIN SET FRONTEND CONFIG ("disable_balance"="false")`,
				Critical: false,
			})
		}
	}
	
	return hooks
}

// getFEConnection establishes connection to FE
func (uhc *UpgradeHookController) getFEConnection(ctx context.Context, src *srapi.StarRocksCluster) (*sql.DB, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController")
	
	// Build FE service name (follows operator convention: {cluster-name}-fe-service)
	feServiceName := fmt.Sprintf("%s-fe-service", src.Name)
	
	// Build connection string
	host := fmt.Sprintf("%s.%s.svc.cluster.local", feServiceName, src.Namespace)
	port := "9030" // Default FE query port
	if src.Spec.StarRocksFeSpec != nil && src.Spec.StarRocksFeSpec.Service != nil {
		for _, p := range src.Spec.StarRocksFeSpec.Service.Ports {
			if p.Name == "query" {
				port = fmt.Sprintf("%d", p.Port)
				break
			}
		}
	}
	
	dsn := fmt.Sprintf("root@tcp(%s:%s)/", host, port)
	
	logger.Info("Connecting to FE", "dsn", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %v", err)
	}
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping FE: %v", err)
	}
	
	return db, nil
}

// executeHook executes a single hook
func (uhc *UpgradeHookController) executeHook(ctx context.Context, db *sql.DB, hook srapi.UpgradeHook) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController")
	
	// Set timeout
	timeout := 300 * time.Second // Default timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	logger.Info("Executing SQL command", "command", hook.Command)
	_, err := db.ExecContext(ctx, hook.Command)
	if err != nil {
		return fmt.Errorf("failed to execute hook %s: %v", hook.Name, err)
	}
	
	return nil
}

// CleanupAnnotations removes upgrade preparation annotations
func (uhc *UpgradeHookController) CleanupAnnotations(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController")
	
	annotations := src.GetAnnotations()
	if annotations == nil {
		return nil
	}
	
	changed := false
	if _, exists := annotations[AnnotationPrepareUpgrade]; exists {
		delete(annotations, AnnotationPrepareUpgrade)
		changed = true
	}
	
	if changed {
		src.SetAnnotations(annotations)
		logger.Info("Cleaned up upgrade preparation annotations")
	}
	
	return nil
}

// IsUpgradeReady checks if the cluster is ready for upgrade
func (uhc *UpgradeHookController) IsUpgradeReady(ctx context.Context, src *srapi.StarRocksCluster) bool {
	// If no upgrade preparation configured, ready by default
	if !uhc.ShouldExecuteUpgradeHooks(ctx, src) {
		return true
	}
	
	// Check if preparation is completed
	if src.Status.UpgradePreparationStatus != nil {
		return src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationCompleted
	}
	
	return false
}
