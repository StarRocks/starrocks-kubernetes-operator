package pkg

import (
	"context"
	starrockscomv1alpha1 "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
)

type SubController interface {
	Sync(ctx context.Context, src *starrockscomv1alpha1.StarRocksCluster) error
}
