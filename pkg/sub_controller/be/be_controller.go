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

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	subc "github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BeController struct {
	k8sClient client.Client
}

func New(k8sClient client.Client) *BeController {
	return &BeController{
		k8sClient: k8sClient,
	}
}

func (be *BeController) GetControllerName() string {
	return "beController"
}

func (be *BeController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksBeSpec == nil {
		if err := be.ClearResources(ctx, src); err != nil {
			klog.Errorf("beController sync clearResource namespace=%s,srcName=%s, err=%s\n", src.Namespace, src.Name, err.Error())
			return err
		}

		return nil
	}

	if !fe.CheckFEReady(ctx, be.k8sClient, src.Namespace, src.Name) {
		return nil
	}

	beSpec := src.Spec.StarRocksBeSpec

	// get the be configMap for resolve ports.
	// 2. get config for generate statefulset and service.
	config, err := be.GetConfig(ctx, &beSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("BeController Sync ", "resolve cn configmap failed, namespace ", src.Namespace, " configmapName ",
			beSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", beSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	feconfig, _ := be.getFeConfig(ctx, &src.Spec.StarRocksFeSpec.ConfigMapInfo, src.Namespace)
	// add query port from fe config.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	// generate new be external service.
	externalsvc := rutils.BuildExternalService(object.NewFromCluster(src),
		beSpec, config, load.Selector(src.Name, beSpec), load.Labels(src.Name, beSpec))
	// generate internal fe service, update the status of cn on src.
	internalService := be.generateInternalService(ctx, src, &externalsvc, config)

	// create be statefulset.
	podTemplateSpec := be.buildPodTemplate(src, config)
	st := statefulset.MakeStatefulset(object.NewFromCluster(src), beSpec, podTemplateSpec)

	// update the statefulset if feSpec be updated.
	if err = k8sutils.ApplyStatefulSet(ctx, be.k8sClient, &st, func(new *appv1.StatefulSet, est *appv1.StatefulSet) bool {
		return rutils.StatefulSetDeepEqual(new, est, false)
	}); err != nil {
		klog.Errorf("BeController Sync patch statefulset name=%s, namespace=%s, error=%s\n", st.Name, st.Namespace, err.Error())
		return err
	}

	if err = k8sutils.ApplyService(ctx, be.k8sClient, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `cn-domain-search` for internal communicating.
		internalService.Name = st.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		klog.Errorf("BeController Sync patch internal service name=%s, namespace=%s, error=%s\n",
			internalService.Name, internalService.Namespace, err.Error())
		return err
	}

	if err = k8sutils.ApplyService(ctx, be.k8sClient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		klog.Error("BeController Sync ", "patch external service namespace ",
			externalsvc.Namespace, " name ", externalsvc.Name)
		return err
	}

	return err
}

// UpdateClusterStatus update the all resource status about be.
func (be *BeController) UpdateClusterStatus(src *srapi.StarRocksCluster) error {
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

	// update statefulset, if restart operation finished, we should update the annotation value as finished.
	var st appv1.StatefulSet
	statefulSetName := load.Name(src.Name, beSpec)
	if err := be.k8sClient.Get(context.Background(),
		types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); apierrors.IsNotFound(err) {
		klog.Infof("BeController UpdateClusterStatus the statefulset name=%s is not found.\n", statefulSetName)
		return nil
	} else if err != nil {
		return err
	}

	bs.ServiceName = service.ExternalServiceName(src.Name, beSpec)
	bs.ResourceNames = rutils.MergeSlices(bs.ResourceNames, []string{statefulSetName})

	if err := subc.UpdateStatus(&bs.StarRocksComponentStatus, be.k8sClient,
		src.Namespace, load.Name(src.Name, beSpec), pod.Labels(src.Name, beSpec), subc.StatefulSetLoadType); err != nil {
		return err
	}

	return nil
}

func (be *BeController) generateInternalService(ctx context.Context,
	src *srapi.StarRocksCluster, externalService *corev1.Service, config map[string]interface{}) *corev1.Service {
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
	if err := be.k8sClient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: "be-domain-search"}, &esearchSvc); err == nil {
		if rutils.HaveEqualOwnerReference(&esearchSvc, searchSvc) {
			searchSvc.Name = "be-domain-search"
		}
	} else if !apierrors.IsNotFound(err) {
		klog.Errorf("beController generateInternalService get old svc namespace=%s, name=%s,failed, error=%s.\n",
			src.Namespace, "be-domain-search", err.Error())
	}

	return searchSvc
}

func (be *BeController) GetConfig(ctx context.Context,
	configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	configMap, err := k8sutils.GetConfigMap(ctx, be.k8sClient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("BeController GetCnConfig config is not exist namespace ", namespace, " configmapName ", configMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

func (be *BeController) getFeConfig(ctx context.Context,
	feconfigMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	feconfigMap, err := k8sutils.GetConfigMap(ctx, be.k8sClient, namespace, feconfigMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("BeController getFeConfig fe config not exist namespace ", namespace, " configmapName ", feconfigMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}
	res, err := rutils.ResolveConfigMap(feconfigMap, feconfigMapInfo.ResolveKey)
	return res, err
}

func (be *BeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	beSpec := src.Spec.StarRocksBeSpec
	if beSpec != nil {
		return nil
	}

	statefulSetName := load.Name(src.Name, beSpec)
	if err := k8sutils.DeleteStatefulset(ctx, be.k8sClient, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.\n",
			src.Namespace, statefulSetName, err.Error())
		return err
	}

	searchServiceName := service.SearchServiceName(src.Name, beSpec)
	err := k8sutils.DeleteService(ctx, be.k8sClient, src.Namespace, searchServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete search service, namespace=%s,name=%s,error=%s.\n",
			src.Namespace, searchServiceName, err.Error())
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, beSpec)
	err = k8sutils.DeleteService(ctx, be.k8sClient, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete external service, namespace=%s, name=%s,error=%s.\n",
			src.Namespace, externalServiceName, err.Error())
		return err
	}

	return nil
}
