package resource_utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type Labels map[string]string

func New(labels ...Labels) Labels {
	nlabels := Labels{}
	for _, l := range labels {
		for k, v := range l {
			nlabels[k] = v
		}
	}

	return nlabels
}

func (l Labels) Add(key, value string) {
	l[key] = value
}

func (l Labels) AddLabel(label Labels) {
	for k, v := range label {
		l[k] = v
	}
}

// mergeMetadata takes labels and annotations from the old resource and merges
// them into the new resource. If a key is present in both resources, the new
// resource wins. It also copies the ResourceVersion from the old resource to
// the new resource to prevent update conflicts.
func MergeMetadata(new *metav1.ObjectMeta, old metav1.ObjectMeta) {
	new.ResourceVersion = old.ResourceVersion
	new.SetFinalizers(MergeSlices(new.Finalizers, old.Finalizers))
	new.SetLabels(mergeMaps(new.Labels, old.Labels))
	new.SetAnnotations(mergeMaps(new.Annotations, old.Annotations))
	mergeOwnerReferences(new.OwnerReferences, old.OwnerReferences)
}

func MergeSlices(new []string, old []string) []string {
	var set map[string]bool
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

func mergeOwnerReferences(old []metav1.OwnerReference, new []metav1.OwnerReference) []metav1.OwnerReference {
	existing := make(map[metav1.OwnerReference]bool)
	for _, ownerRef := range old {
		existing[ownerRef] = true
	}
	for _, ownerRef := range new {
		if _, ok := existing[ownerRef]; !ok {
			old = append(old, ownerRef)
		}
	}
	return old
}
