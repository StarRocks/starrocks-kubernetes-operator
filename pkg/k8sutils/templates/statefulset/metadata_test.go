package statefulset

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"reflect"
	"testing"
)

func TestMakeName(t *testing.T) {
	type args struct {
		clusterName string
		spec        interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test MakeName for Be",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksBeSpec{},
			},
			want: "test-be",
		},
		{
			name: "test MakeName for Cn",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksCnSpec{},
			},
			want: "test-cn",
		},
		{
			name: "test MakeName for Fe",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: "test-fe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeName(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("MakeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLabels(t *testing.T) {
	type args struct {
		ownerReference string
		spec           interface{}
	}
	tests := []struct {
		name string
		args args
		want resource_utils.Labels
	}{
		{
			name: "test MakeLabels for Be",
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
			name: "test MakeLabels for Cn",
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
			name: "test MakeLabels for Fe",
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
			if got := MakeLabels(tt.args.ownerReference, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
