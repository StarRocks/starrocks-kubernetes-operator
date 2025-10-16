/*
Copyright 2021-present, StarRocks Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

// StarRocksClusterReconciler reconciles a StarRocksCluster object
type StarRocksClusterReconciler struct {
	client.Client
	Recorder          record.EventRecorder
	FeController      subcontrollers.ClusterSubController
	BeController      subcontrollers.ClusterSubController
	CnController      subcontrollers.ClusterSubController
	FeProxyController subcontrollers.ClusterSubController
}

// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",resources=endpoints,verbs=get;watch;list
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
//
//nolint:gocyclo // Complexity is inherent to orchestrating multiple controllers with upgrade sequencing
func (r *StarRocksClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithName("StarRocksClusterReconciler").WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("begin to reconcile StarRocksCluster")

	logger.Info("get StarRocksCluster CR from kubernetes")
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

	// Determine if this is an upgrade scenario and get controllers in appropriate order
	isUpgradeScenario := isUpgrade(ctx, r.Client, src)
	controllers := getControllersInOrder(isUpgradeScenario, r.FeController, r.BeController, r.CnController, r.FeProxyController)

	// subControllers reconcile for create or update component.
	for _, rc := range controllers {
		kvs := []interface{}{"subController", rc.GetControllerName()}
		controllerName := rc.GetControllerName()

		// During upgrades, check BE and CN are ready BEFORE syncing FE
		if isUpgradeScenario && controllerName == "feController" {
			// Check BE readiness if BE exists in spec
			if src.Spec.StarRocksBeSpec != nil && !isComponentReady(ctx, r.Client, src, "be") {
				logger.Info("upgrade: waiting for BE rollout to complete before updating FE",
					"controller", controllerName)
				return ctrl.Result{}, nil
			}
			// Check CN readiness if CN exists in spec
			if src.Spec.StarRocksCnSpec != nil && !isComponentReady(ctx, r.Client, src, "cn") {
				logger.Info("upgrade: waiting for CN rollout to complete before updating FE",
					"controller", controllerName)
				return ctrl.Result{}, nil
			}
		}

		// Sync the controller (create or update resources)
		if err = rc.SyncCluster(ctx, src); err != nil {
			logger.Error(err, "sub controller reconciles spec failed", kvs...)
			handleSyncClusterError(src, rc, err)
			if updateError := r.UpdateStarRocksClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update StarRocksCluster Status")
			}
			return requeueIfError(err)
		}

		// After syncing, check if we need to wait for this component to be ready before proceeding
		// Initial deployment: Wait for FE to be ready before creating BE/CN
		if !isUpgradeScenario && controllerName == "feController" {
			if src.Spec.StarRocksFeSpec != nil && !isComponentReady(ctx, r.Client, src, "fe") {
				logger.Info("initial deployment: waiting for FE to be ready before creating BE/CN", "controller", controllerName)
				return ctrl.Result{}, nil
			}
		}
	}

	for _, rc := range controllers {
		kvs := []interface{}{"subController", rc.GetControllerName()}
		logger.Info("sub controller update status", kvs...)
		if err = rc.UpdateClusterStatus(ctx, src); err != nil {
			logger.Error(err, "sub controller update status failed", kvs...)
			handleSyncClusterError(src, rc, err)
			if updateError := r.UpdateStarRocksClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update StarRocksCluster Status")
			}
			return requeueIfError(err)
		}
	}

	logger.Info("update StarRocksCluster level status")
	r.reconcileStatus(ctx, src)
	err = r.UpdateStarRocksClusterStatus(ctx, src)
	if err != nil {
		logger.Error(err, "update StarRocksCluster status failed")
		return ctrl.Result{}, err
	}
	logger.Info("reconcile StarRocksCluster success")
	return ctrl.Result{}, nil
}

// UpdateStarRocksClusterStatus update the status of src.
func (r *StarRocksClusterReconciler) UpdateStarRocksClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var esrc srapi.StarRocksCluster
		client := r.Client
		if err := client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Name}, &esrc); err != nil {
			return err
		}

		esrc.Status = src.Status
		return r.Client.Status().Update(ctx, &esrc)
	})
}

func (r *StarRocksClusterReconciler) reconcileStatus(_ context.Context, src *srapi.StarRocksCluster) {
	src.Status.Phase = srapi.ClusterRunning
	src.Status.Reason = ""
	var phase srapi.Phase
	if src.Status.StarRocksFeStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.StarRocksFeStatus.StarRocksComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
	if src.Status.StarRocksBeStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.StarRocksBeStatus.StarRocksComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
	if src.Status.StarRocksCnStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.StarRocksCnStatus.StarRocksComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
}

// handleSyncClusterError handle errors from sub-controller, and log it in StarRocksCluster Status
func handleSyncClusterError(src *srapi.StarRocksCluster, subController subcontrollers.ClusterSubController, err error) {
	reason := err.Error()
	switch subController.(type) {
	case *fe.FeController:
		reason = fmt.Sprintf("error from FE controller: %v", reason)
	case *be.BeController:
		reason = fmt.Sprintf("error from BE controller: %v", reason)
	case *cn.CnController:
		reason = fmt.Sprintf("error from CN controller: %v", reason)
	case *feproxy.FeProxyController:
		reason = fmt.Sprintf("error from fe-proxy controller: %v", reason)
	}

	src.Status.Phase = srapi.ClusterFailed
	src.Status.Reason = reason
}

func requeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// GetPhaseFromComponent return the Phase of Cluster or Warehouse based on the component status.
// It returns empty string if not sure the phase.
func GetPhaseFromComponent(componentStatus *srapi.StarRocksComponentStatus) srapi.Phase {
	if componentStatus == nil {
		return ""
	}
	if componentStatus.Phase == srapi.ComponentReconciling {
		return srapi.ClusterReconciling
	}
	if componentStatus.Phase == srapi.ComponentFailed {
		return srapi.ClusterFailed
	}
	return ""
}
