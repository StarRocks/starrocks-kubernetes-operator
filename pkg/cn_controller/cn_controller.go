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

package cn_controller

import (
	"context"
	"errors"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CnController struct {
	k8sclient   client.Client
	k8srecorder record.EventRecorder
}

func New(k8sclient client.Client, k8srecorder record.EventRecorder) *CnController {
	return &CnController{
		k8sclient:   k8sclient,
		k8srecorder: k8srecorder,
	}
}

func (cc *CnController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksCnSpec == nil {
		klog.Info("CnController Sync ", "the cn component is not needed", " namespace ", src.Namespace, " starrocks cluster name ", src.Name)
		return nil
	}

	//generate new cn service.
	//TODO:need
	svc := rutils.BuildExternalService(src, cc.getCnServiceName(src), rutils.CnService)
	cs := srapi.StarRocksCnStatus{ServiceName: svc.Name}
	src.Status.StarRocksCnStatus = &cs
	//create or update fe service, update the status of cn on src.
	if err := k8sutils.CreateOrUpdateService(ctx, cc.k8sclient, &svc); err != nil {
		klog.Error("CnController Sync ", "create or update service namespace ", svc.Namespace, " name ", svc.Name)
		return err
	}

	cnFinalizers := []string{srapi.CN_SERVICE_FINALIZER}
	defer func() {
		rutils.MergeSlices(src.Finalizers, cnFinalizers)
	}()

	//create cn statefulset.
	st := rutils.NewStatefulset(cc.buildStatefulSetParams(src))
	var cst appv1.StatefulSet
	err := cc.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &cst)
	if err != nil && apierrors.IsNotFound(err) {
		cs.ResourceNames = append(cs.ResourceNames, st.Name)
		cnFinalizers = append(cnFinalizers, srapi.CN_STATEFULSET_FINALIZER)
		return k8sutils.CreateClientObject(ctx, cc.k8sclient, &st)
	} else if err != nil {
		return err
	}

	//if the spec is not change, update the status of cn on src.
	if rutils.StatefulSetDeepEqual(&st, cst) {
		cs.ResourceNames = rutils.MergeSlices(cs.ResourceNames, []string{st.Name})
		if err := cc.updateCnStatus(&cs, st); err != nil {
			return err
		}
		//no update
		return nil
	}

	//cn spec changed update the statefulset.
	rutils.MergeStatefulSets(&st, cst)
	if err := k8sutils.UpdateClientObject(ctx, cc.k8sclient, &st); err != nil {
		return err
	}

	return cc.updateCnStatus(&cs, cst)
}

func (cc *CnController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	//if the starrocks is not have cn.
	if src.Status.StarRocksCnStatus == nil {
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

	for _, name := range src.Status.StarRocksCnStatus.ResourceNames {
		if _, err := k8sutils.DeleteClientObject(ctx, cc.k8sclient, src.Namespace, name); err != nil {
			return false, errors.New("cn delete statefulset" + err.Error())
		}
	}

	if count == len(src.Status.StarRocksCnStatus.ResourceNames) {
		fmap[srapi.CN_STATEFULSET_FINALIZER] = true
	}

	if _, ok := fmap[srapi.CN_STATEFULSET_FINALIZER]; !ok {
		return k8sutils.DeleteClientObject(ctx, cc.k8sclient, src.Namespace, src.Status.StarRocksCnStatus.ServiceName)
	}

	return false, nil
}

func (cc *CnController) updateCnStatus(cs *srapi.StarRocksCnStatus, st appv1.StatefulSet) error {
	var podList corev1.PodList
	if err := cc.k8sclient.List(context.Background(), &podList, client.InNamespace(st.Namespace), client.MatchingLabels(st.Spec.Selector.MatchLabels)); err != nil {
		return err
	}

	var creatings, runnings, faileds []string
	podmap := make(map[string]corev1.Pod)
	//get all pod status that controlled by st.
	for _, pod := range podList.Items {
		//TODO: test
		podmap[pod.Name] = pod
		if pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodRunning {
			runnings = append(runnings, pod.Name)
		} else {
			faileds = append(faileds, pod.Name)
		}
	}

	cs.Phase = srapi.ComponentReconciling
	if len(runnings) == int(*st.Spec.Replicas) {
		cs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		cs.Phase = srapi.ComponentFailed
		cs.Reason = podmap[faileds[0]].Status.Reason
	} else if len(creatings) != 0 {
		cs.Phase = srapi.ComponentReconciling
		cs.Reason = podmap[creatings[0]].Status.Reason
	}

	return nil
}

//buildStatefulSetParams generate the params of construct the statefulset.
func (cc *CnController) buildStatefulSetParams(src *srapi.StarRocksCluster) rutils.StatefulSetParams {
	cnSpec := src.Spec.StarRocksCnSpec
	stname := src.Name + "-" + srapi.DEFAULT_CN
	if cnSpec.Name != "" {
		stname = cnSpec.Name
	}

	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_CN
	labels.AddLabel(src.Labels)
	or := metav1.OwnerReference{
		UID:        src.UID,
		Kind:       src.Kind,
		APIVersion: src.APIVersion,
		Name:       src.Name,
	}

	return rutils.StatefulSetParams{
		Name:            stname,
		Namespace:       src.Namespace,
		ServiceName:     cc.getCnServiceName(src),
		PodTemplateSpec: cc.buildPodTemplate(src),
		Labels:          labels,
		Selector:        labels,
		OwnerReferences: []metav1.OwnerReference{or},
	}
}

func (cc *CnController) getCnServiceName(src *srapi.StarRocksCluster) string {
	if src.Spec.StarRocksCnSpec.Service.Name != "" {
		return src.Spec.StarRocksCnSpec.Service.Name + "-" + "service"
	}

	return src.Name + "-" + srapi.DEFAULT_CN + "-" + "service"
}

//buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(src *srapi.StarRocksCluster) corev1.PodTemplateSpec {
	metaname := src.Name + "-" + srapi.DEFAULT_CN
	labels := src.Labels
	cnSpec := src.Spec.StarRocksCnSpec
	labels[srapi.OwnerReference] = cc.getCnServiceName(src)

	opContainers := []corev1.Container{
		{
			Name:    srapi.DEFAULT_FE,
			Image:   cnSpec.Image,
			Command: []string{},
			Args:    []string{"--daemon"},
			Ports: []corev1.ContainerPort{
				{
					Name:          "thrift-port",
					ContainerPort: 9060,
					Protocol:      corev1.ProtocolTCP,
				}, {
					Name:          "webserver-port",
					ContainerPort: 8040,
					Protocol:      corev1.ProtocolTCP,
				}, {
					Name:          "heartbeat-service-port",
					ContainerPort: 9050,
					Protocol:      corev1.ProtocolTCP,
				}, {
					Name:          "brpc-port",
					ContainerPort: 8060,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					},
				}, {
					Name: "POD_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
					},
				}, {
					Name:  srapi.COMPONENT_NAME,
					Value: srapi.DEFAULT_FE,
				}, {
					Name:  srapi.SERVICE_NAME,
					Value: cc.getCnServiceName(src),
				},
			},
			Resources:       cnSpec.ResourceRequirements,
			ImagePullPolicy: corev1.PullIfNotPresent,
			StartupProbe: &corev1.Probe{
				InitialDelaySeconds: 5,
				FailureThreshold:    120,
				PeriodSeconds:       5,
				ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 9010,
				}}},
			},
		},
	}

	initContainers := []corev1.Container{
		{
			//TODO: 设置启动参数
			Command: []string{},
			Name:    "cn-prepare",
			Image:   cnSpec.Image,
		},
	}

	podSpec := corev1.PodSpec{
		InitContainers:     initContainers,
		Containers:         opContainers,
		ServiceAccountName: src.Spec.ServiceAccount,
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Labels:      labels,
			Annotations: src.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: src.APIVersion,
					Kind:       src.Kind,
					Name:       src.Name,
				},
			},
		},
		Spec: podSpec,
	}
}
