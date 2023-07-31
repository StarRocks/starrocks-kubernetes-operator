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
	"reflect"
	"testing"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMakeLivenessProbe(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LivenessProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LivenessProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeReadinessProbe(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadinessProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadinessProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeStartupProbe(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				FailureThreshold: 60,
				PeriodSeconds:    5,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartupProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StartupProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func Test_makeProbeHandler(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want corev1.ProbeHandler
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/api/health2",
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountConfigMapInfo(t *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		cmInfo       v1.ConfigMapInfo
		mountPath    string
	}
	tests := []struct {
		name  string
		args  args
		want  []corev1.Volume
		want1 []corev1.VolumeMount
	}{
		{
			name: "test mount configmap",
			args: args{
				cmInfo:    v1.ConfigMapInfo{ConfigMapName: "cm", ResolveKey: "key"},
				mountPath: "/pkg/mounts/volume",
			},
			want: []corev1.Volume{
				{
					Name: "cm",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "cm",
							},
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "cm",
					MountPath: "/pkg/mounts/volume",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MountConfigMapInfo(tt.args.volumes, tt.args.volumeMounts, tt.args.cmInfo, tt.args.mountPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountConfigMapInfo() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountConfigMapInfo() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMountSecrets(t *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		secrets      []v1.SecretReference
	}
	tests := []struct {
		name  string
		args  args
		want  []corev1.Volume
		want1 []corev1.VolumeMount
	}{
		{
			name: "test mount configmaps",
			args: args{
				secrets: []v1.SecretReference{
					{
						Name:      "s1",
						MountPath: "/pkg/mounts/volumes1",
					},
					{
						Name:      "s2",
						MountPath: "/pkg/mounts/volumes2",
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "s1-4145",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "s1",
						},
					},
				},
				{
					Name: "s2-1955",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "s2",
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "s1-4145",
					MountPath: "/pkg/mounts/volumes1",
				},
				{
					Name:      "s2-1955",
					MountPath: "/pkg/mounts/volumes2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MountSecrets(tt.args.volumes, tt.args.volumeMounts, tt.args.secrets)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountSecrets() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountSecrets() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMountConfigMaps(t *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		configmaps   []v1.ConfigMapReference
	}
	tests := []struct {
		name  string
		args  args
		want  []corev1.Volume
		want1 []corev1.VolumeMount
	}{
		{
			name: "test mount configmaps",
			args: args{
				configmaps: []v1.ConfigMapReference{
					{
						Name:      "s1",
						MountPath: "/pkg/mounts/volumes1",
					},
					{
						Name:      "s2",
						MountPath: "/pkg/mounts/volumes2",
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "s1-4145",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s1",
							},
						},
					},
				},
				{
					Name: "s2-1955",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s2",
							},
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "s1-4145",
					MountPath: "/pkg/mounts/volumes1",
				},
				{
					Name:      "s2-1955",
					MountPath: "/pkg/mounts/volumes2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MountConfigMaps(tt.args.volumes, tt.args.volumeMounts, tt.args.configmaps)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountSecrets() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountSecrets() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMountStorageVolumes(t *testing.T) {
	type args struct {
		spec v1.SpecInterface
	}
	tests := []struct {
		name  string
		args  args
		want  []corev1.Volume
		want1 []corev1.VolumeMount
		want2 map[string]bool
	}{
		{
			name: "test mount storage volumes",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StorageVolumes: []v1.StorageVolume{
							{
								Name:             "s1",
								MountPath:        "/pkg/mounts/volumes1",
								StorageClassName: func() *string { s := "sc1"; return &s }(),
								StorageSize:      "1GB",
							},
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "s1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "s1",
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "s1",
					MountPath: "/pkg/mounts/volumes1",
				},
			},
			want2: map[string]bool{
				"/pkg/mounts/volumes1": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := MountStorageVolumes(tt.args.spec)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountStorageVolumes() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountStorageVolumes() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("MountStorageVolumes() got2 = %v, want %v", got2, tt.want2)
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
						PodLabels: map[string]string{
							"l1": "v1",
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
		{
			Name:  "USER",
			Value: "root",
		},
	}

	type args struct {
		clusterName string
		namespace   string
		spec        v1.SpecInterface
		config      map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
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
					Value: v1.GetExternalServiceName("test", &v1.StarRocksFeSpec{}) + "." + "ns",
				},
			}...),
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
					Value: v1.GetExternalServiceName("test", &v1.StarRocksFeSpec{}),
				},
				{
					Name:  "FE_QUERY_PORT",
					Value: fmt.Sprintf("%v", rutils.DefMap[rutils.QUERY_PORT]),
				},
			}...),
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
					Value: v1.GetExternalServiceName("test", &v1.StarRocksFeSpec{}),
				},
				{
					Name:  "FE_QUERY_PORT",
					Value: fmt.Sprintf("%v", rutils.DefMap[rutils.QUERY_PORT]),
				},
			}...),
		},
	}
	for _, tt := range tests {
		feExternalServiceName := v1.GetExternalServiceName("test", &v1.StarRocksFeSpec{})
		t.Run(tt.name, func(t *testing.T) {
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
		spec                  v1.SpecInterface
		defaultServiceAccount string
		container             corev1.Container
		volumes               []corev1.Volume
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
						ServiceAccount: "test",
					},
				},
				defaultServiceAccount: "default",
				container:             corev1.Container{},
				volumes:               nil,
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
				spec:                  &v1.StarRocksFeSpec{},
				defaultServiceAccount: "default",
				container:             corev1.Container{},
				volumes:               nil,
			},
			want: corev1.PodSpec{
				Containers:                    []corev1.Container{{}},
				ServiceAccountName:            "default",
				TerminationGracePeriodSeconds: rutils.GetInt64ptr(int64(120)),
				AutomountServiceAccountToken:  func() *bool { b := false; return &b }(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Spec(tt.args.spec, tt.args.defaultServiceAccount, tt.args.container, tt.args.volumes); !reflect.DeepEqual(got, tt.want) {
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
		spec               v1.SpecInterface
		clusterAnnotations map[string]string
		now                string
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
						Annotations: map[string]string{"v1": "v1"},
					},
				},
				clusterAnnotations: map[string]string{
					string(v1.AnnotationFERestartKey): "true",
				},
				now: "100",
			},
			want: map[string]string{
				common.KubectlRestartAnnotationKey: "100",
				"v1":                               "v1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Annotations(tt.args.spec, tt.args.clusterAnnotations, tt.args.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Annotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVolumeName(t *testing.T) {
	type args struct {
		mountInfo v1.MountInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test get volume name",
			args: args{
				mountInfo: v1.MountInfo{
					Name:      "test",
					MountPath: "/my/path",
				},
			},
			want: "test-3195",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVolumeName(tt.args.mountInfo); got != tt.want {
				t.Errorf("getVolumeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
