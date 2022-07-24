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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// cr's component status
type ResourceCondition struct {
	Status         metav1.ConditionStatus `json:"status"`
	Type           ReconcileType          `json:"type"`
	LastUpdateTime metav1.Time            `json:"lastUpdateTime,omitempty" protobuf:"bytes,6,opt,name=lastUpdateTime"`
	Message        string                 `json:"message,omitempty"`
}

type CnComponent string

const (
	Deployment CnComponent = "Deployment"
	CronJob    CnComponent = "CronJob"
	Hpa        CnComponent = "HPA"
	Fe         CnComponent = "Fe"
	Reconcile  CnComponent = "Reconcile"
)

type ReconcileType string

const (
	SyncedType  ReconcileType = "Synced"
	UpdatedType ReconcileType = "Updated"
)
