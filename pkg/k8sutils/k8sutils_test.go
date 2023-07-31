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

package k8sutils

import (
	"context"
	"testing"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	groupVersion := schema.GroupVersion{Group: "starrocks.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	chemeBuilder := &scheme.Builder{GroupVersion: groupVersion}
	clientgoscheme.AddToScheme(sch)
	chemeBuilder.Register(&srapi.StarRocksCluster{}, &srapi.StarRocksClusterList{})
	chemeBuilder.AddToScheme(sch)
}

func Test_DeleteAutoscaler(t *testing.T) {
	v1autoscaler := v1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	v2autoscaler := v2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
	}

	v2beta2Autoscaler := v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	k8sclient := NewFakeClient(sch, &v1autoscaler, &v2autoscaler, &v2beta2Autoscaler)
	// confirm the v1autoscaler is exist.
	var cv1autoscaler v1.HorizontalPodAutoscaler
	cerr := k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv1autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", cv1autoscaler.Name)
	delerr := DeleteAutoscaler(context.Background(), k8sclient, "default", "test", srapi.AutoScalerV1)
	require.Equal(t, nil, delerr)
	var ev1autoscaler v1.HorizontalPodAutoscaler
	geterr := k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev1autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))

	var cv2autoscaler v2.HorizontalPodAutoscaler
	cerr = k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv2autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", v2autoscaler.Name)
	delerr = DeleteAutoscaler(context.Background(), k8sclient, "default", "test", srapi.AutoSclaerV2)
	require.Equal(t, nil, delerr)
	var ev2autoscaler v2.HorizontalPodAutoscaler
	geterr = k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev2autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))

	var cv2beta2autoscaler v2beta2.HorizontalPodAutoscaler
	cerr = k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv2beta2autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", cv2beta2autoscaler.Name)
	delerr = DeleteAutoscaler(context.Background(), k8sclient, "default", "test", srapi.AutoScalerV2Beta2)
	require.Equal(t, nil, delerr)
	var ev2beta2autoscaler v2beta2.HorizontalPodAutoscaler
	geterr = k8sclient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev2beta2autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))
}
