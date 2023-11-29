/*
 * Copyright 2021-present, StarRocks Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package pod

import (
	v1 "k8s.io/api/core/v1"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
)

type PodStatus struct {
	Phase  v1.PodPhase
	Reason string
}

// Status return the status of pods.
func Status(podList v1.PodList) map[string]PodStatus {
	pods := podList.Items
	status := make(map[string]PodStatus)
	for i := range pods {
		status[pods[i].Name] = PodStatus{
			Phase:  pods[i].Status.Phase,
			Reason: pods[i].Status.Reason,
		}
	}
	return status
}

func Count(podList v1.PodList) (creating, ready, failed []string) {
	pods := podList.Items
	for i := range pods {
		if b := k8sutils.PodIsReady(&pods[i].Status); b {
			ready = append(ready, pods[i].Name)
		} else if pods[i].Status.Phase == v1.PodRunning || pods[i].Status.Phase == v1.PodPending {
			creating = append(creating, pods[i].Name)
		} else {
			failed = append(failed, pods[i].Name)
		}
	}
	return
}
