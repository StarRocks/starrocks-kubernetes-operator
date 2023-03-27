package be

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (be *BeController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	//match the selector can control numbers.
	return be.beStatefulsetSelector(src)
}
