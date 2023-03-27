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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
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

const (
	BE_SEARCH_SUFFIX = "-search"
)

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

	//get the be configMap for resolve ports.
	//2. get config for generate statefulset and service.
	config, err := be.GetConfig(ctx, &beSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("BeController Sync ", "resolve cn configmap failed, namespace ", src.Namespace, " configmapName ", beSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", beSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	feconfig, _ := be.getFeConfig(ctx, &src.Spec.StarRocksFeSpec.ConfigMapInfo, src.Namespace)
	//annotation: add query port in cnconfig.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	//generate new cn internal service.
	externalsvc := rutils.BuildExternalService(src, srapi.GetBeExternalServiceName(src), rutils.BeService, config, be.generateServiceSelector(src))

	//create or update fe service, update the status of cn on src.
	//3. issue the service.
	internalService := be.generateInternalService(ctx, src, &externalsvc, config)
	if err := k8sutils.ApplyService(ctx, be.k8sclient, internalService, rutils.ServiceDeepEqual); err != nil {
		klog.Errorf("BeController Sync patch internal service name=%s, namespace=%s, error=%s\n", internalService.Name, internalService.Namespace, err.Error())
		return err
	}

	if err := k8sutils.ApplyService(ctx, be.k8sclient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		klog.Error("BeController Sync ", "patch external service namespace ", externalsvc.Namespace, " name ", externalsvc.Name)
		return err
	}

	//4. create cn statefulset.
	st := rutils.NewStatefulset(be.buildStatefulSetParams(src, config, internalService.Name))
	st.Spec.PodManagementPolicy = appv1.ParallelPodManagement

	//5. last update the status.
	//if the spec is changed, update the status of be on src.
	err = k8sutils.ApplyStatefulSet(ctx, be.k8sclient, &st, func(new *appv1.StatefulSet, est *appv1.StatefulSet) bool {
		//exclude the restart annotation interference. annotation
		_, ok := est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey]
		if !be.statefulsetNeedRolloutRestart(src.Annotations, est.Annotations) && ok {
			// when restart we add `AnnotationRestart` to annotation. so we should add again when we equal the exsit statefulset and new statefulset.
			anno := rutils.Annotations{}
			anno.Add(common.KubectlRestartAnnotationKey, est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey])
			new.Spec.Template.Annotations = anno
		}

		// if have restart annotation, we should exclude the interference for comparison.
		return rutils.StatefulSetDeepEqual(new, est, false)
	})

	return err
}

//check statefuslet rolling restart is needed.
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
	//if spec is not exist, status is empty. but before clear status we must clear all resource about be used by ClearResources.
	if src.Spec.StarRocksBeSpec == nil {
		src.Status.StarRocksBeStatus = nil
		return nil
	}

	bs := &srapi.StarRocksBeStatus{
		Phase: srapi.ComponentReconciling,
	}
	if src.Status.StarRocksBeStatus != nil {
		bs = src.Status.StarRocksBeStatus.DeepCopy()
	}
	src.Status.StarRocksBeStatus = bs

	// update statefulset, if restart operation finished, we should update the annotation value as finished.
	var st appv1.StatefulSet
	if err := be.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.BeStatefulSetName(src)}, &st); apierrors.IsNotFound(err) {
		klog.Infof("BeController UpdateStatus the statefulset name=%s is not found.\n", srapi.BeStatefulSetName(src))
		return nil
	} else if err != nil {
		return err

	}

	bs.ServiceName = srapi.GetBeExternalServiceName(src)
	bs.ResourceNames = rutils.MergeSlices(bs.ResourceNames, []string{srapi.BeStatefulSetName(src)})

	if err := be.updateBeStatus(bs, be.bePodLabels(src), src.Namespace, *src.Spec.StarRocksBeSpec.Replicas); err != nil {
		return err
	}

	//if have pod not running that the operation is not finished, we don't need update statefulset annotation.
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
	if err := be.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.BeStatefulSetName(src)}, &st); err != nil {
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
	//1. wait for fe ok.
	endpoints := corev1.Endpoints{}
	if err := be.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: srapi.GetFeExternalServiceName(src)}, &endpoints); err != nil {
		klog.Infof("BeController Sync wait fe service name %s available occur failed %s\n", srapi.GetFeExternalServiceName(src), err.Error())
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
	searchSvc := &corev1.Service{}
	externalService.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	searchSvc.Name = be.getBeSearchServiceName(src)
	searchSvc.Spec = corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{
			{
				Name:       "heartbeat",
				Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
				TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
			},
		},
		Selector: externalService.Spec.Selector,

		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}
	fs := (rutils.Finalizers)(searchSvc.Finalizers)
	fs.AddFinalizer(srapi.SERVICE_FINALIZER)
	searchSvc.Finalizers = fs
	svcList := corev1.ServiceList{}
	opts := client.ListOptions{
		Namespace: externalService.Namespace,
	}
	matchLabels := client.MatchingLabels{}
	//for compatible use version <=1.3
	matchLabels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	matchLabels.ApplyToList(&opts)
	if err := be.k8sclient.List(ctx, &svcList, &opts); err == nil {
		for i, _ := range svcList.Items {
			if rutils.HaveEqualOwnerReference(searchSvc, &svcList.Items[i]) && svcList.Items[i].Spec.PublishNotReadyAddresses {
				searchSvc.Name = svcList.Items[i].Name
				break
			}
		}
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

//updateCnStatus update the src status about cn status.
func (be *BeController) updateBeStatus(bs *srapi.StarRocksBeStatus, labels map[string]string, namespace string, replicas int32) error {
	var podList corev1.PodList
	if err := be.k8sclient.List(context.Background(), &podList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
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
		} else {
			faileds = append(faileds, pod.Name)
		}
	}

	bs.Phase = srapi.ComponentReconciling
	if len(readys) == int(replicas) {
		bs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		bs.Phase = srapi.ComponentFailed
		bs.Reason = podmap[faileds[0]].Status.Reason
	} else if len(creatings) != 0 {
		bs.Reason = podmap[creatings[0]].Status.Reason
	}

	bs.RunningInstances = readys
	bs.CreatingInstances = creatings
	bs.FailedInstances = faileds
	return nil
}

func (be *BeController) clearFinalizersOnBeResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	cleared := true
	var err error
	beStatus := src.Status.StarRocksBeStatus
	if beStatus == nil {
		return true, nil
	}

	if len(beStatus.ResourceNames) != 0 {
		cleared, err = k8sutils.ClearFinalizersOnStatefulset(ctx, be.k8sclient, src.Namespace, beStatus.ResourceNames[0])
		if err != nil {
			klog.Errorf("beController clearFinalizersOnBeResources clearFinalizersOnStatefulset namespace=%s, name=%s,failed, error=%s.\n", src.Namespace, beStatus.ResourceNames[0], err.Error())
			return cleared, err
		}
	}

	if beStatus.ServiceName != "" {
		exist, serr := k8sutils.ClearFinalizersOnServices(ctx, be.k8sclient, src.Namespace, []string{srapi.GetBeExternalServiceName(src), be.getBeSearchServiceName(src)})
		if serr != nil {
			return exist, serr
		}
		cleared = cleared && exist
	}

	return cleared, nil
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

	cleared, err := be.clearFinalizersOnBeResources(ctx, src)
	if err != nil || !cleared {
		return cleared, err
	}

	stName := srapi.CnStatefulSetName(src)
	if err := k8sutils.DeleteStatefulset(ctx, be.k8sclient, src.Namespace, stName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.\n", src.Namespace, stName, err.Error())
		return false, err
	}

	if err := k8sutils.DeleteService(ctx, be.k8sclient, src.Namespace, be.getBeSearchServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete search service, namespace=%s,name=%s,error=%s.\n", src.Namespace, be.getBeSearchServiceName(src), err.Error())
		return false, err
	}
	if err := k8sutils.DeleteService(ctx, be.k8sclient, src.Namespace, srapi.GetBeExternalServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("beController ClearResources delete external service, namespace=%s, name=%s,error=%s.\n", src.Namespace, srapi.GetFeExternalServiceName(src), err.Error())
		return false, err
	}

	return true, nil
}

func (be *BeController) getBeSearchServiceName(src *srapi.StarRocksCluster) string {
	return src.Name + "-be" + BE_SEARCH_SUFFIX
}
