/*
Copyright 2022 StarRocks.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// PodPolicy reserve expansibility for cr
type PodPolicy struct {
	Labels       map[string]string `json:"labels,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Resources specifies the resource requirements of a pod to run in the cluster
	Resources        corev1.ResourceRequirements   `json:"resources,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Affinity         *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext  *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	// +optional
	Command []string `json:"command,omitempty"`
	// +optional
	Args []string `json:"args,omitempty"`
}
