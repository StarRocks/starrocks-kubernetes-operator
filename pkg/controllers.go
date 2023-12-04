package pkg

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/feproxy"
)

func SetupClusterReconciler(mgr ctrl.Manager) error {
	subcs := make(map[string]sub_controller.ClusterSubController)
	feController := fe.New(mgr.GetClient())
	subcs[feControllerName] = feController
	cnController := cn.New(mgr.GetClient())
	subcs[cnControllerName] = cnController
	beController := be.New(mgr.GetClient())
	subcs[beControllerName] = beController
	feProxyController := feproxy.New(mgr.GetClient())
	subcs[feProxyControllerName] = feProxyController

	reconciler := &StarRocksClusterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor(name),
		Scs:      subcs,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		klog.Error(err, " unable to create controller ", "controller ", "StarRocksCluster ")
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

func SetupWarehouseReconciler(mgr ctrl.Manager) error {
	// check StarRocksWarehouse CRD exists or not
	if err := mgr.GetAPIReader().List(context.Background(), &srapi.StarRocksWarehouseList{}); err != nil {
		if meta.IsNoMatchError(err) {
			klog.Infof("StarRocksWarehouse CRD is not found, skip StarRocksWarehouseReconciler")
			return nil
		}
		return err
	}

	reconciler := &StarRocksWarehouseReconciler{
		Client:         mgr.GetClient(),
		recorder:       mgr.GetEventRecorderFor(name),
		subControllers: []sub_controller.WarehouseSubController{cn.New(mgr.GetClient())},
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		klog.Error(err, "failed to setup StarRocksWarehouseReconciler")
		return err
	}
	return nil
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
