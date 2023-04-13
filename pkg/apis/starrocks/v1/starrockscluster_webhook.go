package v1

import (
	"errors"
	"fmt"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-starrocks-com-v1-starrockscluster,mutating=false,failurePolicy=ignore,resources=starrockscluster,verbs=create;update,versions=v1;v1alpha1,name=starrockscluster.starrocks.com,groups=starrocks.com,sideEffects=None,admissionReviewVersions=v1;v1alpha1,matchPolicy=Exact

var _ webhook.Validator = &StarRocksCluster{}

func (src *StarRocksCluster) ValidateCreate() error {

	return nil
}

func (src *StarRocksCluster) ValidateUpdate(old runtime.Object) error {
	fmt.Println("test")
	oldsr, ok := old.(*StarRocksCluster)
	if !ok {
		return errors.New("cannot cast old object to StarRocksCluster type")
	}
	if err := src.validateUpdateFE(oldsr.Spec.StarRocksFeSpec); err != nil {
		return err
	}
	if err := src.validateUpdateBE(oldsr.Spec.StarRocksBeSpec); err != nil {
		return err
	}
	if err := src.validateUpdateCN(oldsr.Spec.StarRocksCnSpec); err != nil {
		return err
	}

	return nil
}

func (src *StarRocksCluster) validateUpdateFE(oldFESpec *StarRocksFeSpec) error {
	if oldFESpec == nil || src.Spec.StarRocksFeSpec == nil {
		return nil
	}

	if apiequality.Semantic.DeepEqual(oldFESpec.StorageVolumes, src.Spec.StarRocksFeSpec.StorageVolumes) {
		return nil
	}

	return errors.New("the starRocksFeSpec not allow update storageVolumes")
}

func (src *StarRocksCluster) validateUpdateBE(oldBeSpec *StarRocksBeSpec) error {
	if oldBeSpec == nil || src.Spec.StarRocksBeSpec == nil {
		return nil
	}

	if apiequality.Semantic.DeepEqual(oldBeSpec.StorageVolumes, src.Spec.StarRocksBeSpec.StorageVolumes) {
		return nil
	}

	return errors.New("the starRocksBeSpec not allow update storageVolumes")
}

func (src *StarRocksCluster) validateUpdateCN(oldCnSpec *StarRocksCnSpec) error {

	return nil
}

func (src *StarRocksCluster) ValidateDelete() error {
	return nil
}
