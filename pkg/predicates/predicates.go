package predicates

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// ignoredAnnotation is the annotation key that marks an object as ignored by the operator
	ignoredAnnotation = "starrocks.com/ignored"
)

// GenericPredicates implements predicate.Predicate for filtering events
type GenericPredicates struct {
	predicate.Funcs
}

// Create returns true if the Create event should be processed
func (GenericPredicates) Create(e event.CreateEvent) bool {
	return shouldReconcile(e.Object)
}

// Update returns true if the Update event should be processed
func (GenericPredicates) Update(e event.UpdateEvent) bool {
	return shouldReconcile(e.ObjectNew)
}

// Delete returns true if the Delete event should be processed
func (GenericPredicates) Delete(e event.DeleteEvent) bool {
	return shouldReconcile(e.Object)
}

// Generic returns true if the Generic event should be processed
func (GenericPredicates) Generic(e event.GenericEvent) bool {
	return shouldReconcile(e.Object)
}

// shouldReconcile checks if an object should be reconciled based on namespace and annotation filters
func shouldReconcile(obj client.Object) bool {
	if obj == nil {
		return false
	}

	// Check namespace deny list
	if !ignoreNamespacePredicate(obj) {
		return false
	}

	// Check ignored annotation
	if !ignoreIgnoredObjectPredicate(obj) {
		return false
	}

	return true
}

// ignoreNamespacePredicate returns false if the object is in a denied namespace
func ignoreNamespacePredicate(obj client.Object) bool {
	namespaces := getEnvAsSlice("DENY_LIST", nil, ",")
	logger := log.Log.WithName("predicates")

	for _, namespace := range namespaces {
		if obj.GetNamespace() == namespace {
			logger.Info("starrocks operator will not reconcile namespace, alter DENY_LIST to reconcile",
				"namespace", obj.GetNamespace())
			return false
		}
	}
	return true
}

// ignoreIgnoredObjectPredicate returns false if the object has the ignored annotation set to "true"
func ignoreIgnoredObjectPredicate(obj client.Object) bool {
	if ignoredStatus := obj.GetAnnotations()[ignoredAnnotation]; ignoredStatus == "true" {
		objType := "StarRocks resource"
		if runtimeObj, ok := obj.(runtime.Object); ok {
			objType = fmt.Sprintf("%T", runtimeObj)
		}
		logger := log.Log.WithName("predicates")
		logger.Info("starrocks operator will not reconcile ignored resource, remove annotation to reconcile",
			"type", objType,
			"namespace", obj.GetNamespace(),
			"name", obj.GetName())
		return false
	}
	return true
}

// getEnvAsSlice returns an environment variable as a slice of strings
func getEnvAsSlice(name string, defaultVal []string, separator string) []string {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}

	// Split by separator and trim spaces
	parts := strings.Split(valStr, separator)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
