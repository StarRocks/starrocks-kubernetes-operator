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
	"k8s.io/client-go/tools/record"
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
	Client   client.Client
	Recorder record.EventRecorder
}

// New construct a FeController.
func New(k8sClient client.Client, recorderFor subcontrollers.GetEventRecorderForFunc) *FeController {
	controller := &FeController{
		Client: k8sClient,
	}
	controller.Recorder = recorderFor(controller.GetControllerName())
	return controller
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

	var err error
	defer func() {
		// we do not record an event if the error is nil, because this will cause too many events to be recorded.
		if err != nil {
			fc.Recorder.Event(src, corev1.EventTypeWarning, "SyncFeFailed", err.Error())
		}
	}()

	feSpec := src.Spec.StarRocksFeSpec
	if err = fc.Validating(feSpec); err != nil {
		return err
	}

	// get the fe configMap for resolve ports
	logger.V(log.DebugLevel).Info("get fe configMap to resolve ports", "ConfigMapInfo", feSpec.ConfigMapInfo)
	feConfig, err := GetFEConfig(ctx, fc.Client, feSpec, src.Namespace)
	if err != nil {
		logger.Error(err, "get fe config failed", "ConfigMapInfo", feSpec.ConfigMapInfo)
		return err
	}

	// generate new fe service.
	logger.V(log.DebugLevel).Info("build fe service", "StarRocksCluster", src)
	object := object.NewFromCluster(src)
	defaultLabels := load.Labels(src.Name, feSpec)
	svc := rutils.BuildExternalService(object, feSpec, feConfig, load.Selector(src.Name, feSpec), defaultLabels)
	searchServiceName := service.SearchServiceName(src.Name, feSpec)
	internalService := service.MakeSearchService(searchServiceName, &svc, []corev1.ServicePort{
		{
			Name:        "query-port",
			Port:        rutils.GetPort(feConfig, rutils.QUERY_PORT),
			TargetPort:  intstr.FromInt(int(rutils.GetPort(feConfig, rutils.QUERY_PORT))),
			AppProtocol: func() *string { mysql := "mysql"; return &mysql }(),
		},
	}, defaultLabels)

	// first deploy statefulset for compatible v1.5, apply statefulset for update pod.
	podTemplateSpec, err := fc.buildPodTemplate(src, feConfig)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}
	expectSts := statefulset.MakeStatefulset(object, feSpec, podTemplateSpec)

	drSpec := src.Spec.DisasterRecovery
	drStatus := src.Status.DisasterRecoveryStatus
	shouldEnterDRMode, queryPort := ShouldEnterDisasterRecoveryMode(drSpec, drStatus, feConfig)
	if shouldEnterDRMode {
		logger.Info("should enter disaster recovery mode")
		if drStatus == nil {
			drStatus = srapi.NewDisasterRecoveryStatus(drSpec.Generation)
			src.Status.DisasterRecoveryStatus = drStatus
		}
		if err = EnterDisasterRecoveryMode(ctx, fc.Client, src, &expectSts, queryPort); err != nil {
			logger.Error(err, "enter disaster recovery mode failed")
			return err
		}
		logger.Info("deploy statefulset", "statefulset", expectSts)
	}

	if err = k8sutils.ApplyStatefulSetWithPVCExpansion(ctx, fc.Client, &expectSts, shouldEnterDRMode,
		func(new *appv1.StatefulSet, actual *appv1.StatefulSet) bool {
			return rutils.StatefulSetDeepEqual(new, actual, false)
		}, feSpec.GetStorageVolumes()); err != nil {
		logger.Error(err, "deploy statefulset with PVC expansion failed")
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.Client, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `fe-domain-search` for internal communicating.
		internalService.Name = expectSts.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		logger.Error(err, "deploy search service failed", "internalService", internalService)
		fc.Recorder.Event(src, corev1.EventTypeWarning, "DeployFeFailed", err.Error())
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

func (fc *FeController) Validating(feSpec *srapi.StarRocksFeSpec) error {
	for i := range feSpec.StorageVolumes {
		if err := feSpec.StorageVolumes[i].Validate(); err != nil {
			return err
		}
	}
	if err := srapi.ValidUpdateStrategy(feSpec.UpdateStrategy); err != nil {
		return err
	}
	return nil
}

// CheckFEReady check the fe cluster is ok.
// Note:
// When user upgrade the cluster, and the statefulset controller has not begun to update the statefulset,
// CheckFEReady will use the old status to check whether FE is ready.
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
