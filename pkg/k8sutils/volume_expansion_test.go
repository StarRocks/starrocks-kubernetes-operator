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

package k8sutils_test

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
)

var selector = map[string]string{"app": "cn"}

func claimTemplate(name, size string) corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(size)},
			},
		},
	}
}

func claim(name, size string) *corev1.PersistentVolumeClaim {
	c := claimTemplate(name, size)
	c.Namespace, c.Labels = "default", selector
	return &c
}

func sts(serviceName string, templates ...corev1.PersistentVolumeClaim) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "kube-cn", Namespace: "default", Annotations: map[string]string{}},
		Spec: appsv1.StatefulSetSpec{
			Selector:             &metav1.LabelSelector{MatchLabels: selector},
			ServiceName:          serviceName,
			VolumeClaimTemplates: templates,
		},
	}
}

func storageOf(t *testing.T, c client.Client, name string) string {
	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: name}, &pvc); err != nil {
		t.Fatal(err)
	}
	return pvc.Spec.Resources.Requests.Storage().String()
}

func stsExists(t *testing.T, c client.Client) bool {
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "kube-cn"}, &appsv1.StatefulSet{})
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	}
	return err == nil
}

func TestApplyStatefulSetExpandsVolumeClaims(t *testing.T) {
	ctx := context.Background()
	c := fake.NewClientBuilder().WithScheme(srapi.Scheme).WithObjects(
		sts("kube-cn-search", claimTemplate("cn-data", "10Gi"), claimTemplate("cn-log", "1Gi")),
		claim("cn-data-kube-cn-0", "10Gi"), claim("cn-data-kube-cn-1", "10Gi"), claim("cn-log-kube-cn-0", "1Gi"),
		claim("cn-data-kube-cn-other-0", "10Gi"), // another StatefulSet's claim with the same labels
	).Build()

	expect := sts("kube-cn-search", claimTemplate("cn-data", "20Gi"), claimTemplate("cn-log", "1Gi"))
	if err := k8sutils.ApplyStatefulSet(ctx, c, expect, true, rutils.StatefulSetDeepEqual); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"cn-data-kube-cn-0": "20Gi", "cn-data-kube-cn-1": "20Gi", "cn-log-kube-cn-0": "1Gi", "cn-data-kube-cn-other-0": "10Gi",
	} {
		if got := storageOf(t, c, name); got != want {
			t.Errorf("%s: got %s, want %s", name, got, want)
		}
	}
	if stsExists(t, c) {
		t.Fatal("statefulset should be deleted so it can be recreated with the new templates")
	}

	// the next reconcile recreates it with the new size
	expect = sts("kube-cn-search", claimTemplate("cn-data", "20Gi"), claimTemplate("cn-log", "1Gi"))
	if err := k8sutils.ApplyStatefulSet(ctx, c, expect, true, rutils.StatefulSetDeepEqual); err != nil {
		t.Fatal(err)
	}
	var got appsv1.StatefulSet
	if err := c.Get(ctx, types.NamespacedName{Namespace: "default", Name: "kube-cn"}, &got); err != nil {
		t.Fatal(err)
	}
	if size := got.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage().String(); size != "20Gi" {
		t.Fatalf("recreated statefulset has %s, want 20Gi", size)
	}
}

func TestApplyStatefulSetRefusesVolumeExpansion(t *testing.T) {
	tests := []struct {
		name   string
		expect *appsv1.StatefulSet
	}{
		{name: "decrease", expect: sts("kube-cn-search", claimTemplate("cn-data", "5Gi"))},
		{name: "legacy service name", expect: sts("kube-cn-search-new", claimTemplate("cn-data", "20Gi"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(srapi.Scheme).WithObjects(
				sts("kube-cn-search", claimTemplate("cn-data", "10Gi")), claim("cn-data-kube-cn-0", "10Gi"),
			).Build()
			if err := k8sutils.ApplyStatefulSet(context.Background(), c, tt.expect, true, rutils.StatefulSetDeepEqual); err == nil {
				t.Fatal("expected an error")
			}
			if got := storageOf(t, c, "cn-data-kube-cn-0"); got != "10Gi" {
				t.Errorf("claim changed to %s", got)
			}
			if !stsExists(t, c) {
				t.Error("statefulset must not be deleted")
			}
		})
	}
}

func TestApplyStatefulSetLeavesOtherTemplateChanges(t *testing.T) {
	for name, expect := range map[string]*appsv1.StatefulSet{
		"unchanged":     sts("kube-cn-search", claimTemplate("cn-data", "10Gi")),
		"volume added":  sts("kube-cn-search", claimTemplate("cn-data", "20Gi"), claimTemplate("cn-spill", "10Gi")),
		"volume rename": sts("kube-cn-search", claimTemplate("cn-cache", "20Gi")),
	} {
		c := fake.NewClientBuilder().WithScheme(srapi.Scheme).WithObjects(
			sts("kube-cn-search", claimTemplate("cn-data", "10Gi")), claim("cn-data-kube-cn-0", "10Gi"),
		).Build()
		if err := k8sutils.ApplyStatefulSet(context.Background(), c, expect, true, rutils.StatefulSetDeepEqual); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := storageOf(t, c, "cn-data-kube-cn-0"); got != "10Gi" || !stsExists(t, c) {
			t.Errorf("%s: claim %s, statefulset kept %v", name, got, stsExists(t, c))
		}
	}
}

func TestApplyStatefulSetKeepsOrphanFinalizer(t *testing.T) {
	deleting := sts("kube-cn-search")
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	deleting.Finalizers = []string{metav1.FinalizerOrphanDependents, "starrocks.com.starrockswarehouse/protection"}
	c := fake.NewClientBuilder().WithScheme(srapi.Scheme).WithObjects(deleting).Build()

	for i := 0; i < 2; i++ {
		if err := k8sutils.ApplyStatefulSet(context.Background(), c, sts("kube-cn-search"), true, rutils.StatefulSetDeepEqual); err != nil {
			t.Fatal(err)
		}
	}
	var got appsv1.StatefulSet
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "kube-cn"}, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Finalizers) != 1 || got.Finalizers[0] != metav1.FinalizerOrphanDependents {
		t.Fatalf("finalizers = %v, want only %s", got.Finalizers, metav1.FinalizerOrphanDependents)
	}
}
