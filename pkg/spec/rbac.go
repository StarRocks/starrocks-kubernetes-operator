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
package spec

import (
	"github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// build a role binding base on cn
// is for cn offline job to sync pod on fe
func MakeRoleBinding(cn *v1alpha1.ComputeNodeGroup) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cn.Name,
			Namespace: cn.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cn, cn.GroupVersionKind()),
			},
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      cn.Name,
				Namespace: cn.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     common.CnOfflineJobRole,
		},
	}
}

// build a svc account base on cn
// is for cn offline job to sync pod on fe
func MakeServiceAccount(cn *v1alpha1.ComputeNodeGroup) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cn.Name,
			Namespace: cn.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cn, cn.GroupVersionKind()),
			},
		},
	}
}
