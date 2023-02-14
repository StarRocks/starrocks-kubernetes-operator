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
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/be_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/cn_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/fe_controller"
	appv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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
	Scs      map[string]SubController
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
	klog.FromContext(ctx)
	klog.Info("StarRocksClusterReconciler reconciler the update crd name ", req.Name, " namespace ", req.Namespace)
	var esrc srapi.StarRocksCluster
	if err := r.Client.Get(ctx, req.NamespacedName, &esrc); err != nil {
		klog.Error(err, " the req kind is not exists ", req.NamespacedName, " name ", req.Name)
		return requeueIfError(err)
	}

	src := esrc.DeepCopy()
	//if the src deleted, clean all resource ownerreference to src.
	clean := func() (res ctrl.Result, err error) {
		if err := r.CleanSubResources(ctx, src); err != nil {
			klog.Error("StarRocksClusterReconciler reconciler", "update faield, message ", err)
			return requeueIfError(err)
		}
		//update the sr finalizers and status.
		defer func() {
			err = r.Client.Update(ctx, src)
			if err != nil {
				klog.Error("StarRocksClusterReconciler reconciler ", "update resource failed ", "namespace ", src.Namespace, " name ", src.Name, " error ", err)
			}
		}()

		if len(src.Finalizers) == 0 {
			return res, nil
		}

		//wait for finalizers be cleaned clear.
		klog.Info("StarRocksClusterReconciler reconciler ", "have sub resosurce to cleaned ", "namespace ", src.Namespace, " starrockscluster ", src.Name, " resources finalizers ", src.Finalizers)

		return requeueAfter(10*time.Second, nil)
	}

	//reconcile deleted
	if !src.DeletionTimestamp.IsZero() {
		return clean()
	}

	//subControllers reconcile for create or update sub resource.
	for _, rc := range r.Scs {
		if err := rc.Sync(ctx, src); err != nil {
			klog.Error("StarRocksClusterReconciler reconciler ", " sub resource reconcile failed ", "namespace ", src.Namespace, " name ", src.Name, "faield ", err)
			return requeueIfError(err)
		}
	}

	//generate the status.
	r.reconcileStatus(ctx, src)
	if err := r.UpdateStarRocksClusterStatus(ctx, src); err != nil {
		klog.Error(err, "failed to update cluster status name ", src.Name, " namespace ", src.Namespace)
		return requeueIfError(err)
	}

	if src.Status.Phase != srapi.ClusterRunning {
		//TODO: or reconcile now
		return requeueAfter(10*time.Second, nil)
	}
	klog.Info("")
	return noRequeue()
}

func (r *StarRocksClusterReconciler) reconcileStatus(ctx context.Context, src *srapi.StarRocksCluster) {
	klog.Info("StarRocksClusterReconciler reconcile status.")
	//calculate the status of starrocks cluster by subresource's status.
	//clear resources when sub resource deleted.
	for _, rc := range r.Scs {
		rc.ClearResources(ctx, src)
	}

	smap := make(map[string]bool)
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
		if cnStatus != nil && cnStatus.Phase == srapi.ComponentWaiting {
			smap[srapi.ClusterPending] = true
		} else if cnStatus != nil && cnStatus.Phase == srapi.ComponentReconciling {
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

func (r *StarRocksClusterReconciler) UpdateStarRocksClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	klog.Info("StarRocksClusterReconciler reconciler ", "namespace ", src.Namespace, " name ", src.Name)

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Client.Status().Update(ctx, src)
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *StarRocksClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&srapi.StarRocksCluster{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&v2.HorizontalPodAutoscaler{}).
		Complete(r)
}

//Init initial the StarRocksClusterReconciler for reconcile.
func (r *StarRocksClusterReconciler) Init(mgr ctrl.Manager) {
	subcs := make(map[string]SubController)
	fc := fe_controller.New(mgr.GetClient(), mgr.GetEventRecorderFor(feControllerName))
	subcs[feControllerName] = fc
	cc := cn_controller.New(mgr.GetClient(), mgr.GetEventRecorderFor(cnControllerName))
	subcs[cnControllerName] = cc
	be := be_controller.New(mgr.GetClient(), mgr.GetEventRecorderFor(beControllerName))
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

//CleanSubResources clean all sub resources ownerreference to src.
func (r *StarRocksClusterReconciler) CleanSubResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	for _, c := range r.Scs {
		if _, err := c.ClearResources(ctx, src); err != nil {
			return err
		}
	}

	return nil
}
