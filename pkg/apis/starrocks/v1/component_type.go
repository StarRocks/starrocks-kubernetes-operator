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

import (
	"errors"
	"strings"
	"time"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type StarRocksComponentSpec struct {
	StarRocksLoadSpec `json:",inline"`

	// RunAsNonRoot is used to determine whether to run starrocks as a normal user.
	// If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
	// default: nil
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`

	// refer to https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container
	// grant certain privileges to a process without granting all the privileges of the root user
	// +optional
	Capabilities *corev1.Capabilities `json:"capabilities,omitempty"`

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

	// Sidecars is an optional list of containers that are run in the same pod as the starrocks component.
	// You can use this field to launch helper containers that provide additional functionality to the main container.
	// See https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Container for how to configure a container.
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`

	// InitContainers is an optional list of containers that are run in the same pod as the starrocks component.
	// You can use this field to launch helper containers that run before the main container starts.
	// See https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Container for how to configure a container.
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// Entrypoint array. Not executed within a shell.
	// If this is not provided, it will use default entrypoint for different components:
	//	1. For FE, it will use /opt/starrocks/fe_entrypoint.sh as the entrypoint.
	//  2. For BE, it will use /opt/starrocks/be_entrypoint.sh as the entrypoint.
	//  3. For CN, it will use /opt/starrocks/cn_entrypoint.sh as the entrypoint.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	Command []string `json:"command,omitempty"`

	// Arguments to the entrypoint.
	// If this is not provided, it will use $(FE_SERVICE_NAME) for all components.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	Args []string `json:"args,omitempty"`

	// StarRocksCluster use StatefulSet to deploy FE/BE/CN components.
	// UpdateStrategy indicates the StatefulSetUpdateStrategy that will be
	// employed to update Pods in the StatefulSet when a revision is made to
	// Template. See https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#rolling-updates for more details.
	// Note: The maxUnavailable field is in Alpha stage and it is honored only by API servers that are running with the
	//       MaxUnavailableStatefulSet feature gate enabled.
	// +optional
	UpdateStrategy *appv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// Whether this container has a read-only root filesystem.
	// Default is false.
	// Note that:
	// 	1. This field cannot be set when spec.os.name is windows.
	//	2. The FE/BE/CN container should support read-only root filesystem. The newest version of FE/BE/CN is 3.3.6,
	//     and does not support read-only root filesystem
	// +optional
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty" protobuf:"varint,6,opt,name=readOnlyRootFilesystem"`
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

	// ResourceNames the statefulset names of fe.
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

// MountInfo
// The reason why we do not support defaultMode is that we use hash.HashObject to
// calculate the actual volume name. This volume name is used in pod template of statefulset,
// and if this MountInfo type has been changed, the volume name will be changed too, and
// that will make pods restart.
// The default mode is 0644, and in order to support to set permission information for a configMap
// or secret, we add should specify the subPath and specify a command or args in the container.
// And It will be set 0755.
type MountInfo struct {
	// This must match the Name of a ConfigMap or Secret in the same namespace, and
	// the length of name must not more than 50 characters.
	Name string `json:"name,omitempty"`

	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath,omitempty"`

	// SubPath within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +optional
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

func (spec *StarRocksComponentSpec) GetCapabilities() *corev1.Capabilities {
	return spec.Capabilities
}

func (spec *StarRocksComponentSpec) GetSidecars() []corev1.Container {
	return spec.Sidecars
}

func (spec *StarRocksComponentSpec) GetInitContainers() []corev1.Container {
	return spec.InitContainers
}

func (spec *StarRocksComponentSpec) GetCommand() []string {
	return spec.Command
}

func (spec *StarRocksComponentSpec) GetArgs() []string {
	return spec.Args
}

func (spec *StarRocksComponentSpec) GetUpdateStrategy() *appv1.StatefulSetUpdateStrategy {
	if spec.UpdateStrategy == nil {
		const defaultRollingUpdateStartPod int32 = 0
		return &appv1.StatefulSetUpdateStrategy{
			Type: appv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appv1.RollingUpdateStatefulSetStrategy{
				Partition: func(v int32) *int32 { return &v }(defaultRollingUpdateStartPod),
			},
		}
	}
	return spec.UpdateStrategy
}

func ValidUpdateStrategy(updateStrategy *appv1.StatefulSetUpdateStrategy) error {
	if updateStrategy != nil {
		if (updateStrategy.Type == "" || updateStrategy.Type == appv1.RollingUpdateStatefulSetStrategyType) &&
			updateStrategy.RollingUpdate != nil {
			rollingUpdate := updateStrategy.RollingUpdate
			if rollingUpdate.MaxUnavailable != nil {
				s := rollingUpdate.MaxUnavailable.String()
				if strings.HasPrefix(s, "0") {
					return errors.New("maxUnavailable field should > 0")
				}
			}
		}
	}
	return nil
}

func (spec *StarRocksComponentSpec) IsReadOnlyRootFilesystem() *bool {
	if spec.ReadOnlyRootFilesystem == nil {
		b := false
		return &b
	}
	return spec.ReadOnlyRootFilesystem
}

// DisasterRecovery is used to determine whether to enter disaster recovery mode.
type DisasterRecovery struct {
	// Enabled is used to determine whether to enter disaster recovery mode.
	Enabled bool `json:"enabled,omitempty"`

	// Generation records the generation of disaster recovery. If you want to trigger disaster recovery, you should
	// increase the generation.
	Generation int64 `json:"generation,omitempty"`
}

// DisasterRecoveryStatus represents the status of disaster recovery.
// Note: you should create a new instance of DisasterRecoveryStatus by NewDisasterRecoveryStatus.
type DisasterRecoveryStatus struct {
	// the available phase include: todo, doing, done
	Phase DRPhase `json:"phase,omitempty"`

	// the reason of disaster recovery.
	Reason string `json:"reason,omitempty"`

	// the unix time of starting disaster recovery.
	StartTimestamp int64 `json:"startTimestamp,omitempty"`

	// the unix time of ending disaster recovery.
	EndTimestamp int64 `json:"endTimestamp,omitempty"`

	// the observed generation of disaster recovery.
	// If the observed generation is less than the generation, it will trigger disaster recovery.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// NewDisasterRecoveryStatus creates a new disaster recovery status which the phase is todo.
func NewDisasterRecoveryStatus(generation int64) *DisasterRecoveryStatus {
	return &DisasterRecoveryStatus{
		Phase:              DRPhaseTodo,
		StartTimestamp:     time.Now().Unix(),
		ObservedGeneration: generation,
	}
}

type DRPhase string

const (
	DRPhaseTodo  DRPhase = "todo"
	DRPhaseDoing DRPhase = "doing"
	DRPhaseDone  DRPhase = "done"
)
