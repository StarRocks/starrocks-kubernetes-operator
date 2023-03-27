package cn

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (cc *CnController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	return cc.cnStatefulsetSelector(src)
}
