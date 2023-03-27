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

const (
	//STATEFULSET_FINALIZER pre hook wait for fe statefulset deleted.
	STATEFULSET_FINALIZER = "starrocks.com.statefulset/protection"

	SERVICE_FINALIZER = "starrocks.com.service/protection"

	STARROCKS_FINALIZER = "starrocks.com.starrockscluster/protection"

	AUTOSCALER_FINALIZER = "starrocks.com.autoscaler/protection"
)
