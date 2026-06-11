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

package resource_utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
)

func TestBuildFeIngress(t *testing.T) {
	className := "nginx"
	src := &srapi.StarRocksCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       object.StarRocksClusterKind,
			APIVersion: "starrocks.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
	}
	feIngress := &srapi.FeIngress{
		IngressClassName: &className,
		Host:             "test-cluster.example.com",
		Annotations:      map[string]string{"a": "b"},
	}

	got := BuildFeIngress(object.NewFromCluster(src), feIngress, map[string]string{"l": "v"})

	require.Equal(t, "test-cluster-fe-ingress", got.Name)
	require.Equal(t, "default", got.Namespace)
	require.Equal(t, "b", got.Annotations["a"])
	require.NotNil(t, got.Spec.IngressClassName)
	require.Equal(t, "nginx", *got.Spec.IngressClassName)
	require.Len(t, got.Spec.Rules, 1)
	require.Equal(t, "test-cluster.example.com", got.Spec.Rules[0].Host)
	require.NotNil(t, got.Spec.Rules[0].HTTP)

	paths := got.Spec.Rules[0].HTTP.Paths
	require.Len(t, paths, 1)
	backend := paths[0].Backend.Service
	require.NotNil(t, backend)
	// Backend must target the FE external service's "http" (web UI) port, never the
	// MySQL query port.
	require.Equal(t, "test-cluster-fe-service", backend.Name)
	require.Equal(t, FeHTTPPortName, backend.Port.Name)

	require.Len(t, got.OwnerReferences, 1)
	owner := got.OwnerReferences[0]
	require.Equal(t, object.StarRocksClusterKind, owner.Kind)
	require.Equal(t, "test-cluster", owner.Name)
	require.NotNil(t, owner.Controller)
	require.True(t, *owner.Controller, "owner reference must be a controller reference for garbage collection")
}

func TestBuildFeIngress_TLS(t *testing.T) {
	src := &srapi.StarRocksCluster{
		TypeMeta:   metav1.TypeMeta{Kind: object.StarRocksClusterKind, APIVersion: "starrocks.com/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "test-cluster", Namespace: "default"},
	}
	feIngress := &srapi.FeIngress{
		Host: "test-cluster.example.com",
		TLS: []networkingv1.IngressTLS{{
			Hosts:      []string{"test-cluster.example.com"},
			SecretName: "test-cluster-tls",
		}},
	}

	got := BuildFeIngress(object.NewFromCluster(src), feIngress, nil)

	require.Len(t, got.Spec.TLS, 1)
	require.Equal(t, "test-cluster-tls", got.Spec.TLS[0].SecretName)
	require.Equal(t, []string{"test-cluster.example.com"}, got.Spec.TLS[0].Hosts)
}
