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

package statefulset

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
)

func TestMakeStatefulset_SkipFECleanup(t *testing.T) {
	replicas := int32(1)
	spec := &v1.StarRocksCnSpec{StarRocksComponentSpec: v1.StarRocksComponentSpec{
		StarRocksLoadSpec: v1.StarRocksLoadSpec{Replicas: &replicas},
	}}
	warehouse := func(annotations map[string]string) *v1.StarRocksWarehouse {
		return &v1.StarRocksWarehouse{
			ObjectMeta: metav1.ObjectMeta{Name: "wh", Namespace: "namespace", Annotations: annotations},
			Spec:       v1.StarRocksWarehouseSpec{StarRocksCluster: "test"},
		}
	}
	cluster := &v1.StarRocksCluster{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "namespace",
		Annotations: map[string]string{object.SkipFECleanupAnnotation: "true"}}}

	optedOut := func(t *testing.T, got appsv1.StatefulSet) {
		require.Empty(t, got.Finalizers)
		require.Equal(t, "true", got.Annotations[object.SkipFECleanupAnnotation])
	}
	notOptedOut := func(t *testing.T, got appsv1.StatefulSet) {
		require.Equal(t, []string{STARROCKS_WAREHOUSE_FINALIZER}, got.Finalizers)
		require.NotContains(t, got.Annotations, object.SkipFECleanupAnnotation)
	}

	tests := []struct {
		name  string
		obj   object.StarRocksObject
		check func(t *testing.T, got appsv1.StatefulSet)
	}{
		{"warehouse without the annotation renders exactly as before", object.NewFromWarehouse(warehouse(nil)), notOptedOut},
		{"annotation true drops the finalizer and is mirrored",
			object.NewFromWarehouse(warehouse(map[string]string{object.SkipFECleanupAnnotation: "true"})), optedOut},
		{"annotation value is case-insensitive",
			object.NewFromWarehouse(warehouse(map[string]string{object.SkipFECleanupAnnotation: "TRUE"})), optedOut},
		{"annotation false keeps the finalizer",
			object.NewFromWarehouse(warehouse(map[string]string{object.SkipFECleanupAnnotation: "false"})), notOptedOut},
		{"any other value keeps the finalizer",
			object.NewFromWarehouse(warehouse(map[string]string{object.SkipFECleanupAnnotation: "yes"})), notOptedOut},
		{"annotation on a cluster changes nothing", object.NewFromCluster(cluster), func(t *testing.T, got appsv1.StatefulSet) {
			require.Empty(t, got.Finalizers)
			require.NotContains(t, got.Annotations, object.SkipFECleanupAnnotation)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, MakeStatefulset(tt.obj, spec, &corev1.PodTemplateSpec{}))
		})
	}

	// The opt-out only touches metadata. If it reached the spec, turning it on would roll every CN pod.
	podTemplate := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "cn", Image: "cn:latest"}}},
	}
	plain := MakeStatefulset(object.NewFromWarehouse(warehouse(nil)), spec, podTemplate.DeepCopy())
	annotated := MakeStatefulset(object.NewFromWarehouse(warehouse(map[string]string{object.SkipFECleanupAnnotation: "true"})),
		spec, podTemplate.DeepCopy())
	require.Equal(t, plain.Spec, annotated.Spec)
	require.Equal(t, plain.Labels, annotated.Labels)
}
