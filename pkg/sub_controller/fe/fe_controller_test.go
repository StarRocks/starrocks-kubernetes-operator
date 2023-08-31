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

package fe

import (
	"context"
	"fmt"
	"testing"
	"time"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
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
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	sch = runtime.NewScheme()
)

func init() {
	groupVersion := schema.GroupVersion{Group: "starrocks.com", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	schemeBuilder := &scheme.Builder{GroupVersion: groupVersion}
	_ = clientgoscheme.AddToScheme(sch)
	schemeBuilder.Register(&srapi.StarRocksCluster{}, &srapi.StarRocksClusterList{})
	_ = schemeBuilder.AddToScheme(sch)
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

	fc := New(k8sutils.NewFakeClient(sch, src, svc, ssvc))
	err := fc.ClearResources(context.Background(), src)
	require.Equal(t, nil, err)

	var est appv1.StatefulSet
	err = fc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &est)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var aesvc corev1.Service
	err = fc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-fe-access", Namespace: "default"}, &aesvc)
	require.True(t, err == nil || apierrors.IsNotFound(err))
	var resvc corev1.Service
	err = fc.k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-fe-search", Namespace: "default"}, &resvc)
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

	fc := New(k8sutils.NewFakeClient(sch, src))

	err := fc.SyncCluster(context.Background(), src)
	require.Equal(t, nil, err)
	err = fc.UpdateClusterStatus(src)
	require.Equal(t, nil, err)
	festatus := src.Status.StarRocksFeStatus
	require.Equal(t, nil, err)
	require.Equal(t, festatus.Phase, srapi.ComponentReconciling)
	require.Equal(t, festatus.ServiceName, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec))

	var st appv1.StatefulSet
	var asvc corev1.Service
	var rsvc corev1.Service
	spec := src.Spec.StarRocksFeSpec
	require.NoError(t, fc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), Namespace: "default"}, &asvc))
	require.Equal(t, service.ExternalServiceName(src.Name, src.Spec.StarRocksFeSpec), asvc.Name)
	require.NoError(t, fc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: service.SearchServiceName(src.Name, spec), Namespace: "default"}, &rsvc))
	require.Equal(t, service.SearchServiceName(src.Name, spec), rsvc.Name)
	require.NoError(t, fc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: load.Name(src.Name, spec), Namespace: "default"}, &st))
	// validate service selector matches statefulset selector
	require.Equal(t, asvc.Spec.Selector, st.Spec.Selector.MatchLabels)
}
