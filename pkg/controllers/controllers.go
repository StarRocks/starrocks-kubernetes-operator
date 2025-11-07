package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feproxy"
)

const (
	componentTypeFE = "fe"
	componentTypeBE = "be"
	componentTypeCN = "cn"
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
	isUpgradeScenario bool,
	fe, be, cn, feproxy subcontrollers.ClusterSubController,
) []subcontrollers.ClusterSubController {
	if isUpgradeScenario {
		return []subcontrollers.ClusterSubController{be, cn, fe, feproxy}
	}

	// default order
	return []subcontrollers.ClusterSubController{fe, be, cn, feproxy}
}

// isUpgrade determines if the current reconciliation is an upgrade scenario.
// Returns true only if StatefulSets exist AND there are image changes detected.
func isUpgrade(ctx context.Context, kubeClient client.Client, cluster *srapi.StarRocksCluster) bool {
	logger := logr.FromContextOrDiscard(ctx)

	// Check FE first (always required in StarRocks)
	feSts := &appsv1.StatefulSet{}
	feExists := kubeClient.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name + "-fe",
	}, feSts) == nil

	beSts := &appsv1.StatefulSet{}
	beExists := kubeClient.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name + "-be",
	}, beSts) == nil

	// Corrupted state safeguard: BE exists but FE doesn't (invalid configuration).
	// Treat as initial deployment so FE is reconciled first.
	// Rationale: FE is a prerequisite for BE/CN; prioritizing FE allows recovery without misordering.
	if beExists && !feExists {
		logger.Info("WARNING: BE StatefulSet exists without FE - treating as initial deployment to recreate FE first")
		return false
	}

	if !feExists {
		return false
	}

	return checkForImageChanges(ctx, kubeClient, cluster)
}

// checkForImageChanges compares the desired StarRocks component images
// against the currently deployed StatefulSet images.
// Returns true if any component image differs.
func checkForImageChanges(ctx context.Context, kubeClient client.Client, cluster *srapi.StarRocksCluster) bool {
	if cluster.Spec.StarRocksFeSpec != nil {
		desiredImage := cluster.Spec.StarRocksFeSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-fe")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	if cluster.Spec.StarRocksBeSpec != nil {
		desiredImage := cluster.Spec.StarRocksBeSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-be")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	if cluster.Spec.StarRocksCnSpec != nil {
		desiredImage := cluster.Spec.StarRocksCnSpec.Image
		currentImage := getCurrentImageFromStatefulSet(ctx, kubeClient, cluster.Namespace, cluster.Name+"-cn")
		if currentImage != "" && desiredImage != currentImage {
			return true
		}
	}

	return false
}

// getCurrentImageFromStatefulSet returns the container image used in a StatefulSet.
// Returns an empty string if the StatefulSet is missing or has no containers.
func getCurrentImageFromStatefulSet(ctx context.Context, kubeClient client.Client, namespace, name string) string {
	var sts appsv1.StatefulSet
	if err := kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &sts); err != nil {
		// StatefulSet does not exist
		return ""
	}

	containers := sts.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return ""
	}

	// Prefer a named match for known StarRocks components
	for _, c := range containers {
		switch c.Name {
		case "fe", "be", "cn":
			return c.Image
		}
	}

	// Fallback for backward compatibility
	return containers[0].Image
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

// isComponentReady checks if a component is ready by verifying:
// 1. Its service endpoints have ready addresses (pods are healthy)
// 2. Its StatefulSet rollout is complete (no pending updates)
func isComponentReady(ctx context.Context, k8sClient client.Client, cluster *srapi.StarRocksCluster, componentType string) bool {
	logger := logr.FromContextOrDiscard(ctx)

	var serviceName string
	var statefulSetName string

	switch componentType {
	case componentTypeFE:
		if cluster.Spec.StarRocksFeSpec == nil {
			return true // Component not configured, consider it ready
		}
		serviceName = rutils.ExternalServiceName(cluster.Name, cluster.Spec.StarRocksFeSpec)
		statefulSetName = cluster.Name + "-fe"
	case componentTypeBE:
		if cluster.Spec.StarRocksBeSpec == nil {
			return true
		}
		serviceName = rutils.ExternalServiceName(cluster.Name, cluster.Spec.StarRocksBeSpec)
		statefulSetName = cluster.Name + "-be"
	case componentTypeCN:
		if cluster.Spec.StarRocksCnSpec == nil {
			return true
		}
		serviceName = rutils.ExternalServiceName(cluster.Name, cluster.Spec.StarRocksCnSpec)
		statefulSetName = cluster.Name + "-cn"
	default:
		return true
	}

	// Check 1: Service endpoints must have ready addresses
	endpoints := corev1.Endpoints{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      serviceName,
	}, &endpoints); err != nil {
		logger.V(5).Info("get component service endpoints failed", "component", componentType, "serviceName", serviceName, "error", err)
		return false
	}

	hasReadyEndpoints := false
	for _, sub := range endpoints.Subsets {
		if len(sub.Addresses) > 0 {
			hasReadyEndpoints = true
			break
		}
	}

	if !hasReadyEndpoints {
		logger.Info("component not ready: no ready endpoints", "component", componentType, "serviceName", serviceName)
		return false
	}

	// Check 2: StatefulSet rollout must be complete (currentRevision == updateRevision)
	sts := &appsv1.StatefulSet{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      statefulSetName,
	}, sts); err != nil {
		logger.V(5).Info("get component StatefulSet failed", "component", componentType, "statefulSetName", statefulSetName, "error", err)
		return false
	}

	// Check if StatefulSet controller has observed our latest spec change
	if sts.Generation != sts.Status.ObservedGeneration {
		logger.Info("component not ready: StatefulSet spec change not yet observed",
			"component", componentType,
			"statefulSetName", statefulSetName,
			"generation", sts.Generation,
			"observedGeneration", sts.Status.ObservedGeneration)
		return false
	}

	// Check if rollout is complete
	if sts.Status.CurrentRevision != sts.Status.UpdateRevision {
		logger.Info("component not ready: StatefulSet rollout in progress",
			"component", componentType,
			"statefulSetName", statefulSetName,
			"currentRevision", sts.Status.CurrentRevision,
			"updateRevision", sts.Status.UpdateRevision)
		return false
	}

	// Check if all replicas are ready
	if sts.Status.ReadyReplicas != *sts.Spec.Replicas {
		logger.Info("component not ready: waiting for replicas",
			"component", componentType,
			"statefulSetName", statefulSetName,
			"readyReplicas", sts.Status.ReadyReplicas,
			"desiredReplicas", *sts.Spec.Replicas)
		return false
	}

	logger.Info("component is ready",
		"component", componentType,
		"serviceName", serviceName,
		"readyAddresses", len(endpoints.Subsets[0].Addresses),
		"revision", sts.Status.CurrentRevision)
	return true
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
