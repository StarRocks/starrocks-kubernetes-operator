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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

// hashStatefulsetObject contains the info for hash comparison.
type hashStatefulsetObject struct {
	name                                 string
	namespace                            string
	labels                               map[string]string
	finalizers                           []string
	selector                             metav1.LabelSelector
	podTemplate                          corev1.PodTemplateSpec
	serviceName                          string
	volumeClaimTemplates                 []corev1.PersistentVolumeClaim
	replicas                             int32
	updateStrategy                       appsv1.StatefulSetUpdateStrategy
	persistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy
}

// statefulSetHashObject construct the hash spec for deep equals to exist statefulset.
func statefulSetHashObject(sts *appsv1.StatefulSet) hashStatefulsetObject {
	replicas := int32(-1)
	if sts.Spec.Replicas != nil {
		replicas = *sts.Spec.Replicas
	}

	selector := metav1.LabelSelector{}
	if sts.Spec.Selector != nil {
		selector = *sts.Spec.Selector
	}

	return hashStatefulsetObject{
		name:                                 sts.Name,
		namespace:                            sts.Namespace,
		labels:                               sts.Labels,
		finalizers:                           sts.Finalizers,
		selector:                             selector,
		podTemplate:                          sts.Spec.Template,
		serviceName:                          sts.Spec.ServiceName,
		volumeClaimTemplates:                 sts.Spec.VolumeClaimTemplates,
		replicas:                             replicas,
		updateStrategy:                       sts.Spec.UpdateStrategy,
		persistentVolumeClaimRetentionPolicy: sts.Spec.PersistentVolumeClaimRetentionPolicy,
	}
}

// StatefulSetDeepEqual judge two statefulset equal or not, and it will not change the statefulset instance.
// This function will always return a new hash value of the expected statefulset
func StatefulSetDeepEqual(expect *appsv1.StatefulSet, actual *appsv1.StatefulSet) (string, bool) {
	var newHashv, oldHashv string

	newHso := statefulSetHashObject(expect)
	if _, ok := expect.Annotations[srapi.ComponentResourceHash]; ok {
		newHashv = expect.Annotations[srapi.ComponentResourceHash]
	} else {
		newHashv = hash.HashObject(newHso)
	}

	// The hash value calculated from a statefulset instance in k8s may never equal to the hash value from
	// the starrocks cluster. Because statefulset may be updated by k8s controller manager.
	// Every time you update the statefulset, a new reconciling will be triggered.
	if _, ok := actual.Annotations[srapi.ComponentResourceHash]; ok {
		oldHashv = actual.Annotations[srapi.ComponentResourceHash]
	} else {
		oldHso := statefulSetHashObject(actual)
		oldHashv = hash.HashObject(oldHso)
	}

	return newHashv, newHashv == oldHashv && expect.Namespace == actual.Namespace
}

// MergeStatefulSets merge exist statefulset and new statefulset.
func MergeStatefulSets(new *appsv1.StatefulSet, old appsv1.StatefulSet) {
	MergeMetadata(&new.ObjectMeta, old.ObjectMeta)
}
