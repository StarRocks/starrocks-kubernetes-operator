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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SubController interface {
	// Sync reconcile for sub controller. bool represent the component have updated.
	Sync(ctx context.Context, src *srapi.StarRocksCluster) error

	// ClearResources clear all resource about sub-component.
	ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error)

	// GetControllerName return the controller name, beController, feController,cnController for log.
	GetControllerName() string

	// UpdateStatus update the component status on src.
	UpdateStatus(src *srapi.StarRocksCluster) error

	// SyncRestartStatus sync the status of restart.
	SyncRestartStatus(src *srapi.StarRocksCluster) error
}

func UpdateStatefulSetStatus(componentStatus *srapi.StarRocksComponentStatus, k8sClient client.Client,
	namespace string, name string, podLabels map[string]string) error {
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
		componentStatus.Reason = podsStatus[failed[0]].Reason
	} else {
		// even all pods is ready, that does not mean the fe is ready, maybe fe is upgrading
		componentStatus.Phase = srapi.ComponentReconciling
		var sts appsv1.StatefulSet
		err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &sts)
		if err != nil {
			return err
		}
		reason, done, err := statefulset.Status(&sts)
		if err != nil {
			return err
		}
		if done {
			componentStatus.Phase = srapi.ComponentRunning
		} else {
			componentStatus.Phase = srapi.ComponentReconciling
			componentStatus.Reason = reason
		}
	}
	return nil
}
