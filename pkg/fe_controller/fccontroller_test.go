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
	"fmt"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestFeController_updateStatus(t *testing.T) {
	var creatings, readys, faileds []string
	podmap := make(map[string]corev1.Pod)
	//get all pod status that controlled by st.
	var podList corev1.PodList
	podList.Items = append(podList.Items, corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}})

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

	fmt.Printf("the ready len %d, the creatings len %d, the faileds %d", len(readys), len(creatings), len(faileds))
}
