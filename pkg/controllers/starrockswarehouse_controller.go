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

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/cn"
)

// CelerDataWarehouseReconciler reconciles a CelerDataWarehouse object
type CelerDataWarehouseReconciler struct {
	client.Client
	recorder       record.EventRecorder
	subControllers []subcontrollers.WarehouseSubController
	denyList       string
}

// +kubebuilder:rbac:groups=celerdata.com,resources=celerdatawarehouses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celerdata.com,resources=celerdatawarehouses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=celerdata.com,resources=celerdatawarehouses/finalizers,verbs=update
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
func (r *CelerDataWarehouseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("CelerDataWarehouseReconciler").
		WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("begin to reconcile CelerDataWarehouse")

	logger.Info("get CelerDataWarehouse CR from kubernetes")
	warehouse := &cdapi.CelerDataWarehouse{}
	client := r.Client
	err := client.Get(ctx, req.NamespacedName, warehouse)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("CelerDataWarehouse CR is not found, maybe deleted, begin to clear warehouse")
			for _, controller := range r.subControllers {
				kvs := []interface{}{"subController", controller.GetControllerName()}
				logger.Info("sub controller begin to clear warehouse", kvs...)
				if err = controller.ClearWarehouse(ctx, req.Namespace, req.Name); err != nil {
					logger.Error(err, "failed to clear warehouse", kvs...)
				}
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get CelerDataWarehouse CR failed")
		return ctrl.Result{}, err
	}

	if warehouse.Status.WarehouseComponentStatus == nil {
		warehouse.Status.WarehouseComponentStatus = &cdapi.CelerDataCnStatus{
			CelerDataComponentStatus: cdapi.CelerDataComponentStatus{
				Phase: cdapi.ComponentReconciling,
			},
		}
	}

	for _, controller := range r.subControllers {
		kvs := []interface{}{"subController", controller.GetControllerName()}
		logger.Info("sub controller sync spec", kvs...)
		if err = controller.SyncWarehouse(ctx, warehouse); err != nil {
			handled := handleSyncWarehouseError(ctx, err, warehouse)
			if updateError := r.UpdateCelerDataWarehouseStatus(ctx, warehouse); updateError != nil {
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
			warehouse.Status.Phase = cdapi.ComponentFailed
			warehouse.Status.Reason = err.Error()
			if updateError := r.UpdateCelerDataWarehouseStatus(ctx, warehouse); updateError != nil {
				logger.Error(err, "failed to update warehouse status")
			}
			return requeueIfError(err)
		}
	}

	logger.Info("update CelerDataWarehouse status")
	err = r.UpdateCelerDataWarehouseStatus(ctx, warehouse)
	if err != nil {
		logger.Error(err, "update CelerDataWarehouse status failed")
		return ctrl.Result{}, err
	}

	logger.Info("reconcile CelerDataWarehouse success")
	return ctrl.Result{}, nil
}

// UpdateCelerDataWarehouseStatus update the status of warehouse.
func (r *CelerDataWarehouseReconciler) UpdateCelerDataWarehouseStatus(ctx context.Context, warehouse *cdapi.CelerDataWarehouse) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		actualWarehouse := &cdapi.CelerDataWarehouse{}
		client := r.Client
		if err := client.Get(ctx, types.NamespacedName{Namespace: warehouse.Namespace, Name: warehouse.Name}, actualWarehouse); err != nil {
			return err
		}
		actualWarehouse.Status = warehouse.Status
		return client.Status().Update(ctx, actualWarehouse)
	})
}

// handleSyncWarehouseError handles the error returned from SyncWarehouse, and update warehouse status.
// If handled return true, else return false
func handleSyncWarehouseError(ctx context.Context, err error, warehouse *cdapi.CelerDataWarehouse) bool {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Error(err, "sub controller reconciles spec failed")
	warehouse.Status.Phase = cdapi.ComponentFailed
	warehouse.Status.Reason = err.Error()
	if errors.Is(err, cn.ErrSpecIsMissing) || errors.Is(err, cn.ErrCelerDataClusterIsMissing) ||
		errors.Is(err, cn.ErrFeIsNotReady) || errors.Is(err, cn.ErrFailedToGetFeFeatureList) {
		return true
	}
	return false
}
