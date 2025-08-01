/*
Copyright 2021-present, StarRocks Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"context"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestStarRocksClusterValidator_validateCreate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)

	validator := &StarRocksClusterValidator{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
	}

	tests := []struct {
		name    string
		cluster *srapi.StarRocksCluster
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cluster with FE only",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "starrocks/fe-ubuntu:latest",
								Replicas: &[]int32{1}[0],
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cluster without FE",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksBeSpec: &srapi.StarRocksBeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image: "starrocks/be-ubuntu:latest",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "StarRocks Frontend (FE) specification is required",
		},
		{
			name: "invalid FE with even replicas for HA",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "starrocks/fe-ubuntu:latest",
								Replicas: &[]int32{4}[0],
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "for HA deployment, FE replicas must be an odd number",
		},
		{
			name: "invalid FE with less than 3 replicas for HA",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "starrocks/fe-ubuntu:latest",
								Replicas: &[]int32{2}[0],
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "for HA deployment, FE replicas must be at least 3",
		},
		{
			name: "valid HA cluster with 3 FE replicas",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "starrocks/fe-ubuntu:latest",
								Replicas: &[]int32{3}[0],
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cluster name too long",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "this-is-a-very-long-cluster-name-that-exceeds-the-maximum-allowed-length-of-63-characters",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "starrocks/fe-ubuntu:latest",
								Replicas: &[]int32{1}[0],
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "cluster name cannot exceed 63 characters",
		},
		{
			name: "invalid FE without image",
			cluster: &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: srapi.StarRocksClusterSpec{
					StarRocksFeSpec: &srapi.StarRocksFeSpec{
						StarRocksComponentSpec: srapi.StarRocksComponentSpec{
							StarRocksLoadSpec: srapi.StarRocksLoadSpec{
								Image:    "",
								Replicas: &[]int32{1}[0],
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "FE image is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCreate(context.Background(), tt.cluster)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateCreate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateCreate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCreate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestStarRocksClusterValidator_validateAutoScalingPolicy(t *testing.T) {
	validator := &StarRocksClusterValidator{}

	tests := []struct {
		name    string
		policy  *srapi.AutoScalingPolicy
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid autoscaling policy",
			policy: &srapi.AutoScalingPolicy{
				MaxReplicas: 10,
				MinReplicas: &[]int32{1}[0],
				HPAPolicy: &srapi.HPAPolicy{
					Metrics: []autoscalingv2beta2.MetricSpec{
						{
							Type: autoscalingv2beta2.ResourceMetricSourceType,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid maxReplicas zero",
			policy: &srapi.AutoScalingPolicy{
				MaxReplicas: 0,
			},
			wantErr: true,
			errMsg:  "maxReplicas must be greater than 0",
		},
		{
			name: "invalid minReplicas negative",
			policy: &srapi.AutoScalingPolicy{
				MaxReplicas: 10,
				MinReplicas: &[]int32{-1}[0],
			},
			wantErr: true,
			errMsg:  "minReplicas cannot be negative",
		},
		{
			name: "invalid minReplicas greater than maxReplicas",
			policy: &srapi.AutoScalingPolicy{
				MaxReplicas: 5,
				MinReplicas: &[]int32{10}[0],
			},
			wantErr: true,
			errMsg:  "minReplicas cannot be greater than maxReplicas",
		},
		{
			name: "invalid HPA without metrics",
			policy: &srapi.AutoScalingPolicy{
				MaxReplicas: 10,
				HPAPolicy: &srapi.HPAPolicy{
					Metrics: []autoscalingv2beta2.MetricSpec{},
				},
			},
			wantErr: true,
			errMsg:  "at least one metric must be specified for HPA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAutoScalingPolicy(tt.policy)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateAutoScalingPolicy() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateAutoScalingPolicy() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateAutoScalingPolicy() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestStarRocksClusterValidator_Handle(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = srapi.AddToScheme(scheme)

	validator := &StarRocksClusterValidator{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
	}

	decoder, err := admission.NewDecoder(scheme)
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	_ = validator.InjectDecoder(decoder)

	validCluster := &srapi.StarRocksCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: srapi.StarRocksClusterSpec{
			StarRocksFeSpec: &srapi.StarRocksFeSpec{
				StarRocksComponentSpec: srapi.StarRocksComponentSpec{
					StarRocksLoadSpec: srapi.StarRocksLoadSpec{
						Image:    "starrocks/fe-ubuntu:latest",
						Replicas: &[]int32{1}[0],
					},
				},
			},
		},
	}

	codec := serializer.NewCodecFactory(scheme).LegacyCodec(srapi.GroupVersion)
	validClusterRaw, err := runtime.Encode(codec, validCluster)
	if err != nil {
		t.Fatalf("Failed to encode cluster: %v", err)
	}

	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: validClusterRaw,
			},
		},
	}

	resp := validator.Handle(context.Background(), req)
	if !resp.Allowed {
		t.Errorf("Expected admission to be allowed, got denied with message: %s", resp.Result.Message)
	}
}
