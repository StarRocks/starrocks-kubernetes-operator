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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/be"
)

func TestMain(_ *testing.M) {
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

	st := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.StatefulSetKind,
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.ServiceKind,
			APIVersion: appsv1.SchemeGroupVersion.String(),
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
	err := bc.ClearCluster(context.Background(), src)
	require.Equal(t, nil, err)

	var est appsv1.StatefulSet
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
	var st appsv1.StatefulSet
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

func TestBeController_GetBeConfig(t *testing.T) {
	type args struct {
		ctx       context.Context
		beSpec    *srapi.StarRocksBeSpec
		namespace string
	}
	type fields struct {
		k8sClient client.Client
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "get BE config from ConfigMapInfo",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "be-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"be.conf": "aa = bb",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				beSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							ConfigMapInfo: srapi.ConfigMapInfo{
								ConfigMapName: "be-configMap",
								ResolveKey:    "be.conf",
							},
						},
						ConfigMaps: nil,
					},
				},
				namespace: "default",
			},
			want: map[string]interface{}{
				"aa": "bb",
			},
		},
		{
			name: "get BE config from configMaps, with matching subpath",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "be-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"be.conf": "cc = dd",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				beSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						ConfigMaps: []srapi.ConfigMapReference{
							{
								Name:      "be-configMap",
								MountPath: "/opt/starrocks/be/conf/be.conf",
								SubPath:   "be.conf",
							},
						},
					},
				},
				namespace: "default",
			},
			want: map[string]interface{}{
				"cc": "dd",
			},
		},
		{
			name: "get BE config from configMap 2, without subpath",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "be-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"be.conf": "cc = dd",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				beSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						ConfigMaps: []srapi.ConfigMapReference{
							{
								Name:      "be-configMap",
								MountPath: "/opt/starrocks/be/conf",
							},
						},
					},
				},
				namespace: "default",
			},
			want: map[string]interface{}{
				"cc": "dd",
			},
		},
		{
			name: "get BE empty config",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:       context.Background(),
				beSpec:    &srapi.StarRocksBeSpec{},
				namespace: "default",
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := &be.BeController{
				Client: tt.fields.k8sClient,
			}
			got, err := be.GetBeConfig(tt.args.ctx, tt.args.beSpec, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBeConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
