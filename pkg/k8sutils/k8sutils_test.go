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

package k8sutils_test

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
)

func init() {
	srapi.Register()
}

// fake client can not delete a resource when resource version is wrong
func Test_DeleteAutoscaler(t *testing.T) {
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
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
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
			got, err := k8sutils.GetValueFromConfigmap(context.Background(), tt.args.k8sClient, tt.args.namespace, tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueFromConfigmap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValueFromConfigmap() got = %v, want %v", got, tt.want)
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
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.Secret{
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
			got, err := k8sutils.GetValueFromSecret(context.Background(), tt.args.k8sClient, tt.args.namespace, tt.args.name, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueFromSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValueFromSecret() got = %v, want %v", got, tt.want)
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
				k8sClient: fake.NewFakeClient(srapi.Scheme),
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
				k8sClient: fake.NewFakeClient(srapi.Scheme,
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
				k8sClient: fake.NewFakeClient(srapi.Scheme,
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
			got, err := k8sutils.GetEnvVarValue(context.Background(), tt.args.k8sClient, tt.args.namespace, tt.args.envVar)
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

func TestService(t *testing.T) {
	type args struct {
		ctx       context.Context
		k8sClient client.Client
		svc       *corev1.Service
		equal     k8sutils.ServiceEqual
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test apply service which not exist",
			args: args{
				ctx:       context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme),
				svc: &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.ServiceKind,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service",
						Namespace: "default",
						// use label because it is used to calculate service hash
						Labels: map[string]string{
							"test": "test",
						},
					},
				},
				equal: rutils.ServiceDeepEqual,
			},
		},
		{
			name: "test apply service which has been created",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.ServiceKind,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service",
						Namespace: "default",
					},
				}),
				svc: &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       rutils.ServiceKind,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service",
						Namespace: "default",
						// use label because it is used to calculate service hash
						Labels: map[string]string{
							"test": "test",
						},
					},
				},
				equal: rutils.ServiceDeepEqual,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := k8sutils.ApplyService(tt.args.ctx, tt.args.k8sClient, tt.args.svc, tt.args.equal); (err != nil) != tt.wantErr {
				t.Errorf("ApplyService() error = %v, wantErr %v", err, tt.wantErr)
			}
			service := &corev1.Service{}
			if err := tt.args.k8sClient.Get(context.Background(),
				types.NamespacedName{
					Name:      tt.args.svc.Name,
					Namespace: tt.args.svc.Namespace,
				},
				service,
			); err != nil {
				t.Errorf("Get Service error = %v", err)
			}
			if service.Labels["test"] != "test" {
				t.Errorf("Object does not have test annotation")
			}

			if err := k8sutils.DeleteService(context.Background(), tt.args.k8sClient, tt.args.svc.Namespace, tt.args.svc.Name); err != nil {
				t.Errorf("Delete Service error = %v", err)
			}
		})
	}
}

func TestApplyDeployment(t *testing.T) {
	type args struct {
		ctx       context.Context
		k8sClient client.Client
		deploy    *appsv1.Deployment
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test apply deployment which not exist",
			args: args{
				ctx:       context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme),
				deploy: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
				},
			},
		},
		{
			name: "test apply deployment which has been created",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
					},
				}),
				deploy: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := k8sutils.ApplyDeployment(tt.args.ctx, tt.args.k8sClient, tt.args.deploy); (err != nil) != tt.wantErr {
				t.Errorf("ApplyDeployment() error = %v, wantErr %v", err, tt.wantErr)
			}

			deployment := &appsv1.Deployment{}
			if err := tt.args.k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      tt.args.deploy.Name,
				Namespace: tt.args.deploy.Namespace},
				deployment,
			); err != nil {
				t.Errorf("Get Deployment error = %v", err)
			}
			if deployment.Annotations["test"] != "test" {
				t.Errorf("Object does not have test annotation")
			}

			if err := k8sutils.DeleteDeployment(context.Background(), tt.args.k8sClient, tt.args.deploy.Namespace, tt.args.deploy.Name); err != nil {
				t.Errorf("Delete Deployment error = %v", err)
			}
		})
	}
}

func TestApplyConfigMap(t *testing.T) {
	type args struct {
		ctx       context.Context
		k8sClient client.Client
		configmap *corev1.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test apply configmap which not exist",
			args: args{
				ctx:       context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme),
				configmap: &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-configmap",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
				},
			},
		},
		{
			name: "test apply configmap which has been created",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-configmap",
						Namespace: "default",
					},
				}),
				configmap: &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-configmap",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
					// add Data to make sure the hash value will be changed
					Data: map[string]string{
						"key": "value",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := k8sutils.ApplyConfigMap(tt.args.ctx, tt.args.k8sClient, tt.args.configmap); (err != nil) != tt.wantErr {
				t.Errorf("ApplyConfigMap() error = %v, wantErr %v", err, tt.wantErr)
			}

			configMap := &corev1.ConfigMap{}
			if err := tt.args.k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      tt.args.configmap.Name,
				Namespace: tt.args.configmap.Namespace},
				configMap,
			); err != nil {
				t.Errorf("Get ConfigMap error = %v", err)
			}
			if configMap.Annotations["test"] != "test" {
				t.Errorf("Object does not have test annotation")
			}

			if err := k8sutils.DeleteConfigMap(context.Background(), tt.args.k8sClient, tt.args.configmap.Namespace, tt.args.configmap.Name); err != nil {
				t.Errorf("Delete Configmap error = %v", err)
			}
		})
	}
}

func TestApplyStatefulSet(t *testing.T) {
	type args struct {
		ctx       context.Context
		k8sClient client.Client
		sts       *appsv1.StatefulSet
		equal     k8sutils.StatefulSetEqual
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test apply sts which not exist",
			args: args{
				ctx:       context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme),
				sts: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sts",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
				},
				equal: func(new *appsv1.StatefulSet, actual *appsv1.StatefulSet) bool {
					return rutils.StatefulSetDeepEqual(new, actual, false)
				},
			},
		},
		{
			name: "test apply sts which has been created",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(srapi.Scheme, &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sts",
						Namespace: "default",
					},
				}),
				sts: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: appsv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sts",
						Namespace: "default",
						Annotations: map[string]string{
							"test": "test",
						},
					},
					// add spec to make sure the hash value will be changed
					Spec: appsv1.StatefulSetSpec{
						Replicas: func() *int32 { v := int32(1); return &v }(),
					},
				},
				equal: func(new *appsv1.StatefulSet, actual *appsv1.StatefulSet) bool {
					return rutils.StatefulSetDeepEqual(new, actual, false)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := k8sutils.ApplyStatefulSet(tt.args.ctx, tt.args.k8sClient, tt.args.sts, tt.args.equal); (err != nil) != tt.wantErr {
				t.Errorf("ApplyStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
			}
			sts := &appsv1.StatefulSet{}
			if err := tt.args.k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      tt.args.sts.Name,
				Namespace: tt.args.sts.Namespace},
				sts,
			); err != nil {
				t.Errorf("Get StatefulSet error = %v", err)
			}
			if sts.Annotations["test"] != "test" {
				t.Errorf("Object does not have test annotation")
			}

			if err := k8sutils.DeleteStatefulset(context.Background(), tt.args.k8sClient, tt.args.sts.Namespace, tt.args.sts.Name); err != nil {
				t.Errorf("Delete Statfulset error = %v", err)
			}
		})
	}
}
