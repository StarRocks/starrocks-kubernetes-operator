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

package fe

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
)

const (
	FE_SEARCH_SUFFIX = "-search"
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

func (fc *FeController) GetControllerName() string {
	return "feController"
}

//Sync starRocksCluster spec to fe statefulset and service.
func (fc *FeController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksFeSpec == nil {
		klog.Infof("FeController Sync: the fe component is not needed, namespace = %v, starrocks cluster name = %v", src.Namespace, src.Name)
		return nil
	}

	feSpec := src.Spec.StarRocksFeSpec
	//get the fe configMap for resolve ports.
	config, err := fc.GetFeConfig(ctx, &feSpec.ConfigMapInfo, src.Namespace)
	if err != nil {
		klog.Errorf("FeController Sync: get fe configmap failed, "+
			"namespace = %v, configmapName = %v, configmapKey = %v, error = %v",
			src.Namespace, feSpec.ConfigMapInfo.ConfigMapName, feSpec.ConfigMapInfo.ResolveKey, err)
		return err
	}

	//generate new fe service.
	svc := rutils.BuildExternalService(src, srapi.GetFeExternalServiceName(src), rutils.FeService, config, fc.generateServiceSelector(src), fc.generateServiceLabels(src))
	//create or update fe external and domain search service, update the status of fe on src.
	internalService := fc.generateInternalService(src, &svc, config)

	// first deploy statefulset for compatible v1.5, apply statefulset for update pod.
	st := rutils.NewStatefulset(fc.buildStatefulSetParams(src, config, internalService.Name))
	if err = k8sutils.ApplyStatefulSet(ctx, fc.k8sclient, &st, func(new *appv1.StatefulSet, est *appv1.StatefulSet) bool {
		//exclude the restart annotation interference,
		_, ok := est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey]
		if !fc.statefulsetNeedRolloutRestart(src.Annotations, est.Annotations) && ok {
			// when restart we add `AnnotationRestart` to annotation. so we should add again when we equal the exsit statefulset and new statefulset.
			anno := rutils.Annotations{}
			anno.Add(common.KubectlRestartAnnotationKey, est.Spec.Template.Annotations[common.KubectlRestartAnnotationKey])
			new.Spec.Template.Annotations = anno
		}

		// if have restart annotation, we should exclude the interference for comparison.
		return rutils.StatefulSetDeepEqual(new, est, false)
	}); err != nil {
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.k8sclient, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		//for compatible v1.5, we use `fe-domain-search` for internal communicating.
		internalService.Name = st.Spec.ServiceName

		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		klog.Error("FeController Sync ", "create or patch internal service namespace ", internalService.Namespace, " name ", internalService.Name, " failed, message ", err.Error())
		return err
	}

	if err = k8sutils.ApplyService(ctx, fc.k8sclient, &svc, rutils.ServiceDeepEqual); err != nil {
		klog.Error("FeController Sync ", "create or patch external service namespace ", svc.Namespace, " name ", svc.Name, " failed, message ", err.Error())
		return err
	}

	return nil
}

func (fc *FeController) statefulsetNeedRolloutRestart(srcAnnotations map[string]string, existStatefulsetAnnotations map[string]string) bool {
	srcRestartValue := srcAnnotations[string(srapi.AnnotationFERestartKey)]
	statefulsetRestartValue := existStatefulsetAnnotations[string(srapi.AnnotationFERestartKey)]
	if srcRestartValue == string(srapi.AnnotationRestart) && (statefulsetRestartValue == "" || statefulsetRestartValue == string(srapi.AnnotationRestartFinished)) {
		return true
	}

	return false
}

// UpdateStatus update the all resource status about fe.
func (fc *FeController) UpdateStatus(src *srapi.StarRocksCluster) error {
	//if spec is not exist, status is empty. but before clear status we must clear all resource about be used by ClearResources.
	if src.Spec.StarRocksFeSpec == nil {
		src.Status.StarRocksFeStatus = nil
		return nil
	}

	fs := &srapi.StarRocksFeStatus{
		Phase: srapi.ComponentReconciling,
	}

	if src.Status.StarRocksFeStatus != nil {
		fs = src.Status.StarRocksFeStatus.DeepCopy()
	}

	src.Status.StarRocksFeStatus = fs
	fs.ServiceName = srapi.GetFeExternalServiceName(src)
	fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{srapi.FeStatefulSetName(src)})

	if err := fc.updateFeStatus(fs, fc.fePodLabels(src), src.Namespace, *src.Spec.StarRocksFeSpec.Replicas); err != nil {
		return err
	}

	var st appv1.StatefulSet
	if err := fc.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.FeStatefulSetName(src)}, &st); err != nil {
		return err
	}

	//if have pod not running that the operation is not finished, we don't need update statefulset annotation.
	if fs.Phase != srapi.ComponentRunning {
		operationValue := st.Annotations[string(srapi.AnnotationFERestartKey)]
		if string(srapi.AnnotationRestart) == operationValue {
			st.Annotations[string(srapi.AnnotationFERestartKey)] = string(srapi.AnnotationRestarting)
			return k8sutils.UpdateClientObject(context.Background(), fc.k8sclient, &st)
		}

		return nil
	}

	if value := st.Annotations[string(srapi.AnnotationFERestartKey)]; value == string(srapi.AnnotationRestarting) {
		st.Annotations[string(srapi.AnnotationFERestartKey)] = string(srapi.AnnotationRestartFinished)
		if err := k8sutils.UpdateClientObject(context.Background(), fc.k8sclient, &st); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (fc *FeController) SyncRestartStatus(src *srapi.StarRocksCluster) error {
	// update statefulset, if restart operation finished, we should update the annotation value as finished.
	var st appv1.StatefulSet
	if err := fc.k8sclient.Get(context.Background(), types.NamespacedName{Namespace: src.Namespace, Name: srapi.FeStatefulSetName(src)}, &st); err != nil {
		klog.Infof("FeController SyncRestartStatus the statefulset name=%s, namespace=%s get error=%s\n.")
		return err
	}

	stValue := st.Annotations[string(srapi.AnnotationFERestartKey)]
	srcValue := src.Annotations[string(srapi.AnnotationFERestartKey)]
	if (srcValue == string(srapi.AnnotationRestart) && stValue == string(srapi.AnnotationRestarting)) ||
		(srcValue == string(srapi.AnnotationRestarting) && stValue == string(srapi.AnnotationRestartFinished)) {
		src.Annotations[string(srapi.AnnotationFERestartKey)] = stValue
	}

	return nil
}

//UpdateFeStatus update the starrockscluster fe status.
func (fc *FeController) updateFeStatus(fs *srapi.StarRocksFeStatus, labels map[string]string, namespace string, replicas int32) error {
	var podList corev1.PodList
	if err := fc.k8sclient.List(context.Background(), &podList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
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

	fs.Phase = srapi.ComponentReconciling
	if len(readys) == int(replicas) {
		fs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		fs.Phase = srapi.ComponentFailed
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
func (fc *FeController) GetFeConfig(ctx context.Context, configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	if configMapInfo.ConfigMapName == "" || configMapInfo.ResolveKey == "" {
		return make(map[string]interface{}), nil
	}
	configMap, err := k8sutils.GetConfigMap(ctx, fc.k8sclient, namespace, configMapInfo.ConfigMapName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("the FeController get fe config is not exist, namespace = %s configmapName = %s", namespace, configMapInfo.ConfigMapName)
			return make(map[string]interface{}), nil
		}
		klog.Errorf("error occurred when FeController get fe config, namespace = %s configmapName = %s", namespace, configMapInfo.ConfigMapName)
		return nil, err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

//ClearResources clear resource about fe.
func (fc *FeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	//if the starrocks is not have fe.
	if src.Status.StarRocksFeStatus == nil {
		return true, nil
	}

	if src.DeletionTimestamp.IsZero() {
		return true, nil
	}

	//check all resource's finalizers have deleted, if not delete its.
	cleared, err := fc.clearFinalizersOnFeResources(ctx, src)
	if err != nil || !cleared {
		return cleared, err
	}

	stName := srapi.FeStatefulSetName(src)
	if err := k8sutils.DeleteStatefulset(ctx, fc.k8sclient, src.Namespace, stName); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("feController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.", src.Namespace, stName, err.Error())
		return false, err
	}

	if err := k8sutils.DeleteService(ctx, fc.k8sclient, src.Namespace, fc.getSearchServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("feController ClearResources delete search service, namespace=%s,name=%s,error=%s.", src.Namespace, fc.getSearchServiceName(src), err.Error())
		return false, err
	}
	if err := k8sutils.DeleteService(ctx, fc.k8sclient, src.Namespace, srapi.GetFeExternalServiceName(src)); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("feController ClearResources delete external service, namespace=%s, name=%s,error=%s.", src.Namespace, srapi.GetFeExternalServiceName(src), err.Error())
		return false, err
	}

	return true, nil
}

//clearFinalizersOnFeResources if resources have finalizers delete it or return true.
func (fc *FeController) clearFinalizersOnFeResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	cleared := true
	var err error
	feStatus := src.Status.StarRocksFeStatus
	if feStatus == nil {
		return true, nil
	}

	if len(feStatus.ResourceNames) != 0 {
		cleared, err = k8sutils.ClearFinalizersOnStatefulset(ctx, fc.k8sclient, src.Namespace, srapi.FeStatefulSetName(src))
		if err != nil {
			klog.Errorf("feController clearFinalizersOnFeResources clearFinalizersOnStatefulset namespace=%s, name=%s,failed, error=%s.", src.Namespace, srapi.FeStatefulSetName(src), err.Error())
			return cleared, err
		}
	}

	if feStatus.ServiceName != "" {
		exist, serr := k8sutils.ClearFinalizersOnServices(ctx, fc.k8sclient, src.Namespace, []string{srapi.GetFeExternalServiceName(src), fc.getSearchServiceName(src)})
		if serr != nil {
			return exist, serr
		}
		cleared = cleared && exist
	}

	return cleared, nil
}

//getSearchServiceName get the domain service name, the domain service for statefulset.
//domain service have PublishNotReadyAddresses. while used PublishNotReadyAddresses, the fe start need all instance domain can resolve.
func (fc *FeController) getSearchServiceName(src *srapi.StarRocksCluster) string {
	return src.Name + "-fe" + FE_SEARCH_SUFFIX
}

func (fc *FeController) generateInternalService(src *srapi.StarRocksCluster, externalService *corev1.Service, config map[string]interface{}) *corev1.Service {
	searchSvc := &corev1.Service{}
	externalService.ObjectMeta.DeepCopyInto(&searchSvc.ObjectMeta)
	fs := (rutils.Finalizers)(searchSvc.Finalizers)
	fs.AddFinalizer(srapi.SERVICE_FINALIZER)
	searchSvc.Finalizers = fs
	searchSvc.Name = fc.getSearchServiceName(src)
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
		Selector: externalService.Spec.Selector,

		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	return searchSvc
}
