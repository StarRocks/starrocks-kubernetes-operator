package object

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StarRocksClusterKind   = "StarRocksCluster"
	StarRocksWarehouseKind = "StarRocksWarehouse"
)

type StarRocksObject struct {
	*metav1.TypeMeta
	*metav1.ObjectMeta

	// ClusterName is the name of StarRocksCluster.
	ClusterName string

	// Kind is StarRocksCluster or StarRocksWarehouse.
	// The reason why we need this field is that we can't make sure ObjectMeta.Kind is filled.
	Kind string

	// AliasName represents the prefix of subresource names for cn component. The reason is that when the name of
	// StarRocksWarehouse is the same as the name of StarRocksCluster, operator should avoid to create the same name
	// StatefulSet, Service, etc.
	AliasName string
}

func NewFromCluster(cluster *srapi.StarRocksCluster) StarRocksObject {
	return StarRocksObject{
		TypeMeta:    &cluster.TypeMeta,
		ObjectMeta:  &cluster.ObjectMeta,
		ClusterName: cluster.Name,
		Kind:        StarRocksClusterKind,
		AliasName:   cluster.Name,
	}
}

func NewFromWarehouse(warehouse *srapi.StarRocksWarehouse) StarRocksObject {
	return StarRocksObject{
		TypeMeta:    &warehouse.TypeMeta,
		ObjectMeta:  &warehouse.ObjectMeta,
		ClusterName: warehouse.Spec.StarRocksCluster,
		Kind:        StarRocksWarehouseKind,
		// add warehouse suffix to avoid name conflict with cluster
		AliasName: GetAliasName(warehouse.Name),
	}
}

func GetAliasName(warehouseName string) string {
	return warehouseName + "-warehouse"
}
