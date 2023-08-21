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

package load

import (
	"reflect"
	"testing"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func TestMakeName(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test Name for Be",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksBeSpec{},
			},
			want: "test-be",
		},
		{
			name: "test Name for Cn",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksCnSpec{},
			},
			want: "test-cn",
		},
		{
			name: "test Name for Fe",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: "test-fe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Name(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLabels(t *testing.T) {
	type args struct {
		ownerReference string
		spec           v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want resource_utils.Labels
	}{
		{
			name: "test Labels for Be",
			args: args{
				ownerReference: "test",
				spec:           &v1.StarRocksBeSpec{},
			},
			want: resource_utils.Labels{
				v1.OwnerReference:    "test",
				v1.ComponentLabelKey: "be",
			},
		},
		{
			name: "test Labels for Cn",
			args: args{
				ownerReference: "test",
				spec:           &v1.StarRocksCnSpec{},
			},
			want: resource_utils.Labels{
				v1.OwnerReference:    "test",
				v1.ComponentLabelKey: "cn",
			},
		},
		{
			name: "test Labels for Fe",
			args: args{
				ownerReference: "test",
				spec:           &v1.StarRocksFeSpec{},
			},
			want: resource_utils.Labels{
				v1.OwnerReference:    "test",
				v1.ComponentLabelKey: "fe",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Labels(tt.args.ownerReference, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Labels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnotations(t *testing.T) {
	tests := []struct {
		name string
		want map[string]string
	}{
		{
			name: "test Annotations",
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Annotations(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Annotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
