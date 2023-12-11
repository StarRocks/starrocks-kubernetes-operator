package fake

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

type WithCRD func() *apiextensionsv1.CustomResourceDefinition

var (
	clusterCRD = &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "starrocksclusters.starrocks.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "starrocks.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "starrocksclusters",
				Singular: "starrockscluster",
				Kind:     "StarRocksCluster",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}

	warehouseCRD = &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "starrockswarehouses.starrocks.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "starrocks.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "starrockswarehouses",
				Singular: "starrockswarehouse",
				Kind:     "StarRocksWarehouse",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}
)

// NewEnvironment is used to create a fake manager, code inspired from pkg/manager/internal/integration/manager_test.go
// see https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/manager/internal/integration/manager_test.go
// for more information
func NewEnvironment(crds ...WithCRD) *envtest.Environment {
	env := &envtest.Environment{
		Scheme:            v1.Scheme,
		CRDInstallOptions: envtest.CRDInstallOptions{},
	}
	for i := range crds {
		env.CRDInstallOptions.CRDs = append(env.CRDInstallOptions.CRDs, crds[i]())
	}
	return env
}

func WithClusterCRD() WithCRD {
	return func() *apiextensionsv1.CustomResourceDefinition {
		return clusterCRD
	}
}

func WithWarehouseCRD() WithCRD {
	return func() *apiextensionsv1.CustomResourceDefinition {
		return warehouseCRD
	}
}

func NewManager(env *envtest.Environment) ctrl.Manager {
	_, err := env.Start()
	if err != nil {
		panic(err)
	}

	mgr, err := manager.New(env.Config, manager.Options{
		MetricsBindAddress: "0",
		Scheme:             v1.Scheme,
	})
	if err != nil {
		panic(err)
	}
	return mgr
}
