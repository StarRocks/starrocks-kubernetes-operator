// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cn

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"testing"
)

var (
	sch = runtime.NewScheme()
)

func init() {
	groupVersion := schema.GroupVersion{Group: "starrocks.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	chemeBuilder := &scheme.Builder{GroupVersion: groupVersion}
	clientgoscheme.AddToScheme(sch)
	chemeBuilder.Register(&srapi.StarRocksCluster{}, &srapi.StarRocksClusterList{})
	chemeBuilder.AddToScheme(sch)
}

func Test_ClearResources(t *testing.T) {
	src := srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksCnSpec: &srapi.StarRocksCnSpec{},
		},
		Status: srapi.StarRocksClusterStatus{
			StarRocksCnStatus: &srapi.StarRocksCnStatus{
				StarRocksComponentStatus: srapi.StarRocksComponentStatus{
					ResourceNames: []string{"test-cn"},
					ServiceName:   "test-cn-access",
				},
			},
		},
	}

	st := appv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.StatefulSetKind,
			APIVersion: appv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cn",
			Namespace: "default",
		},
		Spec: appv1.StatefulSetSpec{},
	}

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cn-service",
			Namespace: "default",
		},
	}

	ssvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cn-search",
			Namespace: "default",
		},
	}
	cc := New(k8sutils.NewFakeClient(sch, &src, &st, &svc, &ssvc), record.NewFakeRecorder(10))
	cleared, err := cc.ClearResources(context.Background(), &src)
	require.Equal(t, true, cleared)
	require.Equal(t, nil, err)

	var est appv1.StatefulSet
	err = cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test-cn", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))

	var aesvc corev1.Service
	err = cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test-cn-service", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))

	var resvc corev1.Service
	err = cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test-cn-search", Namespace: "default"}, &resvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
}

func Test_Sync(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					Image:    "test.image",
					Replicas: rutils.GetInt32Pointer(3),
				},
			},
		},
	}

	ep := corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{
				IP:       "172.0.0.1",
				Hostname: "test-fe-access-01.cluster.local",
			}},
		}},
	}

	cc := New(k8sutils.NewFakeClient(sch, src, &ep), record.NewFakeRecorder(10))
	err := cc.Sync(context.Background(), src)
	cc.UpdateStatus(src)
	require.Equal(t, nil, err)
	ccStatus := src.Status.StarRocksCnStatus
	require.Equal(t, srapi.ComponentReconciling, ccStatus.Phase)

	var st appv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	cnSpec := src.Spec.StarRocksCnSpec
	require.NoError(t, cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: srapi.GetExternalServiceName(src.Name, cnSpec), Namespace: "default"}, &asvc))
	require.Equal(t, srapi.GetExternalServiceName(src.Name, cnSpec), asvc.Name)
	require.NoError(t, cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: cc.getCnSearchServiceName(src), Namespace: "default"}, &rsvc))
	require.Equal(t, cc.getCnSearchServiceName(src), rsvc.Name)
	require.NoError(t, cc.k8sclient.Get(context.Background(), types.NamespacedName{Name: statefulset.Name(src.Name, cnSpec), Namespace: "default"}, &st))
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}
