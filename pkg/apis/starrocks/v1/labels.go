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

package v1

// the labels key
const (
	// ComponentLabelKey is Kubernetes recommended label key, it represents the component within the architecture
	ComponentLabelKey string = "app.kubernetes.io/component"
	// NameLabelKey is Kubernetes recommended label key, it represents the name of the application
	NameLabelKey string = "app.kubernetes.io/name"

	// OwnerReference list object depended by this object
	OwnerReference string = "app.starrocks.ownerreference/name"

	// ComponentsResourceHash the component hash
	ComponentResourceHash string = "app.starrocks.components/hash"

	ComponentReplicasEmpty string = "app.starrocks.components/replica/empty"

	// ComponentGeneration record for last update generation for compare with new spec.
	// ComponentGeneration string = "app.starrocks.components/generation"
)

// the labels value. default statefulset name
const (
	DEFAULT_FE       = "fe"
	DEFAULT_BE       = "be"
	DEFAULT_CN       = "cn"
	DEFAULT_FE_PROXY = "fe-proxy"
)

// the env of container
const (
	COMPONENT_NAME  = "COMPONENT_NAME"
	FE_SERVICE_NAME = "FE_SERVICE_NAME"
)
