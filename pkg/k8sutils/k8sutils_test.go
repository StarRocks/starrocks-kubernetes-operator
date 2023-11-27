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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	k8sClient := NewFakeClient(sch, &v1autoscaler, &v2autoscaler, &v2beta2Autoscaler)
	// confirm the v1.autoscaler exist.
	var cv1autoscaler v1.HorizontalPodAutoscaler
	cerr := k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv1autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", cv1autoscaler.Name)
	delerr := DeleteAutoscaler(context.Background(), k8sClient, "default", "test", srapi.AutoScalerV1)
	require.Equal(t, nil, delerr)
	var ev1autoscaler v1.HorizontalPodAutoscaler
	geterr := k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev1autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))

	var cv2autoscaler v2.HorizontalPodAutoscaler
	cerr = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv2autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", v2autoscaler.Name)
	delerr = DeleteAutoscaler(context.Background(), k8sClient, "default", "test", srapi.AutoScalerV2)
	require.Equal(t, nil, delerr)
	var ev2autoscaler v2.HorizontalPodAutoscaler
	geterr = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev2autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))

	var cv2beta2autoscaler v2beta2.HorizontalPodAutoscaler
	cerr = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &cv2beta2autoscaler)
	require.Equal(t, nil, cerr)
	require.Equal(t, "test", cv2beta2autoscaler.Name)
	delerr = DeleteAutoscaler(context.Background(), k8sClient, "default", "test", srapi.AutoScalerV2Beta2)
	require.Equal(t, nil, delerr)
	var ev2beta2autoscaler v2beta2.HorizontalPodAutoscaler
	geterr = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, &ev2beta2autoscaler)
	require.True(t, apierrors.IsNotFound(geterr))
}

func Test_getValueFromConfigmap(t *testing.T) {
	type args struct {
		k8sClient client.Client
		namespace string
		name      string
		key       string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get value from configmap",
			args: args{
				k8sClient: NewFakeClient(sch, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Data: map[string]string{
						"file.txt": "hell world",
					},
				}),
				namespace: "default",
				name:      "test",
				key:       "file.txt",
			},
			want:    "hell world",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueFromConfigmap(tt.args.k8sClient, tt.args.namespace, tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueFromConfigmap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueFromConfigmap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValueFromSecret(t *testing.T) {
	type args struct {
		k8sClient client.Client
		namespace string
		name      string
		key       string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get value from secret",
			args: args{
				k8sClient: NewFakeClient(sch, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"file.txt": []byte("hell world"),
					},
				}),
				namespace: "default",
				name:      "test",
				key:       "file.txt",
			},
			want:    "hell world",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueFromSecret(tt.args.k8sClient, tt.args.namespace, tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueFromSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueFromSecret() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvVarValue(t *testing.T) {
	type args struct {
		k8sClient client.Client
		namespace string
		envVar    corev1.EnvVar
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get value",
			args: args{
				k8sClient: NewFakeClient(sch),
				namespace: "default",
				envVar: corev1.EnvVar{
					Name:  "test",
					Value: "hw",
				},
			},
			want:    "hw",
			wantErr: false,
		},
		{
			name: "get value from configmap",
			args: args{
				k8sClient: NewFakeClient(sch,
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
						Data: map[string]string{
							"configmap": "hello",
						},
					}),
				namespace: "default",
				envVar: corev1.EnvVar{
					Name: "test",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test",
							},
							Key: "configmap",
						},
					},
				},
			},
			want:    "hello",
			wantErr: false,
		},
		{
			name: "get value from secret",
			args: args{
				k8sClient: NewFakeClient(sch,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
						Data: map[string][]byte{
							"secret": []byte("world"),
						},
					}),
				namespace: "default",
				envVar: corev1.EnvVar{
					Name: "test",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test",
							},
							Key: "secret",
						},
					},
				},
			},
			want:    "world",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEnvVarValue(tt.args.k8sClient, tt.args.namespace, tt.args.envVar)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEnvVarValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEnvVarValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}
