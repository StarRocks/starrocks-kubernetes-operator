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
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
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

	src.Spec.StarRocksFeSpec = &srapi.StarRocksFeSpec{
		ObserverSpec: &srapi.StarRocksFeObserverSpec{
			Enabled: true,
		},
	}

	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-observer",
			Namespace: "default",
		},
	}

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-observer-service",
			Namespace: "default",
		},
	}

	ssvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-observer-search",
			Namespace: "default",
		},
	}

	fc := feobserver.New(fake.NewFakeClient(srapi.Scheme, src, &sts, &svc, &ssvc), fake.GetEventRecorderFor(nil))
	err := fc.ClearCluster(context.Background(), src)
	require.NoError(t, err)

	var est appsv1.StatefulSet
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: sts.Name, Namespace: "default"}, &est)
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
						Replicas: rutils.GetInt32Pointer(3),
						Image:    "starrocks.com/fe:2.40",
						ResourceRequirements: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory: resource.MustParse("16G"),
							},
						},
					},
				},
				ObserverSpec: &srapi.StarRocksFeObserverSpec{
					Enabled:        true,
					ObserverNumber: rutils.GetInt32Pointer(2),
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
	require.Equal(t, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), status.ServiceName)

	var sts appsv1.StatefulSet
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: "test-fe-observer", Namespace: "default"}, &sts))
	require.Equal(t, int32(2), *sts.Spec.Replicas)
	var observerExternalSvc corev1.Service
	err = fc.Client.Get(context.Background(),
		types.NamespacedName{Name: "test-fe-observer-service", Namespace: "default"}, &observerExternalSvc)
	require.True(t, apierrors.IsNotFound(err))
	var observerSearchSvc corev1.Service
	err = fc.Client.Get(context.Background(),
		types.NamespacedName{Name: "test-fe-observer-search", Namespace: "default"}, &observerSearchSvc)
	require.True(t, apierrors.IsNotFound(err))

	require.Equal(t, service.SearchServiceName(src.Name, src.Spec.StarRocksFeSpec), sts.Spec.ServiceName)
	require.Equal(t, map[string]string{
		srapi.OwnerReference:    "test-fe-observer",
		srapi.ComponentLabelKey: srapi.DEFAULT_FE_OBSERVER,
	}, sts.Spec.Selector.MatchLabels)
	require.Equal(t, "starrocks.com/fe:2.40", sts.Spec.Template.Spec.Containers[0].Image)
	require.Equal(t, corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("4"),
		corev1.ResourceMemory: resource.MustParse("16G"),
	}, sts.Spec.Template.Spec.Containers[0].Resources.Requests)
	require.Equal(t, []string{"$(FE_SERVICE_NAME)"}, sts.Spec.Template.Spec.Containers[0].Args)
}

func TestFeObserver_SyncClusterDisabled(t *testing.T) {
	src := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image: "starrocks.com/fe:2.40",
					},
				},
				ObserverSpec: &srapi.StarRocksFeObserverSpec{
					Enabled: false,
				},
			},
		},
	}

	fc := feobserver.New(fake.NewFakeClient(srapi.Scheme, src), fake.GetEventRecorderFor(nil))
	err := fc.SyncCluster(context.Background(), src)
	require.NoError(t, err)
	err = fc.UpdateClusterStatus(context.Background(), src)
	require.NoError(t, err)
	require.Nil(t, src.Status.StarRocksFeObserverStatus)

	var sts appsv1.StatefulSet
	err = fc.Client.Get(context.Background(),
		types.NamespacedName{Name: "test-fe-observer", Namespace: "default"}, &sts)
	require.True(t, apierrors.IsNotFound(err))
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
