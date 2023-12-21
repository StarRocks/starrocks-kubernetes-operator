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
	"errors"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
)

// StarRocksWarehouseReconciler reconciles a StarRocksWarehouse object
type StarRocksWarehouseReconciler struct {
	client.Client
	recorder       record.EventRecorder
	subControllers []subcontrollers.WarehouseSubController
}

// +kubebuilder:rbac:groups=starrocks.com,resources=starrockswarehouses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=starrocks.com,resources=starrockswarehouses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=starrocks.com,resources=starrockswarehouses/finalizers,verbs=update
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
// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *StarRocksWarehouseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("StarRocksWarehouseReconciler").
		WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("begin to reconcile StarRocksWarehouse")

	logger.Info("get StarRocksWarehouse CR from kubernetes")
	warehouse := &srapi.StarRocksWarehouse{}
	err := r.Client.Get(ctx, req.NamespacedName, warehouse)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("StarRocksWarehouse CR is not found, maybe deleted, begin to clear warehouse")
			for _, controller := range r.subControllers {
				kvs := []interface{}{"subController", controller.GetControllerName()}
				logger.Info("sub controller begin to clear warehouse", kvs...)
				if err = controller.ClearWarehouse(ctx, req.Namespace, req.Name); err != nil {
					logger.Error(err, "failed to clear warehouse", kvs...)
				}
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get StarRocksWarehouse CR failed")
		return ctrl.Result{}, err
	}

	if warehouse.Status.WarehouseComponentStatus == nil {
		warehouse.Status.WarehouseComponentStatus = &srapi.StarRocksCnStatus{
			StarRocksComponentStatus: srapi.StarRocksComponentStatus{
				Phase: srapi.ComponentReconciling,
			},
		}
	}

	for _, controller := range r.subControllers {
		kvs := []interface{}{"subController", controller.GetControllerName()}
		logger.Info("sub controller sync spec", kvs...)
		if err = controller.SyncWarehouse(ctx, warehouse); err != nil {
			handled := handleSyncWarehouseError(ctx, err, warehouse)
			if updateError := r.UpdateStarRocksWarehouseStatus(ctx, warehouse); updateError != nil {
				return ctrl.Result{}, updateError
			}
			if handled {
				return ctrl.Result{}, nil
			} else {
				return ctrl.Result{}, err
			}
		}
	}

	for _, controller := range r.subControllers {
		kvs := []interface{}{"subController", controller.GetControllerName()}
		logger.Info("sub controller update warehouse status", kvs...)
		if err = controller.UpdateWarehouseStatus(ctx, warehouse); err != nil {
			logger.Error(err, "update warehouse status failed", kvs...)
			warehouse.Status.Phase = srapi.ComponentFailed
			warehouse.Status.Reason = err.Error()
			if updateError := r.UpdateStarRocksWarehouseStatus(ctx, warehouse); updateError != nil {
				logger.Error(err, "failed to update warehouse status")
			}
			return requeueIfError(err)
		}
	}

	logger.Info("update StarRocksWarehouse status")
	err = r.UpdateStarRocksWarehouseStatus(ctx, warehouse)
	if err != nil {
		logger.Error(err, "update StarRocksWarehouse status failed")
		return ctrl.Result{}, err
	}

	logger.Info("reconcile StarRocksWarehouse success")
	return ctrl.Result{}, nil
}

// UpdateStarRocksWarehouseStatus update the status of warehouse.
func (r *StarRocksWarehouseReconciler) UpdateStarRocksWarehouseStatus(ctx context.Context, warehouse *srapi.StarRocksWarehouse) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		actualWarehouse := &srapi.StarRocksWarehouse{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: warehouse.Namespace, Name: warehouse.Name}, actualWarehouse); err != nil {
			return err
		}
		actualWarehouse.Status = warehouse.Status
		return r.Client.Status().Update(ctx, actualWarehouse)
	})
}

// handleSyncWarehouseError handles the error returned from SyncWarehouse, and update warehouse status.
// If handled return true, else return false
func handleSyncWarehouseError(ctx context.Context, err error, warehouse *srapi.StarRocksWarehouse) bool {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Error(err, "sub controller reconciles spec failed")
	warehouse.Status.Phase = srapi.ComponentFailed
	warehouse.Status.Reason = err.Error()
	if errors.Is(err, cn.SpecMissingError) || errors.Is(err, cn.StarRocksClusterMissingError) ||
		errors.Is(err, cn.FeNotReadyError) || errors.Is(err, cn.GetFeFeatureInfoError) {
		return true
	}
	return false
}
