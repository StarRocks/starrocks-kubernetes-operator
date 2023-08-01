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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BeController struct {
	k8sclient   client.Client
	k8srecorder record.EventRecorder
}

func New(k8sclient client.Client, k8srecorder record.EventRecorder) *BeController {
	return &BeController{
		k8sclient:   k8sclient,
		k8srecorder: k8srecorder,
	}
}

func (be *BeController) GetControllerName() string {
	return "beController"
}

func (be *BeController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksBeSpec == nil {
		if _, err := be.ClearResources(ctx, src); err != nil {
			klog.Errorf("beController sync clearResource namespace=%s,srcName=%s, err=%s\n", src.Namespace, src.Name, err.Error())
			return err
		}

		return nil
	}

	if !be.checkFEOK(ctx, src) {
		return nil
	}

	beSpec := src.Spec.StarRocksBeSpec

	// get the be configMap for resolve ports.
	// 2. get config for generate statefulset and service.
	config, err := be.GetConfig(ctx, &beSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("BeController Sync ", "resolve cn configmap failed, namespace ", src.Namespace, " configmapName ", beSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", beSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	feconfig, _ := be.getFeConfig(ctx, &src.Spec.StarRocksFeSpec.ConfigMapInfo, src.Namespace)
	// annotation: add query port in cnconfig.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	// generate new cn external service.
	externalsvc := rutils.BuildExternalService(src, srapi.GetExternalServiceName(src.Name, beSpec), rutils.BeService, config,
		statefulset.Selector(src.Name, beSpec), statefulset.Labels(src.Name, beSpec))
	// generate internal fe service, update the status of cn on src.
	internalService := be.generateInternalService(ctx, src, &externalsvc, config)

	// create cn statefulset.
	podTemplateSpec := be.buildPodTemplate(src, config)
	st := statefulset.MakeStatefulset(statefulset.MakeParams(src, beSpec, podTemplateSpec))

	// update the statefulset if feSpec be updated.
	if err = k8sutils.ApplyStatefulSet(ctx, be.k8sclient, &st, func(new *appv1.StatefulSet, est *appv1.StatefulSet) bool {
		// exclude the restart annotation interference. annotation
		_, ok := est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey]
		if !be.statefulsetNeedRolloutRestart(src.Annotations, est.Annotations) && ok {
			// when restart we add `AnnotationRestart` to annotation. so we should add again when we equal the exsit statefulset and new statefulset.
			anno := rutils.Annotations{}
			anno.Add(common.KubectlRestartAnnotationKey, est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey])
			new.Spec.Template.Annotations = anno
		}

		// if have restart annotation, we should exclude the interference for comparison.
		return rutils.StatefulSetDeepEqual(new, est, false)
	}); err != nil {
		klog.Errorf("BeController Sync patch statefulset name=%s, namespace=%s, error=%s\n", st.Name, st.Namespace, err.Error())
		return err
	}

	if err := k8sutils.ApplyService(ctx, be.k8sclient, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `cn-domain-search` for internal communicating.
		internalService.Name = st.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		klog.Errorf("BeController Sync patch internal service name=%s, namespace=%s, error=%s\n", internalService.Name, internalService.Namespace, err.Error())
		return err
	}

	if err := k8sutils.ApplyService(ctx, be.k8sclient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		klog.Error("BeController Sync ", "patch external service namespace ", externalsvc.Namespace, " name ", externalsvc.Name)
		return err
	}

	return err
}

// check statefulset rolling restart is needed.
func (be *BeController) statefulsetNeedRolloutRestart(srcAnnotations map[string]string, existStatefulsetAnnotations map[string]string) bool {
	srcRestartValue := srcAnnotations[string(srapi.AnnotationBERestartKey)]
	statefulsetRestartValue := existStatefulsetAnnotations[string(srapi.AnnotationBERestartKey)]
	if srcRestartValue == string(srapi.AnnotationRestart) && (statefulsetRestartValue == "" || statefulsetRestartValue == string(srapi.AnnotationRestartFinished)) {
		return true
	}

	return false
}

// UpdateStatus update the all resource status about be.
func (be *BeController) UpdateStatus(src *srapi.StarRocksCluster) error {
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
	statefulSetName := statefulset.Name(src.Name, beSpec)
	if err := be.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); apierrors.IsNotFound(err) {
		klog.Infof("BeController UpdateStatus the statefulset name=%s is not found.\n", statefulSetName)
		return nil
	} else if err != nil {
		return err

	}

	bs.ServiceName = srapi.GetExternalServiceName(src.Name, beSpec)
	bs.ResourceNames = rutils.MergeSlices(bs.ResourceNames, []string{statefulSetName})

	if err := sub_controller.UpdateStatefulSetStatus(&bs.StarRocksComponentStatus, be.k8sclient,
		src.Namespace, statefulset.Name(src.Name, beSpec), pod.Labels(src.Name, beSpec)); err != nil {
		return err
	}

	// if have pod not running that the operation is not finished, we don't need update statefulset annotation.
	if bs.Phase != srapi.ComponentRunning {
		operationValue := st.Annotations[string(srapi.AnnotationBERestartKey)]
		if string(srapi.AnnotationRestart) == operationValue {
			st.Annotations[string(srapi.AnnotationBERestartKey)] = string(srapi.AnnotationRestarting)
			return k8sutils.UpdateClientObject(context.Background(), be.k8sclient, &st)
		}

		return nil
	}

	if value := st.Annotations[string(srapi.AnnotationBERestartKey)]; value == string(srapi.AnnotationRestarting) {
		st.Annotations[string(srapi.AnnotationBERestartKey)] = string(srapi.AnnotationRestartFinished)
		if err := k8sutils.UpdateClientObject(context.Background(), be.k8sclient, &st); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (be *BeController) SyncRestartStatus(src *srapi.StarRocksCluster) error {
	// update statefulset, if restart operation finished, we should update the annotation value as finished.
	var st appv1.StatefulSet
	statefulSetName := statefulset.Name(src.Name, src.Spec.StarRocksBeSpec)
	if err := be.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: statefulSetName}, &st); err != nil {
		klog.Infof("BeController SyncRestartStatus the statefulset name=%s, namespace=%s get error=%s\n.")
		return err
	}

	stValue := st.Annotations[string(srapi.AnnotationBERestartKey)]
	srcValue := src.Annotations[string(srapi.AnnotationBERestartKey)]
	if (srcValue == string(srapi.AnnotationRestart) && stValue == string(srapi.AnnotationRestarting)) ||
		(srcValue == string(srapi.AnnotationRestarting) && stValue == string(srapi.AnnotationRestartFinished)) {
		src.Annotations[string(srapi.AnnotationBERestartKey)] = stValue
	}

	return nil
}

func (be *BeController) checkFEOK(ctx context.Context, src *srapi.StarRocksCluster) bool {
	// 1. wait for fe ok.
	endpoints := corev1.Endpoints{}
	if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: srapi.GetExternalServiceName(src.Name, src.Spec.StarRocksFeSpec)}, &endpoints); err != nil {
		klog.Infof("BeController Sync wait fe service name %s available occur failed %s\n", srapi.GetExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), err.Error())
		return false
	}

	for _, sub := range endpoints.Subsets {
		if len(sub.Addresses) > 0 {
			return true
		}
	}
	return false
}

func (be *BeController) generateInternalService(ctx context.Context, src *srapi.StarRocksCluster, externalService *corev1.Service, config map[string]interface{}) *corev1.Service {
	spec := src.Spec.StarRocksBeSpec
	searchServiceName := service.SearchServiceName(src.Name, spec)
	searchSvc := service.MakeSearchService(searchServiceName, externalService, []corev1.ServicePort{
		{
			Name:       "heartbeat",
			Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
			TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
		},
	})

	// for compatible verison < v1.5
	var esearchSvc corev1.Service
	if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: "be-domain-search"}, &esearchSvc); err == nil {
		if rutils.HaveEqualOwnerReference(&esearchSvc, searchSvc) {
			searchSvc.Name = "be-domain-search"
		}
	} else if !apierrors.IsNotFound(err) {
		klog.Errorf("beController generateInternalService get old svc namespace=%s, name=%s,failed, error=%s.\n", src.Namespace, "be-domain-search", err.Error())
	}

	return searchSvc
}

func (be *BeController) GetConfig(ctx context.Context, configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	configMap, err := k8sutils.GetConfigMap(ctx, be.k8sclient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("BeController GetCnConfig config is not exist namespace ", namespace, " configmapName ", configMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

func (be *BeController) getFeConfig(ctx context.Context, feconfigMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {

	feconfigMap, err := k8sutils.GetConfigMap(ctx, be.k8sclient, namespace, feconfigMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("BeController getFeConfig fe config not exist namespace ", namespace, " configmapName ", feconfigMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}
	res, err := rutils.ResolveConfigMap(feconfigMap, feconfigMapInfo.ResolveKey)
	return res, err
}

func (be *BeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	// if the starrocks is not have cn.
	if src.Status.StarRocksBeStatus == nil {
		return true, nil
	}

	// no delete.
	if src.DeletionTimestamp.IsZero() {
		return true, nil
	}

	spec := src.Spec.StarRocksBeSpec
	statefulSetName := statefulset.Name(src.Name, spec)
	if err := k8sutils.DeleteStatefulset(ctx, be.k8sclient, src.Namespace, statefulSetName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.\n", src.Namespace, statefulSetName, err.Error())
		return false, err
	}

	searchServiceName := service.SearchServiceName(src.Name, spec)
	if err := k8sutils.DeleteService(ctx, be.k8sclient, src.Namespace, searchServiceName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete search service, namespace=%s,name=%s,error=%s.\n", src.Namespace, searchServiceName, err.Error())
		return false, err
	}
	if err := k8sutils.DeleteService(ctx, be.k8sclient, src.Namespace, srapi.GetExternalServiceName(src.Name, src.Spec.StarRocksBeSpec)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete external service, namespace=%s, name=%s,error=%s.\n", src.Namespace, srapi.GetExternalServiceName(src.Name, src.Spec.StarRocksBeSpec), err.Error())
		return false, err
	}

	return true, nil
}
