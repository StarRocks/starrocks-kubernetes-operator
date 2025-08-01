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
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

func TestMain(_ *testing.M) {
	srapi.Register()
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

	st := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       rutils.StatefulSetKind,
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cn",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{},
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
	cc := New(fake.NewFakeClient(srapi.Scheme, &src, &st, &svc, &ssvc), fake.GetEventRecorderFor(nil))
	err := cc.ClearCluster(context.Background(), &src)
	require.Equal(t, nil, err)

	var est appsv1.StatefulSet
	err = cc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-cn", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))

	var aesvc corev1.Service
	err = cc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-cn-service", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))

	var resvc corev1.Service
	err = cc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-cn-search", Namespace: "default"}, &resvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
}

func Test_SyncCluster(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image:    "test.image",
						Replicas: rutils.GetInt32Pointer(3),
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
				Hostname: "test-fe-access-01.cluster.local",
			}},
		}},
	}

	cc := New(fake.NewFakeClient(srapi.Scheme, src, &ep), fake.GetEventRecorderFor(nil))
	err := cc.SyncCluster(context.Background(), src)
	require.Equal(t, nil, err)
	err = cc.UpdateClusterStatus(context.Background(), src)
	require.Equal(t, nil, err)
	ccStatus := src.Status.StarRocksCnStatus
	require.Equal(t, srapi.ComponentReconciling, ccStatus.Phase)

	var st appsv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	cnSpec := src.Spec.StarRocksCnSpec
	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: service.ExternalServiceName(src.Name, cnSpec), Namespace: "default"}, &asvc))
	require.Equal(t, service.ExternalServiceName(src.Name, cnSpec), asvc.Name)
	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: service.SearchServiceName(src.Name, (*srapi.StarRocksCnSpec)(nil)), Namespace: "default"}, &rsvc))
	require.Equal(t, service.SearchServiceName(src.Name, (*srapi.StarRocksCnSpec)(nil)), rsvc.Name)
	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: load.Name(src.Name, cnSpec), Namespace: "default"}, &st))
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}

func Test_SyncWarehouse(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						ConfigMapInfo: srapi.ConfigMapInfo{
							ConfigMapName: "fe-configMap",
							ResolveKey:    "fe.conf",
						},
					},
				},
			},
			StarRocksCnSpec: &srapi.StarRocksCnSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image:    "test.image",
						Replicas: rutils.GetInt32Pointer(3),
					},
				},
			},
		},
	}

	// fe should run in shared_data mode
	feConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fe-configMap",
			Namespace: "default",
		},
		Data: map[string]string{
			"fe.conf": "run_mode = shared_data",
		},
	}

	warehouse := &srapi.StarRocksWarehouse{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksWarehouseSpec{
			StarRocksCluster: "test",
			Template: &srapi.WarehouseComponentSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image:    "test.image",
						Replicas: rutils.GetInt32Pointer(3),
					},
				},
				EnvVars:           nil,
				AutoScalingPolicy: nil,
			},
		},
		Status: srapi.StarRocksWarehouseStatus{WarehouseComponentStatus: &srapi.WarehouseComponentStatus{}},
	}

	ep := &corev1.Endpoints{
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

	srapi.Register()

	cc := New(fake.NewFakeClient(srapi.Scheme, src, feConfigMap, warehouse, ep), fake.GetEventRecorderFor(nil))
	cc.addEnvForWarehouse = true

	err := cc.SyncWarehouse(context.Background(), warehouse)
	require.Equal(t, nil, err)
	err = cc.UpdateWarehouseStatus(context.Background(), warehouse)
	require.Equal(t, nil, err)
	require.Equal(t, srapi.ComponentReconciling, warehouse.Status.Phase)

	var sts appsv1.StatefulSet
	var externalService corev1.Service
	var searchService corev1.Service
	object := object.NewFromWarehouse(warehouse)
	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{
			Name:      service.ExternalServiceName(object.SubResourcePrefixName, (*srapi.StarRocksCnSpec)(nil)),
			Namespace: "default",
		},
		&externalService),
	)
	require.Equal(t, "test-warehouse-cn-service", externalService.Name)

	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{
			Name:      service.SearchServiceName(object.SubResourcePrefixName, (*srapi.StarRocksCnSpec)(nil)),
			Namespace: "default"},
		&searchService),
	)
	require.Equal(t, "test-warehouse-cn-search", searchService.Name)

	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{
			Name:      load.Name(object.SubResourcePrefixName, (*srapi.StarRocksCnSpec)(nil)),
			Namespace: "default"},
		&sts),
	)
	require.Equal(t, "test-warehouse-cn", sts.Name)
}

func TestCnController_UpdateStatus(t *testing.T) {
	type fields struct {
		k8sClient client.Client
	}
	type args struct {
		object   object.StarRocksObject
		cnSpec   *srapi.StarRocksCnSpec
		cnStatus *srapi.StarRocksCnStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update the status of cluster",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme,
					&appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-cn",
							Namespace: "default",
						},
						Spec: appsv1.StatefulSetSpec{
							UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
								Type: appsv1.RollingUpdateStatefulSetStrategyType,
							},
						},
						Status: appsv1.StatefulSetStatus{
							ObservedGeneration: 1,
						},
					},
					&srapi.StarRocksCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
						Spec:   srapi.StarRocksClusterSpec{},
						Status: srapi.StarRocksClusterStatus{},
					},
				),
			},
			args: args{
				object: object.StarRocksObject{
					ObjectMeta: &metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					ClusterName:           "test",
					Kind:                  object.StarRocksClusterKind,
					SubResourcePrefixName: "test",
				},
				cnSpec:   &srapi.StarRocksCnSpec{},
				cnStatus: &srapi.StarRocksCnStatus{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CnController{
				k8sClient: tt.fields.k8sClient,
			}
			if err := cc.UpdateStatus(context.Background(), tt.args.object, tt.args.cnSpec, tt.args.cnStatus); (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			spew.Dump(tt.args.cnStatus)
		})
	}
}

func TestCnController_generateAutoScalerName(t *testing.T) {
	type fields struct {
		k8sClient client.Client
	}
	type args struct {
		srcName string
		cnSpec  srapi.SpecInterface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				k8sClient: nil,
			},
			args: args{
				srcName: "test",
				cnSpec:  (*srapi.StarRocksCnSpec)(nil),
			},
			want: "test-cn-autoscaler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CnController{
				k8sClient: tt.fields.k8sClient,
			}
			if got := cc.generateAutoScalerName(tt.args.srcName, tt.args.cnSpec); got != tt.want {
				t.Errorf("generateAutoScalerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCnController_GetCnConfig(t *testing.T) {
	type args struct {
		ctx       context.Context
		cnSpec    *srapi.StarRocksCnSpec
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
			name: "get CN config from ConfigMapInfo",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cn-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"cn.conf": "aa = bb",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				cnSpec: &srapi.StarRocksCnSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							ConfigMapInfo: srapi.ConfigMapInfo{
								ConfigMapName: "cn-configMap",
								ResolveKey:    "cn.conf",
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
			name: "get CN config from configMaps, with matching subpath",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cn-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"cn.conf": "cc = dd",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				cnSpec: &srapi.StarRocksCnSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						ConfigMaps: []srapi.ConfigMapReference{
							{
								Name:      "cn-configMap",
								MountPath: "/opt/starrocks/cn/conf/cn.conf",
								SubPath:   "cn.conf",
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
			name: "get CN config from configMap 2, without subpath",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cn-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"cn.conf": "cc = dd",
					},
				}),
			},
			args: args{
				ctx: context.Background(),
				cnSpec: &srapi.StarRocksCnSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						ConfigMaps: []srapi.ConfigMapReference{
							{
								Name:      "cn-configMap",
								MountPath: "/opt/starrocks/cn/conf",
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
			name: "get CN empty config",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:       context.Background(),
				cnSpec:    &srapi.StarRocksCnSpec{},
				namespace: "default",
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CnController{
				k8sClient: tt.fields.k8sClient,
			}
			got, err := cc.GetCnConfig(tt.args.ctx, tt.args.cnSpec, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCnConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCnConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCnController_SyncComputeNodesInFE(t *testing.T) {
	type fields struct {
		k8sClient client.Client
	}
	type args struct {
		ctx          context.Context
		object       object.StarRocksObject
		expectSTS    *appsv1.StatefulSet
		actualSTS    *appsv1.StatefulSet
		actualCNPods *corev1.PodList
		db           *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "the replicas is not equal between expected sts and actual sts",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:    context.Background(),
				object: object.StarRocksObject{},
				expectSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(3),
					},
				},
				actualSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(4),
					},
				},
				actualCNPods: nil,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ErrReplicasNotEqual) {
					t.Errorf("SyncComputeNodesInFE() expected ErrReplicasNotEqual, got %v, args: %v", err, i)
					return false
				}
				return true
			},
		},
		{
			name: "the replicas is not equal between expected pods and actual pods",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:    context.Background(),
				object: object.StarRocksObject{},
				expectSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
				actualSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
				actualCNPods: &corev1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-cn-0",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-cn-1",
							},
						},
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ErrReplicasNotEqual) {
					t.Errorf("SyncComputeNodesInFE() expected ErrReplicasNotEqual, got %v, args: %v", err, i)
					return false
				}
				return true
			},
		},
		{
			name: "the hash value is not equal between expected sts and actual sts",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:    context.Background(),
				object: object.StarRocksObject{},
				expectSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
				actualSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
						Labels: map[string]string{
							"extra-label": "value",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ErrHashValueNotEqual) {
					t.Errorf("SyncComputeNodesInFE() expected ErrHashValueNotEqual, got %v, args: %v", err, i)
					return false
				}
				return true
			},
		},
		{
			name: "the hash value is not equal between expected pods and actual pods",
			fields: fields{
				k8sClient: fake.NewFakeClient(srapi.Scheme),
			},
			args: args{
				ctx:    context.Background(),
				object: object.StarRocksObject{},
				expectSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
				actualSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
					Status: appsv1.StatefulSetStatus{
						UpdateRevision: "v1",
					},
				},
				actualCNPods: &corev1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-cn-0",
								Labels: map[string]string{
									"controller-revision-hash": "v2",
								},
							},
						},
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ErrHashValueNotEqual) {
					t.Errorf("SyncComputeNodesInFE() expected ErrHashValueNotEqual, got %v, args: %v", err, i)
					return false
				}
				return true
			},
		},
		{
			name: "xxx",
			fields: fields{
				k8sClient: fake.NewFakeClient(
					func() *runtime.Scheme {
						schema := runtime.NewScheme()
						_ = clientgoscheme.AddToScheme(schema)
						return schema
					}(),
					&appsv1.StatefulSet{
						TypeMeta: metav1.TypeMeta{
							Kind:       "StatefulSet",
							APIVersion: appsv1.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "wh1-warehouse-cn",
							Namespace: "default",
						},
						Spec: appsv1.StatefulSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Env: []corev1.EnvVar{
												{
													Name:  "MYSQL_PWD",
													Value: "123456",
												},
												{
													Name:  "FE_SERVICE_NAME",
													Value: "fe",
												},
												{
													Name:  "FE_QUERY_PORT",
													Value: "9030",
												},
											},
										},
									},
								},
							},
						},
					},
				),
			},
			args: args{
				ctx: context.Background(),
				object: object.StarRocksObject{
					ObjectMeta: &metav1.ObjectMeta{
						Name:      "wh1",
						Namespace: "default",
					},
					ClusterName:           "cluster",
					SubResourcePrefixName: "wh1-warehouse",
					IsWarehouseObject:     true,
				},
				expectSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "wh1-warehouse-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
				},
				actualSTS: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.StatefulSetKind,
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "wh1-warehouse-cn",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: rutils.GetInt32Pointer(1),
					},
					Status: appsv1.StatefulSetStatus{
						UpdateRevision: "v2",
					},
				},
				actualCNPods: &corev1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "wh1-warehouse-cn-0",
								Labels: map[string]string{
									"controller-revision-hash": "v2",
								},
							},
						},
					},
				},

				db: func() *sql.DB {
					db, mock, err := sqlmock.New()
					require.NoError(t, err)
					mock.ExpectQuery(ShowComputeNodesStatement).WillReturnRows(
						sqlmock.NewRows([]string{"ComputeNodeId", "IP", "WarehouseName"}).AddRow([]byte("1"), []byte("kube-starrocks-fe-0.kube-starrocks-fe-search.default.svc.cluster.local_9010_1751367053833"), []byte("wh1")),
					)
					return db
				}(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if err != nil {
					t.Errorf("SyncComputeNodesInFE() unexpected error: %v, args: %v", err, i)
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CnController{
				k8sClient: tt.fields.k8sClient,
			}

			tt.wantErr(t, cc.SyncComputeNodesInFE(tt.args.ctx, tt.args.object, tt.args.expectSTS, tt.args.actualSTS, tt.args.actualCNPods, tt.args.db), fmt.Sprintf("SyncComputeNodesInFE(%v, %v, %v, %v, %v)", tt.args.ctx, tt.args.object, tt.args.expectSTS, tt.args.actualSTS, tt.args.actualCNPods))
		})
	}
}
