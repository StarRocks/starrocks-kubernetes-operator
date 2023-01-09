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
	//FE_STATEFULSET_FINALIZER pre hook wait for fe statefulset deleted.
	FE_STATEFULSET_FINALIZER = "starrocks.com/fe.statefulset/finalizer"

	//BE_STATEFULSET_FINALIZER pre hook wait for be statefulset deleted.
	BE_STATEFULSET_FINALIZER = "starrocks.com/be.statefulset/finalizer"

	//CN_STATEFULSET_FINALIZER pre hook wait for cn statefulset deleted.
	CN_STATEFULSET_FINALIZER = "starrocks.com/cn.statefulset/finalizer"

	FE_SERVICE_FINALIZER = "starrocks.com/fe.service/finalizer"

	BE_SERVICE_FINALIZER = "starrocks.com/be.service/finalizer"

	CN_SERVICE_FINALIZER = "starrocks.com/cn.service/finalizer"
)

var ResourceTypeFinalizers map[string]string
