package fe

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
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
	src *v1.StarRocksCluster, sts *appsv1.StatefulSet, queryPort int32) error {
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

			// Extract and store cluster state information for future pod restarts
			if drStatus.ClusterUUID == "" {
				clusterUUID := extractClusterStateFromConfigMaps(ctx, k8sClient, src.Namespace, feSpec.ConfigMaps)
				if clusterUUID != "" {
					drStatus.ClusterUUID = clusterUUID
					logger.Info("extracted cluster UUID for state preservation", "clusterUUID", clusterUUID)
				}
			}
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

func rewriteStatefulSetForDisasterRecovery(expectSts *appsv1.StatefulSet, generation int64, queryPort int32) *appsv1.StatefulSet {
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

// ShouldPreserveClusterState determines if cluster state should be preserved in StatefulSet
// This returns true if either disaster recovery is active OR disaster recovery has completed
// and we need to preserve the recovered cluster identity to prevent UUID regeneration
func ShouldPreserveClusterState(drSpec *v1.DisasterRecovery, drStatus *v1.DisasterRecoveryStatus,
	feConfig map[string]interface{}) (bool, int32) {
	// Check if we should enter DR mode (existing logic)
	shouldEnter, queryPort := ShouldEnterDisasterRecoveryMode(drSpec, drStatus, feConfig)
	if shouldEnter {
		return true, queryPort
	}

	// Check if disaster recovery has completed and we have cluster state to preserve
	if drSpec != nil && drSpec.Enabled && drStatus != nil &&
		drStatus.Phase == v1.DRPhaseDone && drStatus.ClusterUUID != "" &&
		IsRunInSharedDataMode(feConfig) {
		return true, rutils.GetPort(feConfig, rutils.QUERY_PORT)
	}

	return false, 0
}

// ExtractClusterUUIDFromSnapshot attempts to extract cluster UUID from cluster snapshot path
// The snapshot path format is typically: s3://bucket/path/<cluster-uuid>/meta/image/snapshot_name
func ExtractClusterUUIDFromSnapshot(snapshotPath string) string {
	if snapshotPath == "" {
		return ""
	}

	// Split by "/" and find the UUID part (should be before /meta/image/)
	parts := strings.Split(snapshotPath, "/")
	for i, part := range parts {
		// Look for the part that comes before "meta"
		if i < len(parts)-2 && parts[i+1] == "meta" && parts[i+2] == "image" {
			// Basic UUID validation (36 characters with dashes in right positions)
			if len(part) == 36 && strings.Count(part, "-") == 4 {
				return part
			}
		}
	}
	return ""
}

// extractClusterStateFromConfigMaps extracts cluster UUID from the cluster_snapshot.yaml ConfigMap
func extractClusterStateFromConfigMaps(ctx context.Context, k8sClient client.Client,
	namespace string, configMaps []v1.ConfigMapReference) string {
	logger := logr.FromContextOrDiscard(ctx)

	// Find the cluster_snapshot.yaml ConfigMap
	for _, cmRef := range configMaps {
		if strings.Contains(cmRef.SubPath, "cluster_snapshot.yaml") {
			// Read the ConfigMap
			cm := &corev1.ConfigMap{}
			err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      cmRef.Name,
				Namespace: namespace,
			}, cm)
			if err != nil {
				logger.Error(err, "failed to read cluster snapshot ConfigMap", "configMap", cmRef.Name)
				continue
			}

			// Parse the cluster_snapshot.yaml content
			if snapshotYaml, exists := cm.Data["cluster_snapshot.yaml"]; exists {
				clusterUUID := extractClusterUUIDFromSnapshotYaml(snapshotYaml)
				if clusterUUID != "" {
					return clusterUUID
				}
			}
		}
	}
	return ""
}

// extractClusterUUIDFromSnapshotYaml parses cluster_snapshot.yaml content to extract cluster UUID
func extractClusterUUIDFromSnapshotYaml(yamlContent string) string {
	// Look for cluster_snapshot_path line in the YAML
	lines := strings.Split(yamlContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "cluster_snapshot_path:") {
			// Extract the path value
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				snapshotPath := strings.TrimSpace(parts[1])
				return ExtractClusterUUIDFromSnapshot(snapshotPath)
			}
		}
	}
	return ""
}

// RewriteStatefulSetForClusterStatePreservation modifies StatefulSet to preserve cluster state
// This is used both during active disaster recovery and after recovery completion
func RewriteStatefulSetForClusterStatePreservation(expectSts *appsv1.StatefulSet,
	drSpec *v1.DisasterRecovery, drStatus *v1.DisasterRecoveryStatus, queryPort int32) {

	podTemplate := &expectSts.Spec.Template
	feContainer := &(podTemplate.Spec.Containers[0])

	if drStatus != nil && drStatus.Phase == v1.DRPhaseDone && drStatus.ClusterUUID != "" {
		// Post-disaster recovery: preserve cluster state without active recovery
		feContainer.Env = append(feContainer.Env,
			corev1.EnvVar{
				Name:  "RECOVERED_CLUSTER_UUID",
				Value: drStatus.ClusterUUID,
			},
			corev1.EnvVar{
				Name:  "USE_RECOVERED_CLUSTER_STATE",
				Value: "true",
			},
		)
		// Use normal probes for post-recovery operation
		if feContainer.ReadinessProbe == nil {
			feContainer.ReadinessProbe = PortReadyProbe(int(queryPort))
		}
	} else {
		// Active disaster recovery: use existing behavior
		rewriteStatefulSetForDisasterRecovery(expectSts, drSpec.Generation, queryPort)
	}
}
