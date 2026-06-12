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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
		logger.Info("fe observer is disabled, clear observer resources")
		if err := fc.clearObserverResources(ctx, src); err != nil {
			logger.Error(err, "clear fe observer resources failed")
			return err
		}
		return nil
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

	podTemplateSpec, err := fc.buildPodTemplate(src, feConfig)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}
	expectSts := statefulset.MakeStatefulset(object, observerSpec, podTemplateSpec)
	expectSts.Spec.ServiceName = feSearchServiceName(src.Name)
	err = k8sutils.ApplyStatefulSet(ctx, fc.Client, &expectSts, false, rutils.StatefulSetDeepEqual)
	if err != nil {
		logger.Error(err, "fe observer statefulset failed", "StarRocksCluster", src)
		return err
	}

	if err = fc.deleteLegacyObserverServices(ctx, src); err != nil {
		logger.Error(err, "delete legacy fe observer services failed")
		return err
	}

	return nil
}

// UpdateClusterStatus update the all resource status about fe observer.
func (fc *FeObserverController) UpdateClusterStatus(_ context.Context, src *srapi.StarRocksCluster) error {
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
	fs.ServiceName = feExternalServiceName(src.Name)
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

	return fc.clearObserverResources(ctx, src)
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
	return nil
}

func (fc *FeObserverController) clearObserverResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	statefulSetName := load.Name(src.Name, (*srapi.StarRocksFeObserverSpec)(nil))
	if err := k8sutils.DeleteStatefulset(ctx, fc.Client, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return fc.deleteLegacyObserverServices(ctx, src)
}

func (fc *FeObserverController) deleteLegacyObserverServices(ctx context.Context, src *srapi.StarRocksCluster) error {
	searchServiceName := src.Name + "-" + srapi.DEFAULT_FE_OBSERVER + "-search"
	if err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, searchServiceName); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	externalServiceName := src.Name + "-" + srapi.DEFAULT_FE_OBSERVER + "-service"
	if err := k8sutils.DeleteService(ctx, fc.Client, src.Namespace, externalServiceName); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func feExternalServiceName(clusterName string) string {
	return clusterName + "-" + srapi.DEFAULT_FE + "-service"
}

func feSearchServiceName(clusterName string) string {
	return clusterName + "-" + srapi.DEFAULT_FE + "-search"
}
