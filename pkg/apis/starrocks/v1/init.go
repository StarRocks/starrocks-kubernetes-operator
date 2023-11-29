package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	Scheme = runtime.NewScheme()
)

func Register() {
	SchemeBuilder.Register(&StarRocksCluster{}, &StarRocksClusterList{})
	SchemeBuilder.Register(&StarRocksWarehouse{}, &StarRocksWarehouseList{})

	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(AddToScheme(Scheme))
}
