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

package cn

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
)

// TestCnController_ClearWarehouse_SkippedByAnnotation: a warehouse that opted out of the FE cleanup
// is left in the FE, and a leftover protection finalizer is released so the deletion is not held up.
// The statefulset carries no FE connection details at all: the skip is decided before they are read.
func TestCnController_ClearWarehouse_SkippedByAnnotation(t *testing.T) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: appsv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "wh-test-warehouse-cn",
			Namespace:   "default",
			Annotations: map[string]string{object.SkipFECleanupAnnotation: "true"},
			Finalizers:  []string{statefulset.STARROCKS_WAREHOUSE_FINALIZER},
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "cn"}}}},
		},
	}
	cc := New(fake.NewFakeClient(srapi.Scheme, sts), fake.GetEventRecorderFor(nil))

	require.NoError(t, cc.ClearWarehouse(context.Background(), "default", "wh-test"))

	var got appsv1.StatefulSet
	require.NoError(t, cc.k8sClient.Get(context.Background(),
		types.NamespacedName{Name: "wh-test-warehouse-cn", Namespace: "default"}, &got))
	require.Empty(t, got.Finalizers)
}

// TestCnController_ClearWarehouse_StatefulSetGoneIsNothingToClear: a warehouse that opted out keeps
// no finalizer, so its statefulset can be collected before the cleanup runs. With no statefulset
// there is no FE connection to clean up with, and that is not an error.
func TestCnController_ClearWarehouse_StatefulSetGoneIsNothingToClear(t *testing.T) {
	cc := New(fake.NewFakeClient(srapi.Scheme), fake.GetEventRecorderFor(nil))
	require.NoError(t, cc.ClearWarehouse(context.Background(), "default", "wh-test"))
}
