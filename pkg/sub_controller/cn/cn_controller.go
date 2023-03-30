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

package cn

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
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

type CnController struct {
	k8sclient   client.Client
	k8srecorder record.EventRecorder
}

const (
	CN_SEARCH_SUFFIX = "-search"
)

func New(k8sclient client.Client, k8srecorder record.EventRecorder) *CnController {
	return &CnController{
		k8sclient:   k8sclient,
		k8srecorder: k8srecorder,
	}
}

func (cc *CnController) GetControllerName() string {
	return "cnController"
}

func (cc *CnController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksCnSpec == nil {
		if _, err := cc.ClearResources(ctx, src); err != nil {
			klog.Errorf("cnController sync namespace=%s, name=%s, err=%s", src.Namespace, src.Name, err.Error())
		}
		return nil
	}

	if !cc.checkFEOk(ctx, src) {
		return nil
	}

	cnSpec := src.Spec.StarRocksCnSpec

	//get the cn configMap for resolve ports.
	//2. get config for generate statefulset and service.
	config, err := cc.GetConfig(ctx, &cnSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Error("CnController Sync ", "resolve cn configmap failed, namespace ", src.Namespace, " configmapName ", cnSpec.ConfigMapInfo.ConfigMapName, " configMapKey ", cnSpec.ConfigMapInfo.ResolveKey, " error ", err)
		return err
	}

	feconfig, _ := cc.getFeConfig(ctx, &src.Spec.StarRocksFeSpec.ConfigMapInfo, src.Namespace)
	//annotation: add query port in cnconfig.
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)

	//generate new cn internal service.
	externalsvc := rutils.BuildExternalService(src, srapi.GetCnExternalServiceName(src), rutils.CnService, config, cc.generateServiceSelector(src), cc.generateServiceLabels(src))
	//create or update fe service, update the status of cn on src.
	//3. publish the service.
	//3.1 patch the internal service for fe and cn connection.
	internalService := cc.generateInternalService(ctx, src, &externalsvc, config)
	if err := k8sutils.ApplyService(ctx, cc.k8sclient, internalService, rutils.ServiceDeepEqual); err != nil {
		klog.Infof("CnController Sync patch internal service namespace=%s, name=%s, error=%s", internalService.Namespace, internalService.Name)
		return err
	}
	//3.2 patch the external service for users access cn service.
	if err := k8sutils.ApplyService(ctx, cc.k8sclient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		klog.Infof("CnController Sync patch external service namespace=%s, name=%s, error=%s", externalsvc.Namespace, externalsvc.Name)
		return err
	}

	//4. create cn statefulset.
	st := rutils.NewStatefulset(cc.buildStatefulSetParams(src, config, internalService.Name))

	//5. create or update the status. create statefulset return, must ensure the
	var est appv1.StatefulSet
	if err := cc.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est); apierrors.IsNotFound(err) {
		return k8sutils.CreateClientObject(ctx, cc.k8sclient, &st)
	} else if err != nil {
		klog.Errorf("CnController Sync create statefulset name=%s, namespace=%s error=%s", st.Name, st.Namespace, err.Error())
		return err
	}
	//if the spec is changed, update the status of cn on src.
	var excludeReplica bool
	//if replicas =0 and not the first time, exclude the hash for autoscaler
	if st.Spec.Replicas == nil {
		if _, ok := est.Annotations[srapi.ComponentReplicasEmpty]; !ok {
			excludeReplica = true
		}
	}
	//exclude the restart annotation interference,
	_, ok := est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey]
	if !cc.statefulsetNeedRolloutRestart(src.Annotations, est.Annotations) && ok {
		// when restart we add `AnnotationRestart` to annotation. so we should add again when we equal the exsit statefulset and new statefulset.
		anno := rutils.Annotations{}
		anno.Add(common.KubectlRestartAnnotationKey, est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey])
		st.Spec.Template.Annotations = anno
	}

	if !rutils.StatefulSetDeepEqual(&st, &est, excludeReplica) {
		//if the replicas not zero, represent user have cancel autoscaler.
		if st.Spec.Replicas != nil {
			if _, ok := est.Annotations[srapi.ComponentReplicasEmpty]; ok {
				rutils.MergeStatefulSets(&st, est)
				delete(st.Annotations, srapi.ComponentReplicasEmpty)
				return k8sutils.UpdateClientObject(ctx, cc.k8sclient, &st)
			}
		}

		st.ResourceVersion = est.ResourceVersion
		return k8sutils.PatchClientObject(ctx, cc.k8sclient, &st)
	}

	//5. create autoscaler.
	if cnSpec.AutoScalingPolicy != nil {
		err = cc.deployAutoScaler(ctx, *cnSpec.AutoScalingPolicy, &st, src)
	} else {
		if src.Status.StarRocksCnStatus == nil || src.Status.StarRocksCnStatus.HorizontalScaler.Name == "" {
			return nil
		}

		hpaInfo := src.Status.StarRocksCnStatus.HorizontalScaler
		if _, err := k8sutils.ClearFinalizersOnAutoscaler(ctx, cc.k8sclient, src.Namespace, hpaInfo.Name, hpaInfo.Version); err != nil {
			return err
		}

		err = cc.deleteAutoScaler(ctx, src)
	}

	return err
}

func (cc *CnController) statefulsetNeedRolloutRestart(srcAnnotations map[string]string, existStatefulsetAnnotations map[string]string) bool {
	srcRestartValue := srcAnnotations[string(srapi.AnnotationCNRestartKey)]
	statefulsetRestartValue := existStatefulsetAnnotations[string(srapi.AnnotationCNRestartKey)]
	if srcRestartValue == string(srapi.AnnotationRestart) && (statefulsetRestartValue == "" || statefulsetRestartValue == string(srapi.AnnotationRestartFinished)) {
		return true
	}

	return false
}

func (cc *CnController) SyncRestartStatus(src *srapi.StarRocksCluster) error {
	// update statefulset, if restart operation finished, we should update the annotation value as finished.
	var st appv1.StatefulSet
	if err := cc.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.CnStatefulSetName(src)}, &st); err != nil {
		klog.Infof("CnController SyncRestartStatus the statefulset name=%s, namespace=%s get error=%s\n.", srapi.CnStatefulSetName(src), src.Namespace, err.Error())
		return err
	}

	stValue := st.Annotations[string(srapi.AnnotationCNRestartKey)]
	srcValue := src.Annotations[string(srapi.AnnotationCNRestartKey)]
	if (srcValue == string(srapi.AnnotationRestart) && stValue == string(srapi.AnnotationRestarting)) ||
		(srcValue == string(srapi.AnnotationRestarting) && stValue == string(srapi.AnnotationRestartFinished)) {
		src.Annotations[string(srapi.AnnotationCNRestartKey)] = stValue
	}

	return nil
}

func (cc *CnController) UpdateStatus(src *srapi.StarRocksCluster) error {
	//if spec is not exist, status is empty. but before clear status we must clear all resource about be used by ClearResources.
	if src.Spec.StarRocksCnSpec == nil {
		src.Status.StarRocksCnStatus = nil
		return nil
	}

	cs := &srapi.StarRocksCnStatus{}
	if src.Status.StarRocksCnStatus != nil {
		cs = src.Status.StarRocksCnStatus.DeepCopy()
	}
	cs.Phase = srapi.ComponentReconciling
	src.Status.StarRocksCnStatus = cs

	var st appv1.StatefulSet
	if err := cc.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.CnStatefulSetName(src)}, &st); apierrors.IsNotFound(err) {
		klog.Infof("CnController UpdateStatus  the statefulset name=%s is not found.\n", srapi.CnStatefulSetName(src))
		return nil
	}

	if src.Spec.StarRocksCnSpec.AutoScalingPolicy != nil {
		cs.HorizontalScaler.Name = cc.generateAutoScalerName(src)
		cs.HorizontalScaler.Version = src.Spec.StarRocksCnSpec.AutoScalingPolicy.Version
	} else {
		cs.HorizontalScaler = srapi.HorizontalScaler{}
	}

	cs.ServiceName = srapi.GetCnExternalServiceName(src)
	cs.ResourceNames = rutils.MergeSlices(cs.ResourceNames, []string{srapi.CnStatefulSetName(src)})

	if err := cc.updateCnStatus(cs, cc.cnPodLabels(src), src.Namespace, *st.Spec.Replicas); err != nil {
		return err
	}

	//if have pod not running that the operation is not finished, we don't need update statefulset annotation.
	if cs.Phase != srapi.ComponentRunning {
		operationValue := st.Annotations[string(srapi.AnnotationCNRestartKey)]
		if string(srapi.AnnotationRestart) == operationValue {
			st.Annotations[string(srapi.AnnotationCNRestartKey)] = string(srapi.AnnotationRestarting)
			return k8sutils.UpdateClientObject(context.Background(), cc.k8sclient, &st)
		}

		return nil
	}

	if value := st.Annotations[string(srapi.AnnotationCNRestartKey)]; value == string(srapi.AnnotationRestarting) {
		st.Annotations[string(srapi.AnnotationCNRestartKey)] = string(srapi.AnnotationRestartFinished)
		if err := k8sutils.UpdateClientObject(context.Background(), cc.k8sclient, &st); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

//updateCnStatus update the src status about cn status.
func (cc *CnController) updateCnStatus(cs *srapi.StarRocksCnStatus, labels map[string]string, namespace string, replicas int32) error {
	var podList corev1.PodList
	if err := cc.k8sclient.List(context.Background(), &podList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
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

	cs.Phase = srapi.ComponentReconciling
	if len(readys) == int(replicas) {
		cs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		cs.Phase = srapi.ComponentFailed
		cs.Reason = podmap[faileds[0]].Status.Reason
	} else if len(creatings) != 0 {
		cs.Reason = podmap[creatings[0]].Status.Reason
	}

	cs.RunningInstances = readys
	cs.CreatingInstances = creatings
	cs.FailedInstances = faileds

	return nil
}

//Deploy autoscaler
func (cc *CnController) deployAutoScaler(ctx context.Context, policy srapi.AutoScalingPolicy, target *appv1.StatefulSet, src *srapi.StarRocksCluster) error {
	params := cc.buildCnAutoscalerParams(policy, target, src)
	autoScaler := rutils.BuildHorizontalPodAutoscaler(params)
	if err := k8sutils.PatchOrCreate(ctx, cc.k8sclient, autoScaler); err != nil {
		klog.Errorf("cnController deployAutoscaler failed, namespace=%s,name=%s,version=%s,error=%s", autoScaler.GetNamespace(), autoScaler.GetNamespace(), policy.Version)
		return err
	}

	return nil
}

//deleteAutoScaler delete the autoscaler.
func (cc *CnController) deleteAutoScaler(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Status.StarRocksCnStatus == nil {
		return nil
	}

	if src.Status.StarRocksCnStatus.HorizontalScaler.Name == "" {
		klog.V(6).Infof("cnController not need delete the autoScaler, namespace=%s, src name=%s.", src.Namespace, src.Name)
		return nil
	}

	autoScalerName := src.Status.StarRocksCnStatus.HorizontalScaler.Name
	version := src.Status.StarRocksCnStatus.HorizontalScaler.Version
	if err := k8sutils.DeleteAutoscaler(ctx, cc.k8sclient, src.Namespace, autoScalerName, version); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController sync deploy or delete failed, namespace=%s, autosclaer name=%s, autoscaler version=%s", src.GetNamespace(), autoScalerName, version)
		return err
	}

	src.Status.StarRocksCnStatus.HorizontalScaler = srapi.HorizontalScaler{}
	return nil
}

//generateInternalService the service for fe communicate with cn.
func (cc *CnController) generateInternalService(ctx context.Context, src *srapi.StarRocksCluster, externalService *corev1.Service, config map[string]interface{}) *corev1.Service {
	searchSvc := &corev1.Service{}
	externalService.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	searchSvc.Name = cc.getCnSearchServiceName(src)
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

	//for compatible verison < v1.5
	var esearchSvc corev1.Service
	if err := cc.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: "cn-domain-search"}, &esearchSvc); err == nil {
		if rutils.HaveEqualOwnerReference(&esearchSvc, searchSvc) {
			searchSvc.Name = "cn-domain-search"
		}
	} else if !apierrors.IsNotFound(err) {
		klog.Errorf("ccController generateInternalService get old svc  namespace=%s, name=%s,failed, error=%s.\n", src.Namespace, "cn-domain-search", err.Error())
	}

	return searchSvc
}

//check the fe cluster is ok for add cn node.
func (cc *CnController) checkFEOk(ctx context.Context, src *srapi.StarRocksCluster) bool {
	endpoints := corev1.Endpoints{}
	//1. wait for fe ok.
	if err := cc.k8sclient.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: srapi.GetFeExternalServiceName(src)}, &endpoints); err != nil {
		klog.Errorf("CnController wait fe available fe service name %s, occur failed %s", srapi.GetFeExternalServiceName(src), err.Error())
		return false
	}

	for _, sub := range endpoints.Subsets {
		if len(sub.Addresses) > 0 {
			return true
		}
	}

	return false
}

func (cc *CnController) clearFinalizersOnCnResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	cleared := true
	var err error
	cnStatus := src.Status.StarRocksCnStatus
	if cnStatus == nil {
		return true, nil
	}

	if len(cnStatus.ResourceNames) != 0 {
		cleared, err = k8sutils.ClearFinalizersOnStatefulset(ctx, cc.k8sclient, src.Namespace, cnStatus.ResourceNames[0])
		if err != nil {
			klog.Errorf("beController clearFinalizersOnBeResources clearFinalizersOnStatefulset namespace=%s, name=%s,failed, error=%s.", src.Namespace, cnStatus.ResourceNames[0], err.Error())
			return cleared, err
		}
	}

	if cnStatus.ServiceName != "" {
		svcCleared, serr := k8sutils.ClearFinalizersOnServices(ctx, cc.k8sclient, src.Namespace, []string{srapi.GetCnExternalServiceName(src), cc.getCnSearchServiceName(src)})
		if serr != nil {
			return svcCleared, serr
		}
		cleared = cleared && svcCleared
	}

	if cnStatus.HorizontalScaler.Name != "" {
		hpaCleared, herr := k8sutils.ClearFinalizersOnAutoscaler(ctx, cc.k8sclient, src.Namespace, cnStatus.HorizontalScaler.Name, cnStatus.HorizontalScaler.Version)
		if herr != nil {
			return hpaCleared, herr
		}
		cleared = cleared && hpaCleared
	}

	return cleared, nil
}

//ClearResources clear the deployed resource about cn. statefulset, services, hpa.
func (cc *CnController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	//if the starrocks is not have cn.
	cnStatus := src.Status.StarRocksCnStatus
	if cnStatus == nil {
		return true, nil
	}

	//no delete
	if src.DeletionTimestamp.IsZero() {
		return true, nil
	}

	cleared, err := cc.clearFinalizersOnCnResources(ctx, src)
	if err != nil || !cleared {
		return cleared, err
	}

	stName := srapi.CnStatefulSetName(src)
	if err := k8sutils.DeleteStatefulset(ctx, cc.k8sclient, src.Namespace, stName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.", src.Namespace, stName, err.Error())
		return false, err
	}

	if err := k8sutils.DeleteService(ctx, cc.k8sclient, src.Namespace, cc.getCnSearchServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete search service, namespace=%s,name=%s,error=%s.", src.Namespace, cc.getCnSearchServiceName(src), err.Error())
		return false, err
	}
	if err := k8sutils.DeleteService(ctx, cc.k8sclient, src.Namespace, srapi.GetCnExternalServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete external service, namespace=%s, name=%s,error=%s.", src.Namespace, srapi.GetCnExternalServiceName(src), err.Error())
		return false, err
	}

	if err := cc.deleteAutoScaler(ctx, src); err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	return true, nil
}

func (cc *CnController) GetConfig(ctx context.Context, configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	configMap, err := k8sutils.GetConfigMap(ctx, cc.k8sclient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Info("CnController GetCnConfig cn config is not exist namespace ", namespace, " configmapName ", configMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) getFeConfig(ctx context.Context, feconfigMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {

	feconfigMap, err := k8sutils.GetConfigMap(ctx, cc.k8sclient, namespace, feconfigMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.V(6).Info("CnController getFeConfig fe config is not exist namespace ", namespace, " configmapName ", feconfigMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}
	res, err := rutils.ResolveConfigMap(feconfigMap, feconfigMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) getCnSearchServiceName(src *srapi.StarRocksCluster) string {
	return src.Name + "-cn" + CN_SEARCH_SUFFIX
}
