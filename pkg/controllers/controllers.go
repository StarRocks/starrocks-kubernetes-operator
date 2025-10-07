package controllers

import (
	"context"

	"github.com/go-logr/logr"
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

	reconciler := &StarRocksClusterReconciler{
		Client:            mgr.GetClient(),
		Recorder:          mgr.GetEventRecorderFor("starrockscluster-controller"),
		FeController:      feController,
		BeController:      beController,
		CnController:      cnController,
		FeProxyController: feProxyController,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

// getControllersInOrder returns controllers in the appropriate order based on deployment scenario
func getControllersInOrder(
	ctx context.Context,
	client client.Client,
	cluster *srapi.StarRocksCluster,
	fe, be, cn, feproxy subcontrollers.ClusterSubController,
) []subcontrollers.ClusterSubController {
	logger := logr.FromContextOrDiscard(ctx)

	// Auto-detect upgrade scenario by checking if this is an image change
	if isUpgrade(ctx, client, cluster) {
		logger.Info("upgrade detected: using BE-first ordering")
		// upgrade order: BE/CN -> FE
		return []subcontrollers.ClusterSubController{be, cn, fe, feproxy}
	}

	logger.Info("initial deployment detected: using FE-first ordering")
	// initial deployment (default) order: FE -> BE -> CN -> FeProxy
	return []subcontrollers.ClusterSubController{fe, be, cn, feproxy}
}

// isUpgrade detects if this reconciliation is due to an upgrade scenario
// An upgrade is detected when:
// 1. The cluster status shows it's already running
// 2. The desired image versions differ from currently deployed versions
func isUpgrade(ctx context.Context, kubeClient client.Client, cluster *srapi.StarRocksCluster) bool {
	logger := logr.FromContextOrDiscard(ctx)

	// Check if cluster is already running by looking at status
	if cluster.Status.Phase != srapi.ClusterRunning {
		logger.Info("cluster not in running state, assuming initial deployment", "currentPhase", cluster.Status.Phase)
		return false
	}

	// Compare desired images with current images by checking StatefulSets
	hasImageChanges := checkForImageChanges(ctx, kubeClient, cluster)
	if hasImageChanges {
		logger.Info("image changes detected in running cluster, this is an upgrade")
	} else {
		logger.Info("no image changes detected")
	}

	return hasImageChanges
}

// checkForImageChanges compares desired spec images with currently deployed StatefulSet images
func checkForImageChanges(ctx context.Context, kubeClient client.Client, cluster *srapi.StarRocksCluster) bool {
	// Check FE image changes
	if cluster.Spec.StarRocksFeSpec != nil {
		desiredImage := cluster.Spec.StarRocksFeSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-fe")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	// Check BE image changes
	if cluster.Spec.StarRocksBeSpec != nil {
		desiredImage := cluster.Spec.StarRocksBeSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-be")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	// Check CN image changes
	if cluster.Spec.StarRocksCnSpec != nil {
		desiredImage := cluster.Spec.StarRocksCnSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-cn")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	return false
}

// getCurrentImageFromStatefulSet gets the current image from a deployed StatefulSet
func getCurrentImageFromStatefulSet(ctx context.Context, kubeClient client.Client, namespace, name string) string {
	var st appsv1.StatefulSet
	err := kubeClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &st)
	if err != nil {
		// StatefulSet doesn't exist yet, this is initial deployment
		return ""
	}

	// Get image from first container (StarRocks container)
	if len(st.Spec.Template.Spec.Containers) > 0 {
		return st.Spec.Template.Spec.Containers[0].Image
	}

	return ""
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
