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

package fe_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

func init() {
	srapi.Register()
}

func TestFeController_updateStatus(_ *testing.T) {
	var creatings, readys, faileds []string
	podmap := make(map[string]corev1.Pod)
	// get all pod status that controlled by st.
	var podList corev1.PodList
	podList.Items = append(podList.Items, corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}})

	for i := range podList.Items {
		pod := &podList.Items[i]
		podmap[pod.Name] = podList.Items[i]
		if ready := k8sutils.PodIsReady(&pod.Status); ready {
			readys = append(readys, pod.Name)
		} else if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			faileds = append(faileds, pod.Name)
		}
	}

	fmt.Printf("the ready len %d, the creatings len %d, the faileds %d", len(readys), len(creatings), len(faileds))
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

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-access",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{},
	}

	ssvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fe-search",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{},
	}

	fc := fe.New(fake.NewFakeClient(srapi.Scheme, src, svc, ssvc))
	err := fc.ClearResources(context.Background(), src)
	require.Equal(t, nil, err)

	var est appv1.StatefulSet
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var aesvc corev1.Service
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: "test-fe-access", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var resvc corev1.Service
	err = fc.Client.Get(context.Background(), types.NamespacedName{Name: "test-fe-search", Namespace: "default"}, &resvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
}

func Test_SyncDeploy(t *testing.T) {
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
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Replicas: rutils.GetInt32Pointer(3),
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

	fc := fe.New(fake.NewFakeClient(srapi.Scheme, src))

	err := fc.SyncCluster(context.Background(), src)
	require.Equal(t, nil, err)
	err = fc.UpdateClusterStatus(context.Background(), src)
	require.Equal(t, nil, err)
	festatus := src.Status.StarRocksFeStatus
	require.Equal(t, nil, err)
	require.Equal(t, festatus.Phase, srapi.ComponentReconciling)
	require.Equal(t, festatus.ServiceName, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec))

	var st appv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	spec := src.Spec.StarRocksFeSpec
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), Namespace: "default"}, &asvc))
	require.Equal(t, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), asvc.Name)
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: service.SearchServiceName(src.Name, spec), Namespace: "default"}, &rsvc))
	require.Equal(t, service.SearchServiceName(src.Name, spec), rsvc.Name)
	require.NoError(t, fc.Client.Get(context.Background(),
		types.NamespacedName{Name: load.Name(src.Name, spec), Namespace: "default"}, &st))
	// validate service selector matches statefulset selector
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}

func TestCheckFEReady(t *testing.T) {
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

func TestGetFeConfig(t *testing.T) {
	type args struct {
		ctx           context.Context
		k8sClient     client.Client
		configMapInfo *srapi.ConfigMapInfo
		namespace     string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "test get FE config",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fe-configMap",
						Namespace: "default",
					},
					Data: map[string]string{
						"fe.config": "aa = bb",
					},
				}),
				configMapInfo: &srapi.ConfigMapInfo{
					ConfigMapName: "fe-configMap",
					ResolveKey:    "fe.config",
				},
				namespace: "default",
			},
			want: map[string]interface{}{
				"aa": "bb",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fe.GetFeConfig(tt.args.ctx, tt.args.k8sClient, tt.args.configMapInfo, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
