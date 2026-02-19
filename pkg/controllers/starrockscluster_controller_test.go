/*
Copyright 2021-present, StarRocks Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cdapi "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
	rutils "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/common/resource_utils"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/k8sutils/fake"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/be"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/cn"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/fe"
	"github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/subcontrollers/feproxy"
)

func newCelerDataClusterController(objects ...runtime.Object) *CelerDataClusterReconciler {
	srcController := &CelerDataClusterReconciler{}
	srcController.Recorder = record.NewFakeRecorder(10)
	srcController.Client = fake.NewFakeClient(cdapi.Scheme, objects...)
	srcController.Scs = []subcontrollers.ClusterSubController{
		fe.New(srcController.Client, fake.GetEventRecorderFor(nil)),
		be.New(srcController.Client, fake.GetEventRecorderFor(srcController.Recorder)),
		cn.New(srcController.Client, fake.GetEventRecorderFor(nil)),
		feproxy.New(srcController.Client, fake.GetEventRecorderFor(nil)),
	}
	return srcController
}

func TestReconcileConstructFeResource(t *testing.T) {
	src := &cdapi.CelerDataCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "celerdatacluster-sample",
			Namespace: "default",
		},
		Spec: cdapi.CelerDataClusterSpec{
			CelerDataFeSpec: &cdapi.CelerDataFeSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "celerdata.com/fe:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
						StorageVolumes: []cdapi.StorageVolume{
							{
								Name:             "fe-meta",
								StorageClassName: rutils.GetStringPointer("shard-data"),
								MountPath:        "/data/fe/meta",
								StorageSize:      "10Gi",
							},
						},
					},
				},
			},
			CelerDataCnSpec: &cdapi.CelerDataCnSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "test",
					},
				},
			},
			CelerDataBeSpec: &cdapi.CelerDataBeSpec{
				CelerDataComponentSpec: cdapi.CelerDataComponentSpec{
					CelerDataLoadSpec: cdapi.CelerDataLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "test",
					},
				},
			},
		},
	}

	r := newCelerDataClusterController(src)
	res, err := r.Reconcile(context.Background(),
		reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "celerdatacluster-sample"}})
	require.NoError(t, err)
	require.Equal(t, reconcile.Result{}, res)
}
