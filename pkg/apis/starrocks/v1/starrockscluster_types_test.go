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

package v1

import (
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestStorageVolumeValidate(t *testing.T) {
	hostPathClass := HostPath
	otherClass := "standard"
	csiClass := CSI
	csiClassUpper := "CSI"
	emptyDirClass := EmptyDir
	tests := []struct {
		name    string
		volume  StorageVolume
		wantErr error
	}{
		{
			name:    "hostPath class without hostPath is invalid",
			volume:  StorageVolume{StorageClassName: &hostPathClass},
			wantErr: ErrHostPathRequired,
		},
		{
			name:    "hostPath class with empty path is invalid",
			volume:  StorageVolume{StorageClassName: &hostPathClass, HostPath: &corev1.HostPathVolumeSource{Path: ""}},
			wantErr: ErrHostPathRequired,
		},
		{
			name:    "hostPath class with path is valid",
			volume:  StorageVolume{StorageClassName: &hostPathClass, HostPath: &corev1.HostPathVolumeSource{Path: "/data"}},
			wantErr: nil,
		},
		{
			name:    "no class but hostPath with empty path is invalid",
			volume:  StorageVolume{HostPath: &corev1.HostPathVolumeSource{Path: ""}},
			wantErr: ErrHostPathRequired,
		},
		{
			name:    "no class with hostPath path is valid",
			volume:  StorageVolume{HostPath: &corev1.HostPathVolumeSource{Path: "/data"}},
			wantErr: nil,
		},
		{
			name:    "normal storage class is valid",
			volume:  StorageVolume{StorageClassName: &otherClass},
			wantErr: nil,
		},
		{
			name:    "csi class without csi is invalid",
			volume:  StorageVolume{StorageClassName: &csiClass},
			wantErr: ErrCSIRequired,
		},
		{
			name:    "csi class with empty driver is invalid",
			volume:  StorageVolume{StorageClassName: &csiClass, CSI: &corev1.CSIVolumeSource{Driver: ""}},
			wantErr: ErrCSIRequired,
		},
		{
			name:    "csi class with driver is valid",
			volume:  StorageVolume{StorageClassName: &csiClass, CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
			wantErr: nil,
		},
		{
			name:    "csi class is case insensitive",
			volume:  StorageVolume{StorageClassName: &csiClassUpper, CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
			wantErr: nil,
		},
		{
			name:    "no class with csi driver is valid",
			volume:  StorageVolume{CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
			wantErr: nil,
		},
		{
			name:    "no class with empty csi driver is invalid",
			volume:  StorageVolume{CSI: &corev1.CSIVolumeSource{Driver: ""}},
			wantErr: ErrCSIRequired,
		},
		{
			name:    "csi with a real storage class is invalid",
			volume:  StorageVolume{StorageClassName: &otherClass, CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
			wantErr: ErrCSIStorageClass,
		},
		{
			name:    "csi with emptyDir class is invalid",
			volume:  StorageVolume{StorageClassName: &emptyDirClass, CSI: &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"}},
			wantErr: ErrCSIStorageClass,
		},
		{
			name: "csi together with hostPath is invalid",
			volume: StorageVolume{
				CSI:      &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
				HostPath: &corev1.HostPathVolumeSource{Path: "/data"},
			},
			wantErr: ErrCSIConflict,
		},
		{
			name: "csi conflict with hostPath takes precedence over a real storage class",
			volume: StorageVolume{
				StorageClassName: &otherClass,
				CSI:              &corev1.CSIVolumeSource{Driver: "csi.spiffe.io"},
				HostPath:         &corev1.HostPathVolumeSource{Path: "/data"},
			},
			wantErr: ErrCSIConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.volume.Validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
