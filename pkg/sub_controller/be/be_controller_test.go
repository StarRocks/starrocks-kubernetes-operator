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

package be

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"testing"
	"time"
)

var (
	sch = runtime.NewScheme()
)

func init() {
	groupVersion := schema.GroupVersion{Group: "starrocks.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	schemeBuilder := &scheme.Builder{GroupVersion: groupVersion}
	clientgoscheme.AddToScheme(sch)
	schemeBuilder.Register(&srapi.StarRocksCluster{}, &srapi.StarRocksClusterList{})
	schemeBuilder.AddToScheme(sch)
}

func Test_ClearResources(t *testing.T) {
	now := metav1.NewTime(time.Now())
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
		Spec: srapi.StarRocksClusterSpec{},
		Status: srapi.StarRocksClusterStatus{
			StarRocksFeStatus: &srapi.StarRocksFeStatus{
				ResourceNames: []string{"test-fe"},
				ServiceName:   "test-fe-access",
			},
		},
	}

	st := appv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.StatefulSetKind,
			APIVersion: appv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: appv1.StatefulSetSpec{},
	}

	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.ServiceKind,
			APIVersion: appv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-be-access",
			Namespace: "default",
		},
	}
	ssvc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.ServiceKind,
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-be-search",
			Namespace: "default",
		},
	}

	bc := New(k8sutils.NewFakeClient(sch, src, &st, &svc, &ssvc), record.NewFakeRecorder(10))
	cleared, err := bc.ClearResources(context.Background(), src)
	require.Equal(t, true, cleared)
	require.Equal(t, nil, err)

	var est appv1.StatefulSet
	err = bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var aesvc corev1.Service
	err = bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test-be-access", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var resvc corev1.Service
	err = bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: "test-be-search", Namespace: "default"}, &resvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
}

func Test_Sync(t *testing.T) {
	requests := map[corev1.ResourceName]resource.Quantity{}
	requests["cpu"] = resource.MustParse("4")
	requests["memory"] = resource.MustParse("4Gi")
	labels := map[string]string{}
	labels["test"] = "test"
	labels["test1"] = "test1"

	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{},
			StarRocksBeSpec: &srapi.StarRocksBeSpec{
				Replicas:       rutils.GetInt32Pointer(3),
				Image:          "test.image",
				ServiceAccount: "test-sa",
				ResourceRequirements: corev1.ResourceRequirements{
					Requests: requests,
				},
				PodLabels: labels,
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
				Hostname: "test-fe-service-01.cluster.local",
			}},
		}},
	}

	bc := New(k8sutils.NewFakeClient(sch, src, &ep), record.NewFakeRecorder(10))
	err := bc.Sync(context.Background(), src)
	bc.UpdateStatus(src)
	beStatus := src.Status.StarRocksBeStatus
	require.Equal(t, beStatus.Phase, srapi.ComponentReconciling)
	require.Equal(t, nil, err)
	var st appv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	spec := src.Spec.StarRocksBeSpec
	searchServiceName := service.SearchServiceName(src.Name, spec)
	require.NoError(t, bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: srapi.GetBeExternalServiceName(src), Namespace: "default"}, &asvc))
	require.Equal(t, srapi.GetBeExternalServiceName(src), asvc.Name)
	require.NoError(t, bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: searchServiceName, Namespace: "default"}, &rsvc))
	require.Equal(t, searchServiceName, rsvc.Name)
	require.NoError(t, bc.k8sclient.Get(context.Background(), types.NamespacedName{Name: statefulset.MakeName(src.Name, src.Spec.StarRocksBeSpec), Namespace: "default"}, &st))
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}
