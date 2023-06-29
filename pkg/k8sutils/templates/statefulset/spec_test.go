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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestMakePVCList(t *testing.T) {
	type args struct {
		volumes []v1.StorageVolume
	}
	tests := []struct {
		name string
		args args
		want []corev1.PersistentVolumeClaim
	}{
		{
			name: "test PVCList",
			args: args{
				volumes: []v1.StorageVolume{
					{
						Name:             "test",
						StorageClassName: func() *string { name := "test"; return &name }(),
						StorageSize:      "1Gi",
					},
				},
			},
			want: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						StorageClassName: &[]string{"test"}[0],
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PVCList(tt.args.volumes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PVCList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeSelector(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want resource_utils.Labels
	}{
		{
			name: "test Selector",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: resource_utils.Labels{
				v1.OwnerReference:    "test-fe",
				v1.ComponentLabelKey: "fe",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Selector(tt.args.clusterName, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Selector() = %v, want %v", got, tt.want)
			}
		})
	}
}
