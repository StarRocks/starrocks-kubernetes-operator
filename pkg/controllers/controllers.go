package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

func SetupClusterReconciler(mgr ctrl.Manager) error {
	feController := fe.New(mgr.GetClient(), mgr.GetEventRecorderFor)
	beController := be.New(mgr.GetClient(), mgr.GetEventRecorderFor)
	cnController := cn.New(mgr.GetClient(), mgr.GetEventRecorderFor)
	feProxyController := feproxy.New(mgr.GetClient(), mgr.GetEventRecorderFor)
	subcs := []subcontrollers.ClusterSubController{
		feController, beController, cnController, feProxyController,
	}

	reconciler := &StarRocksClusterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("starrockscluster-controller"),
		Scs:      subcs,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// cannot add Owns(&v2.HorizontalPodAutoscaler{}), because if a kubernetes version is lower than 1.23,
	// v2.HorizontalPodAutoscaler does not exist.
	// todo(yandongxiao): watch the HPA resource
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// SetupWarehouseReconciler
// Why do we need a namespace parameter?
//  1. Warehouse CRD is an optional feature, and user may not install it.
//  2. We try to use list Warehouses operation to check if Warehouse CRD exists or not.
//  3. By Default, It needs the cluster scope permission.
func SetupWarehouseReconciler(mgr ctrl.Manager, namespace string) error {
	var listOpts []client.ListOption
	if namespace != "" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}
	// check StarRocksWarehouse CRD exists or not
	if err := mgr.GetAPIReader().List(context.Background(), &srapi.StarRocksWarehouseList{}, listOpts...); err != nil {
		if meta.IsNoMatchError(err) {
			// StarRocksWarehouse CRD is not found, skip StarRocksWarehouseReconciler
			return nil
		}
		return err
	}

	reconciler := &StarRocksWarehouseReconciler{
		Client:         mgr.GetClient(),
		recorder:       mgr.GetEventRecorderFor("starrockswarehouse-controller"),
		subControllers: []subcontrollers.WarehouseSubController{cn.New(mgr.GetClient(), mgr.GetEventRecorderFor)},
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksWarehouseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// cannot add Owns(&v2.HorizontalPodAutoscaler{}), because if a kubernetes version is lower than 1.23,
	// v2.HorizontalPodAutoscaler does not exist.
	// todo(yandongxiao): watch the HPA resource
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksWarehouse{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
