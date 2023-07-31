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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Labels map[string]string

func (l Labels) Add(key, value string) {
	l[key] = value
}

func (l Labels) AddLabel(label Labels) {
	if label == nil {
		return
	}

	for k, v := range label {
		l[k] = v
	}
}

type Annotations map[string]string

func (a Annotations) Add(key, value string) {
	a[key] = value
}

func (a Annotations) AddAnnotation(annotation Annotations) {
	for k, v := range annotation {
		a[k] = v
	}
}

// MergeMetadata takes labels and annotations from the old resource and merges
// them into the new resource. If a key is present in both resources, the new
// resource wins. It also copies the ResourceVersion from the old resource to
// the new resource to prevent update conflicts.
func MergeMetadata(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.ResourceVersion = old.ResourceVersion
	new.SetFinalizers(MergeSlices(new.Finalizers, old.Finalizers))
	new.SetLabels(mergeMaps(new.Labels, old.Labels))
	new.SetAnnotations(mergeMaps(new.Annotations, old.Annotations))
	new.OwnerReferences = mergeOwnerReferences(new.OwnerReferences, old.OwnerReferences)
}

func MergeSlices(new []string, old []string) []string {
	set := make(map[string]bool)
	for _, s := range new {
		set[s] = true
	}
	for _, os := range old {
		if _, ok := set[os]; ok {
			continue
		}
		new = append(new, os)
	}
	return new
}

func mergeMaps(new map[string]string, old map[string]string) map[string]string {
	return mergeMapsByPrefix(new, old, "")
}

func mergeMapsByPrefix(from map[string]string, to map[string]string, prefix string) map[string]string {
	if to == nil {
		to = make(map[string]string)
	}

	if from == nil {
		from = make(map[string]string)
	}

	for k, v := range from {
		if strings.HasPrefix(k, prefix) {
			to[k] = v
		}
	}

	return to
}

func mergeOwnerReferences(new []metav1.OwnerReference, old []metav1.OwnerReference) []metav1.OwnerReference {
	var references []metav1.OwnerReference
	existing := make(map[metav1.OwnerReference]bool)
	for _, ownerRef := range new {
		existing[ownerRef] = true
		references = append(references, ownerRef)
	}
	for _, ownerRef := range old {
		if _, ok := existing[ownerRef]; !ok {
			references = append(references, ownerRef)
		}
	}
	return references
}

func GetInt32Pointer(v int32) *int32 {
	return &v
}

func GetStringPointer(s string) *string {
	return &s
}

func GetInt64ptr(n int64) *int64 {
	return &n
}
