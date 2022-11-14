package resource_utils

import (
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strconv"
)

type StatefulSetType string

type StatefulSetParams struct {
	Name                 string
	Namespace            string
	ServiceName          string
	StatefulSetType      string
	Selector             map[string]string
	Labels               Labels
	OwnerReferences      []metav1.OwnerReference
	Annotations          map[string]string
	PodTemplateSpec      corev1.PodTemplateSpec
	RevisionHistoryLimit *int32
	Replicas             *int32
	VolumeClaimTemplates []corev1.PersistentVolumeClaim
}

//NewStatefulset  statefulset
func NewStatefulset(params StatefulSetParams) appv1.StatefulSet {
	// TODO: statefulset only allow update 'replicas', 'template',  'updateStrategy'
	st := appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            params.Name,
			Namespace:       params.Namespace,
			Labels:          params.Labels,
			Annotations:     params.Annotations,
			OwnerReferences: params.OwnerReferences,
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: params.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: params.Selector,
			},
			UpdateStrategy: appv1.StatefulSetUpdateStrategy{
				Type: appv1.RollingUpdateStatefulSetStrategyType,
			},
			Template:             params.PodTemplateSpec,
			ServiceName:          params.ServiceName,
			VolumeClaimTemplates: params.VolumeClaimTemplates,
		},
	}

	hst := statefulSetHashObject(&st, false)
	hvalue := hash.HashObject(hst)

	anno := Annotations{}
	anno.AddAnnotation(params.Annotations)
	anno[srapi.ComponentResourceHash] = hvalue
	anno[srapi.ComponentGeneration] = "1"
	st.Annotations = anno
	return st
}

//hashStatefulsetObject contains the info for hash comparison.
type hashStatefulsetObject struct {
	name                 string
	namespace            string
	labels               map[string]string
	selector             metav1.LabelSelector
	podSpec              corev1.PodSpec
	serviceName          string
	volumeClaimTemplates []corev1.PersistentVolumeClaim
	replicas             int32
}

//StatefulsetHashObject construct the hash spec for deep equals to exist statefulset.
func statefulSetHashObject(st *appv1.StatefulSet, excludeReplica bool) hashStatefulsetObject {
	//set -1 for the initial is zero.
	replicas := int32(-1)
	if !excludeReplica {
		if st.Spec.Replicas != nil {
			replicas = *st.Spec.Replicas
		}
	}
	selector := metav1.LabelSelector{}
	if st.Spec.Selector != nil {
		selector = *st.Spec.Selector
	}

	return hashStatefulsetObject{
		name:                 st.Name,
		namespace:            st.Namespace,
		labels:               st.Labels,
		selector:             selector,
		podSpec:              st.Spec.Template.Spec,
		serviceName:          st.Spec.ServiceName,
		volumeClaimTemplates: st.Spec.VolumeClaimTemplates,
		replicas:             replicas,
	}
}

//StatefulSetDeepEqual judge two statefulset equal or not.
func StatefulSetDeepEqual(new *appv1.StatefulSet, old appv1.StatefulSet, excludeReplicas bool) bool {
	var newHashv, oldHashv string

	if _, ok := new.Annotations[srapi.ComponentResourceHash]; ok {
		newHashv = new.Annotations[srapi.ComponentResourceHash]
	} else {
		newHso := statefulSetHashObject(new, excludeReplicas)
		newHashv = hash.HashObject(newHso)
	}

	if _, ok := old.Annotations[srapi.ComponentResourceHash]; ok {
		oldHashv = old.Annotations[srapi.ComponentResourceHash]
	} else {
		oldHso := statefulSetHashObject(&old, excludeReplicas)
		oldHashv = hash.HashObject(oldHso)
	}

	var oldGeneration int64
	if _, ok := old.Annotations[srapi.ComponentGeneration]; ok {
		oldGeneration, _ = strconv.ParseInt(old.Annotations[srapi.ComponentGeneration], 10, 64)
	}

	new.Annotations[srapi.ComponentGeneration] = strconv.FormatInt(old.Generation+1, 10)
	klog.Info("the statefulset new hash value ", newHashv, " old have value ", oldHashv, " oldGeneration ", oldGeneration, " new Generation ", old.Generation)
	//avoid the update from kubectl.
	return newHashv == oldHashv && new.Name == old.Name &&
		new.Namespace == old.Namespace &&
		oldGeneration == old.Generation
}

//MergeStatefulSets merge exist statefulset and new statefulset.
func MergeStatefulSets(new *appv1.StatefulSet, old appv1.StatefulSet) {
	MergeMetadata(&new.ObjectMeta, old.ObjectMeta)
}
