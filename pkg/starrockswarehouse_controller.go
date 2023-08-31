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

package pkg

import (
	"context"
	"os"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/cn"
	appv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Controllers = append(Controllers, &StarRocksWarehouseReconciler{})
}

// StarRocksWarehouseReconciler reconciles a StarRocksWarehouse object
type StarRocksWarehouseReconciler struct {
	client.Client
	recorder       record.EventRecorder
	subControllers []sub_controller.WarehouseSubController
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *StarRocksWarehouseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("StarRocksWarehouseReconciler reconcile the StarRocksWarehouse CR, namespace=%v, name=%v", req.Namespace, req.Name)

	klog.Infof("get StarRocksWarehouse CR, namespace=%v, name=%v", req.Namespace, req.Name)
	warehouse := &srapi.StarRocksWarehouse{}
	err := r.Client.Get(ctx, req.NamespacedName, warehouse)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("StarRocksWarehouse CR is not found, begin to clear warehouse, namespace=%v, name=%v",
				req.Namespace, req.Name)
			for _, controller := range r.subControllers {
				if err = controller.ClearWarehouse(ctx, req.Namespace, req.Name); err != nil {
					klog.Errorf("failed to clear warehouse %s/%s, error=%v", req.Namespace, req.Name, err)
				}
			}
			return ctrl.Result{}, nil
		}
		klog.Errorf("failed to get StarRocksWarehouse CR, namespace=%v, name=%v, error=%v", req.Namespace, req.Name, err)
		return ctrl.Result{}, err
	}

	warehouse = warehouse.DeepCopy()

	for _, controller := range r.subControllers {
		klog.Infof("StarRocksWarehouseReconciler reconcile component, namespace=%v, name=%v, controller=%v",
			warehouse.Namespace, warehouse.Name, controller.GetControllerName())
		if err := controller.SyncWarehouse(ctx, warehouse); err != nil {
			if err == cn.SpecMissingError {
				klog.Infof("the spec part is invalid, skip sync warehouse %s/%s", warehouse.Namespace, warehouse.Name)
				return ctrl.Result{}, nil
			} else if err == cn.StarRocksClusterMissingError {
				klog.Infof("StarRocksCluster %s/%s not found, skip sync warehouse %s/%s",
					warehouse.Namespace, warehouse.Spec.StarRocksCluster, warehouse.Namespace, warehouse.Name)
				return ctrl.Result{}, nil
			} else if err == cn.FeNotOkError {
				klog.Infof("StarRocksFe is not ok, skip sync warehouse %s/%s", warehouse.Namespace, warehouse.Name)
				return ctrl.Result{}, nil
			} else if err == cn.GetFeFeatureInfoError {
				klog.Infof("failed to get FE feature or FE does not support multi-warehouse, skip sync warehouse %s/%s",
					warehouse.Namespace, warehouse.Name)
				return ctrl.Result{}, nil
			}
			klog.Errorf("failed to reconcile component, namespace=%v, name=%v, controller=%v, error=%v",
				warehouse.Namespace, warehouse.Name, controller.GetControllerName(), err)
			return ctrl.Result{}, err
		}
	}

	for _, controller := range r.subControllers {
		klog.Infof("StarRocksWarehouseReconciler update component status, namespace=%v, name=%v, controller=%v",
			warehouse.Namespace, warehouse.Name, controller.GetControllerName())
		if err := controller.UpdateWarehouseStatus(warehouse); err != nil {
			klog.Infof("failed to reconcile component, namespace=%v, name=%v, controller=%v, error=%v",
				warehouse.Namespace, warehouse.Name, controller.GetControllerName(), err)
			return requeueIfError(err)
		}
	}

	warehouse.Status.Phase = srapi.ClusterRunning
	phase := GetPhaseFromComponent(&warehouse.Status.WarehouseComponentStatus.StarRocksComponentStatus)
	if phase != "" {
		warehouse.Status.Phase = phase
	}

	return ctrl.Result{}, r.UpdateStarRocksWarehouseStatus(ctx, warehouse)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksWarehouseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksWarehouse{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&v2.HorizontalPodAutoscaler{}).
		Complete(r)
}

// Init initial the StarRocksWarehouseReconciler for reconcile.
func (r *StarRocksWarehouseReconciler) Init(mgr ctrl.Manager) {
	if err := (&StarRocksWarehouseReconciler{
		Client:         mgr.GetClient(),
		recorder:       mgr.GetEventRecorderFor(name),
		subControllers: []sub_controller.WarehouseSubController{cn.New(mgr.GetClient())},
	}).SetupWithManager(mgr); err != nil {
		klog.Error(err, "failed to setup StarRocksWarehouseReconciler")
		os.Exit(1)
	}
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
