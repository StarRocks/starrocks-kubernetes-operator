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

package be_controller

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
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

func (be *BeController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksBeSpec == nil {
		klog.Info(" BeControler Sync ", "the be component is not needed", " namespace ", src.Namespace, " starrocks cluster name ", src.Name)
		return nil
	}

	beSpec := src.Spec.StarRocksBeSpec
	bs := &srapi.StarRocksBeStatus{}
	if src.Status.StarRocksBeStatus != nil {
		bs = src.Status.StarRocksBeStatus.DeepCopy()
	}
	bs.Phase = srapi.ComponentWaiting
	src.Status.StarRocksBeStatus = bs
	endpoints := corev1.Endpoints{}
	//1. wait for fe ok.
	if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: srapi.GetFeExternalServiceName(src)}, &endpoints); apierrors.IsNotFound(err) || len(endpoints.Subsets) == 0 {
		klog.Info("BeController Sync wait fe available fe service name ", srapi.GetFeExternalServiceName(src))
		return nil
	} else if err != nil {
		return err
	}

	feReady := false
	for _, sub := range endpoints.Subsets {
		if len(sub.Addresses) > 0 {
			feReady = true
			break
		}
	}
	if !feReady {
		klog.Info("BeController Sync wait fe available fe service name ", srapi.GetFeExternalServiceName(src), " have not ready fe.")
		return nil
	}

	//get the be configMap for resolve ports.
	//2. get config for generate statefulset and service.
	config, err := be.GetCnConfig(ctx, &beSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("BeController Sync ", "resolve cn configmap failed, namespace ", src.Namespace, " configmapName ", beSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", beSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	feconfig, _ := be.getFeConfig(ctx, &src.Spec.StarRocksFeSpec.ConfigMapInfo, src.Namespace)
	//annotation: add query port in cnconfig.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)

	//generate new cn internal service.
	externalsvc := rutils.BuildExternalService(src, srapi.GetBeExternalServiceName(src), rutils.BeService, config)
	insvc := &corev1.Service{}
	externalsvc.ObjectMeta.DeepCopyInto(&insvc.ObjectMeta)
	insvc.Name = be.getBeDomainService()
	insvc.Spec = corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:       "heartbeat",
				Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
				TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
			},
		},
		Selector: externalsvc.Spec.Selector,

		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	bs.ServiceName = insvc.Name
	//create or update fe service, update the status of cn on src.
	//3. issue the service.
	if err := k8sutils.CreateOrUpdateService(ctx, be.k8sclient, insvc); err != nil {
		klog.Error("BeController Sync ", "create or update service namespace ", insvc.Namespace, " name ", insvc.Name)
		return err
	}

	beFinalizers := []string{srapi.BE_SERVICE_FINALIZER}
	//4. create cn statefulset.
	st := rutils.NewStatefulset(be.buildStatefulSetParams(src, config))
	st.Spec.PodManagementPolicy = appv1.ParallelPodManagement
	defer func() {
		src.Finalizers = rutils.MergeSlices(src.Finalizers, beFinalizers)
		bs.ResourceNames = rutils.MergeSlices(bs.ResourceNames, []string{st.Name})
	}()

	var bst appv1.StatefulSet
	err = be.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &bst)
	if err != nil && apierrors.IsNotFound(err) {
		return k8sutils.CreateClientObject(ctx, be.k8sclient, &st)
	} else if err != nil {
		return err
	}

	//6. last update the status.
	//if the spec is changed, update the status of be on src.
	if !rutils.StatefulSetDeepEqual(&st, bst, false) {
		klog.Info("BeController Sync exist statefulset not equals to new statefuslet")
		rutils.MergeStatefulSets(&st, bst)
		if err := k8sutils.UpdateClientObject(ctx, be.k8sclient, &st); err != nil {
			return err
		}
	}

	//no changed update the status of cn on src.
	return be.updateBeStatus(bs, bst)
}

func (be *BeController) GetCnConfig(ctx context.Context, configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
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

//updateCnStatus update the src status about cn status.
func (be *BeController) updateBeStatus(bs *srapi.StarRocksBeStatus, st appv1.StatefulSet) error {
	var podList corev1.PodList
	if err := be.k8sclient.List(context.Background(), &podList, client.InNamespace(st.Namespace), client.MatchingLabels(st.Spec.Selector.MatchLabels)); err != nil {
		return err
	}

	var creatings, readys, faileds []string
	podmap := make(map[string]corev1.Pod)
	//get all pod status that controlled by st.
	for _, pod := range podList.Items {
		//TODO: test
		podmap[pod.Name] = pod
		if ready := k8sutils.PodIsReady(&pod.Status); ready {
			readys = append(readys, pod.Name)
		} else if pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			faileds = append(faileds, pod.Name)
		}
	}

	bs.Phase = srapi.ComponentReconciling
	if len(readys) == int(*st.Spec.Replicas) {
		bs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		bs.Phase = srapi.ComponentFailed
		bs.Reason = podmap[faileds[0]].Status.Reason
	} else if len(creatings) != 0 {
		bs.Phase = srapi.ComponentReconciling
		bs.Reason = podmap[creatings[0]].Status.Reason
	}

	bs.RunningInstances = readys
	bs.CreatingInstances = creatings
	bs.FailedInstances = faileds
	return nil
}

func (be *BeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	//if the starrocks is not have cn.
	if src.Status.StarRocksBeStatus == nil {
		return true, nil
	}

	//no delete.
	if src.DeletionTimestamp.IsZero() {
		return true, nil
	}

	beStatus := src.Status.StarRocksBeStatus
	if beStatus.ServiceName == "" && len(beStatus.ResourceNames) == 0 {
		src.Status.StarRocksBeStatus = nil
		return true, nil
	}

	fmap := map[string]bool{}
	count := 0
	//clear the be finalizers.
	defer func() {
		var finalizers []string
		for _, f := range src.Finalizers {
			if _, ok := fmap[f]; !ok {
				finalizers = append(finalizers, f)
			}
		}
		src.Finalizers = finalizers
	}()

	for _, name := range src.Status.StarRocksBeStatus.ResourceNames {
		var st appv1.StatefulSet
		if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: name}, &st); err != nil {
			if apierrors.IsNotFound(err) {
				count++
			}
		} else {
			k8sutils.DeleteClientObject(ctx, be.k8sclient, src.Namespace, name)
		}
	}

	if count == len(src.Status.StarRocksBeStatus.ResourceNames) {
		fmap[srapi.BE_STATEFULSET_FINALIZER] = true
	}

	var svc corev1.Service
	if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Status.StarRocksBeStatus.ServiceName}, &svc); err == nil {
		k8sutils.DeleteClientObject(ctx, be.k8sclient, src.Namespace, src.Status.StarRocksBeStatus.ServiceName)
	}

	src.Status.StarRocksBeStatus = nil

	return true, nil
}

//getCnDomainService get the cn service name for dns resolve.
func (be *BeController) getBeDomainService() string {
	return "be-domain-search"
}
