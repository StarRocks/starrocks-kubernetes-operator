package pod

import (
	"os"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestMain(m *testing.M) {
	// Set up the test environment
	config.VolumeNameWithHash = true // because this is a default value

	os.Exit(m.Run())
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
			got, got1 := MountConfigMaps(nil, tt.args.volumes, tt.args.volumeMounts, tt.args.configmaps)
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
		},
		{
			name: "storage volume with csi",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
							StorageVolumes: []v1.StorageVolume{
								{
									Name:      "spiffe-workload-api",
									MountPath: "/spiffe-workload-api",
									SubPath:   "sub",
									CSI: &corev1.CSIVolumeSource{
										Driver:   "csi.spiffe.io",
										ReadOnly: func() *bool { b := true; return &b }(),
									},
								},
							},
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "spiffe-workload-api",
					VolumeSource: corev1.VolumeSource{
						CSI: &corev1.CSIVolumeSource{
							Driver:   "csi.spiffe.io",
							ReadOnly: func() *bool { b := true; return &b }(),
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "spiffe-workload-api",
					MountPath: "/spiffe-workload-api",
					SubPath:   "sub",
				},
			},
		},
		{
			name: "readOnly sets the volumeMount read-only for every volume kind",
			args: args{
				spec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						StarRocksLoadSpec: v1.StarRocksLoadSpec{
							StorageVolumes: []v1.StorageVolume{
								{
									Name:      "ro-csi",
									MountPath: "/ro-csi",
									CSI:       &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
									ReadOnly:  true,
								},
								{
									Name:             "ro-emptydir",
									MountPath:        "/ro-emptydir",
									StorageClassName: func() *string { s := v1.EmptyDir; return &s }(),
									ReadOnly:         true,
								},
								{
									Name:      "ro-hostpath",
									MountPath: "/ro-hostpath",
									HostPath:  &corev1.HostPathVolumeSource{Path: "/host/data"},
									ReadOnly:  true,
								},
								{
									Name:             "ro-pvc",
									MountPath:        "/ro-pvc",
									StorageClassName: func() *string { s := "gp3"; return &s }(),
									ReadOnly:         true,
								},
								{
									Name:      "rw-csi",
									MountPath: "/rw-csi",
									CSI:       &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
								},
							},
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name:         "ro-csi",
					VolumeSource: corev1.VolumeSource{CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
				},
				{
					Name:         "ro-emptydir",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				},
				{
					Name:         "ro-hostpath",
					VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/host/data"}},
				},
				{
					Name: "ro-pvc",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "ro-pvc"},
					},
				},
				{
					Name:         "rw-csi",
					VolumeSource: corev1.VolumeSource{CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
				},
			},
			want1: []corev1.VolumeMount{
				{Name: "ro-csi", MountPath: "/ro-csi", ReadOnly: true},
				{Name: "ro-emptydir", MountPath: "/ro-emptydir", ReadOnly: true},
				{Name: "ro-hostpath", MountPath: "/ro-hostpath", ReadOnly: true},
				{Name: "ro-pvc", MountPath: "/ro-pvc", ReadOnly: true},
				{Name: "rw-csi", MountPath: "/rw-csi"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MountStorageVolumes(tt.args.spec)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountStorageVolumes() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountStorageVolumes() got1 = %v, want %v", got1, tt.want1)
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

func TestMountPersistentVolumeClaim(t *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		volumeName   string
		mountPath    string
		subPath      string
	}
	tests := []struct {
		name  string
		args  args
		want  []corev1.Volume
		want1 []corev1.VolumeMount
	}{
		{
			name: "test mount persistent volume claim",
			args: args{
				volumes:      []corev1.Volume{},
				volumeMounts: []corev1.VolumeMount{},
				volumeName:   "fe-meta",
				mountPath:    "/opt/starrocks/fe/fe-meta",
			},
			want: []corev1.Volume{
				{
					Name: "fe-meta",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "fe-meta",
						},
					},
				},
			},
			want1: []corev1.VolumeMount{
				{
					Name:      "fe-meta",
					MountPath: "/opt/starrocks/fe/fe-meta",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MountPersistentVolumeClaim(tt.args.volumes, tt.args.volumeMounts, tt.args.volumeName, tt.args.mountPath, tt.args.subPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MountPersistentVolumeClaim() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MountPersistentVolumeClaim() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSpecialStorageClassName(t *testing.T) {
	type args struct {
		sv v1.StorageVolume
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test storage volume with empty dir",
			args: args{
				sv: v1.StorageVolume{
					StorageClassName: func() *string { s := v1.EmptyDir; return &s }(),
				},
			},
			want: v1.EmptyDir,
		},
		{
			name: "test storage volume with host path",
			args: args{
				sv: v1.StorageVolume{
					StorageClassName: func() *string { s := v1.HostPath; return &s }(),
				},
			},
			want: v1.HostPath,
		},
		{
			name: "test storage volume with persistent volume claim",
			args: args{
				sv: v1.StorageVolume{
					StorageClassName: func() *string { s := "pvc"; return &s }(),
				},
			},
			want: "",
		},
		{
			name: "test storage volume with host path",
			args: args{
				sv: v1.StorageVolume{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "",
						Type: nil,
					},
				},
			},
			want: v1.HostPath,
		},
		{
			name: "test storage volume with host path and pvc name",
			args: args{
				sv: v1.StorageVolume{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "",
						Type: nil,
					},
					StorageClassName: func() *string { s := "pvc"; return &s }(),
				},
			},
			want: "",
		},
		{
			name: "test storage volume with csi storage class",
			args: args{
				sv: v1.StorageVolume{
					StorageClassName: func() *string { s := v1.CSI; return &s }(),
				},
			},
			want: v1.CSI,
		},
		{
			name: "test storage volume with csi storage class in upper case",
			args: args{
				sv: v1.StorageVolume{
					StorageClassName: func() *string { s := "CSI"; return &s }(),
				},
			},
			want: v1.CSI,
		},
		{
			name: "test storage volume with csi source only",
			args: args{
				sv: v1.StorageVolume{
					CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
				},
			},
			want: v1.CSI,
		},
		{
			name: "test storage volume with csi source and pvc name",
			args: args{
				sv: v1.StorageVolume{
					CSI:              &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
					StorageClassName: func() *string { s := "pvc"; return &s }(),
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SpecialStorageClassName(tt.args.sv); got != tt.want {
				t.Errorf("SpecialStorageClassName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountCSIVolume(t *testing.T) {
	csi := &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}
	volumes, mounts := MountCSIVolume(nil, nil, "data", "/data", "sub", csi)
	if len(volumes) != 1 || volumes[0].Name != "data" || volumes[0].CSI != csi {
		t.Errorf("MountCSIVolume() volumes = %v", volumes)
	}
	if len(mounts) != 1 || mounts[0].Name != "data" || mounts[0].MountPath != "/data" || mounts[0].SubPath != "sub" {
		t.Errorf("MountCSIVolume() mounts = %v", mounts)
	}
}
