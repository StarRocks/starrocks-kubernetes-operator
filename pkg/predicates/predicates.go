package predicates

import (
	"fmt"
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

// GenericPredicates implements predicate.Predicate for filtering events.
// The deny list is cached at initialization time for better performance.
type GenericPredicates struct {
	predicate.Funcs
	denyList map[string]struct{}
}

// NewGenericPredicates creates a new GenericPredicates with the given deny list.
// The denyList parameter is a comma-separated string of namespace names.
func NewGenericPredicates(denyList string) GenericPredicates {
	denyMap := make(map[string]struct{})
	if denyList != "" {
		for _, ns := range strings.Split(denyList, ",") {
			trimmed := strings.TrimSpace(ns)
			if trimmed != "" {
				denyMap[trimmed] = struct{}{}
			}
		}
	}
	return GenericPredicates{denyList: denyMap}
}

// Create returns true if the Create event should be processed
func (gp GenericPredicates) Create(e event.CreateEvent) bool {
	return gp.shouldReconcile(e.Object)
}

// Update returns true if the Update event should be processed
func (gp GenericPredicates) Update(e event.UpdateEvent) bool {
	return gp.shouldReconcile(e.ObjectNew)
}

// Delete returns true if the Delete event should be processed
func (gp GenericPredicates) Delete(e event.DeleteEvent) bool {
	return gp.shouldReconcile(e.Object)
}

// Generic returns true if the Generic event should be processed
func (gp GenericPredicates) Generic(e event.GenericEvent) bool {
	return gp.shouldReconcile(e.Object)
}

// shouldReconcile checks if an object should be reconciled based on namespace and annotation filters
func (gp GenericPredicates) shouldReconcile(obj client.Object) bool {
	if obj == nil {
		return false
	}

	// Check namespace deny list
	if !gp.isNamespaceAllowed(obj) {
		return false
	}

	// Check ignored annotation
	if !isObjectAllowed(obj) {
		return false
	}

	return true
}

// isNamespaceAllowed returns true if the object's namespace is not in the deny list
func (gp GenericPredicates) isNamespaceAllowed(obj client.Object) bool {
	if len(gp.denyList) == 0 {
		return true
	}

	if _, denied := gp.denyList[obj.GetNamespace()]; denied {
		logger := log.Log.WithName("predicates")
		logger.Info("starrocks operator will not reconcile namespace, update --deny-list to reconcile",
			"namespace", obj.GetNamespace())
		return false
	}
	return true
}

// isObjectAllowed returns true if the object does not have the ignored annotation set to "true"
func isObjectAllowed(obj client.Object) bool {
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
