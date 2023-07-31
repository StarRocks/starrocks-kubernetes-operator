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

// +k8s:deepcopy-gen=package
// +groupName=starrocks.com
package sub_controller

import (
	"context"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

type SubController interface {
	// Sync reconcile for sub controller. bool represent the component have updated.
	Sync(ctx context.Context, src *srapi.StarRocksCluster) error

	// ClearResources clear all resource about sub-component.
	ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error)

	// GetControllerName return the controller name, beController, feController,cnController for log.
	GetControllerName() string

	// UpdateStatus update the component status on src.
	UpdateStatus(src *srapi.StarRocksCluster) error

	// SyncRestartStatus sync the status of restart.
	SyncRestartStatus(src *srapi.StarRocksCluster) error
}
