package be

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (be *BeController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	//match the selector can control numbers.
	return be.beStatefulsetSelector(src)
}

//generateServiceLabels generate service labels for user.
func (be *BeController) generateServiceLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_BE
	//once the src labels updated, the statefulset will enter into a can't be modified state.
	//labels.AddLabel(src.Labels)
	return labels
}
