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

package be

import (
	"context"
	"strconv"

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
	subc "github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

type BeController struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func New(k8sClient client.Client, recorderFor subc.GetEventRecorderForFunc) *BeController {
	controller := &BeController{
		Client: k8sClient,
	}
	controller.Recorder = recorderFor(controller.GetControllerName())
	return controller
}

func (be *BeController) GetControllerName() string {
	return "beController"
}

func (be *BeController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(be.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)

	if src.Spec.StarRocksBeSpec == nil {
		if err := be.ClearResources(ctx, src); err != nil {
			logger.Error(err, "clear resource failed")
			return err
		}

		return nil
	}

	var err error
	defer func() {
		// we do not record an event if the error is nil, because this will cause too many events to be recorded.
		if err != nil {
			be.Recorder.Event(src, corev1.EventTypeWarning, "SyncBeFailed", err.Error())
		}
	}()

	beSpec := src.Spec.StarRocksBeSpec
	if err = be.validating(beSpec); err != nil {
		return err
	}

	if !fe.CheckFEReady(ctx, be.Client, src.Namespace, src.Name) {
		return nil
	}

	logger.V(log.DebugLevel).Info("get be/fe config to resolve ports", "ConfigMapInfo", beSpec.ConfigMapInfo)
	config, err := be.GetBeConfig(ctx, beSpec, src.Namespace)
	if err != nil {
		logger.Error(err, "get be config failed", "ConfigMapInfo", beSpec.ConfigMapInfo)
		return err
	}

	feconfig, _ := fe.GetFEConfig(ctx, be.Client, src.Spec.StarRocksFeSpec, src.Namespace)
	// add query port from fe config.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	// generate new be external service.
	externalsvc := rutils.BuildExternalService(object.NewFromCluster(src),
		beSpec, config, load.Selector(src.Name, beSpec), load.Labels(src.Name, beSpec))
	// generate internal fe service, update the status of cn on src.
	internalService := be.generateInternalService(ctx, src, &externalsvc, config)

	// create be statefulset
	podTemplateSpec, err := be.buildPodTemplate(src, config)
	if err != nil {
		logger.Error(err, "build pod template failed")
		return err
	}
	st := statefulset.MakeStatefulset(object.NewFromCluster(src), beSpec, podTemplateSpec)

	// update the statefulset if feSpec be updated.
	if err = k8sutils.ApplyStatefulSet(ctx, be.Client, &st, true, func(new *appv1.StatefulSet, actual *appv1.StatefulSet) bool {
		return rutils.StatefulSetDeepEqual(new, actual, false)
	}); err != nil {
		logger.Error(err, "apply statefulset failed")
		return err
	}

	if err = k8sutils.ApplyService(ctx, be.Client, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `cn-domain-search` for internal communicating.
		internalService.Name = st.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		logger.Error(err, "apply internal service failed", "internalService", internalService)
		return err
	}

	if err = k8sutils.ApplyService(ctx, be.Client, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		logger.Error(err, "apply external service failed", "externalService", externalsvc)
		return err
	}

	return err
}

// UpdateClusterStatus update the all resource status about be.
func (be *BeController) UpdateClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(be.GetControllerName()).WithValues(log.ActionKey, log.ActionUpdateClusterStatus)
	ctx = logr.NewContext(ctx, logger)

	// if spec is not exist, status is empty. but before clear status we must clear all resource about be used by ClearResources.
	beSpec := src.Spec.StarRocksBeSpec
	if beSpec == nil {
		src.Status.StarRocksBeStatus = nil
		return nil
	}

	bs := &srapi.StarRocksBeStatus{
		StarRocksComponentStatus: srapi.StarRocksComponentStatus{
			Phase: srapi.ComponentReconciling,
		},
	}
	if src.Status.StarRocksBeStatus != nil {
		bs = src.Status.StarRocksBeStatus.DeepCopy()
	}
	src.Status.StarRocksBeStatus = bs

	var st appv1.StatefulSet
	statefulSetName := load.Name(src.Name, beSpec)
	if err := be.Client.Get(ctx,
		types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); apierrors.IsNotFound(err) {
		logger.Info("statefulset is not found")
		return nil
	} else if err != nil {
		return err
	}

	bs.ServiceName = service.ExternalServiceName(src.Name, beSpec)
	bs.ResourceNames = rutils.MergeSlices(bs.ResourceNames, []string{statefulSetName})

	if err := subc.UpdateStatus(&bs.StarRocksComponentStatus, be.Client,
		src.Namespace, load.Name(src.Name, beSpec), pod.Labels(src.Name, beSpec), subc.StatefulSetLoadType); err != nil {
		return err
	}

	return nil
}

func (be *BeController) generateInternalService(ctx context.Context,
	src *srapi.StarRocksCluster, externalService *corev1.Service, config map[string]interface{}) *corev1.Service {
	logger := logr.FromContextOrDiscard(ctx)
	spec := src.Spec.StarRocksBeSpec
	searchServiceName := service.SearchServiceName(src.Name, spec)
	searchSvc := service.MakeSearchService(searchServiceName, externalService, []corev1.ServicePort{
		{
			Name:       "heartbeat",
			Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
			TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
		},
	})

	// for compatible version < v1.5
	var esearchSvc corev1.Service
	if err := be.Client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: "be-domain-search"}, &esearchSvc); err == nil {
		if rutils.HaveEqualOwnerReference(&esearchSvc, searchSvc) {
			searchSvc.Name = "be-domain-search"
		}
	} else if !apierrors.IsNotFound(err) {
		logger.Error(err, "get internal service object failed")
	}

	return searchSvc
}

// GetBeConfig get the config of BE from configmap.
func (be *BeController) GetBeConfig(ctx context.Context,
	beSpec *srapi.StarRocksBeSpec, namespace string) (map[string]interface{}, error) {
	return k8sutils.GetConfig(ctx, be.Client, beSpec.ConfigMapInfo,
		beSpec.ConfigMaps, pod.GetConfigDir(beSpec), "be.conf", namespace)
}

func (be *BeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(be.GetControllerName()).WithValues(log.ActionKey, log.ActionClearResources)
	ctx = logr.NewContext(ctx, logger)

	beSpec := src.Spec.StarRocksBeSpec
	if beSpec != nil {
		return nil
	}

	statefulSetName := load.Name(src.Name, beSpec)
	if err := k8sutils.DeleteStatefulset(ctx, be.Client, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete statefulset failed", "statefulset", statefulSetName)
		return err
	}

	searchServiceName := service.SearchServiceName(src.Name, beSpec)
	err := k8sutils.DeleteService(ctx, be.Client, src.Namespace, searchServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete search service failed", "searchService", searchServiceName)
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, beSpec)
	err = k8sutils.DeleteService(ctx, be.Client, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "delete external service failed", "externalService", externalServiceName)
		return err
	}

	return nil
}

func (be *BeController) validating(beSpec *srapi.StarRocksBeSpec) error {
	for i := range beSpec.StorageVolumes {
		if err := beSpec.StorageVolumes[i].Validate(); err != nil {
			return err
		}
	}
	if err := srapi.ValidUpdateStrategy(beSpec.UpdateStrategy); err != nil {
		return err
	}
	return nil
}
