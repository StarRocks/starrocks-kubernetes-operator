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

package feobserver_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/feobserver"
)

func TestMain(m *testing.M) {
	srapi.Register()
	os.Exit(m.Run())
}

func TestFeObserver_ClearCluster(t *testing.T) {
	now := metav1.NewTime(time.Now())
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
		Spec: srapi.StarRocksClusterSpec{},
		Status: srapi.StarRocksClusterStatus{
			StarRocksFeObserverStatus: &srapi.StarRocksFeObserverStatus{
				StarRocksComponentStatus: srapi.StarRocksComponentStatus{
					ResourceNames: []string{"test-fe-observer"},
					ServiceName:   "test-fe-observer-access",
				},
			},
		},
	}

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      load.Name(src.Name, src.Spec.StarRocksFeObserverSpec),
			Namespace: "default",
		},
	}

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ExternalServiceName(src.Name, src.Spec.StarRocksFeObserverSpec),
			Namespace: "default",
		},
	}

	ssvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.SearchServiceName(src.Name, src.Spec.StarRocksFeObserverSpec),
			Namespace: "default",
		},
	}

	fc := feobserver.New(fake.NewFakeClient(srapi.Scheme, src, &deploy, &svc, &ssvc), fake.GetEventRecorderFor(nil))
	err := fc.ClearCluster(context.Background(), src)
	require.NoError(t, err)

	var ed appsv1.Deployment
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: deploy.Name, Namespace: "default"}, &ed)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var esvc corev1.Service
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: svc.Name, Namespace: "default"}, &esvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var essvc corev1.Service
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: ssvc.Name, Namespace: "default"}, &essvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
}

func TestFeObserver_SyncCluster(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image: "starrocks.com/cn:2.40",
					},
				},
			},
			StarRocksFeObserverSpec: &srapi.StarRocksFeObserverSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(1),
						Image:    "starrocks.com/cn:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
					},
				},
			},
		},
	}

	fc := feobserver.New(fake.NewFakeClient(srapi.Scheme, src), fake.GetEventRecorderFor(nil))
	err := fc.SyncCluster(context.Background(), src)
	require.NoError(t, err)
	err = fc.UpdateClusterStatus(context.Background(), src)
	require.NoError(t, err)

	status := src.Status.StarRocksFeObserverStatus
	require.NotNil(t, status)
	require.Equal(t, srapi.ComponentReconciling, status.Phase)
	require.Equal(t, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeObserverSpec), status.ServiceName)

	var deploy appsv1.Deployment
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: load.Name(src.Name, src.Spec.StarRocksFeObserverSpec), Namespace: "default"}, &deploy))
	var asvc corev1.Service
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: service.ExternalServiceName(src.Name, src.Spec.StarRocksFeObserverSpec), Namespace: "default"}, &asvc))
	var rsvc corev1.Service
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: service.SearchServiceName(src.Name, src.Spec.StarRocksFeObserverSpec), Namespace: "default"}, &rsvc))

	require.Equal(t, asvc.Spec.Selector, deploy.Spec.Selector.MatchLabels)
}

func TestFeObserver_Ready(t *testing.T) {
	type args struct {
		ctx              context.Context
		k8sClient        client.Client
		clusterNamespace string
		clusterName      string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test fe is not ready",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.Endpoints{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Endpoints",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-starrocks-fe-service",
						Namespace: "default",
					},
				}),
				clusterNamespace: "default",
				clusterName:      "kube-starrocks",
			},
			want: false,
		},
		{
			name: "test fe is ready",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.Endpoints{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Endpoints",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-starrocks-fe-service",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "127.0.0.1",
								},
							},
						},
					},
				}),
				clusterNamespace: "default",
				clusterName:      "kube-starrocks",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fe.CheckFEReady(tt.args.ctx, tt.args.k8sClient, tt.args.clusterNamespace, tt.args.clusterName); got != tt.want {
				t.Errorf("CheckFEReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFEObserverConfig(t *testing.T) {
	type args struct {
		ctx          context.Context
		observerSpec *srapi.StarRocksFeObserverSpec
		namespace    string
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
			name: "get config from ConfigMapInfo",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "feobserver-config",
						Namespace: "default",
					},
					Data: map[string]string{
						"fe.conf": "aa = bb",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				observerSpec: &srapi.StarRocksFeObserverSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							ConfigMapInfo: srapi.ConfigMapInfo{
								ConfigMapName: "feobserver-config",
								ResolveKey:    "fe.conf",
							},
						},
					},
				},
				namespace: "default",
			},
			want: map[string]interface{}{
				"aa": "bb",
			},
		},
		{
			name: "get config from configMaps with matching subpath",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "feobserver-config",
						Namespace: "default",
					},
					Data: map[string]string{
						"fe.conf": "cc = dd",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				observerSpec: &srapi.StarRocksFeObserverSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						ConfigMaps: []srapi.ConfigMapReference{
							{
								Name:      "feobserver-config",
								MountPath: "/opt/starrocks/fe/conf/fe.conf",
								SubPath:   "fe.conf",
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
			name: "get empty config",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:          context.Background(),
				observerSpec: &srapi.StarRocksFeObserverSpec{},
				namespace:    "default",
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := feobserver.GetFEObserverConfig(tt.args.ctx, tt.fields.k8sClient, tt.args.observerSpec, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFEObserverConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFEObserverConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
