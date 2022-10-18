package pkg

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
)

type SubController interface {
	Sync(ctx context.Context, src *srapi.StarRocksCluster) error
	ClearResources(ctx context.Context, src *srapi.StarRocksCluster) (bool, error)
}
