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
 * limitations under the License.
 */

package v1

import corev1 "k8s.io/api/core/v1"

type StarRocksComponentSpec struct {
	StarRocksLoadSpec `json:",inline"`

	// RunAsNonRoot is used to determine whether to run starrocks as a normal user.
	// If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
	// default: nil
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`

	// the reference for configMap which allow users to mount any files to container.
	// +optional
	ConfigMaps []ConfigMapReference `json:"configMaps,omitempty"`

	// the reference for secrets.
	// +optional
	Secrets []SecretReference `json:"secrets,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`

	// TerminationGracePeriodSeconds defines duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 120 seconds.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
}

// StarRocksComponentStatus represents the status of a starrocks component.
type StarRocksComponentStatus struct {
	// the name of fe service exposed for user.
	ServiceName string `json:"serviceName,omitempty"`

	// FailedInstances failed pod names.
	FailedInstances []string `json:"failedInstances,omitempty"`

	// CreatingInstances in creating pod names.
	CreatingInstances []string `json:"creatingInstances,omitempty"`

	// RunningInstances in running status pod names.
	RunningInstances []string `json:"runningInstances,omitempty"`

	// ResourceNames the statefulset names of fe in v1alpha1 version.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// Phase the value from all pods of component status. If component have one failed pod phase=failed,
	// also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.
	Phase ComponentPhase `json:"phase"`

	// +optional
	// Reason represents the reason of not running.
	Reason string `json:"reason"`
}

type ConfigMapInfo struct {
	// the config info for start progress.
	ConfigMapName string `json:"configMapName,omitempty"`

	// the config response key in configmap.
	ResolveKey string `json:"resolveKey,omitempty"`
}

type ConfigMapReference MountInfo

type SecretReference MountInfo

type MountInfo struct {
	// This must match the Name of a ConfigMap or Secret in the same namespace, and
	// the length of name must not more than 50 characters.
	Name string `json:"name,omitempty"`

	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath,omitempty"`

	// SubPath within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	SubPath string `json:"subPath,omitempty"`
}

func (spec *StarRocksComponentSpec) GetHostAliases() []corev1.HostAlias {
	return spec.HostAliases
}

func (spec *StarRocksComponentSpec) GetRunAsNonRoot() (*int64, *int64) {
	runAsNonRoot := spec.RunAsNonRoot
	if runAsNonRoot == nil || !*runAsNonRoot {
		return nil, nil
	}

	var userID int64 = 1000
	var groupID int64 = 1000
	return &userID, &groupID
}

func (spec *StarRocksComponentSpec) GetTerminationGracePeriodSeconds() *int64 {
	var defaultSeconds int64 = 120
	if spec.TerminationGracePeriodSeconds == nil {
		return &defaultSeconds
	}
	return spec.TerminationGracePeriodSeconds
}
