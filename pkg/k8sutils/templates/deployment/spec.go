// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployment

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

// MakeDeployment make deployment
func MakeDeployment(cluster *v1.StarRocksCluster, spec v1.SpecInterface, podTemplateSpec corev1.PodTemplateSpec) *appv1.Deployment {
	or := metav1.NewControllerRef(cluster, cluster.GroupVersionKind())
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            load.Name(cluster.Name, spec),
			Namespace:       cluster.Namespace,
			Labels:          load.Labels(cluster.Name, spec),
			Annotations:     load.Annotations(),
			OwnerReferences: []metav1.OwnerReference{*or},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: spec.GetReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: load.Selector(cluster.Name, spec),
			},
			Template: podTemplateSpec,
		},
	}

	return deployment
}
