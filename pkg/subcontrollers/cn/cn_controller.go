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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	addEnvForWarehouse bool
}

func New(k8sClient client.Client) *CnController {
	return &CnController{
		k8sClient: k8sClient,
	}
}

func (cc *CnController) GetControllerName() string {
	return "cnController"
}

var SpecMissingError = errors.New("spec.template or spec.starRocksCluster is missing")
var StarRocksClusterMissingError = errors.New("custom resource StarRocksCluster is missing")
var FeNotReadyError = errors.New("component fe is not ready")
var StarRocksClusterRunModeError = errors.New("StarRocks Cluster should run in shared_data mode")
var GetFeFeatureInfoError = errors.New("failed to invoke FE /api/v2/feature or FE does not support multi-warehouse feature")

func (cc *CnController) SyncWarehouse(ctx context.Context, warehouse *srapi.StarRocksWarehouse) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncWarehouse)
	ctx = logr.NewContext(ctx, logger)

	template := warehouse.Spec.Template
	if warehouse.Spec.StarRocksCluster == "" || template == nil {
		return SpecMissingError
	}

	logger.Info("get StarRocksCluster CR from kubernetes")
	_, err := cc.getStarRocksCluster(ctx, warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return StarRocksClusterMissingError
		}
		return err
	}

	logger.Info("get fe config to make sure StarRocks run in shared_data mode")
	feconfig, err := cc.getFeConfig(ctx, warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		return err
	}
	if val := feconfig["run_mode"]; val == nil || !strings.Contains(val.(string), "shared_data") {
		return StarRocksClusterRunModeError
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, warehouse.Namespace, warehouse.Spec.StarRocksCluster) {
		return FeNotReadyError
	}

	return cc.SyncCnSpec(ctx, object.NewFromWarehouse(warehouse), template.ToCnSpec())
}

func (cc *CnController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)

	if src.Spec.StarRocksCnSpec == nil {
		if err := cc.ClearResources(ctx, src); err != nil {
			logger.Error(err, "clear resource failed")
		}
		return nil
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, src.Namespace, src.Name) {
		return nil
	}

	return cc.SyncCnSpec(ctx, object.NewFromCluster(src), src.Spec.StarRocksCnSpec)
}

func (cc *CnController) SyncCnSpec(ctx context.Context, object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec) error {
	logger := logr.FromContextOrDiscard(ctx)

	if err := cc.mutating(cnSpec); err != nil {
		return err
	}

	if err := cc.validating(cnSpec); err != nil {
		return err
	}

	logger.V(log.DebugLevel).Info("get cn config to resolve ports", "ConfigMapInfo", cnSpec.ConfigMapInfo)
	config, err := cc.GetConfig(ctx, &cnSpec.ConfigMapInfo, object.Namespace)
	if err != nil {
		return err
	}
	logger.V(log.DebugLevel).Info("get fe config to resolve ports", "ConfigMapInfo", cnSpec.ConfigMapInfo)
	feconfig, err := cc.getFeConfig(ctx, object.Namespace, object.ClusterName)
	if err != nil {
		return err
	}
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	config[rutils.HTTP_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.HTTP_PORT)), 10)

	// build and deploy statefulset
	podTemplateSpec, err := cc.buildPodTemplate(ctx, object, cnSpec, config)
	if err != nil {
		return err
	}
	sts := statefulset.MakeStatefulset(object, cnSpec, *podTemplateSpec)
	if err = k8sutils.ApplyStatefulSet(ctx, cc.k8sClient, &sts, func(st1 *appv1.StatefulSet, st2 *appv1.StatefulSet) bool {
		return rutils.StatefulSetDeepEqual(st1, st2, true)
	}); err != nil {
		return err
	}

	// build and deploy service
	externalsvc := rutils.BuildExternalService(object, cnSpec, config,
		load.Selector(object.AliasName, cnSpec), load.Labels(object.AliasName, cnSpec))
	searchServiceName := service.SearchServiceName(object.AliasName, cnSpec)
	internalService := service.MakeSearchService(searchServiceName, &externalsvc, []corev1.ServicePort{
		{
			Name:       "heartbeat",
			Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
			TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
		},
	})

	if err := k8sutils.ApplyService(ctx, cc.k8sClient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "sync CN external service failed")
		return err
	}

	if err := k8sutils.ApplyService(ctx, cc.k8sClient, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `cn-domain-search` for internal communicating.
		internalService.Name = sts.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		logger.Error(err, "sync CN search service failed")
		return err
	}

	// build and deploy HPA
	if cnSpec.AutoScalingPolicy != nil {
		return cc.deployAutoScaler(ctx, object, cnSpec, *cnSpec.AutoScalingPolicy, &sts)
	} else {
		// If the HPA policy is nil, delete the HPA resource.
		return cc.deleteAutoScaler(ctx, object)
	}
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
	var st appv1.StatefulSet
	logger := logr.FromContextOrDiscard(ctx)

	// todo(yandongxiao): delete it
	statefulSetName := load.Name(object.AliasName, cnSpec)
	namespacedName := types.NamespacedName{Namespace: object.Namespace, Name: statefulSetName}
	if err := cc.k8sClient.Get(ctx, namespacedName, &st); apierrors.IsNotFound(err) {
		logger.Error(err, "get statefulset failed")
		return nil
	}

	if cnSpec.AutoScalingPolicy != nil {
		cnStatus.HorizontalScaler.Name = cc.generateAutoScalerName(object.AliasName, cnSpec)
		cnStatus.HorizontalScaler.Version = cnSpec.AutoScalingPolicy.Version.Complete(k8sutils.KUBE_MAJOR_VERSION,
			k8sutils.KUBE_MINOR_VERSION)
	} else {
		cnStatus.HorizontalScaler = srapi.HorizontalScaler{}
	}

	cnStatus.ServiceName = service.ExternalServiceName(object.AliasName, cnSpec)
	cnStatus.ResourceNames = rutils.MergeSlices(cnStatus.ResourceNames, []string{statefulSetName})

	if err := subc.UpdateStatus(&cnStatus.StarRocksComponentStatus, cc.k8sClient,
		object.Namespace, load.Name(object.AliasName, cnSpec), pod.Labels(object.AliasName, cnSpec), subc.StatefulSetLoadType); err != nil {
		return err
	}

	return nil
}

// ClearWarehouse clear the warehouse resource. It is different from ClearResources, which need to clear the
// CN related resources of StarRocksCluster. ClearWarehouse only has CN related resources, when the warehouse CR
// is deleted, sub resources of CN will be deleted by k8s.
func (cc *CnController) ClearWarehouse(ctx context.Context, namespace string, name string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(cc.GetControllerName()).WithValues(log.ActionKey, log.ActionClearWarehouse)
	ctx = logr.NewContext(ctx, logger)

	executor, err := NewSQLExecutor(ctx, cc.k8sClient, namespace, object.GetAliasName(name))
	if err != nil {
		logger.Error(err, "new SQL executor failed")
		return err
	}

	warehouseName := strings.ReplaceAll(name, "-", "_")
	err = executor.Execute(ctx, nil, fmt.Sprintf("DROP WAREHOUSE %s", warehouseName))
	if err != nil {
		logger.Error(err, "drop warehouse failed", "warehouse", warehouseName)
		// we do not return error here, because we want to delete the statefulset anyway.
	}

	// Remove the finalizer from cn statefulset
	var sts appv1.StatefulSet
	if err = cc.k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      load.Name(object.GetAliasName(name), (*srapi.StarRocksCnSpec)(nil)),
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
func (cc *CnController) deployAutoScaler(ctx context.Context, object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec,
	policy srapi.AutoScalingPolicy, target *appv1.StatefulSet) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update k8s hpa resource")

	labels := rutils.Labels{}
	labels.AddLabel(target.Labels)
	labels.Add(srapi.ComponentLabelKey, "autoscaler")
	autoscalerParams := &rutils.PodAutoscalerParams{
		Namespace:       target.Namespace,
		Name:            cc.generateAutoScalerName(object.AliasName, cnSpec),
		Labels:          labels,
		AutoscalerType:  cnSpec.AutoScalingPolicy.Version, // cnSpec.AutoScalingPolicy can not be nil
		TargetName:      target.Name,
		OwnerReferences: target.OwnerReferences,
		ScalerPolicy:    &policy,
	}

	expectHPA := rutils.BuildHorizontalPodAutoscaler(autoscalerParams, "")
	expectHPA.SetAnnotations(make(map[string]string))

	actualHPA := autoscalerParams.AutoscalerType.CreateEmptyHPA(k8sutils.KUBE_MAJOR_VERSION, k8sutils.KUBE_MINOR_VERSION)
	if err := cc.k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: autoscalerParams.Namespace,
			Name:      autoscalerParams.Name,
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
func (cc *CnController) deleteAutoScaler(ctx context.Context, object object.StarRocksObject) error {
	logger := logr.FromContextOrDiscard(ctx)

	autoScalerName := cc.generateAutoScalerName(object.AliasName, (*srapi.StarRocksCnSpec)(nil))
	if err := k8sutils.DeleteAutoscaler(ctx, cc.k8sClient, object.Namespace, autoScalerName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete autoscaler failed")
		return err
	}
	return nil
}

// ClearResources clear the deployed resource about cn. statefulset, services, hpa.
func (cc *CnController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
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

	if err := cc.deleteAutoScaler(ctx, object.NewFromCluster(src)); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete autoscaler failed")
		return err
	}

	return nil
}

func (cc *CnController) GetConfig(ctx context.Context,
	configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	configMap, err := k8sutils.GetConfigMap(ctx, cc.k8sClient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) getFeConfig(ctx context.Context,
	clusterNamespace string, clusterName string) (map[string]interface{}, error) {
	src, err := cc.getStarRocksCluster(ctx, clusterNamespace, clusterName)
	if err != nil {
		return nil, err
	}
	feconfigMapInfo := &src.Spec.StarRocksFeSpec.ConfigMapInfo

	feconfigMap, err := k8sutils.GetConfigMap(ctx, cc.k8sClient, clusterNamespace, feconfigMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}
	res, err := rutils.ResolveConfigMap(feconfigMap, feconfigMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) mutating(cnSpec *srapi.StarRocksCnSpec) error {
	// Mutating because of the autoscaling policy.
	// When the HPA policy with a fixed replica count is set: every time the starrockscluster CR is
	// applied, the replica count of the StatefulSet object in K8S will be reset to the value
	// specified by the 'Replicas' field, erasing the value previously set by HPA.
	policy := cnSpec.AutoScalingPolicy
	if policy != nil {
		cnSpec.Replicas = nil
	}
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
