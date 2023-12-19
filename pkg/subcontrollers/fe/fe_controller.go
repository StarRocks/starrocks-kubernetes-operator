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

package fe

import (
	"context"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/log"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
)

type FeController struct {
	Client client.Client
}

// New construct a FeController.
func New(k8sClient client.Client) *FeController {
	return &FeController{
		Client: k8sClient,
	}
}

func (fc *FeController) GetControllerName() string {
	return "feController"
}

// SyncCluster starRocksCluster spec to fe statefulset and service.
func (fc *FeController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(fc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)
	if src.Spec.StarRocksFeSpec == nil {
		logger.Info("src.Spec.StarRocksFeSpec == nil, skip sync fe")
		return nil
	}

	// get the fe configMap for resolve ports
	feSpec := src.Spec.StarRocksFeSpec
	logger.V(log.DebugLevel).Info("get fe configMap to resolve ports", "ConfigMapInfo", feSpec.ConfigMapInfo)
	config, err := GetFeConfig(ctx, fc.Client, &feSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		logger.Error(err, "get fe config failed", "ConfigMapInfo", feSpec.ConfigMapInfo)
		return err
	}

	// generate new fe service.
	logger.V(log.DebugLevel).Info("build fe service", "StarRocksCluster", src)
	object := object.NewFromCluster(src)
	svc := rutils.BuildExternalService(object, feSpec, config, load.Selector(src.Name, feSpec), load.Labels(src.Name, feSpec))
	searchServiceName := service.SearchServiceName(src.Name, feSpec)
	internalService := service.MakeSearchService(searchServiceName, &svc, []corev1.ServicePort{
		{
			Name:        "query-port",
			Port:        rutils.GetPort(config, rutils.QUERY_PORT),
			TargetPort:  intstr.FromInt(int(rutils.GetPort(config, rutils.QUERY_PORT))),
			AppProtocol: func() *string { mysql := "mysql"; return &mysql }(),
		},
	})

	// first deploy statefulset for compatible v1.5, apply statefulset for update pod.
	podTemplateSpec := fc.buildPodTemplate(src, config)
	st := statefulset.MakeStatefulset(object, feSpec, podTemplateSpec)
	if err = k8sutils.ApplyStatefulSet(ctx, fc.Client, &st, func(new *appv1.StatefulSet, actual *appv1.StatefulSet) bool {
		return rutils.StatefulSetDeepEqual(new, actual, false)
	}); err != nil {
		logger.Error(err, "deploy statefulset failed")
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.Client, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `fe-domain-search` for internal communicating.
		internalService.Name = st.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		logger.Error(err, "deploy search service failed", "internalService", internalService)
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.Client, &svc, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "deploy external service failed", "externalService", svc)
		return err
	}

	return nil
}

// UpdateClusterStatus update the all resource status about fe.
func (fc *FeController) UpdateClusterStatus(_ context.Context, src *srapi.StarRocksCluster) error {
	// if spec is not exist, status is empty. but before clear status we must clear all resource about be used by ClearResources.
	feSpec := src.Spec.StarRocksFeSpec
	if feSpec == nil {
		src.Status.StarRocksFeStatus = nil
		return nil
	}

	fs := &srapi.StarRocksFeStatus{
		StarRocksComponentStatus: srapi.StarRocksComponentStatus{
			Phase: srapi.ComponentReconciling,
		},
	}

	if src.Status.StarRocksFeStatus != nil {
		fs = src.Status.StarRocksFeStatus.DeepCopy()
	}

	src.Status.StarRocksFeStatus = fs
	fs.ServiceName = service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)
	statefulSetName := load.Name(src.Name, src.Spec.StarRocksFeSpec)
	fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{statefulSetName})

	if err := subcontrollers.UpdateStatus(&fs.StarRocksComponentStatus, fc.Client,
		src.Namespace, load.Name(src.Name, feSpec), pod.Labels(src.Name, feSpec), subcontrollers.StatefulSetLoadType); err != nil {
		return err
	}

	var st appv1.StatefulSet
	if err := fc.Client.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); err != nil {
		return err
	}

	return nil
}

// GetFeConfig get the fe start config.
func GetFeConfig(ctx context.Context,
	k8sClient client.Client, configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	if configMapInfo.ConfigMapName == "" || configMapInfo.ResolveKey == "" {
		return make(map[string]interface{}), nil
	}
	configMap, err := k8sutils.GetConfigMap(ctx, k8sClient, namespace, configMapInfo.ConfigMapName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

// ClearResources clear resource about fe.
func (fc *FeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(fc.GetControllerName()).WithValues(log.ActionKey, log.ActionClearResources)
	ctx = logr.NewContext(ctx, logger)

	// if the starrocks is not have fe.
	if src.Status.StarRocksFeStatus == nil {
		return nil
	}

	if src.DeletionTimestamp.IsZero() {
		return nil
	}

	statefulSetName := load.Name(src.Name, src.Spec.StarRocksFeSpec)
	if err := k8sutils.DeleteStatefulset(ctx, fc.Client, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete statefulset failed", "statefulsetName", statefulSetName)
		return err
	}

	feSpec := src.Spec.StarRocksFeSpec
	searchServiceName := service.SearchServiceName(src.Name, feSpec)
	if err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, searchServiceName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete search service failed", "searchServiceName", searchServiceName)
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, feSpec)
	err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete external service failed", "externalServiceName", externalServiceName)
		return err
	}

	return nil
}

// CheckFEReady check the fe cluster is ok.
func CheckFEReady(ctx context.Context, k8sClient client.Client, clusterNamespace, clusterName string) bool {
	logger := logr.FromContextOrDiscard(ctx)
	endpoints := corev1.Endpoints{}
	serviceName := service.ExternalServiceName(clusterName, (*srapi.StarRocksFeSpec)(nil))
	// 1. wait for FE ready.
	if err := k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: clusterNamespace,
			Name:      serviceName,
		},
		&endpoints); err != nil {
		logger.Error(err, "get fe service failed", "serviceName", serviceName)
		return false
	}

	for _, sub := range endpoints.Subsets {
		if len(sub.Addresses) > 0 {
			return true
		}
	}

	return false
}
