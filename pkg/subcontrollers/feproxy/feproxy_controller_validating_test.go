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

package feproxy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// TestFeProxyController_validating covers the gap where feproxy's storageVolumes were rendered by
// the shared pod.MountStorageVolumes renderer but never validated, so a csi block conflicting with
// a real storageClassName silently fell back to a PersistentVolumeClaim instead of erroring out.
func TestFeProxyController_validating(t *testing.T) {
	csiStorageClass := srapi.CSI
	realStorageClass := "gp3"
	hostPathStorageClass := srapi.HostPath

	tests := []struct {
		name        string
		feProxySpec *srapi.StarRocksFeProxySpec
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "valid csi volume passes",
			feProxySpec: &srapi.StarRocksFeProxySpec{
				StarRocksLoadSpec: srapi.StarRocksLoadSpec{
					StorageVolumes: []srapi.StorageVolume{
						{
							Name:             "spiffe-workload-api",
							StorageClassName: &csiStorageClass,
							MountPath:        "/spiffe-workload-api",
							CSI: &corev1.CSIVolumeSource{
								Driver: "csi.spiffe.io",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "csi together with a real storageClassName returns ErrCSIStorageClass",
			feProxySpec: &srapi.StarRocksFeProxySpec{
				StarRocksLoadSpec: srapi.StarRocksLoadSpec{
					StorageVolumes: []srapi.StorageVolume{
						{
							Name:             "spiffe-workload-api",
							StorageClassName: &realStorageClass,
							MountPath:        "/spiffe-workload-api",
							CSI: &corev1.CSIVolumeSource{
								Driver: "csi.spiffe.io",
							},
						},
					},
				},
			},
			wantErr:   true,
			wantErrIs: srapi.ErrCSIStorageClass,
		},
		{
			name: "hostPath with an empty path returns ErrHostPathRequired",
			feProxySpec: &srapi.StarRocksFeProxySpec{
				StarRocksLoadSpec: srapi.StarRocksLoadSpec{
					StorageVolumes: []srapi.StorageVolume{
						{
							Name:             "tmp-data",
							StorageClassName: &hostPathStorageClass,
							MountPath:        "/tmp",
							HostPath:         &corev1.HostPathVolumeSource{},
						},
					},
				},
			},
			wantErr:   true,
			wantErrIs: srapi.ErrHostPathRequired,
		},
		{
			name: "no storage volumes still validates the service as before",
			feProxySpec: &srapi.StarRocksFeProxySpec{
				StarRocksLoadSpec: srapi.StarRocksLoadSpec{
					Service: &srapi.StarRocksService{
						Type:                  corev1.ServiceTypeClusterIP,
						ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &FeProxyController{}
			err := controller.validating(tt.feProxySpec)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.True(t, errors.Is(err, tt.wantErrIs))
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
