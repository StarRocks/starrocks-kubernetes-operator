package fe

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (fc *FeController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	return fc.fePodLabels(src)
}
