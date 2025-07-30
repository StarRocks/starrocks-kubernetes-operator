package object

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

const (
	StarRocksClusterKind   = "StarRocksCluster"
	StarRocksWarehouseKind = "StarRocksWarehouse"
)

// StarRocksObject is a wrapper of metav1.TypeMeta and metav1.ObjectMeta for StarRocksCluster and StarRocksWarehouse.
type StarRocksObject struct {
	*metav1.TypeMeta
	*metav1.ObjectMeta

	// ClusterName is the name of StarRocksCluster.
	ClusterName string

	// Kind is StarRocksCluster or StarRocksWarehouse.
	// The reason why we need this field is that we can't make sure ObjectMeta.Kind is filled.
	Kind string

	// SubResourcePrefixName represents the prefix of subresource names for cn component. The reason is that when the name of
	// StarRocksWarehouse is the same as the name of StarRocksCluster, operator should avoid to create the same name
	// StatefulSet, Service, etc.
	SubResourcePrefixName string

	// IsWarehouseObject indicates whether this object is a StarRocksWarehouse object.
	IsWarehouseObject bool
}

func NewFromCluster(cluster *srapi.StarRocksCluster) StarRocksObject {
	return StarRocksObject{
		TypeMeta:              &cluster.TypeMeta,
		ObjectMeta:            &cluster.ObjectMeta,
		ClusterName:           cluster.Name,
		Kind:                  StarRocksClusterKind,
		SubResourcePrefixName: cluster.Name,
		IsWarehouseObject:     false,
	}
}

func NewFromWarehouse(warehouse *srapi.StarRocksWarehouse) StarRocksObject {
	return StarRocksObject{
		TypeMeta:              &warehouse.TypeMeta,
		ObjectMeta:            &warehouse.ObjectMeta,
		ClusterName:           warehouse.Spec.StarRocksCluster,
		Kind:                  StarRocksWarehouseKind,
		SubResourcePrefixName: GetPrefixNameForWarehouse(warehouse.Name),
		IsWarehouseObject:     true,
	}
}

func GetPrefixNameForWarehouse(warehouseName string) string {
	return warehouseName + "-warehouse"
}

// Name The reason why we need this method is that we don't want user to use object.Name directly.
// In a warehouse situation, most of the time you should use SubResourcePrefixName, not Name.
func (object *StarRocksObject) Name() string {
	return object.ObjectMeta.Name
}

func (object *StarRocksObject) GetCNStatefulSetName() string {
	return load.Name(object.SubResourcePrefixName, (*srapi.StarRocksCnSpec)(nil))
}

func (object *StarRocksObject) GetWarehouseNameInFE() string {
	return GetWarehouseNameInFE(object.Name())
}

// GetWarehouseNameInFE warehouseName is the name defined in k8s
func GetWarehouseNameInFE(warehouseName string) string {
	return strings.ReplaceAll(warehouseName, "-", "_")
}
