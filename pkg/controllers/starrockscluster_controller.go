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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/be"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/cn"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/fe"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/feproxy"
)

// CelerDataClusterReconciler reconciles a CelerDataCluster object
type CelerDataClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scs      []subcontrollers.ClusterSubController
	denyList string
}

// +kubebuilder:rbac:groups=celerdata.com,resources=celerdataclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celerdata.com,resources=celerdataclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=celerdata.com,resources=celerdataclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",resources=endpoints,verbs=get;watch;list
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *CelerDataClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithName("CelerDataClusterReconciler").WithValues("name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)
	logger.Info("begin to reconcile CelerDataCluster")

	logger.Info("get CelerDataCluster CR from kubernetes")
	var esrc cdapi.CelerDataCluster
	client := r.Client
	err := client.Get(ctx, req.NamespacedName, &esrc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get CelerDataCluster object failed")
		return requeueIfError(err)
	}
	src := esrc.DeepCopy()

	// reconcile src deleted
	if !src.DeletionTimestamp.IsZero() {
		logger.Info("deletion timestamp is not zero, clear CelerDataCluster related resources")
		return ctrl.Result{}, nil
	}

	// subControllers reconcile for create or update component.
	for _, rc := range r.Scs {
		kvs := []interface{}{"subController", rc.GetControllerName()}
		logger.Info("sub controller sync spec", kvs...)
		if err = rc.SyncCluster(ctx, src); err != nil {
			logger.Error(err, "sub controller reconciles spec failed", kvs...)
			handleSyncClusterError(src, rc, err)
			if updateError := r.UpdateCelerDataClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update CelerDataCluster Status")
			}
			return requeueIfError(err)
		}
	}

	for _, rc := range r.Scs {
		kvs := []interface{}{"subController", rc.GetControllerName()}
		logger.Info("sub controller update status", kvs...)
		if err = rc.UpdateClusterStatus(ctx, src); err != nil {
			logger.Error(err, "sub controller update status failed", kvs...)
			handleSyncClusterError(src, rc, err)
			if updateError := r.UpdateCelerDataClusterStatus(ctx, src); updateError != nil {
				logger.Error(updateError, "failed to update CelerDataCluster Status")
			}
			return requeueIfError(err)
		}
	}

	logger.Info("update CelerDataCluster level status")
	r.reconcileStatus(ctx, src)
	err = r.UpdateCelerDataClusterStatus(ctx, src)
	if err != nil {
		logger.Error(err, "update CelerDataCluster status failed")
		return ctrl.Result{}, err
	}
	logger.Info("reconcile CelerDataCluster success")
	return ctrl.Result{}, nil
}

// UpdateCelerDataClusterStatus update the status of src.
func (r *CelerDataClusterReconciler) UpdateCelerDataClusterStatus(ctx context.Context, src *cdapi.CelerDataCluster) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var esrc cdapi.CelerDataCluster
		client := r.Client
		if err := client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Name}, &esrc); err != nil {
			return err
		}

		esrc.Status = src.Status
		return r.Client.Status().Update(ctx, &esrc)
	})
}

func (r *CelerDataClusterReconciler) reconcileStatus(_ context.Context, src *cdapi.CelerDataCluster) {
	src.Status.Phase = cdapi.ClusterRunning
	src.Status.Reason = ""
	var phase cdapi.Phase
	if src.Status.CelerDataFeStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.CelerDataFeStatus.CelerDataComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
	if src.Status.CelerDataBeStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.CelerDataBeStatus.CelerDataComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
	if src.Status.CelerDataCnStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.CelerDataCnStatus.CelerDataComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
}

// handleSyncClusterError handle errors from sub-controller, and log it in CelerDataCluster Status
func handleSyncClusterError(src *cdapi.CelerDataCluster, subController subcontrollers.ClusterSubController, err error) {
	reason := err.Error()
	switch subController.(type) {
	case *fe.FeController:
		reason = fmt.Sprintf("error from FE controller: %v", reason)
	case *be.BeController:
		reason = fmt.Sprintf("error from BE controller: %v", reason)
	case *cn.CnController:
		reason = fmt.Sprintf("error from CN controller: %v", reason)
	case *feproxy.FeProxyController:
		reason = fmt.Sprintf("error from fe-proxy controller: %v", reason)
	}

	src.Status.Phase = cdapi.ClusterFailed
	src.Status.Reason = reason
}

func requeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// GetPhaseFromComponent return the Phase of Cluster or Warehouse based on the component status.
// It returns empty string if not sure the phase.
func GetPhaseFromComponent(componentStatus *cdapi.CelerDataComponentStatus) cdapi.Phase {
	if componentStatus == nil {
		return ""
	}
	if componentStatus.Phase == cdapi.ComponentReconciling {
		return cdapi.ClusterReconciling
	}
	if componentStatus.Phase == cdapi.ComponentFailed {
		return cdapi.ClusterFailed
	}
	return ""
}
