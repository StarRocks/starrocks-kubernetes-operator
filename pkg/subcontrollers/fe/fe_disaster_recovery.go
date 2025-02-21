package fe

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

func ShouldEnterDisasterRecoveryMode(drSpec *v1.DisasterRecovery,
	drStatus *v1.DisasterRecoveryStatus, feConfig map[string]interface{}) (bool, int32) {
	if !IsRunInSharedDataMode(feConfig) {
		// not run in shared data mode, not in disaster recovery mode
		return false, 0
	}

	if drSpec == nil {
		return false, 0
	}

	if !drSpec.Enabled {
		return false, 0
	}

	// we separate the if condition in order to fix: unnecessary leading newline (whitespace)
	if drStatus == nil {
		return true, rutils.GetPort(feConfig, rutils.QUERY_PORT)
	}

	og := drStatus.ObservedGeneration
	if drSpec.Generation > og || (drSpec.Generation == og && drStatus.Phase != v1.DRPhaseDone) {
		return true, rutils.GetPort(feConfig, rutils.QUERY_PORT)
	}

	return false, 0
}

func EnterDisasterRecoveryMode(ctx context.Context, k8sClient client.Client,
	src *v1.StarRocksCluster, sts *appv1.StatefulSet, queryPort int32) error {
	feSpec := src.Spec.StarRocksFeSpec
	drSpec := src.Spec.DisasterRecovery
	drStatus := src.Status.DisasterRecoveryStatus
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info("enter disaster recovery mode")
	if drStatus == nil || drSpec.Generation > drStatus.ObservedGeneration {
		drStatus = v1.NewDisasterRecoveryStatus(drSpec.Generation)
		src.Status.DisasterRecoveryStatus = drStatus
	}

	switch drStatus.Phase {
	case v1.DRPhaseTodo:
		if !hasClusterSnapshotConf(feSpec.ConfigMaps) {
			drStatus.Phase = v1.DRPhaseTodo
			reason := "cluster_snapshot.yaml is not mounted"
			drStatus.Reason = reason
			return errors.New(reason)
		}
		// rewrite the statefulset
		rewriteStatefulSetForDisasterRecovery(sts, drSpec.Generation, queryPort)
		drStatus.Phase = v1.DRPhaseDoing
		drStatus.Reason = "has changed to pod template for disaster recovery"
	case v1.DRPhaseDoing:
		// check whether the pod is ready
		rewriteStatefulSetForDisasterRecovery(sts, drSpec.Generation, queryPort)
		if !CheckFEReadyInDisasterRecovery(ctx, k8sClient, src.Namespace, src.Name, drSpec.Generation) {
			drStatus.Reason = "disaster recovery is in progress"
		} else {
			drStatus.Phase = v1.DRPhaseDone
			drStatus.Reason = "disaster recovery is done"
			drStatus.EndTimestamp = time.Now().Unix()
		}
	}
	return nil
}

func hasClusterSnapshotConf(configMaps []v1.ConfigMapReference) bool {
	// check all the mount paths, to make sure cluster_snapshot.yaml is mounted
	hasConf := false
	for _, sv := range configMaps {
		if strings.Contains(sv.SubPath, "cluster_snapshot.yaml") ||
			strings.HasSuffix(sv.MountPath, "fe/conf") ||
			strings.HasSuffix(sv.MountPath, "fe/conf/") {
			hasConf = true
			break
		}
	}
	return hasConf
}

func rewriteStatefulSetForDisasterRecovery(expectSts *appv1.StatefulSet, generation int64, queryPort int32) *appv1.StatefulSet {
	// set replicas to 1
	expectSts.Spec.Replicas = func(i int32) *int32 { return &i }(1)

	// set the pod template
	podTemplate := &expectSts.Spec.Template
	feContainer := &(podTemplate.Spec.Containers[0])
	feContainer.StartupProbe = nil
	feContainer.LivenessProbe = nil
	feContainer.ReadinessProbe = PortReadyProbe(int(queryPort))

	feContainer.Env = append(feContainer.Env,
		corev1.EnvVar{
			Name:  "RESTORE_CLUSTER_GENERATION",
			Value: strconv.FormatInt(generation, 10),
		},
		corev1.EnvVar{
			Name:  "RESTORE_CLUSTER_SNAPSHOT",
			Value: "true",
		},
	)
	return expectSts
}

// PortReadyProbe detect whether the port is ready
func PortReadyProbe(port int) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(port),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// CheckFEReadyInDisasterRecovery check whether the FE pod is ready.
// When user upgrade the cluster, and the statefulset controller has not begun to update the statefulset, CheckFEReady
// will use the old status to check whether FE is ready.
// CheckFEReadyInDisasterRecovery will check the following things:
//  1. make sure the value of environment RESTORE_CLUSTER_GENERATION equals to the generation of disaster recovery.
//  2. make sure the pod is ready.
func CheckFEReadyInDisasterRecovery(ctx context.Context, k8sClient client.Client,
	clusterNamespace string, clusterName string, generation int64) bool {
	logger := logr.FromContextOrDiscard(ctx)

	podList := corev1.PodList{}
	if err := k8sClient.List(ctx, &podList, client.InNamespace(clusterNamespace),
		client.MatchingLabels{
			v1.ComponentLabelKey: v1.DEFAULT_FE,
			v1.OwnerReference:    load.Name(clusterName, (*v1.StarRocksFeSpec)(nil)),
		}); err != nil {
		logger.Error(err, "list fe pod failed")
		return false
	} else if len(podList.Items) == 0 {
		return false
	}

	for i := range podList.Items {
		pod := &podList.Items[i]

		for j := range pod.Spec.Containers {
			container := &pod.Spec.Containers[j]
			if container.Name != v1.DEFAULT_FE {
				continue
			}
			hasExpectedGeneration := false
			for _, env := range container.Env {
				if env.Value == strconv.FormatInt(generation, 10) {
					hasExpectedGeneration = true
				}
			}
			if !hasExpectedGeneration {
				return false
			}
		}

		if len(pod.Status.ContainerStatuses) == 0 {
			return false
		}
		for j := range pod.Status.ContainerStatuses {
			containerStatus := &pod.Status.ContainerStatuses[j]
			if containerStatus.Name != v1.DEFAULT_FE {
				continue
			}
			if !containerStatus.Ready {
				return false
			}
		}
	}
	return true
}
