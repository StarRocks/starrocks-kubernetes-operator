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

package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/utils"
	v1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2beta2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	//batchv1 "k8s.io/api/batch/v1"
	"time"

	"github.com/StarRocks/starrocks-kubernetes-operator/internal/fe"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/spec"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	starrocksv1alpha1 "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	k8s_error "k8s.io/apimachinery/pkg/api/errors"
)

/*func init() {
	Controllers = append(Controllers, &ComputeNodeGroupReconciler{})
}*/

// ComputeNodeGroupReconciler reconciles a ComputeNodeGroup object
type ComputeNodeGroupReconciler struct {
	Client  client.Client
	Rclient client.Reader
	Scheme  *runtime.Scheme
}

//+kubebuilder:rbac:groups=starrocks.com,resources=computenodegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=starrocks.com,resources=computenodegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=starrocks.com,resources=computenodegroups/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ComputeNodeGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ComputeNodeGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	_ = log.FromContext(ctx)
	klog.Infof("Reconcile ComputeNodeGroup: %s:%s", req.Namespace, req.Name)
	// TODO(user): your logic here
	cn := &starrocksv1alpha1.ComputeNodeGroup{}
	err = r.Client.Get(ctx, req.NamespacedName, cn)
	if err != nil {
		if k8s_error.IsNotFound(err) {
			return utils.OK()
		}
		return utils.Failed(err)
	}
	cn = cn.DeepCopy()

	state := &CnState{
		Req:  req,
		Inst: cn,
	}
	if state.Inst.Status.Conditions == nil {
		state.Inst.Status.Conditions = make(map[starrocksv1alpha1.CnComponent]starrocksv1alpha1.ResourceCondition)
	}
	err = r.handleFinalizer(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "handle finalizer", err)
	}
	err = r.observeDeployment(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "observe deployment status", err)
	}
	err = r.observeCronJob(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "observe cronjob status", err)
	}
	err = r.observeRbac(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "observe rbac status", err)
	}
	err = r.observeHPA(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "observe hpa status", err)
	}
	err = r.Client.Status().Update(ctx, state.Inst)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "update status status", err)
	}
	err = r.applyDeployment(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "apply deployment", err)
	}
	err = r.applyCronJob(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "apply cronjob", err)
	}
	err = r.applyRbac(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "apply rbac", err)
	}
	err = r.applyHPA(ctx, state)
	if err != nil {
		return r.reconcileFailed(ctx, state.Inst, "apply hpa", err)
	}

	for component, condition := range state.Inst.Status.Conditions {
		if component != starrocksv1alpha1.Reconcile {
			if condition.Status == metav1.ConditionFalse {
				return r.reconcileInProgress(ctx, state.Inst, fmt.Sprintf("%s is not ready", component))
			}
		}
	}

	return r.reconcileSuccess(ctx, state.Inst)
}

func (r *ComputeNodeGroupReconciler) handleFinalizer(ctx context.Context, state *CnState) error {
	if state.Inst.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(state.Inst.ObjectMeta.Finalizers, common.CnFinalizerName) {
			state.Inst.Finalizers = append(state.Inst.Finalizers, common.CnFinalizerName)
			inst := state.Inst.DeepCopy()
			if err := r.Client.Update(ctx, inst); err != nil {
				return err
			}
		}
	} else {
		if containsString(state.Inst.ObjectMeta.Finalizers, common.CnFinalizerName) {
			err := r.cleanPods(ctx, state.Inst)
			if err != nil {
				return err
			}
			cleanUp, err := r.isPodCleanUp(ctx, state.Inst)
			if err != nil {
				return err
			}
			if !cleanUp {
				return utils.ErrFinalizerUnfinished
			}
			err = r.cleanCnOnFe(ctx, state.Inst)
			if err != nil {
				return err
			}
			state.Inst.ObjectMeta.Finalizers = removeString(state.Inst.ObjectMeta.Finalizers, common.CnFinalizerName)
			inst := state.Inst.DeepCopy()
			if err := r.Client.Update(ctx, inst); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ComputeNodeGroupReconciler) cleanPods(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup) error {
	deploy := &v1.Deployment{}
	err := r.Rclient.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Name,
	}, deploy)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		return nil
	}
	*deploy.Spec.Replicas = 0

	return r.Client.Update(ctx, deploy)
}

func (r *ComputeNodeGroupReconciler) isPodCleanUp(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup) (bool, error) {
	deploy := &v1.Deployment{}
	err := r.Rclient.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Name,
	}, deploy)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return false, err
		}
		return true, nil
	}
	pods := &corev1.PodList{}
	err = r.Rclient.List(ctx, pods, client.InNamespace(deploy.Namespace), client.MatchingLabels(deploy.Spec.Selector.MatchLabels))
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			return false, nil
		}
	}
	return true, nil
}

func (r *ComputeNodeGroupReconciler) cleanCnOnFe(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup) error {
	if cn.Status.Servers.Available+cn.Status.Servers.Unavailable == 0 {
		return nil
	}

	feUsr, fePwd, err := r.getFeAccount(ctx, cn)
	fePick := fe.PickFe(cn.Spec.FeInfo.Addresses)
	nodes, err := fe.GetNodes(fePick, feUsr, fePwd)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if node.Alive {
			continue
		}
		cnAddr := fmt.Sprintf("%s:%s", node.Ip, common.CnHeartBeatPort)
		err := fe.DropNode(fePick, feUsr, fePwd, cnAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *ComputeNodeGroupReconciler) applyRbac(ctx context.Context, state *CnState) error {
	if state.ServiceAccount == nil {
		err := r.Client.Create(ctx, spec.MakeServiceAccount(state.Inst))
		if err != nil {
			return err
		}
	}
	if state.RoleBinding == nil {
		err := r.Client.Create(ctx, spec.MakeRoleBinding(state.Inst))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ComputeNodeGroupReconciler) applyDeployment(ctx context.Context, state *CnState) error {
	current := state.Deployment
	desired := spec.MakeCnDeployment(state.Inst)
	if state.Deployment == nil { // create a new one
		err := r.Client.Create(ctx, desired)
		if err != nil {
			return err
		}
	} else {
		spec.SyncDeploymentChanged(current, desired)
		err := r.Client.Update(ctx, current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ComputeNodeGroupReconciler) applyCronJob(ctx context.Context, state *CnState) error {
	current := state.CronJob
	desired := spec.MakeCnCronJob(state.Inst)
	if state.CronJob == nil {
		err := r.Client.Create(ctx, desired)
		if err != nil {
			return err
		}
	} else {
		spec.SyncCronJobChanged(current, desired)
		err := r.Client.Update(ctx, current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ComputeNodeGroupReconciler) applyHPA(ctx context.Context, state *CnState) error {
	current := state.HPA
	desired := spec.MakeCnHPA(state.Inst)
	if desired == nil {
		if current != nil {
			return r.Client.Delete(ctx, current)
		}
		return nil
	}
	if state.HPA == nil {
		err := r.Client.Create(ctx, desired)
		if err != nil {
			return err
		}
	} else {
		spec.SyncHPAChanged(current, desired)
		err := r.Client.Update(ctx, current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ComputeNodeGroupReconciler) observeCronJob(ctx context.Context, state *CnState) error {
	cronJob := &batchv1beta1.CronJob{}
	err := r.Rclient.Get(ctx, state.Req.NamespacedName, cronJob)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		cronJob = nil
	}
	state.CronJob = cronJob
	return nil
}

func (r *ComputeNodeGroupReconciler) observeHPA(ctx context.Context, state *CnState) error {
	hpa := &v2.HorizontalPodAutoscaler{}
	err := r.Rclient.Get(ctx, state.Req.NamespacedName, hpa)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		hpa = nil
	}

	state.HPA = hpa
	return nil
}

func (r *ComputeNodeGroupReconciler) observeRbac(ctx context.Context, state *CnState) error {
	roleBinding := &rbacv1.ClusterRoleBinding{}
	err := r.Rclient.Get(ctx, state.Req.NamespacedName, roleBinding)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		roleBinding = nil
	}
	serviceAccount := &corev1.ServiceAccount{}
	err = r.Rclient.Get(ctx, state.Req.NamespacedName, serviceAccount)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		serviceAccount = nil
	}

	state.RoleBinding = roleBinding
	state.ServiceAccount = serviceAccount
	return nil
}

func (r *ComputeNodeGroupReconciler) observeDeployment(ctx context.Context, state *CnState) error {
	deploy := &v1.Deployment{}
	err := r.Rclient.Get(ctx, state.Req.NamespacedName, deploy)
	if err != nil {
		if !k8s_error.IsNotFound(err) {
			return err
		}
		deploy = nil
	}
	state.Deployment = deploy
	if deploy != nil {
		pods := &corev1.PodList{}
		err := r.Rclient.List(ctx, pods, client.InNamespace(state.Req.Namespace), client.MatchingLabels(state.Deployment.Spec.Selector.MatchLabels))
		if err != nil {
			return err
		}
		err = r.observeFeSyncWithPods(ctx, state, pods)
		if err != nil {
			return err
		}
		// for hpa
		selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
		if err != nil {
			return err
		}
		state.Inst.Status.Replicas = deploy.Status.Replicas
		state.Inst.Status.LabelSelector = selector.String()
	}

	condition := starrocksv1alpha1.ResourceCondition{
		Type:   starrocksv1alpha1.SyncedType,
		Status: metav1.ConditionFalse,
		LastUpdateTime: metav1.Time{
			Time: time.Now(),
		},
	}

	if state.Deployed() {
		condition.Status = metav1.ConditionTrue
	}

	state.Inst.Status.Conditions[starrocksv1alpha1.Deployment] = condition

	return nil
}

func (r *ComputeNodeGroupReconciler) getFeAccount(ctx context.Context, cn *starrocksv1alpha1.ComputeNodeGroup) (usr, pwd string, err error) {
	sec := &corev1.Secret{}

	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: cn.Namespace,
		Name:      cn.Spec.FeInfo.AccountSecret,
	}, sec)
	if err != nil {
		return "", "", err
	}

	usrBytes, ok := sec.Data[common.EnvKeyFeUsr]
	if !ok {
		err = errors.New(fmt.Sprintf("can not get usr in secret"))
		return "", "", err
	}
	pwdBytes, ok := sec.Data[common.EnvKeyFePwd]
	if !ok {
		err = errors.New(fmt.Sprintf("can not get pwd in secret"))
		return "", "", err
	}
	return string(usrBytes), string(pwdBytes), nil
}

func (r *ComputeNodeGroupReconciler) observeFeSyncWithPods(ctx context.Context, state *CnState, pods *corev1.PodList) error {
	if state.Deployment != nil {
		registedIps := sets.NewString()
		podIps := sets.NewString()

		available := int32(0)
		unavailable := int32(0)
		useless := int32(0)
		unregisted := int32(0)

		fePick := fe.PickFe(state.Inst.Spec.FeInfo.Addresses)

		feUsr, fePwd, err := r.getFeAccount(ctx, state.Inst)
		if err != nil {
			return err
		}
		nodes, err := fe.GetNodes(fePick, feUsr, fePwd)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			registedIps = registedIps.Insert(node.Ip)
			if node.Alive {
				available++
			} else {
				unavailable++
			}
		}
		for _, pod := range pods.Items {
			podIps = podIps.Insert(pod.Status.PodIP)
		}
		for _, ip := range podIps.List() {
			if !registedIps.Has(ip) {
				unregisted++
			}
		}
		for _, ip := range registedIps.List() {
			if !podIps.Has(ip) {
				useless++
			}
		}

		state.Inst.Status.Servers.Available = available
		state.Inst.Status.Servers.Unavailable = unavailable
		state.Inst.Status.Servers.Unregistered = unregisted
		state.Inst.Status.Servers.Useless = useless

		condition := starrocksv1alpha1.ResourceCondition{
			Status: metav1.ConditionFalse,
			Type:   starrocksv1alpha1.SyncedType,
			LastUpdateTime: metav1.Time{
				Time: time.Now(),
			},
		}

		if state.SyncedWithFe() {
			condition.Status = metav1.ConditionTrue
		}

		state.Inst.Status.Conditions[starrocksv1alpha1.Fe] = condition
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComputeNodeGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 3}).
		For(&starrocksv1alpha1.ComputeNodeGroup{}).
		Owns(&v1.Deployment{}).
		Owns(&batchv1beta1.CronJob{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *ComputeNodeGroupReconciler) Init(mgr ctrl.Manager) {
	if err := (&ComputeNodeGroupReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Rclient: mgr.GetAPIReader(),
	}).SetupWithManager(mgr); err != nil {
		klog.Error(err, "unable to create controller", "controller", "ComputeNodeGroup")
		os.Exit(1)
	}
}
