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
	starrocksv1alpha1 "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2beta2"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// represent state of cn which reconciled
type CnState struct {
	Result         ctrl.Result
	Req            ctrl.Request
	Inst           *starrocksv1alpha1.ComputeNodeGroup
	Deployment     *appv1.Deployment
	HPA            *v2.HorizontalPodAutoscaler
	CronJob        *batchv1beta1.CronJob
	RoleBinding    *rbacv1.ClusterRoleBinding
	ServiceAccount *corev1.ServiceAccount
	ConfigMap      *corev1.ConfigMap
}

// is deployment synced
func (c *CnState) Deployed() bool {
	return c.Deployment != nil &&
		c.Deployment.Status.ObservedGeneration == c.Deployment.Generation &&
		c.Deployment.Status.ReadyReplicas == c.Deployment.Status.Replicas
}

//func (c *CnState) ImageUpdated() bool {
//	return c.De
//}

// is fe synced
func (c *CnState) SyncedWithFe() bool {
	return c.Inst.Status.Servers.Useless == 0 &&
		c.Inst.Status.Servers.Unregistered == 0 &&
		c.Inst.Status.Servers.Available+c.Inst.Status.Servers.Unavailable == c.Inst.Status.Replicas
}

// is cron job created
func (c *CnState) JobCreated() bool {
	return c.CronJob != nil
}

// is rolebinding created
func (c *CnState) RoleBindingCreated() bool {
	return c.RoleBinding != nil
}

// is svc account created
func (c *CnState) ServiceAccountCreated() bool {
	return c.ServiceAccount != nil
}

// is hpa created
func (c *CnState) HpaCreated() bool {
	return c.HPA != nil
}
