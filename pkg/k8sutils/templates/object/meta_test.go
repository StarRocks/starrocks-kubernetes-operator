package object

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
)

func TestNewFromCluster(t *testing.T) {
	type args struct {
		cluster *cdapi.CelerDataCluster
	}
	tests := []struct {
		name string
		args args
		want StarRocksObject
	}{
		{
			name: "test NewFromCluster",
			args: args{
				cluster: &cdapi.CelerDataCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "CelerDataCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "celerdata",
					},
				},
			},
			want: StarRocksObject{
				TypeMeta: &metav1.TypeMeta{
					Kind: "CelerDataCluster",
				},
				ObjectMeta: &metav1.ObjectMeta{
					Name: "celerdata",
				},
				ClusterName:           "celerdata",
				Kind:                  "CelerDataCluster",
				SubResourcePrefixName: "celerdata",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFromCluster(tt.args.cluster); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFromWarehouse(t *testing.T) {
	type args struct {
		warehouse *cdapi.CelerDataWarehouse
	}
	tests := []struct {
		name string
		args args
		want StarRocksObject
	}{
		{
			name: "test NewFromWarehouse",
			args: args{
				warehouse: &cdapi.CelerDataWarehouse{
					TypeMeta: metav1.TypeMeta{
						Kind: "CelerDataWarehouse",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "celerdata",
					},
					Spec: cdapi.CelerDataWarehouseSpec{
						CelerDataCluster: "celerdata",
					},
				},
			},
			want: StarRocksObject{
				TypeMeta: &metav1.TypeMeta{
					Kind: "CelerDataWarehouse",
				},
				ObjectMeta: &metav1.ObjectMeta{
					Name: "celerdata",
				},
				ClusterName:           "celerdata",
				Kind:                  "CelerDataWarehouse",
				SubResourcePrefixName: "celerdata-warehouse",
				IsWarehouseObject:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFromWarehouse(tt.args.warehouse); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromWarehouse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStarRocksObject_GetCNStatefulSetName(t *testing.T) {
	type fields struct {
		ClusterName       string
		Kind              string
		AliasName         string
		IsWarehouseObject bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test GetCNStatefulSetName for CelerDataCluster",
			fields: fields{
				ClusterName:       "cluster",
				Kind:              "CelerDataCluster",
				AliasName:         "cluster",
				IsWarehouseObject: false,
			},
			want: "cluster-cn",
		},
		{
			name: "test GetCNStatefulSetName for Warehouse",
			fields: fields{
				ClusterName:       "cluster",
				Kind:              "CelerDataCluster",
				AliasName:         "wh1-warehouse",
				IsWarehouseObject: true,
			},
			want: "wh1-warehouse-cn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object := &StarRocksObject{
				ClusterName:           tt.fields.ClusterName,
				Kind:                  tt.fields.Kind,
				SubResourcePrefixName: tt.fields.AliasName,
				IsWarehouseObject:     tt.fields.IsWarehouseObject,
			}
			if got := object.GetCNStatefulSetName(); got != tt.want {
				t.Errorf("GetCNStatefulSetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStarRocksObject_GetWarehouseNameInFE(t *testing.T) {
	type fields struct {
		ObjectMeta        *metav1.ObjectMeta
		ClusterName       string
		Kind              string
		AliasName         string
		IsWarehouseObject bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				ObjectMeta: &metav1.ObjectMeta{
					Name: "wh-1",
				},
				ClusterName:       "cluster",
				Kind:              "CelerDataCluster",
				AliasName:         "wh-1-warehouse",
				IsWarehouseObject: true,
			},
			want: "wh_1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object := &StarRocksObject{
				ObjectMeta:            tt.fields.ObjectMeta,
				ClusterName:           tt.fields.ClusterName,
				Kind:                  tt.fields.Kind,
				SubResourcePrefixName: tt.fields.AliasName,
				IsWarehouseObject:     tt.fields.IsWarehouseObject,
			}
			if got := object.GetWarehouseNameInFE(); got != tt.want {
				t.Errorf("GetWarehouseNameInFE() = %v, want %v", got, tt.want)
			}
		})
	}
}
