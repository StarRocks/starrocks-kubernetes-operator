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

package fe_controller

import (
	"context"
	"errors"
	v1alpha12 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
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
)

type FeController struct {
	k8sclient   client.Client
	k8srecorder record.EventRecorder
	feConfig    map[string]interface{}
}

//New construct a FeController.
func New(k8sclient client.Client, k8sRecorder record.EventRecorder) *FeController {
	return &FeController{
		k8sclient:   k8sclient,
		k8srecorder: k8sRecorder,
	}
}

//Sync starRocksCluster spec to fe statefulset and service.
func (fc *FeController) Sync(ctx context.Context, src *v1alpha12.StarRocksCluster) error {
	if src.Spec.StarRocksFeSpec == nil {
		klog.Info("FeController Sync ", "the fe component is not needed ", "namespace ", src.Namespace, " starrocks cluster name ", src.Name)
		return nil
	}

	feSpec := src.Spec.StarRocksFeSpec
	//get the fe configMap for resolve ports.
	config, err := fc.GetFeConfig(ctx, &feSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("FeController Sync ", "resolve fe configmap failed, namespace ", src.Namespace, " configmapName ", feSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", feSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	//generate new fe service.
	svc := rutils.BuildExternalService(src, v1alpha12.GetFeExternalServiceName(src), rutils.FeService, config)
	fs := &v1alpha12.StarRocksFeStatus{ServiceName: svc.Name, Phase: v1alpha12.ComponentReconciling}
	src.Status.StarRocksFeStatus = fs
	//create or update fe external and domain search service, update the status of fe on src.
	if err := fc.applyService(ctx, &svc, config); err != nil {
		klog.Error("FeController Sync ", "create or update service namespace ", svc.Namespace, " name ", svc.Name, " failed, message ", err.Error())
		return err
	}
	feFinalizers := []string{v1alpha12.FE_SERVICE_FINALIZER}
	//create fe statefulset.
	st := rutils.NewStatefulset(fc.buildStatefulSetParams(src, config))
	defer func() {
		src.Finalizers = rutils.MergeSlices(src.Finalizers, feFinalizers)
		fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{st.Name})
	}()

	var est appv1.StatefulSet
	err = fc.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est)
	if err != nil && apierrors.IsNotFound(err) {
		if err := k8sutils.CreateClientObject(ctx, fc.k8sclient, &st); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	//if the spec is changed, merge old and new statefulset. update the status of fe on src.
	if !rutils.StatefulSetDeepEqual(&st, est, false) {
		klog.Info("FeController Sync exist statefulset not equals to new statefuslet")
		//fe spec changed update the statefulset.
		rutils.MergeStatefulSets(&st, est)
		if err := k8sutils.UpdateClientObject(ctx, fc.k8sclient, &st); err != nil {
			return err
		}
	}

	//no changed update the status of fe on src.l
	return fc.updateFeStatus(fs, st)
}

//UpdateFeStatus update the starrockscluster fe status.
func (fc *FeController) updateFeStatus(fs *v1alpha12.StarRocksFeStatus, st appv1.StatefulSet) error {
	var podList corev1.PodList
	if err := fc.k8sclient.List(context.Background(), &podList, client.InNamespace(st.Namespace), client.MatchingLabels(st.Spec.Selector.MatchLabels)); err != nil {
		return err
	}

	var creatings, readys, faileds []string
	podmap := make(map[string]corev1.Pod)
	//get all pod status that controlled by st.
	for _, pod := range podList.Items {
		podmap[pod.Name] = pod
		if ready := k8sutils.PodIsReady(&pod.Status); ready {
			readys = append(readys, pod.Name)
		} else if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			faileds = append(faileds, pod.Name)
		}
	}

	fs.Phase = v1alpha12.ComponentReconciling
	if st.Spec.Replicas != nil && len(readys) == int(*st.Spec.Replicas) {
		fs.Phase = v1alpha12.ComponentRunning
	} else if len(faileds) != 0 {
		fs.Phase = v1alpha12.ComponentFailed
		fs.Reason = podmap[faileds[0]].Status.Message
	} else if len(creatings) != 0 {
		fs.Reason = podmap[creatings[0]].Status.Message
	}

	fs.RunningInstances = readys
	fs.FailedInstances = faileds
	fs.CreatingInstances = creatings
	return nil
}

//GetFeConfig get the fe start config.
func (fc *FeController) GetFeConfig(ctx context.Context, configMapInfo *v1alpha12.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	if configMapInfo.ConfigMapName == "" || configMapInfo.ResolveKey == "" {
		return make(map[string]interface{}), nil
	}
	configMap, err := k8sutils.GetConfigMap(ctx, fc.k8sclient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("the FeController get fe config is not exist namespace ", namespace, " configmapName ", configMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

//ClearResources clear resource about fe.
func (fc *FeController) ClearResources(ctx context.Context, src *v1alpha12.StarRocksCluster) (bool, error) {
	//if the starrocks is not have fe.
	if src.Status.StarRocksFeStatus == nil {
		return true, nil
	}

	if src.DeletionTimestamp.IsZero() {
		return true, nil
	}

	feStatus := src.Status.StarRocksFeStatus
	if feStatus.ServiceName == "" && len(feStatus.ResourceNames) == 0 {
		src.Status.StarRocksFeStatus = nil
		return true, nil
	}

	fmap := map[string]bool{}
	count := 0
	defer func() {
		finalizers := []string{}
		for _, f := range src.Finalizers {
			if _, ok := fmap[f]; !ok {
				finalizers = append(finalizers, f)
			}
		}
		src.Finalizers = finalizers
	}()

	for _, name := range src.Status.StarRocksFeStatus.ResourceNames {
		var st appv1.StatefulSet
		if err := fc.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: name}, &st); err != nil {
			if apierrors.IsNotFound(err) {
				count++
			}
		} else {
			k8sutils.DeleteClientObject(ctx, fc.k8sclient, src.Namespace, name)
		}
	}

	if count == len(src.Status.StarRocksFeStatus.ResourceNames) {
		fmap[v1alpha12.FE_STATEFULSET_FINALIZER] = true
	}

	var svc corev1.Service
	if err := fc.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Status.StarRocksFeStatus.ServiceName}, &svc); err == nil {
		k8sutils.DeleteClientObject(ctx, fc.k8sclient, src.Namespace, src.Status.StarRocksFeStatus.ServiceName)
	}

	return false, nil
}

//getSearchService get the domain service name, the domain service for statefulset.
//domain service have PublishNotReadyAddresses. while used PublishNotReadyAddresses, the fe start need all instance domain can resolve.
func (fc *FeController) getSearchService() string {
	return "fe-domain-search"
}

func (fc *FeController) applyService(ctx context.Context, svc *corev1.Service, config map[string]interface{}) error {
	//need create domain dns service.
	searchSvc := &corev1.Service{}
	svc.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	searchSvc.Name = fc.getSearchService()
	searchSvc.Spec = corev1.ServiceSpec{
		//for compatible kube-dns

		ClusterIP: "None",
		Ports: []corev1.ServicePort{
			{
				Name:       "query-port",
				Port:       rutils.GetPort(config, rutils.QUERY_PORT),
				TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.QUERY_PORT))),
			},
		},
		Selector: svc.Spec.Selector,

		//
		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	if err := k8sutils.ApplyService(ctx, fc.k8sclient, searchSvc); err != nil {
		return errors.New("create or update domain service " + err.Error())
	}
	return k8sutils.ApplyService(ctx, fc.k8sclient, svc)
}
