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

package cn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/log"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	subc "github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

type CnController struct {
	k8sClient          client.Client
	Recorder           record.EventRecorder
	addEnvForWarehouse bool
}

func New(k8sClient client.Client, recorderFor subc.GetEventRecorderForFunc) *CnController {
	controller := &CnController{
		k8sClient: k8sClient,
	}
	controller.Recorder = recorderFor(controller.GetControllerName())
	return controller
}

func (cc *CnController) GetControllerName() string {
	return "cnController"
}

var ErrWarehouseNameIsNotAllowed = errors.New("warehouse name should not equal to cluster name")
var ErrSpecIsMissing = errors.New("spec.template or spec.starRocksCluster is missing")
var ErrStarRocksClusterIsMissing = errors.New("custom resource StarRocksCluster is missing")
var ErrFeIsNotReady = errors.New("component fe is not ready")
var ErrShouldRunInSharedDataMode = errors.New("StarRocks Cluster should run in shared_data mode")
var ErrFailedToGetFeFeatureList = errors.New("failed to invoke FE /api/v2/feature or FE does not support multi-warehouse feature")

func (cc *CnController) SyncWarehouse(ctx context.Context, warehouse *srapi.StarRocksWarehouse) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncWarehouse)
	ctx = logr.NewContext(ctx, logger)

	template := warehouse.Spec.Template
	if warehouse.Spec.StarRocksCluster == "" || template == nil {
		return ErrSpecIsMissing
	}

	if warehouse.Name == warehouse.Spec.StarRocksCluster {
		return ErrWarehouseNameIsNotAllowed
	}

	logger.Info("get StarRocksCluster CR from kubernetes")
	_, err := cc.getStarRocksCluster(ctx, warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ErrStarRocksClusterIsMissing
		}
		return err
	}

	logger.Info("get fe config to make sure StarRocks run in shared_data mode")
	feConfig, err := cc.getFeConfig(ctx, warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		return err
	}
	if !fe.IsRunInSharedDataMode(feConfig) {
		return ErrShouldRunInSharedDataMode
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, warehouse.Namespace, warehouse.Spec.StarRocksCluster) {
		return ErrFeIsNotReady
	}

	return cc.SyncCnSpec(ctx, object.NewFromWarehouse(warehouse), template.ToCnSpec(), warehouse.Status.WarehouseComponentStatus)
}

func (cc *CnController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)

	if src.Spec.StarRocksCnSpec == nil {
		if err := cc.ClearCluster(ctx, src); err != nil {
			logger.Error(err, "clear resource failed")
		}
		return nil
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, src.Namespace, src.Name) {
		return nil
	}

	feConfig, err := cc.getFeConfig(ctx, src.Namespace, src.Name)
	if err != nil {
		return err
	}
	drSpec := src.Spec.DisasterRecovery
	drStatus := src.Status.DisasterRecoveryStatus
	if b, _ := fe.ShouldEnterDisasterRecoveryMode(drSpec, drStatus, feConfig); b {
		// return nil because in disaster recovery mode, we do not need to sync the CN.
		return nil
	}

	err = cc.SyncCnSpec(ctx, object.NewFromCluster(src), src.Spec.StarRocksCnSpec, src.Status.StarRocksCnStatus)
	defer func() {
		// we do not record an event if the error is nil, because this will cause too many events to be recorded.
		if err != nil {
			cc.Recorder.Event(src, corev1.EventTypeWarning, "SyncCnFailed", err.Error())
		}
	}()
	return err
}

//nolint:gocyclo
func (cc *CnController) SyncCnSpec(ctx context.Context, object object.StarRocksObject,
	cnSpec *srapi.StarRocksCnSpec, cnStatus *srapi.StarRocksCnStatus) error {
	logger := logr.FromContextOrDiscard(ctx)

	if err := cc.mutating(cnSpec); err != nil {
		return err
	}

	if err := cc.validating(cnSpec); err != nil {
		return err
	}

	logger.V(log.DebugLevel).Info("get cn config to resolve ports", "ConfigMapInfo", cnSpec.ConfigMapInfo)
	cnConfig, err := cc.GetCnConfig(ctx, cnSpec, object.Namespace)
	if err != nil {
		return err
	}
	logger.V(log.DebugLevel).Info("get fe config to resolve ports", "ConfigMapInfo", cnSpec.ConfigMapInfo)
	feconfig, err := cc.getFeConfig(ctx, object.Namespace, object.ClusterName)
	if err != nil {
		return err
	}
	cnConfig[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	cnConfig[rutils.HTTP_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.HTTP_PORT)), 10)

	// build and deploy statefulset
	podTemplateSpec, err := cc.buildPodTemplate(ctx, object, cnSpec, cnConfig)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}

	expectSTS := statefulset.MakeStatefulset(object, cnSpec, podTemplateSpec)
	if err = k8sutils.ApplyStatefulSet(ctx, cc.k8sClient, &expectSTS, true, rutils.StatefulSetDeepEqual); err != nil {
		return err
	}

	// build and deploy service
	defaultLabels := load.Labels(object.SubResourcePrefixName, cnSpec)
	externalsvc := rutils.BuildExternalService(object, cnSpec, cnConfig,
		load.Selector(object.SubResourcePrefixName, cnSpec), defaultLabels)
	internalService := generateInternalService(object, cnSpec, &externalsvc, cnConfig, defaultLabels)

	if err := k8sutils.ApplyService(ctx, cc.k8sClient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "sync CN external service failed")
		return err
	}

	if err := k8sutils.ApplyService(ctx, cc.k8sClient, internalService, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "sync CN search service failed")
		return err
	}

	// sync autoscaler
	if cnSpec.AutoScalingPolicy != nil {
		err = cc.deployAutoScaler(ctx, object, cnSpec, *cnSpec.AutoScalingPolicy)
	} else {
		// If the HPA policy is nil, delete the HPA resource.
		if cnStatus != nil {
			err = cc.deleteAutoScaler(ctx, object, cnStatus.HorizontalScaler.Version)
		} else {
			err = cc.deleteAutoScaler(ctx, object, "")
		}
	}
	if err != nil {
		logger.Error(err, "sync autoscaler failed")
		return err
	}

	// only take effect for shared-data mode
	if !fe.IsRunInSharedDataMode(feconfig) {
		return nil
	}

	// get actual statefulset object
	var actualSTS appsv1.StatefulSet
	namespacedName := types.NamespacedName{
		Namespace: object.Namespace,
		Name:      object.GetCNStatefulSetName(),
	}
	if err := cc.k8sClient.Get(ctx, namespacedName, &actualSTS); err != nil {
		return err
	}

	// get actual pods
	actualCNPods := corev1.PodList{}
	matchingLabels := client.MatchingLabels{
		srapi.ComponentLabelKey: srapi.DEFAULT_CN,
		srapi.OwnerReference:    object.GetCNStatefulSetName(),
	}

	err = cc.k8sClient.List(ctx, &actualCNPods, client.InNamespace(object.Namespace), matchingLabels)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "list cn pod failed")
		return err
	}
	if err = cc.SyncComputeNodesInFE(ctx, object, &expectSTS, &actualSTS, &actualCNPods, nil); err != nil {
		// Because sync compute nodes error is not a fatal error, we just log the error and return nil.
		logger.Info("sync compute nodes in FE failed", "error", err)
		return nil
	}

	return nil
}

// UpdateWarehouseStatus updates the status of StarRocksWarehouse.
func (cc *CnController) UpdateWarehouseStatus(ctx context.Context, warehouse *srapi.StarRocksWarehouse) error {
	template := warehouse.Spec.Template
	if template == nil {
		warehouse.Status.WarehouseComponentStatus = nil
		return nil
	}

	status := warehouse.Status.WarehouseComponentStatus
	status.Phase = srapi.ComponentReconciling
	return cc.UpdateStatus(ctx, object.NewFromWarehouse(warehouse), template.ToCnSpec(), status)
}

// UpdateClusterStatus update the status of StarRocksCluster.
func (cc *CnController) UpdateClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	cnSpec := src.Spec.StarRocksCnSpec
	if cnSpec == nil {
		src.Status.StarRocksCnStatus = nil
		return nil
	}

	if src.Status.StarRocksCnStatus == nil {
		src.Status.StarRocksCnStatus = &srapi.StarRocksCnStatus{
			StarRocksComponentStatus: srapi.StarRocksComponentStatus{
				Phase: srapi.ComponentReconciling,
			},
		}
	}
	cs := src.Status.StarRocksCnStatus
	cs.Phase = srapi.ComponentReconciling

	return cc.UpdateStatus(ctx, object.NewFromCluster(src), cnSpec, cs)
}

func (cc *CnController) UpdateStatus(ctx context.Context, object object.StarRocksObject,
	cnSpec *srapi.StarRocksCnSpec, cnStatus *srapi.StarRocksCnStatus) error {
	var actualSTS appsv1.StatefulSet
	logger := logr.FromContextOrDiscard(ctx)

	statefulSetName := load.Name(object.SubResourcePrefixName, cnSpec)
	namespacedName := types.NamespacedName{Namespace: object.Namespace, Name: statefulSetName}
	if err := cc.k8sClient.Get(ctx, namespacedName, &actualSTS); apierrors.IsNotFound(err) {
		logger.Info("cn statefulset is not found")
		return nil
	}

	if cnSpec.AutoScalingPolicy != nil {
		cnStatus.HorizontalScaler.Name = cc.generateAutoScalerName(object.SubResourcePrefixName, cnSpec)
		cnStatus.HorizontalScaler.Version = cnSpec.AutoScalingPolicy.Version.Complete(k8sutils.KUBE_MAJOR_VERSION,
			k8sutils.KUBE_MINOR_VERSION)
	} else {
		cnStatus.HorizontalScaler = srapi.HorizontalScaler{}
	}

	cnStatus.ServiceName = service.ExternalServiceName(object.SubResourcePrefixName, cnSpec)
	cnStatus.ResourceNames = rutils.MergeSlices(cnStatus.ResourceNames, []string{statefulSetName})

	// get the selector and replicas field from statefulset
	cnStatus.Replicas = actualSTS.Status.Replicas
	selector, err := metav1.LabelSelectorAsSelector(actualSTS.Spec.Selector)
	if err != nil {
		logger.Error(err, "convert label selector to selector failed", "selector", actualSTS.Spec.Selector)
		return err
	}
	cnStatus.Selector = selector.String()

	if err := subc.UpdateStatus(&cnStatus.StarRocksComponentStatus, cc.k8sClient,
		object.Namespace, load.Name(object.SubResourcePrefixName, cnSpec),
		pod.Labels(object.SubResourcePrefixName, cnSpec), subc.StatefulSetLoadType); err != nil {
		return err
	}

	return nil
}

// ClearWarehouse clear the warehouse resource. It is different from ClearResources, which need to clear the
// CN related resources of StarRocksCluster. ClearWarehouse only has CN related resources, when the warehouse CR
// is deleted, sub resources of CN will be deleted by k8s.
func (cc *CnController) ClearWarehouse(ctx context.Context, namespace string, warehouseName string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionClearWarehouse)
	ctx = logr.NewContext(ctx, logger)

	cnSTSName := load.Name(object.GetPrefixNameForWarehouse(warehouseName), (*srapi.StarRocksCnSpec)(nil))
	executor, err := NewSQLExecutor(ctx, cc.k8sClient, namespace, cnSTSName)
	if err != nil {
		logger.Error(err, "new SQL executor failed")
		return err
	}
	err = executor.ExecuteDropWarehouse(ctx, nil, warehouseName)
	if err != nil {
		logger.Error(err, "drop warehouse failed", "warehouse", warehouseName)
		// we do not return error here, because we want to delete the statefulset anyway.
	}

	// Remove the finalizer from cn statefulset
	var sts appsv1.StatefulSet
	if err = cc.k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      cnSTSName,
		},
		&sts); err != nil {
		return err
	}
	sts.Finalizers = nil
	if err = k8sutils.UpdateClientObject(ctx, cc.k8sClient, &sts); err != nil {
		return err
	}

	// return err
	return err
}

// Deploy autoscaler
func (cc *CnController) deployAutoScaler(ctx context.Context,
	object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec, policy srapi.AutoScalingPolicy) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update k8s hpa resource")

	labels := map[string]string{
		srapi.ComponentLabelKey: "autoscaler",
		srapi.OwnerReference:    object.Name(),
	}
	hpaParams := &rutils.HPAParams{
		Namespace:       object.Namespace,
		Name:            cc.generateAutoScalerName(object.SubResourcePrefixName, cnSpec),
		Labels:          labels,
		Version:         cnSpec.AutoScalingPolicy.Version, // cnSpec.AutoScalingPolicy can not be nil
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(object, object.GroupVersionKind())},
		ScalerPolicy:    &policy,
	}

	expectHPA := rutils.BuildHPA(hpaParams, "")
	expectHPA.SetAnnotations(make(map[string]string))

	actualHPA := hpaParams.Version.CreateEmptyHPA(k8sutils.KUBE_MAJOR_VERSION, k8sutils.KUBE_MINOR_VERSION)
	if err := cc.k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: hpaParams.Namespace,
			Name:      hpaParams.Name,
		},
		actualHPA,
	); err != nil {
		if apierrors.IsNotFound(err) {
			return cc.k8sClient.Create(ctx, expectHPA)
		}
		return err
	}

	var expectHash, actualHash string
	expectHash = hash.HashObject(expectHPA)
	if v, ok := actualHPA.GetAnnotations()[srapi.ComponentResourceHash]; ok {
		actualHash = v
	} else {
		actualHash = hash.HashObject(actualHPA)
	}

	if expectHash == actualHash {
		logger.Info("expectHash == actualHash, no need to update HPA resource")
		return nil
	}
	expectHPA.GetAnnotations()[srapi.ComponentResourceHash] = expectHash
	return cc.k8sClient.Update(ctx, expectHPA)
}

// deleteAutoScaler delete the autoscaler.
func (cc *CnController) deleteAutoScaler(ctx context.Context, object object.StarRocksObject,
	autoScalerVersion srapi.AutoScalerVersion) error {
	logger := logr.FromContextOrDiscard(ctx)

	autoScalerName := cc.generateAutoScalerName(object.SubResourcePrefixName, (*srapi.StarRocksCnSpec)(nil))
	if err := k8sutils.DeleteAutoscaler(ctx, cc.k8sClient, object.Namespace, autoScalerName,
		autoScalerVersion); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete autoscaler failed")
		return err
	}
	return nil
}

// ClearResources clear the deployed resource about cn. statefulset, services, hpa.
func (cc *CnController) ClearCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx)

	if src.Spec.StarRocksCnSpec != nil {
		return nil
	}

	cnSpec := src.Spec.StarRocksCnSpec
	statefulSetName := load.Name(src.Name, cnSpec)
	err := k8sutils.DeleteStatefulset(ctx, cc.k8sClient, src.Namespace, statefulSetName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete statefulset failed")
		return err
	}

	searchServiceName := service.SearchServiceName(src.Name, cnSpec)
	err = k8sutils.DeleteService(ctx, cc.k8sClient, src.Namespace, searchServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete search service failed")
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, cnSpec)
	err = k8sutils.DeleteService(ctx, cc.k8sClient, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete external service failed")
		return err
	}

	var version srapi.AutoScalerVersion
	if src.Status.StarRocksCnStatus != nil {
		version = src.Status.StarRocksCnStatus.HorizontalScaler.Version
	}
	if err := cc.deleteAutoScaler(ctx, object.NewFromCluster(src), version); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete autoscaler failed")
		return err
	}

	return nil
}

func (cc *CnController) GetCnConfig(ctx context.Context,
	cnSpec *srapi.StarRocksCnSpec, namespace string) (map[string]interface{}, error) {
	return k8sutils.GetConfig(ctx, cc.k8sClient, cnSpec.ConfigMapInfo,
		cnSpec.ConfigMaps, pod.GetConfigDir(cnSpec), "cn.conf",
		namespace)
}

func (cc *CnController) getFeConfig(ctx context.Context,
	clusterNamespace string, clusterName string) (map[string]interface{}, error) {
	src, err := cc.getStarRocksCluster(ctx, clusterNamespace, clusterName)
	if err != nil {
		return nil, err
	}
	return fe.GetFEConfig(ctx, cc.k8sClient, src.Spec.StarRocksFeSpec, src.Namespace)
}

func (cc *CnController) mutating(_ *srapi.StarRocksCnSpec) error {
	return nil
}

func (cc *CnController) validating(cnSpec *srapi.StarRocksCnSpec) error {
	// validating the auto scaling policy
	policy := cnSpec.AutoScalingPolicy
	if policy != nil {
		minReplicas := int32(1) // default value
		if policy.MinReplicas != nil {
			minReplicas = *policy.MinReplicas
			if minReplicas < 1 {
				return fmt.Errorf("the min replicas must not be smaller than 1")
			}
		}

		maxReplicas := policy.MaxReplicas
		if maxReplicas < minReplicas {
			return fmt.Errorf("the MaxReplicas must not be smaller than MinReplicas")
		}
	}

	for i := range cnSpec.StorageVolumes {
		if err := cnSpec.StorageVolumes[i].Validate(); err != nil {
			return err
		}
	}

	if err := srapi.ValidUpdateStrategy(cnSpec.UpdateStrategy); err != nil {
		return err
	}

	return nil
}

// getStarRocksCluster get the StarRocksCluster object by namespace and name.
func (cc *CnController) getStarRocksCluster(ctx context.Context, namespace, name string) (*srapi.StarRocksCluster, error) {
	src := &srapi.StarRocksCluster{}
	err := cc.k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, src)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func (cc *CnController) generateAutoScalerName(srcName string, cnSpec srapi.SpecInterface) string {
	return load.Name(srcName, cnSpec) + "-autoscaler"
}

var (
	ErrReplicasNotEqual  = errors.New("replicas not equal")
	ErrHashValueNotEqual = errors.New("hash value not equal")
)

// SyncComputeNodesInFE sync the compute nodes in StarRocks.
func (cc *CnController) SyncComputeNodesInFE(ctx context.Context, object object.StarRocksObject,
	expectSTS *appsv1.StatefulSet, actualSTS *appsv1.StatefulSet,
	actualCNPods *corev1.PodList, db *sql.DB) error {
	logger := logr.FromContextOrDiscard(ctx)

	var expectReplicas int32
	if expectSTS.Spec.Replicas != nil {
		expectReplicas = *expectSTS.Spec.Replicas
	} else {
		expectReplicas = 1
	}

	var stsReplicas int32
	if actualSTS.Spec.Replicas != nil {
		stsReplicas = *actualSTS.Spec.Replicas
	} else {
		stsReplicas = 1
	}

	// compare the replicas between the expected value and the actual value in StatefulSet.
	if expectReplicas != stsReplicas {
		logger.Info("expect replicas is not equal to statefulset replicas", "expectReplicas", expectReplicas, "stsReplicas", stsReplicas)
		return ErrReplicasNotEqual
	}
	if expectHashValue, b := rutils.StatefulSetDeepEqual(expectSTS, actualSTS); !b {
		logger.Info("the actual Statefulset is not operator expected", "expectHashValue", expectHashValue)
		return ErrHashValueNotEqual
	}

	// compare the replicas between the expected value in Statefulset and the actual value in Pods.
	if len(actualCNPods.Items) != int(expectReplicas) {
		logger.Info("the expected number of pods is not equal to the actual number",
			"expectReplicas", expectReplicas, "actualReplicas", len(actualCNPods.Items))
		return ErrReplicasNotEqual
	}
	const controllerRevisionHashKey = "controller-revision-hash"
	for _, item := range actualCNPods.Items {
		if item.Labels[controllerRevisionHashKey] == "" ||
			item.Labels[controllerRevisionHashKey] != actualSTS.Status.UpdateRevision {
			logger.Info("there is old pod, continue to wait for the statefulset to be ready")
			return ErrHashValueNotEqual
		}
	}

	// now, all the new pods have been created, we can check the number of CN from FE
	executor, err := NewSQLExecutor(ctx, cc.k8sClient, object.Namespace, object.GetCNStatefulSetName())
	if err != nil {
		logger.Error(err, "new SQL executor failed")
		return err
	}
	result, err := executor.QueryShowComputeNodes(ctx, db)
	if err != nil {
		logger.Error(err, "query SHOW COMPUTE NODES failed", "sql", "SHOW COMPUTE NODES")
		return err
	}
	warehouseNameInFE := "default_warehouse"
	if object.IsWarehouseObject {
		warehouseNameInFE = object.GetWarehouseNameInFE()
	}
	computeNodes := result.ComputeNodesByWarehouse[warehouseNameInFE]
	if len(computeNodes) > int(expectReplicas) {
		for i := len(computeNodes) - 1; i >= int(expectReplicas); i-- {
			err = executor.ExecuteDropComputeNode(ctx, db, computeNodes[i])
			if err != nil {
				logger.Error(err, "drop compute node failed", "computeNode", computeNodes[i])
				return err
			}
			logger.Info("drop compute node success", "computeNode", computeNodes[i])
		}
	}

	return nil
}

func generateInternalService(object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec,
	externalService *corev1.Service, cnConfig map[string]interface{}, labels map[string]string) *corev1.Service {
	searchServiceName := service.SearchServiceName(object.SubResourcePrefixName, cnSpec)
	searchSvc := service.MakeSearchService(searchServiceName, externalService, []corev1.ServicePort{
		{
			Name:       "heartbeat",
			Port:       rutils.GetPort(cnConfig, rutils.HEARTBEAT_SERVICE_PORT),
			TargetPort: intstr.FromInt(int(rutils.GetPort(cnConfig, rutils.HEARTBEAT_SERVICE_PORT))),
		},
	}, labels)

	arrowFlightPort := rutils.GetPort(cnConfig, rutils.ARROW_FLIGHT_PORT)
	if arrowFlightPort != 0 {
		searchSvc.Spec.Ports = append(searchSvc.Spec.Ports, corev1.ServicePort{
			Name:       rutils.CnArrowFlightPortName,
			Port:       arrowFlightPort,
			TargetPort: intstr.FromInt(int(arrowFlightPort)),
		})
	}

	return searchSvc
}
