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

package pod

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	corev1 "k8s.io/api/core/v1"
)

func TestMakeLifeCycle(t *testing.T) {
	type args struct {
		preStopScriptPath string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Lifecycle
	}{
		{
			name: "test",
			args: args{
				preStopScriptPath: "/scripts/pre-stop.sh",
			},
			want: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"/scripts/pre-stop.sh"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := LifeCycle(tt.args.preStopScriptPath)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("LifeCycle() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestLabels(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "test labels",
			args: args{
				clusterName: "test",
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
							PodLabels: map[string]string{
								"l1": "v1",
							},
						},
					},
				},
			},
			want: map[string]string{
				"l1":                 "v1",
				v1.OwnerReference:    "test-fe",
				v1.ComponentLabelKey: v1.DEFAULT_FE,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Labels(tt.args.clusterName, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Labels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvs(t *testing.T) {
	envsWithoutIP := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
			},
		},
		{
			Name:  "HOST_TYPE",
			Value: "FQDN",
		},
	}

	envs := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.hostIP"},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
			},
		},
		{
			Name:  "HOST_TYPE",
			Value: "FQDN",
		},
	}

	type args struct {
		clusterName string
		namespace   string
		spec        v1.SpecInterface
		config      map[string]interface{}
	}
	tests := []struct {
		name            string
		args            args
		want            []corev1.EnvVar
		unsupportedEnvs string
	}{
		{
			name: "test envs for fe",
			args: args{
				clusterName: "test",
				namespace:   "ns",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: append(append([]corev1.EnvVar(nil), envs...), []corev1.EnvVar{
				{
					Name:  v1.COMPONENT_NAME,
					Value: v1.DEFAULT_FE,
				},
				{
					Name:  v1.FE_SERVICE_NAME,
					Value: service.ExternalServiceName("test", &v1.StarRocksFeSpec{}) + "." + "ns",
				},
			}...),
			unsupportedEnvs: "",
		},
		{
			name: "test envs for be",
			args: args{
				clusterName: "test",
				namespace:   "ns",
				spec:        &v1.StarRocksBeSpec{},
			},
			want: append(append([]corev1.EnvVar(nil), envs...), []corev1.EnvVar{
				{
					Name:  v1.COMPONENT_NAME,
					Value: v1.DEFAULT_BE,
				},
				{
					Name:  v1.FE_SERVICE_NAME,
					Value: service.ExternalServiceName("test", &v1.StarRocksFeSpec{}),
				},
				{
					Name:  "FE_QUERY_PORT",
					Value: fmt.Sprintf("%v", rutils.DefMap[rutils.QUERY_PORT]),
				},
			}...),
			unsupportedEnvs: "",
		},
		{
			name: "test envs for cn",
			args: args{
				clusterName: "test",
				namespace:   "ns",
				spec:        &v1.StarRocksCnSpec{},
			},
			want: append(append([]corev1.EnvVar(nil), envs...), []corev1.EnvVar{
				{
					Name:  v1.COMPONENT_NAME,
					Value: v1.DEFAULT_CN,
				},
				{
					Name:  v1.FE_SERVICE_NAME,
					Value: service.ExternalServiceName("test", &v1.StarRocksFeSpec{}),
				},
				{
					Name:  "FE_QUERY_PORT",
					Value: fmt.Sprintf("%v", rutils.DefMap[rutils.QUERY_PORT]),
				},
			}...),
			unsupportedEnvs: "",
		},
		{
			name: "test envs for be with unsupport envs",
			args: args{
				clusterName: "test",
				namespace:   "ns",
				spec:        &v1.StarRocksBeSpec{},
			},
			want: append(append([]corev1.EnvVar(nil), envsWithoutIP...), []corev1.EnvVar{
				{
					Name:  v1.COMPONENT_NAME,
					Value: v1.DEFAULT_BE,
				},
				{
					Name:  v1.FE_SERVICE_NAME,
					Value: service.ExternalServiceName("test", &v1.StarRocksFeSpec{}),
				},
				{
					Name:  "FE_QUERY_PORT",
					Value: fmt.Sprintf("%v", rutils.DefMap[rutils.QUERY_PORT]),
				},
			}...),
			unsupportedEnvs: "HOST_IP,POD_IP",
		},
	}
	for _, tt := range tests {
		feExternalServiceName := service.ExternalServiceName("test", &v1.StarRocksFeSpec{})
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KUBE_STARROCKS_UNSUPPORTED_ENVS", tt.unsupportedEnvs)
			defer func() {
				os.Setenv("KUBE_STARROCKS_UNSUPPORTED_ENVS", "")
			}()
			got := Envs(tt.args.spec, tt.args.config, feExternalServiceName, tt.args.namespace, nil)
			if len(got) != len(tt.want) {
				t.Errorf("Envs() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i].ValueFrom != nil {
					if got[i].ValueFrom.FieldRef.FieldPath != tt.want[i].ValueFrom.FieldRef.FieldPath {
						t.Errorf("Envs() = %v, want %v", got[i], tt.want[i])
					}
				} else if got[i] != tt.want[i] {
					t.Errorf("Envs() = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSpec(t *testing.T) {
	type args struct {
		spec      v1.SpecInterface
		container corev1.Container
		volumes   []corev1.Volume
	}
	tests := []struct {
		name string
		args args
		want corev1.PodSpec
	}{
		{
			name: "test service account name in spec",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
							ServiceAccount: "test",
						},
					},
				},
				container: corev1.Container{},
				volumes:   nil,
			},
			want: corev1.PodSpec{
				Containers:                    []corev1.Container{{}},
				ServiceAccountName:            "test",
				TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
				AutomountServiceAccountToken:  func() *bool { b := false; return &b }(),
			},
		},
		{
			name: "test service account name 2 in spec",
			args: args{
				spec:      &v1.StarRocksFeSpec{},
				container: corev1.Container{},
				volumes:   nil,
			},
			want: corev1.PodSpec{
				Containers:                    []corev1.Container{{}},
				TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
				AutomountServiceAccountToken:  func() *bool { b := false; return &b }(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Spec(tt.args.spec, tt.args.container, tt.args.volumes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Spec() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSecurityContext(t *testing.T) {
	onrootMismatch := corev1.FSGroupChangeOnRootMismatch

	type args struct {
		spec v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want *corev1.PodSecurityContext
	}{
		{
			name: "test security context",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{},
				},
			},
			want: &corev1.PodSecurityContext{
				FSGroupChangePolicy: &onrootMismatch,
			},
		},
		{
			name: "test security context 2",
			args: args{
				spec: &v1.StarRocksFeSpec{},
			},
			want: &corev1.PodSecurityContext{
				FSGroupChangePolicy: &onrootMismatch,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PodSecurityContext(tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodSecurityContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnnotations(t *testing.T) {
	type args struct {
		spec v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "test annotations",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
							Annotations: map[string]string{"v1": "v1"},
						},
					},
				},
			},
			want: map[string]string{
				"v1": "v1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Annotations(tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Annotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
