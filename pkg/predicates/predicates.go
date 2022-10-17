/*
Copyright 2022 StarRocks.

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
package predicates

import (
	"reflect"

	"github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// OnlyOnSpecChange returns a set of predicates indicating
// that reconciliations should only happen on changes to the Spec of the resource.
func OnlyOnSpecChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldResource := e.ObjectOld.(*v1alpha1.ComputeNodeGroup)
			newResource := e.ObjectNew.(*v1alpha1.ComputeNodeGroup)
			specChanged := !reflect.DeepEqual(oldResource.Spec, newResource.Spec)
			return specChanged
		},
	}
}
