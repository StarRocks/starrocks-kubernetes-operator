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

package statefulset

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	srobject "github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const STARROCKS_WAREHOUSE_FINALIZER = "starrocks.com.starrockswarehouse/protection"

func PVCList(volumes []v1.StorageVolume) []corev1.PersistentVolumeClaim {
	var pvcs []corev1.PersistentVolumeClaim
	for _, vm := range volumes {
		if pod.IsSpecialStorageClass(vm.StorageClassName) {
			continue
		}
		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: vm.Name},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: vm.StorageClassName,
			},
		}
		if vm.StorageSize != "" {
			pvc.Spec.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(vm.StorageSize),
				},
			}
		}
		pvcs = append(pvcs, pvc)
	}
	return pvcs
}

// MakeStatefulset  statefulset
func MakeStatefulset(object srobject.StarRocksObject, spec v1.SpecInterface, podTemplateSpec corev1.PodTemplateSpec) appv1.StatefulSet {
	const defaultRollingUpdateStartPod int32 = 0
	// TODO: statefulset only allow update 'replicas', 'template',  'updateStrategy'
	or := metav1.NewControllerRef(object, object.GroupVersionKind())
	st := appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            load.Name(object.AliasName, spec),
			Namespace:       object.Namespace,
			Annotations:     load.Annotations(),
			Labels:          load.Labels(object.AliasName, spec),
			OwnerReferences: []metav1.OwnerReference{*or},
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: spec.GetReplicas(),
			Selector: &metav1.LabelSelector{
				MatchLabels: load.Selector(object.AliasName, spec),
			},
			UpdateStrategy: appv1.StatefulSetUpdateStrategy{
				Type: appv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appv1.RollingUpdateStatefulSetStrategy{
					Partition: rutils.GetInt32Pointer(defaultRollingUpdateStartPod),
				},
			},
			Template:             podTemplateSpec,
			ServiceName:          service.SearchServiceName(object.AliasName, spec),
			VolumeClaimTemplates: PVCList(spec.GetStorageVolumes()),
			PodManagementPolicy:  appv1.ParallelPodManagement,
		},
	}

	// When Warehouse CR is deleted, operator need to get some environments from the statefulset to
	// execute dropping warehouse statement.
	if object.Kind == srobject.StarRocksWarehouseKind {
		st.Finalizers = append(st.Finalizers, STARROCKS_WAREHOUSE_FINALIZER)
	}

	return st
}
