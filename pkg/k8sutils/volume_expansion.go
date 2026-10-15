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

package k8sutils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// expandVolumeClaims grows the claims of a grown storageSize, then orphan-deletes the StatefulSet to recreate it.
func expandVolumeClaims(ctx context.Context, k8sClient client.Client, expect, actual *appsv1.StatefulSet) (bool, error) {
	grown, err := grownClaimTemplates(expect.Spec.VolumeClaimTemplates, actual.Spec.VolumeClaimTemplates)
	if err != nil || len(grown) == 0 {
		return false, err
	}

	// the recreated StatefulSet is built from expect, so it must keep what the pods depend on
	if !equality.Semantic.DeepEqual(expect.Spec.Selector, actual.Spec.Selector) ||
		expect.Spec.ServiceName != actual.Spec.ServiceName ||
		expect.Spec.PodManagementPolicy != actual.Spec.PodManagementPolicy {
		return true, fmt.Errorf("can not expand the volumes of statefulset %s: recreating it would change "+
			"its selector, serviceName or podManagementPolicy, which were set by an older operator version", actual.Name)
	}

	logger := logr.FromContextOrDiscard(ctx)
	var pvcs corev1.PersistentVolumeClaimList
	if err := k8sClient.List(ctx, &pvcs, client.InNamespace(actual.Namespace),
		client.MatchingLabels(actual.Spec.Selector.MatchLabels)); err != nil {
		return true, err
	}
	for i := range pvcs.Items {
		pvc := &pvcs.Items[i]
		size, ok := grown[claimTemplateName(pvc.Name, actual.Name, grown)]
		if !ok || pvc.Spec.Resources.Requests.Storage().Cmp(size) >= 0 {
			continue
		}
		patch := client.MergeFrom(pvc.DeepCopy())
		if pvc.Spec.Resources.Requests == nil {
			pvc.Spec.Resources.Requests = corev1.ResourceList{}
		}
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = size
		logger.Info("expand persistent volume claim", "name", pvc.Name, "size", size.String())
		// the API server refuses this unless the claim's StorageClass sets allowVolumeExpansion
		if err := k8sClient.Patch(ctx, pvc, patch); err != nil {
			return true, fmt.Errorf("expand persistent volume claim %s: %w", pvc.Name, err)
		}
	}

	logger.Info("delete statefulset with orphan propagation to update its volumeClaimTemplates", "name", actual.Name)
	return true, k8sClient.Delete(ctx, actual, client.PropagationPolicy(metav1.DeletePropagationOrphan),
		client.Preconditions{UID: &actual.UID, ResourceVersion: &actual.ResourceVersion})
}

// grownClaimTemplates returns the new size of each grown template; nothing if the templates differ in anything else.
func grownClaimTemplates(expect, actual []corev1.PersistentVolumeClaim) (map[string]resource.Quantity, error) {
	if len(expect) != len(actual) {
		return nil, nil
	}
	grown := map[string]resource.Quantity{}
	for i := range expect {
		e, a := &expect[i], &actual[i]
		if e.Name != a.Name || !equality.Semantic.DeepEqual(e.Spec.StorageClassName, a.Spec.StorageClassName) {
			return nil, nil
		}
		want, have := e.Spec.Resources.Requests.Storage(), a.Spec.Resources.Requests.Storage()
		switch want.Cmp(*have) {
		case 1:
			grown[e.Name] = want.DeepCopy()
		case -1:
			return nil, fmt.Errorf("storageSize of volume %s can not be decreased from %s to %s",
				e.Name, have.String(), want.String())
		}
	}
	return grown, nil
}

// claimTemplateName returns the template a claim was created from, named <template>-<statefulset>-<ordinal>.
func claimTemplateName(claimName, stsName string, templates map[string]resource.Quantity) string {
	for name := range templates {
		ordinal, found := strings.CutPrefix(claimName, name+"-"+stsName+"-")
		if _, err := strconv.Atoi(ordinal); found && err == nil {
			return name
		}
	}
	return ""
}
