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

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

const minFeObserverImageVersion = "4.1.0"

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

// SyncCluster starRocksCluster spec to fe observer statefulset and service.
func (fc *FeObserverController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(fc.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)
	feSpec := src.Spec.StarRocksFeSpec
	observerSpec := feSpec.ToObserverSpec()
	if observerSpec == nil {
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

	if err = fc.Validating(feSpec); err != nil {
		return err
	}

	if !fe.CheckFEReady(ctx, fc.Client, src.Namespace, src.Name) {
		logger.Info("FE is not ready, stop sync fe observer")
		return nil
	}
	// get the fe configMap for resolve ports
	logger.V(log.DebugLevel).Info("get fe configMap to resolve ports", "ConfigMapInfo", feSpec.ConfigMapInfo)
	feConfig, err := fe.GetFEConfig(ctx, fc.Client, feSpec, src.Namespace)
	if err != nil {
		logger.Error(err, "get fe config failed", "ConfigMapInfo", feSpec.ConfigMapInfo)
		return err
	}

	// generate new fe observer statefulset.
	logger.V(log.DebugLevel).Info("build fe observer statefulset", "StarRocksCluster", src)
	object := object.NewFromCluster(src)
	defaultLabels := load.Labels(src.Name, observerSpec)
	svc := rutils.BuildExternalService(object, observerSpec, feConfig, load.Selector(src.Name, observerSpec), defaultLabels)
	searchServiceName := service.SearchServiceName(src.Name, observerSpec)
	internalService := service.MakeSearchService(searchServiceName, &svc, []corev1.ServicePort{
		{
			Name:        "query-port",
			Port:        rutils.GetPort(feConfig, rutils.QUERY_PORT),
			TargetPort:  intstr.FromInt(int(rutils.GetPort(feConfig, rutils.QUERY_PORT))),
			AppProtocol: func() *string { mysql := "mysql"; return &mysql }(),
		},
	}, defaultLabels)
	podTemplateSpec, err := fc.buildPodTemplate(src, feConfig)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}
	expectSts := statefulset.MakeStatefulset(object, observerSpec, podTemplateSpec)
	err = k8sutils.ApplyStatefulSet(ctx, fc.Client, &expectSts, false, rutils.StatefulSetDeepEqual)
	if err != nil {
		logger.Error(err, "fe observer statefulset failed", "StarRocksCluster", src)
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
func (fc *FeObserverController) UpdateClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	feSpec := src.Spec.StarRocksFeSpec
	observerSpec := feSpec.ToObserverSpec()
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
	fs.ServiceName = service.ExternalServiceName(src.Name, feSpec)
	statefulSetName := load.Name(src.Name, observerSpec)
	fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{statefulSetName})

	if err := subcontrollers.UpdateStatus(&fs.StarRocksComponentStatus, fc.Client,
		src.Namespace, statefulSetName, pod.Labels(src.Name, observerSpec), subcontrollers.StatefulSetLoadType); err != nil {
		return err
	}

	var st appsv1.StatefulSet
	if err := fc.Client.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); err != nil {
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

	observerSpec := src.Spec.StarRocksFeSpec.ToObserverSpec()
	statefulSetName := load.Name(src.Name, observerSpec)
	if err := k8sutils.DeleteStatefulset(ctx, fc.Client, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
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

func (fc *FeObserverController) Validating(feSpec *srapi.StarRocksFeSpec) error {
	for i := range feSpec.StorageVolumes {
		if err := feSpec.StorageVolumes[i].Validate(); err != nil {
			return err
		}
	}
	if err := srapi.ValidUpdateStrategy(feSpec.UpdateStrategy); err != nil {
		return err
	}
	if feSpec.ToObserverSpec() != nil {
		if err := validateObserverImageVersion(feSpec.Image); err != nil {
			return err
		}
	}
	return nil
}

func validateObserverImageVersion(image string) error {
	imageVersion := pod.GetImageVersion(image)
	_, err := pod.IsLowerThanAny(imageVersion, []string{"4.1.0"})
	if err != nil {
		return fmt.Errorf("fe observer requires StarRocks FE image version >= %s "+
			"Current image version %s does not enable FE observer", minFeObserverImageVersion, imageVersion)
	}
	return nil
}
