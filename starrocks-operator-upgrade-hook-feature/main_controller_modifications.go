// Add this to starrockscluster_controller.go

import (
	// ... existing imports ...
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
)

// Add to StarRocksClusterReconciler struct
type StarRocksClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scs      []subcontrollers.ClusterSubController
	// Add upgrade hook controller
	UpgradeHookController *subcontrollers.UpgradeHookController
}

// Modify the Reconcile function to include upgrade hook logic
func (r *StarRocksClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithName("StarRocksClusterReconciler").WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("begin to reconcile StarRocksCluster")

	var esrc srapi.StarRocksCluster
	client := r.Client
	err := client.Get(ctx, req.NamespacedName, &esrc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get StarRocksCluster object failed")
		return requeueIfError(err)
	}

	src := esrc.DeepCopy()

	// reconcile src deleted
	if !src.DeletionTimestamp.IsZero() {
		logger.Info("deletion timestamp is not zero, clear StarRocksCluster related resources")
		return ctrl.Result{}, nil
	}

	// NEW: Check and execute upgrade hooks if needed
	if r.UpgradeHookController != nil && r.UpgradeHookController.ShouldExecuteUpgradeHooks(ctx, src) {
		logger.Info("Executing upgrade preparation hooks")
		if err = r.UpgradeHookController.ExecuteUpgradeHooks(ctx, src); err != nil {
			logger.Error(err, "upgrade hooks execution failed")
			// Update status with error
			if updateError := r.UpdateStarRocksClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update StarRocksCluster Status")
			}
			return requeueIfError(err)
		}

		// Update status after hook execution
		if err = r.UpdateStarRocksClusterStatus(ctx, src); err != nil {
			logger.Error(err, "failed to update StarRocksCluster Status after hooks")
			return requeueIfError(err)
		}

		// If hooks are still running, requeue
		if src.Status.UpgradePreparationStatus != nil && 
			src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationRunning {
			return ctrl.Result{RequeueAfter: time.Second * 30}, nil
		}
	}

	// Original subControllers reconcile logic continues...
	for _, rc := range r.Scs {
		kvs := []interface{}{"subController", rc.GetControllerName()}
		logger.Info("sub controller sync spec", kvs...)
		if err = rc.SyncCluster(ctx, src); err != nil {
			logger.Error(err, "sub controller reconciles spec failed", kvs...)
			handleSyncClusterError(src, rc, err)
			if updateError := r.UpdateStarRocksClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update StarRocksCluster Status")
			}
			return requeueIfError(err)
		}
	}

	// ... rest of original reconcile logic ...

	// NEW: Cleanup upgrade preparation if upgrade completed successfully
	if src.Status.UpgradePreparationStatus != nil && 
		src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationCompleted {

		// Check if all components are running (upgrade completed)
		allRunning := true
		if src.Status.StarRocksFeStatus != nil && 
			src.Status.StarRocksFeStatus.Phase != srapi.ComponentRunning {
			allRunning = false
		}
		if src.Status.StarRocksBeStatus != nil && 
			src.Status.StarRocksBeStatus.Phase != srapi.ComponentRunning {
			allRunning = false
		}
		if src.Status.StarRocksCnStatus != nil && 
			src.Status.StarRocksCnStatus.Phase != srapi.ComponentRunning {
			allRunning = false
		}

		if allRunning {
			logger.Info("Upgrade completed, cleaning up preparation state")
			if err = r.UpgradeHookController.CleanupUpgradePreparation(ctx, src); err != nil {
				logger.Error(err, "failed to cleanup upgrade preparation")
				// Don't fail reconciliation for cleanup errors
			} else {
				// Reset the upgrade preparation status
				src.Status.UpgradePreparationStatus = nil
			}
		}
	}

	logger.Info("reconcile StarRocksCluster success")
	return ctrl.Result{}, nil
}

// Add initialization in main.go or controller setup
func setupUpgradeHookController(mgr ctrl.Manager) error {
	upgradeHookController := subcontrollers.NewUpgradeHookController(mgr.GetClient())

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Complete(&StarRocksClusterReconciler{
			Client:                mgr.GetClient(),
			Recorder:              mgr.GetEventRecorderFor("StarRocksCluster"),
			Scs:                   getSubControllers(mgr), // your existing sub controllers
			UpgradeHookController: upgradeHookController,
		}); err != nil {
		return err
	}

	return nil
}
