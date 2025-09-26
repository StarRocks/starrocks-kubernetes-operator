/*
Copyright 2021-present, StarRocks Inc.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package subcontrollers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/log"
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

	// Check if spec has upgrade preparation enabled
	if src.Spec.UpgradePreparation != nil && src.Spec.UpgradePreparation.Enabled {
		return true
	}

	return false
}

// ExecuteUpgradeHooks executes pre-upgrade hooks
func (uhc *UpgradeHookController) ExecuteUpgradeHooks(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController").WithValues(log.ActionKey, "ExecuteUpgradeHooks")

	logger.Info("Starting upgrade preparation hooks")

	// Initialize status if needed
	if src.Status.UpgradePreparationStatus == nil {
		src.Status.UpgradePreparationStatus = &srapi.UpgradePreparationStatus{
			Phase:     srapi.UpgradePreparationRunning,
			StartTime: &metav1.Time{Time: time.Now()},
		}
	} else {
		src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationRunning
		if src.Status.UpgradePreparationStatus.StartTime == nil {
			src.Status.UpgradePreparationStatus.StartTime = &metav1.Time{Time: time.Now()}
		}
	}

	// Get hooks to execute
	hooks := uhc.getHooksToExecute(src)
	if len(hooks) == 0 {
		logger.Info("No hooks to execute")
		src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationCompleted
		src.Status.UpgradePreparationStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		return nil
	}

	// Execute hooks
	for _, hook := range hooks {
		logger.Info("Executing upgrade hook", "hookName", hook.Name, "command", hook.Command)

		if err := uhc.executeHook(ctx, src, hook); err != nil {
			logger.Error(err, "Failed to execute upgrade hook", "hookName", hook.Name)
			src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationFailed
			src.Status.UpgradePreparationStatus.Reason = fmt.Sprintf("Hook %s failed: %v", hook.Name, err)
			src.Status.UpgradePreparationStatus.FailedHooks = append(src.Status.UpgradePreparationStatus.FailedHooks, hook.Name)

			if hook.Critical {
				return fmt.Errorf("critical hook %s failed: %v", hook.Name, err)
			}
		} else {
			logger.Info("Successfully executed upgrade hook", "hookName", hook.Name)
			src.Status.UpgradePreparationStatus.CompletedHooks = append(src.Status.UpgradePreparationStatus.CompletedHooks, hook.Name)
		}
	}

	// Mark as completed
	src.Status.UpgradePreparationStatus.Phase = srapi.UpgradePreparationCompleted
	src.Status.UpgradePreparationStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	logger.Info("Upgrade preparation completed successfully")

	return nil
}

// getHooksToExecute returns the list of hooks to execute
func (uhc *UpgradeHookController) getHooksToExecute(src *srapi.StarRocksCluster) []srapi.UpgradeHook {
	var hooks []srapi.UpgradeHook

	// Get hooks from spec
	if src.Spec.UpgradePreparation != nil && len(src.Spec.UpgradePreparation.Hooks) > 0 {
		hooks = append(hooks, src.Spec.UpgradePreparation.Hooks...)
	}

	// Get hooks from annotations
	if annotations := src.GetAnnotations(); annotations != nil {
		if hookList, exists := annotations[AnnotationUpgradeHooks]; exists {
			hooks = append(hooks, uhc.parseHookAnnotation(hookList)...)
		}
	}

	// Add default hooks if none specified
	if len(hooks) == 0 {
		hooks = uhc.getDefaultUpgradeHooks()
	}

	return hooks
}

// parseHookAnnotation parses the upgrade-hooks annotation
func (uhc *UpgradeHookController) parseHookAnnotation(hookList string) []srapi.UpgradeHook {
	var hooks []srapi.UpgradeHook

	// Simple parsing for comma-separated hook names
	// In a real implementation, this could be more sophisticated
	switch hookList {
	case "disable-tablet-clone":
		hooks = append(hooks, srapi.UpgradeHook{
			Name:     "disable-tablet-clone",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`,
			Critical: true,
		})
	case "disable-balancer":
		hooks = append(hooks, srapi.UpgradeHook{
			Name:     "disable-balancer",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "0")`,
			Critical: true,
		})
	case "disable-tablet-clone,disable-balancer":
		hooks = append(hooks, 
			srapi.UpgradeHook{
				Name:     "disable-tablet-clone",
				Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`,
				Critical: true,
			},
			srapi.UpgradeHook{
				Name:     "disable-balancer", 
				Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "0")`,
				Critical: true,
			},
		)
	}

	return hooks
}

// getDefaultUpgradeHooks returns the default set of upgrade hooks
func (uhc *UpgradeHookController) getDefaultUpgradeHooks() []srapi.UpgradeHook {
	return []srapi.UpgradeHook{
		{
			Name:     "disable-tablet-clone",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`,
			Critical: true,
		},
		{
			Name:     "disable-balancer",
			Command:  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "0")`,
			Critical: true,
		},
		{
			Name:    "disable-balance-flags",
			Command: `ADMIN SET FRONTEND CONFIG ("disable_balance"="true")`,
			Critical: true,
		},
		{
			Name:    "disable-colocate-balance",
			Command: `ADMIN SET FRONTEND CONFIG ("disable_colocate_balance"="true")`,
			Critical: true,
		},
	}
}

// executeHook executes a single upgrade hook
func (uhc *UpgradeHookController) executeHook(ctx context.Context, src *srapi.StarRocksCluster, hook srapi.UpgradeHook) error {
	logger := logr.FromContextOrDiscard(ctx)

	// Get FE service endpoint
	feEndpoint, err := uhc.getFEEndpoint(ctx, src)
	if err != nil {
		return fmt.Errorf("failed to get FE endpoint: %v", err)
	}

	// Connect to StarRocks FE
	dsn := fmt.Sprintf("root@tcp(%s)/", feEndpoint)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to StarRocks FE: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping StarRocks FE: %v", err)
	}

	// Execute the SQL command
	logger.Info("Executing SQL command", "command", hook.Command)
	_, err = db.Exec(hook.Command)
	if err != nil {
		return fmt.Errorf("failed to execute SQL command: %v", err)
	}

	logger.Info("Successfully executed SQL command", "command", hook.Command)
	return nil
}

// getFEEndpoint gets the FE service endpoint
func (uhc *UpgradeHookController) getFEEndpoint(ctx context.Context, src *srapi.StarRocksCluster) (string, error) {
	// This should get the FE service endpoint from the cluster
	// For now, we'll use a simple approach
	serviceName := fmt.Sprintf("%s-fe-service", src.Name)
	return fmt.Sprintf("%s.%s.svc.cluster.local:9030", serviceName, src.Namespace), nil
}

// CleanupUpgradePreparation removes upgrade preparation annotation and resets hooks
func (uhc *UpgradeHookController) CleanupUpgradePreparation(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("UpgradeHookController").WithValues(log.ActionKey, "CleanupUpgradePreparation")

	// Remove the prepare-upgrade annotation
	if annotations := src.GetAnnotations(); annotations != nil {
		if _, exists := annotations[AnnotationPrepareUpgrade]; exists {
			delete(annotations, AnnotationPrepareUpgrade)
			logger.Info("Removed prepare-upgrade annotation")
		}
	}

	// Execute post-upgrade hooks to re-enable services
	postUpgradeHooks := []srapi.UpgradeHook{
		{
			Name:    "enable-tablet-clone",
			Command: `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "10000")`,
		},
		{
			Name:    "enable-balancer",
			Command: `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "500")`,
		},
		{
			Name:    "enable-balance-flags",
			Command: `ADMIN SET FRONTEND CONFIG ("disable_balance"="false")`,
		},
		{
			Name:    "enable-colocate-balance",
			Command: `ADMIN SET FRONTEND CONFIG ("disable_colocate_balance"="false")`,
		},
	}

	// Execute post-upgrade hooks
	for _, hook := range postUpgradeHooks {
		logger.Info("Executing post-upgrade hook", "hookName", hook.Name)
		if err := uhc.executeHook(ctx, src, hook); err != nil {
			logger.Error(err, "Failed to execute post-upgrade hook", "hookName", hook.Name)
			// Don't fail the cleanup, just log the error
		}
	}

	return nil
}
