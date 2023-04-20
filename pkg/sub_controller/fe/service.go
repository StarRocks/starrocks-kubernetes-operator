package fe

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
)

func (fc *FeController) generateServiceSelector(src *srapi.StarRocksCluster) rutils.Labels {
	return fc.feStatefulsetSelector(src)
}

//GenerateServiceLabels generate service labels for user list.
func (fc *FeController) generateServiceLabels(src *srapi.StarRocksCluster) rutils.Labels {
	labels := rutils.Labels{}
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	//once the labels updated, the statefulset will enter into a not reconcile state.
	//labels.AddLabel(src.Labels)
	return labels
}
