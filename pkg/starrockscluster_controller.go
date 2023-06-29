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
	"errors"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/be"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/cn"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Controllers = append(Controllers, &StarRocksClusterReconciler{})
}

var (
	name             = "starrockscluster-controller"
	feControllerName = "fe-controller"
	cnControllerName = "cn-controller"
	beControllerName = "be-controller"
)

// StarRocksClusterReconciler reconciles a StarRocksCluster object
type StarRocksClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scs      map[string]sub_controller.SubController
}

//+kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=starrocks.com,resources=starrocksclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="core",resources=endpoints,verbs=get;watch;list
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StarRocksCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
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

	//record the src updated or not by process.
	oldHashValue := r.hashStarRocksCluster(src)
	//reconcile src deleted
	if !src.DeletionTimestamp.IsZero() {
		klog.Info("StarRocksClusterReconciler reconcile the src delete namespace=" + req.Namespace + " name= " + req.Name)
		src.Status.Phase = srapi.ClusterDeleting
		//if the src deleted, clean all resource ownerreference to src.
		clean := func() (res ctrl.Result, err error) {
			delres, err := r.CleanSubResources(ctx, src)
			if err != nil {
				klog.Errorf("StarRocksClusterReconciler reconcile update faield, message=%s\n", err.Error())
				return requeueIfError(err)
			}

			if !delres {
				//wait for finalizers be cleaned clear.
				klog.Info("StarRocksClusterReconciler reconcile ", "have sub resosurce to cleaned ", "namespace ", src.Namespace, " starrockscluster ", src.Name)
				return ctrl.Result{}, nil
			}

			klog.Infof("StarRocksClusterReconciler reconcile namespace=%s, name=%s, deleted.\n", src.Namespace, src.Name)

			// all resource clear over, clear starrockcluster finalizers.
			src.Finalizers = nil
			//delete the src will be hooked by finalizer, we should clear the finalizers.
			return ctrl.Result{}, r.UpdateStarRocksCluster(ctx, src)
		}

		return clean()
	}

	//subControllers reconcile for create or update sub resource.
	for _, rc := range r.Scs {
		if err := rc.Sync(ctx, src); err != nil {
			klog.Errorf("StarRocksClusterReconciler reconcile sub resource reconcile failed, "+
				"namespace=%v, name=%v, controller=%v, error=%v", src.Namespace, src.Name, rc.GetControllerName(), err)
			return requeueIfError(err)
		}
	}

	newHashValue := r.hashStarRocksCluster(src)
	if oldHashValue != newHashValue {
		return ctrl.Result{Requeue: true}, r.PatchStarRocksCluster(ctx, src)
	}

	for _, rc := range r.Scs {
		//update component status.
		if err := rc.UpdateStatus(src); err != nil {
			klog.Infof("StarRocksClusterReconciler reconcile update component %s status failed.err=%s\n", rc.GetControllerName(), err.Error())
			return requeueIfError(err)
		}
	}

	//update restart status.
	if r.haveRestartOperation(src.Annotations) {
		r.updateOperationStatus(src)
		newHashValue = r.hashStarRocksCluster(src)
		if oldHashValue != newHashValue {
			return ctrl.Result{Requeue: true}, r.PatchStarRocksCluster(ctx, src)
		}
	}

	//generate the src status.
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

// UpdateStarRocksCluster udpate the starrockscluster metadata, spec.
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

// CleanSubResources clean all sub resources ownerreference to src.
func (r *StarRocksClusterReconciler) CleanSubResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error) {
	var cleanErr error
	res := true
	for _, c := range r.Scs {
		subres, err := c.ClearResources(ctx, src)
		if err != nil {
			cleanErr = errors.New(c.GetControllerName() + "err=" + err.Error())
		}

		res = res && subres
	}

	return res, cleanErr
}

func (r *StarRocksClusterReconciler) updateOperationStatus(src *srapi.StarRocksCluster) {
	for _, rc := range r.Scs {
		rc.SyncRestartStatus(src)
	}
}

// checks if user have restart service need.
func (r *StarRocksClusterReconciler) haveRestartOperation(annos map[string]string) bool {

	if v, ok := annos[string(srapi.AnnotationBERestartKey)]; ok {
		if v != string(srapi.AnnotationRestartFinished) {
			return true
		}
	}
	if v, ok := annos[string(srapi.AnnotationFERestartKey)]; ok {
		if v != string(srapi.AnnotationRestartFinished) {
			return true
		}
	}
	if v, ok := annos[string(srapi.AnnotationCNRestartKey)]; ok {
		if v != string(srapi.AnnotationRestartFinished) {
			return true
		}
	}

	return false
}

func (r *StarRocksClusterReconciler) reconcileStatus(ctx context.Context, src *srapi.StarRocksCluster) {
	//calculate the status of starrocks cluster by subresource's status.
	//clear resources when sub resource deleted. example: deployed fe,be,cn, when cn spec is deleted we should delete cn resources.
	for _, rc := range r.Scs {
		rc.ClearResources(ctx, src)
	}

	smap := make(map[srapi.ClusterPhase]bool)
	src.Status.Phase = srapi.ClusterRunning
	func() {
		feStatus := src.Status.StarRocksFeStatus
		if feStatus != nil && feStatus.Phase == srapi.ComponentReconciling {
			smap[srapi.ClusterPending] = true
		} else if feStatus != nil && feStatus.Phase == srapi.ComponentFailed {
			smap[srapi.ClusterFailed] = true
		}
	}()

	func() {
		cnStatus := src.Status.StarRocksCnStatus
		if cnStatus != nil && cnStatus.Phase == srapi.ComponentReconciling {
			smap[srapi.ClusterPending] = true
		} else if cnStatus != nil && cnStatus.Phase == srapi.ComponentFailed {
			smap[srapi.ClusterFailed] = true
		}
	}()

	if _, ok := smap[srapi.ClusterPending]; ok {
		src.Status.Phase = srapi.ClusterPending
	} else if _, ok := smap[srapi.ClusterFailed]; ok {
		src.Status.Phase = srapi.ClusterFailed
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// Init initial the StarRocksClusterReconciler for reconcile.
func (r *StarRocksClusterReconciler) Init(mgr ctrl.Manager) {
	subcs := make(map[string]sub_controller.SubController)
	fc := fe.New(mgr.GetClient(), mgr.GetEventRecorderFor(feControllerName))
	subcs[feControllerName] = fc
	cc := cn.New(mgr.GetClient(), mgr.GetEventRecorderFor(cnControllerName))
	subcs[cnControllerName] = cc
	be := be.New(mgr.GetClient(), mgr.GetEventRecorderFor(beControllerName))
	subcs[beControllerName] = be

	if err := (&StarRocksClusterReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor(name),
		Scs:      subcs,
	}).SetupWithManager(mgr); err != nil {
		klog.Error(err, " unable to create controller ", "controller ", "StarRocksCluster ")
		os.Exit(1)
	}
}
