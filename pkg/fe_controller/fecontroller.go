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
)

type FeController struct {
	k8sclient   client.Client
	kisrecorder record.EventRecorder
}

//New construct a FeController.
func New(k8sclient client.Client, k8sRecorder record.EventRecorder) *FeController {
	return &FeController{
		k8sclient:   k8sclient,
		kisrecorder: k8sRecorder,
	}
}

//Sync starRocksCluster spec to fe statefulset and service.
func (fc *FeController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksFeSpec == nil {
		klog.Info("FeController Sync ", "the fe component is not needed ", "namespace ", src.Namespace, " starrocks cluster name ", src.Name)
		return nil
	}

	//generate new fe service.
	svc := rutils.BuildExternalService(src, srapi.GetFeExternalServiceName(src), rutils.FeService)
	fs := &srapi.StarRocksFeStatus{ServiceName: svc.Name, Phase: srapi.ComponentReconciling}
	src.Status.StarRocksFeStatus = fs
	//create or update fe external and domain search service, update the status of fe on src.
	if err := fc.createOrUpdateFeService(ctx, &svc); err != nil {
		klog.Error("FeController Sync ", "create or update service namespace ", svc.Namespace, " name ", svc.Name, " failed, message ", err.Error())
		return err
	}
	feFinalizers := []string{srapi.FE_SERVICE_FINALIZER}
	//create fe statefulset.
	st := rutils.NewStatefulset(fc.buildStatefulSetParams(src))
	defer func() {
		src.Finalizers = rutils.MergeSlices(src.Finalizers, feFinalizers)
		fs.ResourceNames = rutils.MergeSlices(fs.ResourceNames, []string{st.Name})
	}()

	var est appv1.StatefulSet
	err := fc.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est)
	if err != nil && apierrors.IsNotFound(err) {
		if err := k8sutils.CreateClientObject(ctx, fc.k8sclient, &st); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	//if the spec is changed, merge old and new statefulset. update the status of fe on src.
	if !rutils.StatefulSetDeepEqual(&st, est) {
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
func (fc *FeController) updateFeStatus(fs *srapi.StarRocksFeStatus, st appv1.StatefulSet) error {
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
		} else if pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			faileds = append(faileds, pod.Name)
		}
	}

	fs.Phase = srapi.ComponentReconciling
	if st.Spec.Replicas != nil && len(readys) == int(*st.Spec.Replicas) {
		fs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		fs.Phase = srapi.ComponentFailed
		fs.Reason = podmap[faileds[0]].Status.Message
	} else if len(creatings) != 0 {
		fs.Reason = podmap[creatings[0]].Status.Message
	}

	fs.RunningInstances = readys
	fs.FailedInstances = creatings
	fs.CreatingInstances = creatings

	return nil
}

//ClearResources clear resource about fe.
func (fc *FeController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	//if the starrocks is not have fe.
	if src.Status.StarRocksFeStatus == nil {
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
		if _, err := k8sutils.DeleteClientObject(ctx, fc.k8sclient, src.Namespace, name); err != nil {
			return false, errors.New("fe delete statefulset" + err.Error())
		}
	}

	if count == len(src.Status.StarRocksFeStatus.ResourceNames) {
		fmap[srapi.FE_STATEFULSET_FINALIZER] = true
	}

	if _, ok := fmap[srapi.FE_STATEFULSET_FINALIZER]; !ok {
		return k8sutils.DeleteClientObject(ctx, fc.k8sclient, src.Namespace, src.Status.StarRocksFeStatus.ServiceName)
	}

	return false, nil
}

//GetFeDomainService get the domain service name, the domain service for statefulset.
//domain service have PublishNotReadyAddresses. while used PublishNotReadyAddresses, the fe start need all instance domain can resolve.
func (fc *FeController) getFeDomainService() string {
	return "fe-domain-search"
}

func (fc *FeController) createOrUpdateFeService(ctx context.Context, svc *corev1.Service) error {
	//need create domain dns service.
	domainSvc := &corev1.Service{}
	svc.ObjectMeta.DeepCopyInto(&domainSvc.ObjectMeta)
	domainSvc.Name = fc.getFeDomainService()
	domainSvc.Spec = corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:       "query-port",
				Port:       9030,
				TargetPort: intstr.FromInt(9030),
			},
		},
		Selector: svc.Spec.Selector,

		//value = true, Pod don't need to become ready that be search by domain.
		PublishNotReadyAddresses: true,
	}

	if err := k8sutils.CreateOrUpdateService(ctx, fc.k8sclient, domainSvc); err != nil {
		return errors.New("create or update domain service " + err.Error())
	}
	return k8sutils.CreateOrUpdateService(ctx, fc.k8sclient, svc)
}
