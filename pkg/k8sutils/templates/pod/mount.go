package pod

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

// SpecialStorageClassName returns the special storage class name of the storage volume, else return "".
// Now we support HostPath, EmptyDir and CSI as special storage class.
func SpecialStorageClassName(sv v1.StorageVolume) string {
	storageClassName := sv.StorageClassName
	if storageClassName != nil {
		if common.EqualsIgnoreCase(*storageClassName, v1.EmptyDir) {
			return v1.EmptyDir
		} else if common.EqualsIgnoreCase(*storageClassName, v1.HostPath) {
			return v1.HostPath
		} else if common.EqualsIgnoreCase(*storageClassName, v1.CSI) {
			return v1.CSI
		}
		return ""
	}

	if sv.HostPath != nil {
		return v1.HostPath
	}

	if sv.CSI != nil {
		return v1.CSI
	}

	return ""
}

// MountStorageVolumes parse StorageVolumes from spec and mount them to pod.
// If StorageClassName is EmptyDir, mount an emptyDir volume to pod.
// If Ephemeral is true, mount a generic ephemeral volume to pod.
func MountStorageVolumes(spec v1.SpecInterface) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	for _, sv := range spec.GetStorageVolumes() {
		if strings.HasPrefix(sv.StorageSize, "0") {
			continue
		}
		switch name := SpecialStorageClassName(sv); name {
		case v1.EmptyDir:
			volumes, volumeMounts = MountEmptyDirVolume(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath)
		case v1.HostPath:
			volumes, volumeMounts = MountHostPathVolume(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath, sv.HostPath)
		case v1.CSI:
			volumes, volumeMounts = MountCSIVolume(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath, sv.CSI)
		default:
			if sv.Ephemeral {
				volumes, volumeMounts = MountEphemeralVolume(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath,
					PersistentVolumeClaimSpec(sv))
			} else {
				volumes, volumeMounts = MountPersistentVolumeClaim(volumes, volumeMounts, sv.Name, sv.MountPath, sv.SubPath)
			}
		}
		// ReadOnly is applied here rather than inside each Mount* helper: those helpers are also
		// called directly by the component pod builders to mount the log and meta directories,
		// which are always writable, so only a user-declared StorageVolume can carry this flag.
		// Every branch above appends exactly one volumeMount, so the last entry is this volume's.
		if sv.ReadOnly && len(volumeMounts) > 0 {
			volumeMounts[len(volumeMounts)-1].ReadOnly = true
		}
	}
	return volumes, volumeMounts
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

// PersistentVolumeClaimSpec builds the claim a storage volume asks for. The StatefulSet declares
// it as a volumeClaimTemplate, and a generic ephemeral volume declares it in the pod spec.
func PersistentVolumeClaimSpec(sv v1.StorageVolume) corev1.PersistentVolumeClaimSpec {
	spec := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		StorageClassName: sv.StorageClassName,
	}
	if sv.StorageSize != "" {
		spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(sv.StorageSize),
			},
		}
	}
	return spec
}

// MountEphemeralVolume mounts a generic ephemeral volume into the pod. The claim template is part of
// the pod spec, so the PersistentVolumeClaim is created with the pod and deleted with it.
func MountEphemeralVolume(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	volumeName string, mountPath string, subPath string,
	claim corev1.PersistentVolumeClaimSpec) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Ephemeral: &corev1.EphemeralVolumeSource{
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
					Spec: claim,
				},
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

func MountHostPathVolume(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	volumeName string, mountPath string, subPath string,
	hostPath *corev1.HostPathVolumeSource) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: hostPath,
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

// MountCSIVolume mounts an ephemeral inline CSI volume into the pod. No PersistentVolumeClaim is
// created for it: the volume is published by the CSI driver at pod start and torn down with the pod.
func MountCSIVolume(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	volumeName string, mountPath string, subPath string,
	csi *corev1.CSIVolumeSource) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes = append(volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			CSI: csi,
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

func MountConfigMaps(spec v1.SpecInterface, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount,
	references []v1.ConfigMapReference) ([]corev1.Volume, []corev1.VolumeMount) {
	prerequisitesOfChangingMode := spec != nil && (spec.GetCommand() != nil || spec.GetArgs() != nil)

	for _, reference := range references {
		volumeName := getVolumeName(v1.MountInfo(reference))
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: reference.Name,
					},
					DefaultMode: func() *int32 {
						if prerequisitesOfChangingMode && reference.SubPath != "" {
							const executionPermission = int32(0755)
							v := executionPermission
							return &v
						}
						return nil
					}(),
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
	if config.VolumeNameWithHash {
		suffixLen := 4
		suffix := hash.HashObject(mountInfo)
		if len(suffix) > suffixLen {
			suffix = suffix[:suffixLen]
		}
		return mountInfo.Name + "-" + suffix
	}
	return mountInfo.Name
}
