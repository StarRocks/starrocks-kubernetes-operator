// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource_utils

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

// hashStatefulsetObject contains the info for hash comparison.
type hashStatefulsetObject struct {
	name                 string
	namespace            string
	labels               map[string]string
	finalizers           []string
	selector             metav1.LabelSelector
	podTemplate          corev1.PodTemplateSpec
	serviceName          string
	volumeClaimTemplates []corev1.PersistentVolumeClaim
	replicas             int32
}

// StatefulsetHashObject construct the hash spec for deep equals to exist statefulset.
func statefulSetHashObject(sts *appv1.StatefulSet, excludeReplica bool) hashStatefulsetObject {
	// set -1 for the initial is zero.
	replicas := int32(-1)
	if !excludeReplica {
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}
	}
	selector := metav1.LabelSelector{}
	if sts.Spec.Selector != nil {
		selector = *sts.Spec.Selector
	}

	return hashStatefulsetObject{
		name:                 sts.Name,
		namespace:            sts.Namespace,
		labels:               sts.Labels,
		finalizers:           sts.Finalizers,
		selector:             selector,
		podTemplate:          sts.Spec.Template,
		serviceName:          sts.Spec.ServiceName,
		volumeClaimTemplates: sts.Spec.VolumeClaimTemplates,
		replicas:             replicas,
	}
}

// StatefulSetDeepEqual judge two statefulset equal or not.
func StatefulSetDeepEqual(new *appv1.StatefulSet, old *appv1.StatefulSet, excludeReplicas bool) bool {
	var newHashv, oldHashv string

	newHso := statefulSetHashObject(new, excludeReplicas)
	if _, ok := new.Annotations[srapi.ComponentResourceHash]; ok {
		newHashv = new.Annotations[srapi.ComponentResourceHash]
	} else {
		newHashv = hash.HashObject(newHso)
	}

	// the hash value calculated from statefulset instance in k8s may will never equal to the hash value from
	// starrocks cluster. Because statefulset may be updated by k8s controller manager.
	// Every time you update the statefulset, a new reconcile will be triggered.
	if _, ok := old.Annotations[srapi.ComponentResourceHash]; ok {
		oldHashv = old.Annotations[srapi.ComponentResourceHash]
	} else {
		oldHso := statefulSetHashObject(old, excludeReplicas)
		oldHashv = hash.HashObject(oldHso)
	}

	anno := Annotations{}
	anno.AddAnnotation(new.Annotations)
	anno.Add(srapi.ComponentResourceHash, newHashv)
	new.Annotations = anno

	// avoid the update from kubectl.
	return newHashv == oldHashv &&
		new.Namespace == old.Namespace /* &&
		oldGeneration == old.Generation*/
}

// MergeStatefulSets merge exist statefulset and new statefulset.
func MergeStatefulSets(new *appv1.StatefulSet, old appv1.StatefulSet) {
	MergeMetadata(&new.ObjectMeta, old.ObjectMeta)
}
