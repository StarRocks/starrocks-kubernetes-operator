// Copyright 2021-present, StarRocks Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This test lives in package k8sutils_test because it drives the real statefulset renderer, and the
// statefulset template package imports k8sutils.
package k8sutils_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
)

// TestApplyStatefulSet_SkipFECleanupRoundTrip drives the real renderer through the annotation's whole
// life: a warehouse statefulset is created, the annotation is turned on, then off again. On: the
// three-way merge takes the protection finalizer off, the annotation is mirrored, and the pod
// template does not move (no CN restart). Off: the protection finalizer comes back and the mirrored
// annotation goes.
func TestApplyStatefulSet_SkipFECleanupRoundTrip(t *testing.T) {
	ctx := context.Background()
	key := types.NamespacedName{Name: "wh-warehouse-cn", Namespace: "default"}
	replicas := int32(1)
	spec := &srapi.StarRocksCnSpec{StarRocksComponentSpec: srapi.StarRocksComponentSpec{
		StarRocksLoadSpec: srapi.StarRocksLoadSpec{Replicas: &replicas},
	}}
	podTemplate := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "cn", Image: "cn:latest"}}},
	}
	warehouse := &srapi.StarRocksWarehouse{
		TypeMeta:   metav1.TypeMeta{Kind: "StarRocksWarehouse", APIVersion: srapi.GroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: "wh", Namespace: "default", UID: "uid-wh"},
		Spec:       srapi.StarRocksWarehouseSpec{StarRocksCluster: "test"},
	}
	render := func(skip bool) *appsv1.StatefulSet {
		wh := warehouse.DeepCopy()
		if skip {
			wh.Annotations = map[string]string{object.SkipFECleanupAnnotation: "true"}
		}
		sts := statefulset.MakeStatefulset(object.NewFromWarehouse(wh), spec, podTemplate.DeepCopy())
		return &sts
	}
	k8sClient := fake.NewFakeClient(srapi.Scheme)
	apply := func(skip bool) appsv1.StatefulSet {
		require.NoError(t, k8sutils.ApplyStatefulSet(ctx, k8sClient, render(skip), true, rutils.StatefulSetDeepEqual))
		var got appsv1.StatefulSet
		require.NoError(t, k8sClient.Get(ctx, key, &got))
		return got
	}

	created := apply(false)
	require.Equal(t, []string{statefulset.STARROCKS_WAREHOUSE_FINALIZER}, created.Finalizers)

	on := apply(true)
	require.Empty(t, on.Finalizers)
	require.Equal(t, "true", on.Annotations[object.SkipFECleanupAnnotation])
	require.Equal(t, created.Spec.Template, on.Spec.Template, "turning the annotation on must not roll the CN pods")

	steady := apply(true)
	require.Equal(t, on.ResourceVersion, steady.ResourceVersion, "a second apply with nothing changed must not write")

	off := apply(false)
	require.Equal(t, []string{statefulset.STARROCKS_WAREHOUSE_FINALIZER}, off.Finalizers)
	require.NotContains(t, off.Annotations, object.SkipFECleanupAnnotation)
	require.Equal(t, created.Spec.Template, off.Spec.Template)
}
