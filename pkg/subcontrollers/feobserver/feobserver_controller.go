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

package feobserver

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/deployment"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
)

type FeObserverController struct {
	Client   client.Client
	Recorder record.EventRecorder
}

// New construct a FeObserverController.
func New(k8sClient client.Client, recorderFor subcontrollers.GetEventRecorderForFunc) *FeObserverController {
	controller := &FeObserverController{
		Client: k8sClient,
	}
	controller.Recorder = recorderFor(controller.GetControllerName())
	return controller
}

func (fc *FeObserverController) GetControllerName() string {
	return "feObserverController"
}

// SyncCluster starRocksCluster spec to fe observer deployment and service.
func (fc *FeObserverController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(fc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)
	if src.Spec.StarRocksFeObserverSpec == nil {
		logger.Info("src.Spec.StarRocksFeObserverSpec == nil, skip sync fe observer")
		return nil
	}
	if src.Spec.StarRocksFeSpec == nil {
		return fmt.Errorf("starRocksFeSpec is required before deploying fe observer")
	}

	var err error
	defer func() {
		// we do not record an event if the error is nil, because this will cause too many events to be recorded.
		if err != nil {
			fc.Recorder.Event(src, corev1.EventTypeWarning, "SyncFeObserverFailed", err.Error())
		}
	}()

	observerSpec := src.Spec.StarRocksFeObserverSpec
	if err = fc.Validating(observerSpec); err != nil {
		return err
	}

	// get the fe observer configMap for resolve ports
	logger.V(log.DebugLevel).Info("get fe observer configMap to resolve ports", "ConfigMapInfo", observerSpec.ConfigMapInfo)
	observerConfig, err := GetFEObserverConfig(ctx, fc.Client, observerSpec, src.Namespace)
	if err != nil {
		logger.Error(err, "get fe observer config failed", "ConfigMapInfo", observerSpec.ConfigMapInfo)
		return err
	}

	// generate new fe observer service.
	logger.V(log.DebugLevel).Info("build fe observer service", "StarRocksCluster", src)
	object := object.NewFromCluster(src)
	defaultLabels := load.Labels(src.Name, observerSpec)
	svc := rutils.BuildExternalService(object, observerSpec, observerConfig, load.Selector(src.Name, observerSpec), defaultLabels)
	searchServiceName := service.SearchServiceName(src.Name, observerSpec)
	internalService := service.MakeSearchService(searchServiceName, &svc, []corev1.ServicePort{
		{
			Name:        "query-port",
			Port:        rutils.GetPort(observerConfig, rutils.QUERY_PORT),
			TargetPort:  intstr.FromInt(int(rutils.GetPort(observerConfig, rutils.QUERY_PORT))),
			AppProtocol: func() *string { mysql := "mysql"; return &mysql }(),
		},
	}, defaultLabels)

	podTemplateSpec, err := fc.buildPodTemplate(src, observerConfig)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}
	expectDeployment := deployment.MakeDeployment(src, observerSpec, *podTemplateSpec)
	err = k8sutils.ApplyDeployment(ctx, fc.Client, expectDeployment)
	if err != nil {
		logger.Error(err, "fe observer deployment failed", "StarRocksCluster", src)
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.Client, internalService, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "deploy search service failed", "internalService", internalService)
		fc.Recorder.Event(src, corev1.EventTypeWarning, "DeployFeObserverFailed", err.Error())
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.Client, &svc, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "deploy external service failed", "externalService", svc)
		return err
	}

	return nil
}

// UpdateClusterStatus update the all resource status about fe observer.
func (fc *FeObserverController) UpdateClusterStatus(_ context.Context, src *srapi.StarRocksCluster) error {
	observerSpec := src.Spec.StarRocksFeObserverSpec
	if observerSpec == nil {
		src.Status.StarRocksFeObserverStatus = nil
		return nil
	}

	fs := &srapi.StarRocksFeObserverStatus{
		StarRocksComponentStatus: srapi.StarRocksComponentStatus{
			Phase: srapi.ComponentReconciling,
		},
	}

	if src.Status.StarRocksFeObserverStatus != nil {
		fs = src.Status.StarRocksFeObserverStatus.DeepCopy()
	}

	src.Status.StarRocksFeObserverStatus = fs
	fs.ServiceName = service.ExternalServiceName(src.Name, observerSpec)
	deploymentName := load.Name(src.Name, observerSpec)
	fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{deploymentName})

	if err := subcontrollers.UpdateStatus(&fs.StarRocksComponentStatus, fc.Client,
		src.Namespace, deploymentName, pod.Labels(src.Name, observerSpec), subcontrollers.DeploymentLoadType); err != nil {
		return err
	}

	var deploy appsv1.Deployment
	if err := fc.Client.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: deploymentName}, &deploy); err != nil {
		return err
	}

	return nil
}

// ClearCluster clears resource about fe observer.
func (fc *FeObserverController) ClearCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(fc.GetControllerName()).WithValues(log.ActionKey, log.ActionCluster)
	ctx = logr.NewContext(ctx, logger)

	if src.Status.StarRocksFeObserverStatus == nil {
		return nil
	}

	if src.DeletionTimestamp.IsZero() {
		return nil
	}

	observerSpec := src.Spec.StarRocksFeObserverSpec
	deploymentName := load.Name(src.Name, observerSpec)
	if err := k8sutils.DeleteDeployment(ctx, fc.Client, src.Namespace, deploymentName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete deployment failed", "deploymentName", deploymentName)
		return err
	}

	searchServiceName := service.SearchServiceName(src.Name, observerSpec)
	if err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, searchServiceName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete search service failed", "searchServiceName", searchServiceName)
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, observerSpec)
	err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete external service failed", "externalServiceName", externalServiceName)
		return err
	}

	return nil
}

func (fc *FeObserverController) Validating(observerSpec *srapi.StarRocksFeObserverSpec) error {
	for i := range observerSpec.StorageVolumes {
		if err := observerSpec.StorageVolumes[i].Validate(); err != nil {
			return err
		}
	}
	if err := srapi.ValidUpdateStrategy(observerSpec.UpdateStrategy); err != nil {
		return err
	}
	return nil
}
