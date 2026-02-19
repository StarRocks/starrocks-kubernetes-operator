package object

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/k8sutils/load"
)

const (
	CelerDataClusterKind   = "CelerDataCluster"
	CelerDataWarehouseKind = "CelerDataWarehouse"
)

// StarRocksObject is a wrapper of metav1.TypeMeta and metav1.ObjectMeta for CelerDataCluster and CelerDataWarehouse.
type StarRocksObject struct {
	*metav1.TypeMeta
	*metav1.ObjectMeta

	// ClusterName is the name of CelerDataCluster.
	ClusterName string

	// Kind is CelerDataCluster or CelerDataWarehouse.
	// The reason why we need this field is that we can't make sure ObjectMeta.Kind is filled.
	Kind string

	// SubResourcePrefixName represents the prefix of subresource names for cn component. The reason is that when the name of
	// CelerDataWarehouse is the same as the name of CelerDataCluster, operator should avoid to create the same name
	// StatefulSet, Service, etc.
	SubResourcePrefixName string

	// IsWarehouseObject indicates whether this object is a CelerDataWarehouse object.
	IsWarehouseObject bool
}

func NewFromCluster(cluster *cdapi.CelerDataCluster) StarRocksObject {
	return StarRocksObject{
		TypeMeta:              &cluster.TypeMeta,
		ObjectMeta:            &cluster.ObjectMeta,
		ClusterName:           cluster.Name,
		Kind:                  CelerDataClusterKind,
		SubResourcePrefixName: cluster.Name,
		IsWarehouseObject:     false,
	}
}

func NewFromWarehouse(warehouse *cdapi.CelerDataWarehouse) StarRocksObject {
	return StarRocksObject{
		TypeMeta:              &warehouse.TypeMeta,
		ObjectMeta:            &warehouse.ObjectMeta,
		ClusterName:           warehouse.Spec.CelerDataCluster,
		Kind:                  CelerDataWarehouseKind,
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
	return load.Name(object.SubResourcePrefixName, (*cdapi.CelerDataCnSpec)(nil))
}

func (object *StarRocksObject) GetWarehouseNameInFE() string {
	return GetWarehouseNameInFE(object.Name())
}

// GetWarehouseNameInFE warehouseName is the name defined in k8s
func GetWarehouseNameInFE(warehouseName string) string {
	return strings.ReplaceAll(warehouseName, "-", "_")
}
