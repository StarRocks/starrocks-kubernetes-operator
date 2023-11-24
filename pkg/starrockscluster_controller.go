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

package pkg

import (
	"context"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/feproxy"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	name                  = "starrockscluster-controller"
	feControllerName      = "fe-controller"
	cnControllerName      = "cn-controller"
	beControllerName      = "be-controller"
	feProxyControllerName = "fe-proxy-controller"
)

// StarRocksClusterReconciler reconciles a StarRocksCluster object
type StarRocksClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scs      map[string]sub_controller.ClusterSubController
}

// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/finalizers,verbs=update
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
func (r *StarRocksClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Info("StarRocksClusterReconciler reconcile the update crd name ", req.Name, " namespace ", req.Namespace)
	var esrc srapi.StarRocksCluster
	err := r.Client.Get(ctx, req.NamespacedName, &esrc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Errorf("the req kind does not exist, namespacedName = %v, name=%v, error = %v",
			req.NamespacedName, req.Name, err)
		return requeueIfError(err)
	}

	src := esrc.DeepCopy()

	// record the src updated or not by process.
	oldHashValue := r.hashStarRocksCluster(src)
	// reconcile src deleted
	if !src.DeletionTimestamp.IsZero() {
		klog.Info("StarRocksClusterReconciler reconcile the src delete namespace=" + req.Namespace + " name= " + req.Name)
		src.Status.Phase = srapi.ClusterDeleting
		// if the src deleted, clean all resource ownerreference to src.
		klog.Infof("StarRocksClusterReconciler reconcile namespace=%s, name=%s, deleted.\n", src.Namespace, src.Name)
		return ctrl.Result{}, nil
	}

	// subControllers reconcile for create or update component.
	for _, rc := range r.Scs {
		if err := rc.SyncCluster(ctx, src); err != nil {
			klog.Errorf("StarRocksClusterReconciler reconcile component failed, "+
				"namespace=%v, name=%v, controller=%v, error=%v", src.Namespace, src.Name, rc.GetControllerName(), err)
			return requeueIfError(err)
		}
	}

	newHashValue := r.hashStarRocksCluster(src)
	if oldHashValue != newHashValue {
		return ctrl.Result{Requeue: true}, r.PatchStarRocksCluster(ctx, src)
	}

	for _, rc := range r.Scs {
		// update component status.
		if err := rc.UpdateClusterStatus(src); err != nil {
			klog.Infof("StarRocksClusterReconciler reconcile update component %s status failed.err=%s\n", rc.GetControllerName(), err.Error())
			return requeueIfError(err)
		}
	}

	// generate the src status.
	r.reconcileStatus(ctx, src)
	return ctrl.Result{}, r.UpdateStarRocksClusterStatus(ctx, src)
}

// UpdateStarRocksClusterStatus update the status of src.
func (r *StarRocksClusterReconciler) UpdateStarRocksClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var esrc srapi.StarRocksCluster
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Name}, &esrc); err != nil {
			return err
		}

		esrc.Status = src.Status
		return r.Client.Status().Update(ctx, &esrc)
	})
}

// PatchStarRocksCluster patch spec, metadata
func (r *StarRocksClusterReconciler) PatchStarRocksCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	klog.Info("StarRocksClusterReconciler reconcile ", "namespace ", src.Namespace, " name ", src.Name)

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var esrc srapi.StarRocksCluster
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Name}, &esrc); err != nil {
			return err
		}

		src.ResourceVersion = esrc.ResourceVersion

		return k8sutils.PatchClientObject(ctx, r.Client, src)
	})
}

// UpdateStarRocksCluster update the starrockscluster metadata, spec.
func (r *StarRocksClusterReconciler) UpdateStarRocksCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var esrc srapi.StarRocksCluster
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: src.Namespace, Name: src.Name}, &esrc); apierrors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}

		src.ResourceVersion = esrc.ResourceVersion
		return k8sutils.UpdateClientObject(ctx, r.Client, src)
	})
}

// hash the starrockscluster for check the crd modified or not.
func (r *StarRocksClusterReconciler) hashStarRocksCluster(src *srapi.StarRocksCluster) string {
	type hashObject struct {
		metav1.ObjectMeta
		Spec srapi.StarRocksClusterSpec
	}
	ho := &hashObject{
		ObjectMeta: src.ObjectMeta,
		Spec:       src.Spec,
	}

	return hash.HashObject(ho)
}

func (r *StarRocksClusterReconciler) reconcileStatus(ctx context.Context, src *srapi.StarRocksCluster) {
	// calculate the status of starrocks cluster by subresource's status.
	// clear resources when component deleted. example: deployed fe,be,cn, when cn spec is deleted we should delete cn resources.
	for _, rc := range r.Scs {
		if err := rc.ClearResources(ctx, src); err != nil {
			klog.Errorf("StarRocksClusterReconciler reconcile clear resource failed, "+
				"namespace=%v, name=%v, controller=%v, error=%v", src.Namespace, src.Name, rc.GetControllerName(), err)
		}
	}

	src.Status.Phase = srapi.ClusterRunning
	phase := GetPhaseFromComponent(&src.Status.StarRocksFeStatus.StarRocksComponentStatus)
	if phase != "" {
		src.Status.Phase = phase
		return
	}
	if src.Status.StarRocksBeStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.StarRocksBeStatus.StarRocksComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
	if src.Status.StarRocksCnStatus != nil {
		phase = GetPhaseFromComponent(&src.Status.StarRocksCnStatus.StarRocksComponentStatus)
		if phase != "" {
			src.Status.Phase = phase
			return
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// can not add Owns(&v2.HorizontalPodAutoscaler{}), because if kubernetes version is lower than 1.23,
	// v2.HorizontalPodAutoscaler does not exist.
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func SetupClusterReconciler(mgr ctrl.Manager) error {
	subcs := make(map[string]sub_controller.ClusterSubController)
	feController := fe.New(mgr.GetClient())
	subcs[feControllerName] = feController
	cnController := cn.New(mgr.GetClient())
	subcs[cnControllerName] = cnController
	beController := be.New(mgr.GetClient())
	subcs[beControllerName] = beController
	feProxyController := feproxy.New(mgr.GetClient())
	subcs[feProxyControllerName] = feProxyController

	reconciler := &StarRocksClusterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor(name),
		Scs:      subcs,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		klog.Error(err, " unable to create controller ", "controller ", "StarRocksCluster ")
		return err
	}
	return nil
}

func requeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}
