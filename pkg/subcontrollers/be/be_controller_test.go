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

package be_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
)

func init() {
	srapi.Register()
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
				StarRocksComponentStatus: srapi.StarRocksComponentStatus{
					ResourceNames: []string{"test-fe"},
					ServiceName:   "test-fe-access",
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

	bc := be.New(fake.NewFakeClient(srapi.Scheme, src, &st, &svc, &ssvc), fake.GetEventRecorderFor(nil))
	err := bc.ClearResources(context.Background(), src)
	require.Equal(t, nil, err)

	var est appv1.StatefulSet
	err = bc.Client.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var aesvc corev1.Service
	err = bc.Client.Get(context.Background(), types.NamespacedName{Name: "test-be-access", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var resvc corev1.Service
	err = bc.Client.Get(context.Background(), types.NamespacedName{Name: "test-be-search", Namespace: "default"}, &resvc)
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
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas:       rutils.GetInt32Pointer(3),
						Image:          "test.image",
						ServiceAccount: "test-sa",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: requests,
						},
						PodLabels: labels,
					},
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
				Hostname: "test-fe-service-01.cluster.local",
			}},
		}},
	}

	bc := be.New(fake.NewFakeClient(srapi.Scheme, src, &ep), fake.GetEventRecorderFor(nil))
	err := bc.SyncCluster(context.Background(), src)
	require.Equal(t, nil, err)
	err = bc.UpdateClusterStatus(context.Background(), src)
	require.Equal(t, nil, err)
	beStatus := src.Status.StarRocksBeStatus
	require.Equal(t, beStatus.Phase, srapi.ComponentReconciling)
	require.Equal(t, nil, err)
	var st appv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	spec := src.Spec.StarRocksBeSpec
	searchServiceName := service.SearchServiceName(src.Name, spec)
	require.NoError(t, bc.Client.Get(context.Background(),
		types.NamespacedName{Name: service.ExternalServiceName(src.Name, spec), Namespace: "default"}, &asvc))
	require.Equal(t, service.ExternalServiceName(src.Name, spec), asvc.Name)
	require.NoError(t, bc.Client.Get(context.Background(),
		types.NamespacedName{Name: searchServiceName, Namespace: "default"}, &rsvc))
	require.Equal(t, searchServiceName, rsvc.Name)
	require.NoError(t, bc.Client.Get(context.Background(),
		types.NamespacedName{Name: load.Name(src.Name, src.Spec.StarRocksBeSpec), Namespace: "default"}, &st))
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}
