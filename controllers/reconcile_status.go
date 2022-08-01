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

package controllers

import (
	"context"
	"fmt"
	starrocksv1alpha1 "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func (r *ComputeNodeGroupReconciler) reconcileFailed(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup, phase string, phaseErr error) (reconcile.Result, error) {
	cn.Status.Conditions[starrocksv1alpha1.Reconcile] = buildReconcileCondition(metav1.ConditionFalse, fmt.Sprintf("%s: %s", phase, phaseErr.Error()))
	statusErr := r.Client.Status().Update(ctx, cn)
	if statusErr != nil {
		return utils.Failed(statusErr)
	}
	return utils.Failed(phaseErr)
}

func (r *ComputeNodeGroupReconciler) reconcileSuccess(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup) (reconcile.Result, error) {
	cn.Status.Conditions[starrocksv1alpha1.Reconcile] = buildReconcileCondition(metav1.ConditionTrue, fmt.Sprintf("Reconciliation succeeded"))
	statusErr := r.Client.Status().Update(ctx, cn)
	if statusErr != nil {
		return utils.Failed(statusErr)
	}
	return utils.OK()
}

func (r *ComputeNodeGroupReconciler) reconcileInProgress(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup, phase string) (reconcile.Result, error) {
	cn.Status.Conditions[starrocksv1alpha1.Reconcile] = buildReconcileCondition(metav1.ConditionFalse, fmt.Sprintf("%s: In progress", phase))
	statusErr := r.Client.Status().Update(ctx, cn)
	if statusErr != nil {
		return utils.Failed(statusErr)
	}
	return utils.Retry(10, nil)
}

func buildReconcileCondition(status metav1.ConditionStatus, message string) starrocksv1alpha1.ResourceCondition {
	return starrocksv1alpha1.ResourceCondition{
		Status: status,
		Type:   starrocksv1alpha1.SyncedType,
		LastUpdateTime: metav1.Time{
			Time: time.Now(),
		},
		Message: message,
	}
}
