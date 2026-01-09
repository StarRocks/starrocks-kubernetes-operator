package feobserver

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
)

// GetFEObserverConfig get the fe observer config from configMap.
func GetFEObserverConfig(ctx context.Context, client client.Client,
	observerSpec *srapi.StarRocksFeObserverSpec, namespace string) (map[string]interface{}, error) {
	return k8sutils.GetConfig(ctx, client, observerSpec.ConfigMapInfo,
		observerSpec.ConfigMaps, pod.GetConfigDir(observerSpec), "fe.conf",
		namespace)
}
