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
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestStarRocksService_Validate(t *testing.T) {
	tests := []struct {
		name    string
		svc     *StarRocksService
		wantErr bool
	}{
		{
			name: "nil service is valid",
			svc:  nil,
		},
		{
			name: "empty service is valid",
			svc:  &StarRocksService{},
		},
		{
			name: "LoadBalancer with Local policy is valid",
			svc: &StarRocksService{
				Type:                  corev1.ServiceTypeLoadBalancer,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "NodePort with Cluster policy is valid",
			svc: &StarRocksService{
				Type:                  corev1.ServiceTypeNodePort,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeCluster,
			},
		},
		{
			name: "explicit ClusterIP with policy is invalid",
			svc: &StarRocksService{
				Type:                  corev1.ServiceTypeClusterIP,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
			wantErr: true,
		},
		{
			name: "default ClusterIP with policy is invalid",
			svc: &StarRocksService{
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
			wantErr: true,
		},
		{
			name: "ClusterIP without policy is valid",
			svc: &StarRocksService{
				Type: corev1.ServiceTypeClusterIP,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.svc.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
