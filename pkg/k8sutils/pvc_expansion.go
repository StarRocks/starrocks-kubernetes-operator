package k8sutils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// PVCExpansionResult represents the result of PVC expansion detection
type PVCExpansionResult struct {
	NeedsExpansion             bool
	NeedsStatefulSetRecreation bool
	RequiresDetachment         bool
	OnlyPVCSizeChanged         bool // True if only PVC sizes changed, no other StatefulSet changes
	PVCsToExpand               []PVCExpansionInfo
	ValidationErrors           []string
}

// PVCExpansionInfo contains information about a PVC that needs expansion
type PVCExpansionInfo struct {
	PVCName     string
	Namespace   string
	CurrentSize resource.Quantity
	NewSize     resource.Quantity
	VolumeName  string
}

// DetectPVCExpansion analyzes if PVC expansion is needed for a StatefulSet
func DetectPVCExpansion(ctx context.Context, k8sClient client.Client,
	namespace, statefulSetName string, newStorageVolumes []v1.StorageVolume) (*PVCExpansionResult, error) {

	logger := log.FromContext(ctx).WithName("pvc-expansion")

	result := &PVCExpansionResult{
		PVCsToExpand:     []PVCExpansionInfo{},
		ValidationErrors: []string{},
	}

	// Get the existing StatefulSet
	var sts appv1.StatefulSet
	err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      statefulSetName,
	}, &sts)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// StatefulSet doesn't exist, no expansion needed
			return result, nil
		}
		return nil, fmt.Errorf("failed to get StatefulSet %s: %w", statefulSetName, err)
	}

	// Create maps for easier comparison
	currentVCTs := make(map[string]corev1.PersistentVolumeClaim)
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		currentVCTs[vct.Name] = vct
	}

	newVCTs := make(map[string]corev1.PersistentVolumeClaim)
	for _, sv := range newStorageVolumes {
		if isSpecialStorageClass(sv) || strings.HasPrefix(sv.StorageSize, "0") {
			continue
		}

		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: sv.Name},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				StorageClassName: sv.StorageClassName,
			},
		}
		if sv.StorageSize != "" {
			pvc.Spec.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(sv.StorageSize),
				},
			}
		}
		newVCTs[sv.Name] = pvc
	}

	// Check for size changes
	for volumeName, newVCT := range newVCTs {
		currentVCT, exists := currentVCTs[volumeName]
		if !exists {
			// New volume, will be handled by normal StatefulSet update
			continue
		}

		currentSize := currentVCT.Spec.Resources.Requests[corev1.ResourceStorage]
		newSize := newVCT.Spec.Resources.Requests[corev1.ResourceStorage]

		if newSize.Cmp(currentSize) > 0 {
			// Size increase detected - validate storage class supports expansion
			err := ValidateStorageClassExpansion(ctx, k8sClient, newVCT.Spec.StorageClassName)
			if err != nil {
				result.ValidationErrors = append(result.ValidationErrors,
					fmt.Sprintf("Storage class validation failed for volume %s: %v", volumeName, err))
				continue
			}

			// Check if storage class requires detachment for expansion
			requiresDetachment, err := CheckStorageClassRequiresDetachment(ctx, k8sClient, newVCT.Spec.StorageClassName)
			if err != nil {
				result.ValidationErrors = append(result.ValidationErrors,
					fmt.Sprintf("Failed to check detachment requirement for volume %s: %v", volumeName, err))
				continue
			}

			if requiresDetachment {
				result.RequiresDetachment = true
			}

			result.NeedsExpansion = true

			// Check existing PVCs for this volume
			pvcs, err := getStatefulSetPVCs(ctx, k8sClient, namespace, statefulSetName, volumeName)
			if err != nil {
				return nil, fmt.Errorf("failed to get PVCs for volume %s: %w", volumeName, err)
			}

			for _, pvc := range pvcs {
				result.PVCsToExpand = append(result.PVCsToExpand, PVCExpansionInfo{
					PVCName:     pvc.Name,
					Namespace:   pvc.Namespace,
					CurrentSize: pvc.Spec.Resources.Requests[corev1.ResourceStorage],
					NewSize:     newSize,
					VolumeName:  volumeName,
				})
			}
		} else if newSize.Cmp(currentSize) < 0 {
			// Size decrease - not allowed
			result.ValidationErrors = append(result.ValidationErrors,
				fmt.Sprintf("Storage size reduction is not allowed for volume %s: %s -> %s",
					volumeName, currentSize.String(), newSize.String()))
		}
	}

	// Check if StatefulSet recreation is needed (for other VCT changes)
	hasNonSizeChanges := hasVolumeClaimTemplateChanges(currentVCTs, newVCTs)
	if hasNonSizeChanges {
		result.NeedsStatefulSetRecreation = true
	}

	// Determine if only PVC sizes changed (important for minimizing disruption)
	if result.NeedsExpansion && !hasNonSizeChanges {
		result.OnlyPVCSizeChanged = true
	}

	logger.Info("PVC expansion detection completed",
		"needsExpansion", result.NeedsExpansion,
		"needsStatefulSetRecreation", result.NeedsStatefulSetRecreation,
		"pvcCount", len(result.PVCsToExpand),
		"validationErrors", len(result.ValidationErrors))

	return result, nil
}

// ExpandPVCs performs the actual PVC expansion
func ExpandPVCs(ctx context.Context, k8sClient client.Client, expansionInfos []PVCExpansionInfo) error {
	logger := log.FromContext(ctx).WithName("pvc-expansion")

	for _, info := range expansionInfos {
		logger.Info("Expanding PVC", "pvc", info.PVCName, "namespace", info.Namespace,
			"currentSize", info.CurrentSize.String(), "newSize", info.NewSize.String())

		var pvc corev1.PersistentVolumeClaim
		err := k8sClient.Get(ctx, types.NamespacedName{
			Namespace: info.Namespace,
			Name:      info.PVCName,
		}, &pvc)
		if err != nil {
			return fmt.Errorf("failed to get PVC %s: %w", info.PVCName, err)
		}

		// Update the PVC size
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = info.NewSize

		err = k8sClient.Update(ctx, &pvc)
		if err != nil {
			return fmt.Errorf("failed to update PVC %s: %w", info.PVCName, err)
		}

		logger.Info("Successfully expanded PVC", "pvc", info.PVCName)
	}

	return nil
}

// ExpandPVCsWithDetachment performs PVC expansion that requires pod detachment
func ExpandPVCsWithDetachment(ctx context.Context, k8sClient client.Client, expect *appv1.StatefulSet,
	expansionInfos []PVCExpansionInfo, enableScaleTo1 bool) error {
	logger := log.FromContext(ctx).WithName("pvc-expansion-detachment")

	if len(expansionInfos) == 0 {
		return nil
	}

	// Get the current StatefulSet
	var current appv1.StatefulSet
	err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: expect.Namespace,
		Name:      expect.Name,
	}, &current)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// StatefulSet doesn't exist, create it with new sizes
			return CreateClientObject(ctx, k8sClient, expect)
		}
		return fmt.Errorf("failed to get current StatefulSet: %w", err)
	}

	// Store current replicas to restore later
	currentReplicas := current.Spec.Replicas

	logger.Info("Starting PVC expansion with detachment",
		"statefulset", expect.Name,
		"namespace", expect.Namespace,
		"pvcCount", len(expansionInfos))

	// Step 1: Delete StatefulSet immediately to detach all PVCs
	// This is necessary because VolumeClaimTemplates are immutable and must be updated
	logger.Info("Deleting StatefulSet to detach PVCs and allow VolumeClaimTemplate updates")
	err = k8sClient.Delete(ctx, &current)
	if err != nil {
		return fmt.Errorf("failed to delete StatefulSet: %w", err)
	}

	// Step 2: Wait for StatefulSet to be fully deleted and all pods terminated
	err = waitForStatefulSetDeletion(ctx, k8sClient, expect.Namespace, expect.Name)
	if err != nil {
		return fmt.Errorf("failed to wait for StatefulSet deletion: %w", err)
	}

	// Step 3: Wait for PVCs to be fully detached
	err = waitForPVCsDetached(ctx, k8sClient, expansionInfos)
	if err != nil {
		return fmt.Errorf("failed to wait for PVCs to be detached: %w", err)
	}

	// Step 4: Expand the PVCs while they are detached
	logger.Info("Expanding PVCs while detached")
	err = ExpandPVCs(ctx, k8sClient, expansionInfos)
	if err != nil {
		return fmt.Errorf("failed to expand detached PVCs: %w", err)
	}

	// Step 5: Wait for PVC expansion to complete
	err = waitForPVCExpansionComplete(ctx, k8sClient, expansionInfos)
	if err != nil {
		return fmt.Errorf("failed to wait for PVC expansion completion: %w", err)
	}

	// Step 6: Restore the original replica count and create new StatefulSet with updated VolumeClaimTemplates
	expect.Spec.Replicas = currentReplicas
	logger.Info("Creating new StatefulSet with expanded PVC templates")
	err = CreateClientObject(ctx, k8sClient, expect)
	if err != nil {
		return fmt.Errorf("failed to create new StatefulSet: %w", err)
	}

	logger.Info("PVC expansion with detachment completed successfully")
	return nil
}

// getStatefulSetPVCs returns all PVCs for a specific volume in a StatefulSet
func getStatefulSetPVCs(ctx context.Context, k8sClient client.Client,
	namespace, statefulSetName, volumeName string) ([]corev1.PersistentVolumeClaim, error) {

	var pvcList corev1.PersistentVolumeClaimList
	err := k8sClient.List(ctx, &pvcList, client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}

	var result []corev1.PersistentVolumeClaim
	for _, pvc := range pvcList.Items {
		// PVC names follow pattern: <volumeName>-<statefulSetName>-<ordinal>
		if strings.HasPrefix(pvc.Name, volumeName+"-"+statefulSetName+"-") {
			result = append(result, pvc)
		}
	}

	// Sort by ordinal for consistent processing
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// hasVolumeClaimTemplateChanges checks if there are non-size changes in VCTs
func hasVolumeClaimTemplateChanges(current, new map[string]corev1.PersistentVolumeClaim) bool {
	if len(current) != len(new) {
		return true
	}

	for name, newVCT := range new {
		currentVCT, exists := current[name]
		if !exists {
			return true
		}

		// Check for changes other than storage size
		if !equalStorageClass(currentVCT.Spec.StorageClassName, newVCT.Spec.StorageClassName) {
			return true
		}

		if !equalAccessModes(currentVCT.Spec.AccessModes, newVCT.Spec.AccessModes) {
			return true
		}

		// Check for other resource requirements (excluding storage)
		currentResources := currentVCT.Spec.Resources.DeepCopy()
		newResources := newVCT.Spec.Resources.DeepCopy()

		// Remove storage from comparison since we handle that separately
		if currentResources.Requests != nil {
			delete(currentResources.Requests, corev1.ResourceStorage)
		}
		if newResources.Requests != nil {
			delete(newResources.Requests, corev1.ResourceStorage)
		}
		if currentResources.Limits != nil {
			delete(currentResources.Limits, corev1.ResourceStorage)
		}
		if newResources.Limits != nil {
			delete(newResources.Limits, corev1.ResourceStorage)
		}

		// Compare non-storage resources
		if !equalResourceRequirements(*currentResources, *newResources) {
			return true
		}
	}

	return false
}

// isSpecialStorageClass checks if the storage volume uses special storage classes
func isSpecialStorageClass(sv v1.StorageVolume) bool {
	if sv.StorageClassName == nil {
		return false
	}
	return *sv.StorageClassName == "emptyDir" || *sv.StorageClassName == "hostPath"
}

// Helper functions for comparison
func equalStorageClass(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalAccessModes(a, b []corev1.PersistentVolumeAccessMode) bool {
	if len(a) != len(b) {
		return false
	}
	for i, mode := range a {
		if mode != b[i] {
			return false
		}
	}
	return true
}

func equalResourceRequirements(a, b corev1.ResourceRequirements) bool {
	return equalResourceList(a.Requests, b.Requests) && equalResourceList(a.Limits, b.Limits)
}

func equalResourceList(a, b corev1.ResourceList) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists || !valueA.Equal(valueB) {
			return false
		}
	}
	return true
}

// ValidateStorageClassExpansion checks if a storage class supports volume expansion
func ValidateStorageClassExpansion(ctx context.Context, k8sClient client.Client, storageClassName *string) error {
	if storageClassName == nil {
		// Use default storage class
		return validateDefaultStorageClassExpansion(ctx, k8sClient)
	}

	// Check if it's a special storage class that doesn't need expansion support
	if *storageClassName == "emptyDir" || *storageClassName == "hostPath" {
		return nil
	}

	var sc storagev1.StorageClass
	err := k8sClient.Get(ctx, types.NamespacedName{Name: *storageClassName}, &sc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("storage class %s not found", *storageClassName)
		}
		return fmt.Errorf("failed to get storage class %s: %w", *storageClassName, err)
	}

	if sc.AllowVolumeExpansion == nil || !*sc.AllowVolumeExpansion {
		return fmt.Errorf("storage class %s does not support volume expansion (allowVolumeExpansion is not set to true)", *storageClassName)
	}

	return nil
}

// validateDefaultStorageClassExpansion checks if the default storage class supports expansion
func validateDefaultStorageClassExpansion(ctx context.Context, k8sClient client.Client) error {
	var scList storagev1.StorageClassList
	err := k8sClient.List(ctx, &scList)
	if err != nil {
		return fmt.Errorf("failed to list storage classes: %w", err)
	}

	var defaultSC *storagev1.StorageClass
	for i := range scList.Items {
		sc := &scList.Items[i]
		if sc.Annotations != nil {
			if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" ||
				sc.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
				defaultSC = sc
				break
			}
		}
	}

	if defaultSC == nil {
		return fmt.Errorf("no default storage class found")
	}

	if defaultSC.AllowVolumeExpansion == nil || !*defaultSC.AllowVolumeExpansion {
		return fmt.Errorf("default storage class %s does not support volume expansion (allowVolumeExpansion is not set to true)", defaultSC.Name)
	}

	return nil
}

// CheckStorageClassRequiresDetachment determines if a storage class requires PVC detachment for expansion
func CheckStorageClassRequiresDetachment(ctx context.Context, k8sClient client.Client, storageClassName *string) (bool, error) {
	if storageClassName == nil {
		// Check default storage class
		return checkDefaultStorageClassRequiresDetachment(ctx, k8sClient)
	}

	// Special storage classes that don't require detachment
	if *storageClassName == "emptyDir" || *storageClassName == "hostPath" {
		return false, nil
	}

	var sc storagev1.StorageClass
	err := k8sClient.Get(ctx, types.NamespacedName{Name: *storageClassName}, &sc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, fmt.Errorf("storage class %s not found", *storageClassName)
		}
		return false, fmt.Errorf("failed to get storage class %s: %w", *storageClassName, err)
	}

	return requiresDetachmentForExpansion(&sc), nil
}

// checkDefaultStorageClassRequiresDetachment checks if the default storage class requires detachment
func checkDefaultStorageClassRequiresDetachment(ctx context.Context, k8sClient client.Client) (bool, error) {
	var scList storagev1.StorageClassList
	err := k8sClient.List(ctx, &scList)
	if err != nil {
		return false, fmt.Errorf("failed to list storage classes: %w", err)
	}

	var defaultSC *storagev1.StorageClass
	for i := range scList.Items {
		sc := &scList.Items[i]
		if sc.Annotations != nil {
			if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" ||
				sc.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
				defaultSC = sc
				break
			}
		}
	}

	if defaultSC == nil {
		return false, fmt.Errorf("no default storage class found")
	}

	return requiresDetachmentForExpansion(defaultSC), nil
}

// requiresDetachmentForExpansion determines if a storage class requires detachment based on provisioner
func requiresDetachmentForExpansion(sc *storagev1.StorageClass) bool {
	// Known provisioners that support online expansion (safe to expand without detachment)
	onlineExpansionSupportedProvisioners := map[string]bool{
		// Google Cloud Persistent Disk
		"kubernetes.io/gce-pd":  true,
		"pd.csi.storage.gke.io": true,
		// AWS EBS
		"kubernetes.io/aws-ebs": true,
		"ebs.csi.aws.com":       true,
		// DigitalOcean Block Storage
		"dobs.csi.digitalocean.com": true,
		// Linode Block Storage
		"linodebs.csi.linode.com": true,
		// OpenEBS (local storage)
		"openebs.io/local": true,
		// Longhorn (distributed storage)
		"driver.longhorn.io": true,
		// Add more known-safe provisioners as needed
	}

	// Check for explicit online expansion mode in parameters
	if sc.Parameters != nil {
		if expansionMode, exists := sc.Parameters["expansion-mode"]; exists {
			if expansionMode == "online" {
				return false // Explicitly supports online expansion
			}
			if expansionMode == "offline" || expansionMode == "detached" {
				return true // Explicitly requires detachment
			}
		}
	}

	// Check if provisioner is known to support online expansion
	if onlineExpansionSupportedProvisioners[sc.Provisioner] {
		return false
	}

	// CONSERVATIVE DEFAULT: If we don't know for sure that the provisioner supports
	// online expansion, default to offline expansion (detachment required) for safety.
	// This prevents potential data corruption or expansion failures.
	// Better to have temporary downtime than risk data loss.
	return true
}

// waitForPVCsDetached waits for PVCs to be fully detached from pods
func waitForPVCsDetached(ctx context.Context, k8sClient client.Client, expansionInfos []PVCExpansionInfo) error {
	logger := log.FromContext(ctx).WithName("pvc-detachment-wait")

	for i := 0; i < 60; i++ { // Wait up to 5 minutes
		allDetached := true

		for _, info := range expansionInfos {
			var pvc corev1.PersistentVolumeClaim
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: info.Namespace,
				Name:      info.PVCName,
			}, &pvc)
			if err != nil {
				return fmt.Errorf("failed to get PVC %s: %w", info.PVCName, err)
			}

			// Check if PVC is still attached to any pod
			if isPVCAttachedToPod(ctx, k8sClient, &pvc) {
				allDetached = false
				break
			}
		}

		if allDetached {
			logger.Info("All PVCs are detached")
			return nil
		}

		logger.V(1).Info("Waiting for PVCs to be detached")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue waiting
		}
	}

	return fmt.Errorf("timeout waiting for PVCs to be detached")
}

// waitForPVCExpansionComplete waits for PVC expansion to complete
func waitForPVCExpansionComplete(ctx context.Context, k8sClient client.Client, expansionInfos []PVCExpansionInfo) error {
	logger := log.FromContext(ctx).WithName("pvc-expansion-wait")

	for i := 0; i < 120; i++ { // Wait up to 10 minutes for expansion
		allExpanded := true

		for _, info := range expansionInfos {
			var pvc corev1.PersistentVolumeClaim
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: info.Namespace,
				Name:      info.PVCName,
			}, &pvc)
			if err != nil {
				return fmt.Errorf("failed to get PVC %s: %w", info.PVCName, err)
			}

			// Check if PVC has been expanded to the desired size
			currentSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			if !currentSize.Equal(info.NewSize) {
				allExpanded = false
				break
			}

			// Check PVC status for expansion completion
			if pvc.Status.Phase != corev1.ClaimBound {
				allExpanded = false
				break
			}

			// Check for expansion-related conditions
			if hasExpansionInProgress(&pvc) {
				allExpanded = false
				break
			}
		}

		if allExpanded {
			logger.Info("All PVC expansions completed")
			return nil
		}

		logger.V(1).Info("Waiting for PVC expansion to complete")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue waiting
		}
	}

	return fmt.Errorf("timeout waiting for PVC expansion to complete")
}

// isPVCAttachedToPod checks if a PVC is currently attached to any pod
func isPVCAttachedToPod(ctx context.Context, k8sClient client.Client, pvc *corev1.PersistentVolumeClaim) bool {
	// List all pods in the namespace
	var podList corev1.PodList
	err := k8sClient.List(ctx, &podList, client.InNamespace(pvc.Namespace))
	if err != nil {
		// If we can't list pods, assume PVC is attached to be safe
		return true
	}

	// Check if any pod is using this PVC
	for _, pod := range podList.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil &&
				volume.PersistentVolumeClaim.ClaimName == pvc.Name {
				// PVC is referenced by a pod
				if pod.DeletionTimestamp == nil {
					// Pod is not being deleted, so PVC is still attached
					return true
				}
			}
		}
	}

	return false
}

// hasExpansionInProgress checks if a PVC has expansion in progress
func hasExpansionInProgress(pvc *corev1.PersistentVolumeClaim) bool {
	// Check PVC conditions for expansion-related status
	for _, condition := range pvc.Status.Conditions {
		switch condition.Type {
		case corev1.PersistentVolumeClaimResizing:
			return condition.Status == corev1.ConditionTrue
		case corev1.PersistentVolumeClaimFileSystemResizePending:
			return condition.Status == corev1.ConditionTrue
		}
	}

	// Check if the actual size matches the requested size
	if pvc.Status.Capacity != nil {
		requestedSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		actualSize := pvc.Status.Capacity[corev1.ResourceStorage]
		return !requestedSize.Equal(actualSize)
	}

	return false
}
