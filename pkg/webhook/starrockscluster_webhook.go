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
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

// StarRocksClusterValidator validates StarRocksCluster resources
type StarRocksClusterValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// Handle validates incoming StarRocksCluster admission requests
func (v *StarRocksClusterValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	cluster := &srapi.StarRocksCluster{}

	err := v.decoder.Decode(req, cluster)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Validate based on operation type
	switch req.Operation {
	case admissionv1.Create:
		if err := v.validateCreate(ctx, cluster); err != nil {
			return admission.Denied(err.Error())
		}
	case admissionv1.Update:
		oldCluster := &srapi.StarRocksCluster{}
		if err := v.decoder.DecodeRaw(req.OldObject, oldCluster); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := v.validateUpdate(ctx, cluster, oldCluster); err != nil {
			return admission.Denied(err.Error())
		}
	case admissionv1.Delete:
		if err := v.validateDelete(ctx, cluster); err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}

// validateCreate validates StarRocksCluster creation
func (v *StarRocksClusterValidator) validateCreate(_ context.Context, cluster *srapi.StarRocksCluster) error {
	// Validate basic cluster configuration
	if err := v.validateClusterSpec(cluster); err != nil {
		return fmt.Errorf("invalid cluster spec: %w", err)
	}

	// Validate FE configuration
	if cluster.Spec.StarRocksFeSpec != nil {
		if err := v.validateFeSpec(cluster.Spec.StarRocksFeSpec); err != nil {
			return fmt.Errorf("invalid FE spec: %w", err)
		}
	}

	// Validate BE configuration
	if cluster.Spec.StarRocksBeSpec != nil {
		if err := v.validateBeSpec(cluster.Spec.StarRocksBeSpec); err != nil {
			return fmt.Errorf("invalid BE spec: %w", err)
		}
	}

	// Validate CN configuration
	if cluster.Spec.StarRocksCnSpec != nil {
		if err := v.validateCnSpec(cluster.Spec.StarRocksCnSpec); err != nil {
			return fmt.Errorf("invalid CN spec: %w", err)
		}
	}

	return nil
}

// validateUpdate validates StarRocksCluster updates
func (v *StarRocksClusterValidator) validateUpdate(ctx context.Context, newCluster, oldCluster *srapi.StarRocksCluster) error {
	// Validate the new cluster spec
	if err := v.validateCreate(ctx, newCluster); err != nil {
		return err
	}

	// Validate immutable fields
	if err := v.validateImmutableFields(newCluster, oldCluster); err != nil {
		return fmt.Errorf("immutable field changed: %w", err)
	}

	return nil
}

// validateDelete validates StarRocksCluster deletion
func (v *StarRocksClusterValidator) validateDelete(_ context.Context, _ *srapi.StarRocksCluster) error {
	// Add any deletion validation logic here
	// For example, check if cluster has persistent data that needs special handling
	return nil
}

// validateClusterSpec validates basic cluster configuration
func (v *StarRocksClusterValidator) validateClusterSpec(cluster *srapi.StarRocksCluster) error {
	// Ensure at least FE is specified
	if cluster.Spec.StarRocksFeSpec == nil {
		return fmt.Errorf("StarRocks Frontend (FE) specification is required")
	}

	// Validate cluster name length
	if len(cluster.Name) > 63 {
		return fmt.Errorf("cluster name cannot exceed 63 characters")
	}

	return nil
}

// validateFeSpec validates Frontend configuration
func (v *StarRocksClusterValidator) validateFeSpec(spec *srapi.StarRocksFeSpec) error {
	// Validate replica count
	if spec.Replicas != nil && *spec.Replicas <= 0 {
		return fmt.Errorf("FE replicas must be greater than 0")
	}

	// For HA deployment, FE replicas should be odd number >= 3
	if spec.Replicas != nil && *spec.Replicas > 1 {
		if *spec.Replicas < 3 {
			return fmt.Errorf("for HA deployment, FE replicas must be at least 3")
		}
		if *spec.Replicas%2 == 0 {
			return fmt.Errorf("for HA deployment, FE replicas must be an odd number")
		}
	}

	// Validate image
	if spec.Image == "" {
		return fmt.Errorf("FE image is required")
	}

	// Validate resource requests and limits
	if err := v.validateResources(&spec.StarRocksLoadSpec); err != nil {
		return fmt.Errorf("FE resource validation failed: %w", err)
	}

	return nil
}

// validateBeSpec validates Backend configuration
func (v *StarRocksClusterValidator) validateBeSpec(spec *srapi.StarRocksBeSpec) error {
	// Validate replica count
	if spec.Replicas != nil && *spec.Replicas <= 0 {
		return fmt.Errorf("BE replicas must be greater than 0")
	}

	// Validate image
	if spec.Image == "" {
		return fmt.Errorf("BE image is required")
	}

	// Validate resource requests and limits
	if err := v.validateResources(&spec.StarRocksLoadSpec); err != nil {
		return fmt.Errorf("BE resource validation failed: %w", err)
	}

	return nil
}

// validateCnSpec validates Compute Node configuration
func (v *StarRocksClusterValidator) validateCnSpec(spec *srapi.StarRocksCnSpec) error {
	// Validate replica count
	if spec.Replicas != nil && *spec.Replicas < 0 {
		return fmt.Errorf("CN replicas cannot be negative")
	}

	// Validate image
	if spec.Image == "" {
		return fmt.Errorf("CN image is required")
	}

	// Validate resource requests and limits
	if err := v.validateResources(&spec.StarRocksLoadSpec); err != nil {
		return fmt.Errorf("CN resource validation failed: %w", err)
	}

	// Validate autoscaling configuration if present
	if spec.AutoScalingPolicy != nil {
		if err := v.validateAutoScalingPolicy(spec.AutoScalingPolicy); err != nil {
			return fmt.Errorf("CN autoscaling validation failed: %w", err)
		}
	}

	return nil
}

// validateResources validates resource requests and limits
func (v *StarRocksClusterValidator) validateResources(spec *srapi.StarRocksLoadSpec) error {
	// Check if limits are greater than or equal to requests
	if spec.Requests != nil && spec.Limits != nil {
		// Validate CPU
		if spec.Requests.Cpu() != nil && spec.Limits.Cpu() != nil {
			if spec.Requests.Cpu().Cmp(*spec.Limits.Cpu()) > 0 {
				return fmt.Errorf("CPU limit cannot be less than CPU request")
			}
		}
		// Validate Memory
		if spec.Requests.Memory() != nil && spec.Limits.Memory() != nil {
			if spec.Requests.Memory().Cmp(*spec.Limits.Memory()) > 0 {
				return fmt.Errorf("memory limit cannot be less than memory request")
			}
		}
	}

	return nil
}

// validateAutoScalingPolicy validates autoscaling configuration
func (v *StarRocksClusterValidator) validateAutoScalingPolicy(policy *srapi.AutoScalingPolicy) error {
	if policy.MaxReplicas <= 0 {
		return fmt.Errorf("maxReplicas must be greater than 0")
	}

	if policy.MinReplicas != nil && *policy.MinReplicas < 0 {
		return fmt.Errorf("minReplicas cannot be negative")
	}

	if policy.MinReplicas != nil && *policy.MinReplicas > policy.MaxReplicas {
		return fmt.Errorf("minReplicas cannot be greater than maxReplicas")
	}

	// Validate HPA behavior
	if policy.HPAPolicy != nil {
		if len(policy.HPAPolicy.Metrics) == 0 {
			return fmt.Errorf("at least one metric must be specified for HPA")
		}
	}

	return nil
}

// validateImmutableFields validates that immutable fields haven't changed
func (v *StarRocksClusterValidator) validateImmutableFields(newCluster, oldCluster *srapi.StarRocksCluster) error {
	// Validate that certain critical fields haven't changed during updates
	// For example, cluster name cannot be changed
	if newCluster.Name != oldCluster.Name {
		return fmt.Errorf("cluster name is immutable")
	}

	// Add more immutable field validations as needed
	return nil
}

// InjectDecoder sets the decoder for the webhook
func (v *StarRocksClusterValidator) InjectDecoder(decoder *admission.Decoder) error {
	v.decoder = decoder
	return nil
}
