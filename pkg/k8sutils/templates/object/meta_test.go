package object

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestNewFromCluster(t *testing.T) {
	type args struct {
		cluster *srapi.StarRocksCluster
	}
	tests := []struct {
		name string
		args args
		want StarRocksObject
	}{
		{
			name: "test NewFromCluster",
			args: args{
				cluster: &srapi.StarRocksCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "StarRocksCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "starrocks",
					},
				},
			},
			want: StarRocksObject{
				TypeMeta: &metav1.TypeMeta{
					Kind: "StarRocksCluster",
				},
				ObjectMeta: &metav1.ObjectMeta{
					Name: "starrocks",
				},
				ClusterName: "starrocks",
				Kind:        "StarRocksCluster",
				AliasName:   "starrocks",
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
		warehouse *srapi.StarRocksWarehouse
	}
	tests := []struct {
		name string
		args args
		want StarRocksObject
	}{
		{
			name: "test NewFromWarehouse",
			args: args{
				warehouse: &srapi.StarRocksWarehouse{
					TypeMeta: metav1.TypeMeta{
						Kind: "StarRocksWarehouse",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "starrocks",
					},
					Spec: srapi.StarRocksWarehouseSpec{
						StarRocksCluster: "starrocks",
					},
				},
			},
			want: StarRocksObject{
				TypeMeta: &metav1.TypeMeta{
					Kind: "StarRocksWarehouse",
				},
				ObjectMeta: &metav1.ObjectMeta{
					Name: "starrocks",
				},
				ClusterName: "starrocks",
				Kind:        "StarRocksWarehouse",
				AliasName:   "starrocks-warehouse",
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
