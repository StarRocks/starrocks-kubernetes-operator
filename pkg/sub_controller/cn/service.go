package cn

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (cc *CnController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	return cc.cnStatefulsetSelector(src)
}

// generateServiceLabels generate service labels for user.
func (cc *CnController) generateServiceLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_CN
	//once src labels updated, the statefulset will enter into a not reconcile state.
	//labels.AddLabel(src.Labels)
	return labels
}
