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

package v1alpha1

//the labels key
const (
	// ComponentLabelKey is Kubernetes recommended label key, it represents the component within the architecture
	ComponentLabelKey string = "app.kubernetes.io/component"
	// NameLabelKey is Kubernetes recommended label key, it represents the name of the application
	NameLabelKey string = "app.kubernetes.io/name"

	//OwnerReference list object depended by this object
	OwnerReference string = "app.starrocks.ownerreference/name"

	//ComponentsResourceHash the component hash
	ComponentResourceHash string = "app.starrocks.components/hash"

	//ComponentGeneration record for last update generation for compare with new spec.
	ComponentGeneration string = "app.starrocks.components/generation"
)

//the labels value. default statefulset name
const (
	DEFAULT_FE = "fe"
	DEFAULT_BE = "be"
	DEFAULT_CN = "cn"
)

//config value
const (
	DEFAULT_FE_CONFIG_NAME = "fe-config"

	DEFAULT_EMPTDIR_NAME = "shard-data"

	INITIAL_VOLUME_PATH = "/pod-data"

	INITIAL_VOLUME_PATH_NAME = "shared-data"

	DEFAULT_START_SCRIPT_NAME = "fe-start-script"

	//TODO: when script need set.
	DEFAULT_START_SCRIPT_PATH = ""

	DEFAULT_FE_SERVICE_NAME = "starrocks-fe-service"

	DEFAULT_BE_SERVICE_NAME = "starrocks-be-service"

	DEFAULT_CN_SERVICE_NAME = "starrocks-cn-service"
)

//the env of container
const (
	COMPONENT_NAME      = "COMPONENT_NAME"
	FE_SERVICE_NAME     = "FE_SERVICE_NAME"
	SEARCH_SERVICE_NAME = "SEARCH_SERVICE_NAME"
)
