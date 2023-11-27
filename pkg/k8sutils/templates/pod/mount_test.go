package pod

import (
	"reflect"
	"testing"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	corev1 "k8s.io/api/core/v1"
)

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
					Name: "s1-1614",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "s1",
						},
					},
				},
				{
					Name: "s2-1229",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "s2",
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "s1-1614",
					MountPath: "/pkg/mounts/volumes1",
				},
				{
					Name:      "s2-1229",
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
					Name: "s1-1614",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s1",
							},
						},
					},
				},
				{
					Name: "s2-1229",
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
					Name:      "s1-1614",
					MountPath: "/pkg/mounts/volumes1",
				},
				{
					Name:      "s2-1229",
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
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
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
			want: "test-1417",
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
