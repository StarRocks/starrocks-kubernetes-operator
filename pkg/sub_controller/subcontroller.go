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

// +k8s:deepcopy-gen=package
// +groupName=starrocks.com
package sub_controller

import (
	"context"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/deployment"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterSubController interface {
	// SyncCluster reconcile for sub controller. bool represent the component have updated.
	SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error

	// ClearResources clear all resource about sub-component.
	ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error

	// GetControllerName return the controller name, beController, feController,cnController for log.
	GetControllerName() string

	// UpdateClusterStatus update the component status on src.
	UpdateClusterStatus(src *srapi.StarRocksCluster) error
}

type WarehouseSubController interface {
	// ClearWarehouse will clear all resource about warehouse.
	ClearWarehouse(ctx context.Context, namespace string, name string) error

	SyncWarehouse(ctx context.Context, src *srapi.StarRocksWarehouse) error

	GetControllerName() string

	UpdateWarehouseStatus(warehouse *srapi.StarRocksWarehouse) error
}

type LoadType string

const (
	DeploymentLoadType  = "Deployment"
	StatefulSetLoadType = "StatefulSet"
)

func UpdateStatus(componentStatus *srapi.StarRocksComponentStatus, k8sClient client.Client,
	namespace string, name string, podLabels map[string]string, loadType LoadType) error {
	ctx := context.TODO()

	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(namespace), client.MatchingLabels(podLabels)); err != nil {
		return err
	}

	creating, ready, failed := pod.Count(podList)
	podsStatus := pod.Status(podList)
	componentStatus.RunningInstances = ready
	componentStatus.FailedInstances = failed
	componentStatus.CreatingInstances = creating

	if len(failed) != 0 {
		componentStatus.Phase = srapi.ComponentFailed
		componentStatus.Reason = podsStatus[failed[0]].Reason
	} else if len(creating) != 0 {
		componentStatus.Phase = srapi.ComponentReconciling
		componentStatus.Reason = podsStatus[creating[0]].Reason
	} else {
		// even all pods is ready, that does not mean the fe is ready, maybe fe is upgrading
		var reason string
		var done bool
		var err error
		if loadType == DeploymentLoadType {
			var load appsv1.Deployment
			componentStatus.Phase = srapi.ComponentReconciling
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &load)
			if err != nil {
				return err
			}
			reason, done, err = deployment.Status(&load)
			if err != nil {
				return err
			}
		} else if loadType == StatefulSetLoadType {
			var sts appsv1.StatefulSet
			componentStatus.Phase = srapi.ComponentReconciling
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &sts)
			if err != nil {
				return err
			}
			reason, done, err = statefulset.Status(&sts)
			if err != nil {
				return err
			}
		}
		if done {
			componentStatus.Phase = srapi.ComponentRunning
			componentStatus.Reason = ""
		} else {
			componentStatus.Phase = srapi.ComponentReconciling
			componentStatus.Reason = reason
		}
	}
	return nil
}
