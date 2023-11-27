package pod

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	corev1 "k8s.io/api/core/v1"
)

const (
	EmptyDir = "emptyDir"
)

func IsSpecialStorageClass(storageClassName *string) bool {
	return storageClassName != nil && (common.EqualsIgnoreCase(*storageClassName, EmptyDir))
}

// MountStorageVolumes parse StorageVolumes from spec and mount them to pod.
// If StorageClassName is EmptyDir, mount an emptyDir volume to pod.
func MountStorageVolumes(spec v1.SpecInterface) ([]corev1.Volume, []corev1.VolumeMount, map[string]bool) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	vexist := make(map[string]bool)
	for _, sv := range spec.GetStorageVolumes() {
		vexist[sv.MountPath] = true
		if IsSpecialStorageClass(sv.StorageClassName) {
			volumes, volumeMounts = MountEmptyDirVolume(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath)
		} else {
			volumes, volumeMounts = MountPersistentVolumeClaim(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath)
		}
	}
	return volumes, volumeMounts, vexist
}

func MountPersistentVolumeClaim(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	volumeName, mountPath, subPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeName,
			},
		},
	})
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
		SubPath:   subPath,
	})
	return volumes, volumeMounts
}

func MountEmptyDirVolume(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	volumeName, mountPath, subPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(
		volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			SubPath:   subPath,
		})
	return volumes, volumeMounts
}

func MountConfigMaps(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	references []v1.ConfigMapReference) ([]corev1.Volume, []corev1.VolumeMount) {
	for _, reference := range references {
		volumeName := getVolumeName(v1.MountInfo(reference))
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: reference.Name,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: reference.MountPath,
			SubPath:   reference.SubPath,
		})
	}
	return volumes, volumeMounts
}

// MountConfigMapInfo parse ConfigMapInfo from spec and mount them to pod.
// Note: we can not reuse MountConfigMaps because it generates a volume name by call getVolumeName,
func MountConfigMapInfo(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	cmInfo v1.ConfigMapInfo, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	if cmInfo.ConfigMapName != "" && cmInfo.ResolveKey != "" {
		volumes = append(volumes, corev1.Volume{
			Name: cmInfo.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cmInfo.ConfigMapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      cmInfo.ConfigMapName,
			MountPath: mountPath,
		})
	}
	return volumes, volumeMounts
}

func MountSecrets(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	references []v1.SecretReference) ([]corev1.Volume, []corev1.VolumeMount) {
	for _, reference := range references {
		volumeName := getVolumeName(v1.MountInfo(reference))
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: reference.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: reference.MountPath,
			SubPath:   reference.SubPath,
		})
	}
	return volumes, volumeMounts
}

func getVolumeName(mountInfo v1.MountInfo) string {
	suffixLen := 4
	suffix := hash.HashObject(mountInfo)
	if len(suffix) > suffixLen {
		suffix = suffix[:suffixLen]
	}
	return mountInfo.Name + "-" + suffix
}
