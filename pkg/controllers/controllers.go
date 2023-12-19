package controllers

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

func SetupClusterReconciler(mgr ctrl.Manager) error {
	subcs := []subcontrollers.ClusterSubController{
		fe.New(mgr.GetClient()),
		be.New(mgr.GetClient()),
		cn.New(mgr.GetClient()),
		feproxy.New(mgr.GetClient()),
	}

	reconciler := &StarRocksClusterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor(_controllerName),
		Scs:      subcs,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// can not add Owns(&v2.HorizontalPodAutoscaler{}), because if kubernetes version is lower than 1.23,
	// v2.HorizontalPodAutoscaler does not exist.
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func SetupWarehouseReconciler(ctx context.Context, mgr ctrl.Manager) error {
	// check StarRocksWarehouse CRD exists or not
	if err := mgr.GetAPIReader().List(ctx, &srapi.StarRocksWarehouseList{}); err != nil {
		if meta.IsNoMatchError(err) {
			// StarRocksWarehouse CRD is not found, skip StarRocksWarehouseReconciler
			return nil
		}
		return err
	}

	reconciler := &StarRocksWarehouseReconciler{
		Client:         mgr.GetClient(),
		recorder:       mgr.GetEventRecorderFor(_controllerName),
		subControllers: []subcontrollers.WarehouseSubController{cn.New(mgr.GetClient())},
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksWarehouseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// can not add Owns(&v2.HorizontalPodAutoscaler{}), because if kubernetes version is lower than 1.23,
	// v2.HorizontalPodAutoscaler does not exist.
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksWarehouse{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
