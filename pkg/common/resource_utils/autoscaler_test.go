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

package resource_utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

const _defaultNamespace = "default"
const _defaultName = "test"

func TestBuildHorizontalPodAutoscalerV1(t *testing.T) {
	labels := map[string]string{}
	labels["cluster"] = _defaultName
	labels["namespace"] = _defaultNamespace
	cluster := srapi.StarRocksCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       StarRocksClusterKind,
			APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      _defaultName,
			Namespace: _defaultNamespace,
		},
	}
	hpaParams := &HPAParams{
		Version:         srapi.AutoScalerV1,
		Namespace:       _defaultNamespace,
		Name:            "test-autoscaler",
		Labels:          labels,
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&cluster, cluster.GroupVersionKind())},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
		},
	}

	ls := make(map[string]string)
	ls["cluster"] = _defaultName
	ls["namespace"] = _defaultNamespace
	ha := &v1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpaParams.Name,
			Namespace: hpaParams.Namespace,
			Labels:    ls,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
				Kind:       StarRocksClusterKind,
				Name:       _defaultName,
				Controller: func() *bool {
					b := true
					return &b
				}(),
				BlockOwnerDeletion: func() *bool {
					b := true
					return &b
				}(),
			}},
		},
		Spec: v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v1.CrossVersionObjectReference{
				Kind:       StarRocksClusterKind,
				Name:       _defaultName,
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
		},
	}
	require.Equal(t, BuildHPA(hpaParams, srapi.AutoScalerV1), ha)
}

func TestBuildHorizontalPodAutoscalerV2beta2(t *testing.T) {
	labels := map[string]string{}
	labels["cluster"] = "test"
	labels["namespace"] = _defaultNamespace
	cluster := srapi.StarRocksCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       StarRocksClusterKind,
			APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      _defaultName,
			Namespace: _defaultNamespace,
		},
	}
	hpaParams := &HPAParams{
		Version:         srapi.AutoScalerV1,
		Namespace:       _defaultNamespace,
		Name:            "test-autoscaler",
		Labels:          labels,
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&cluster, cluster.GroupVersionKind())},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
			HPAPolicy: &srapi.HPAPolicy{
				Metrics: []v2beta2.MetricSpec{{
					Type: v2beta2.PodsMetricSourceType,
					Object: &v2beta2.ObjectMetricSource{
						DescribedObject: v2beta2.CrossVersionObjectReference{
							Kind:       StarRocksClusterKind,
							Name:       _defaultName,
							APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
					},
					Pods: &v2beta2.PodsMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					Resource: &v2beta2.ResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					ContainerResource: &v2beta2.ContainerResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Container: "test",
					},
					External: &v2beta2.ExternalMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
				}},
			},
		},
	}

	expectedHPA := &v2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       AutoscalerKind,
			APIVersion: v2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-autoscaler",
			Namespace: _defaultNamespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
				Kind:       StarRocksClusterKind,
				Name:       _defaultName,
				Controller: func() *bool {
					b := true
					return &b
				}(),
				BlockOwnerDeletion: func() *bool {
					b := true
					return &b
				}(),
			}},
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				Name:       _defaultName,
				Kind:       StarRocksClusterKind,
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
			Metrics: []v2beta2.MetricSpec{{
				Type: v2beta2.PodsMetricSourceType,
				Object: &v2beta2.ObjectMetricSource{
					DescribedObject: v2beta2.CrossVersionObjectReference{
						Kind:       StarRocksClusterKind,
						Name:       _defaultName,
						APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
					},
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Metric: v2beta2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
				},
				Pods: &v2beta2.PodsMetricSource{
					Metric: v2beta2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				Resource: &v2beta2.ResourceMetricSource{
					Name: "test",
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				ContainerResource: &v2beta2.ContainerResourceMetricSource{
					Name: "test",
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Container: "test",
				},
				External: &v2beta2.ExternalMetricSource{
					Metric: v2beta2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
			}},
		},
	}

	require.Equal(t, expectedHPA, BuildHPA(hpaParams, srapi.AutoScalerV2Beta2))
}

func TestBuildHorizontalPodAutoscalerV2(t *testing.T) {
	labels := map[string]string{}
	labels["cluster"] = "test"
	labels["namespace"] = _defaultNamespace

	cluster := srapi.StarRocksCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       StarRocksClusterKind,
			APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      _defaultName,
			Namespace: _defaultNamespace,
		},
	}

	hpaParams := &HPAParams{
		Version:         srapi.AutoScalerV1,
		Namespace:       _defaultNamespace,
		Name:            "test-autoscaler",
		Labels:          labels,
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&cluster, cluster.GroupVersionKind())},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
			HPAPolicy: &srapi.HPAPolicy{
				Metrics: []v2beta2.MetricSpec{{
					Type: v2beta2.PodsMetricSourceType,
					Object: &v2beta2.ObjectMetricSource{
						DescribedObject: v2beta2.CrossVersionObjectReference{
							Kind:       StarRocksClusterKind,
							Name:       _defaultName,
							APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
					},
					Pods: &v2beta2.PodsMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					Resource: &v2beta2.ResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					ContainerResource: &v2beta2.ContainerResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Container: "test",
					},
					External: &v2beta2.ExternalMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
				}},
			},
		},
	}

	expectedHPA := &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       AutoscalerKind,
			APIVersion: v2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-autoscaler",
			Namespace: _defaultNamespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
				Kind:       StarRocksClusterKind,
				Name:       _defaultName,
				Controller: func() *bool {
					b := true
					return &b
				}(),
				BlockOwnerDeletion: func() *bool {
					b := true
					return &b
				}(),
			}},
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Kind:       StarRocksClusterKind,
				Name:       _defaultName,
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
			Metrics: []v2.MetricSpec{{
				Type: v2.PodsMetricSourceType,
				Object: &v2.ObjectMetricSource{
					DescribedObject: v2.CrossVersionObjectReference{
						Kind:       StarRocksClusterKind,
						Name:       _defaultName,
						APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
				},
				Pods: &v2.PodsMetricSource{
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				Resource: &v2.ResourceMetricSource{
					Name: "test",
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				ContainerResource: &v2.ContainerResourceMetricSource{
					Name: "test",
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Container: "test",
				},
				External: &v2.ExternalMetricSource{
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
			}},
		},
	}

	require.Equal(t, expectedHPA, BuildHPA(hpaParams, srapi.AutoScalerV2))
}

func TestBuildHorizontalPodAutoscalerV2_ForWarehouse(t *testing.T) {
	labels := Labels{}
	labels["cluster"] = _defaultName
	labels["namespace"] = _defaultNamespace

	warehouse := srapi.StarRocksWarehouse{
		TypeMeta: metav1.TypeMeta{
			Kind:       StarRocksWarehouseKind,
			APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      _defaultName,
			Namespace: _defaultNamespace,
		},
		Spec: srapi.StarRocksWarehouseSpec{
			StarRocksCluster: _defaultName,
		},
	}

	hpaParams := &HPAParams{
		Version:         srapi.AutoScalerV1,
		Namespace:       _defaultNamespace,
		Name:            "test-autoscaler",
		Labels:          labels,
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&warehouse, warehouse.GroupVersionKind())},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
			HPAPolicy: &srapi.HPAPolicy{
				Metrics: []v2beta2.MetricSpec{{
					Type: v2beta2.PodsMetricSourceType,
					Object: &v2beta2.ObjectMetricSource{
						DescribedObject: v2beta2.CrossVersionObjectReference{
							Kind:       StarRocksWarehouseKind,
							Name:       _defaultName,
							APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
					},
					Pods: &v2beta2.PodsMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					Resource: &v2beta2.ResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					ContainerResource: &v2beta2.ContainerResourceMetricSource{
						Name: "test",
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Container: "test",
					},
					External: &v2beta2.ExternalMetricSource{
						Metric: v2beta2.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
				}},
			},
		},
	}

	expectedHPA := &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       AutoscalerKind,
			APIVersion: v2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-autoscaler",
			Namespace: _defaultNamespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
				Kind:       StarRocksWarehouseKind,
				Name:       _defaultName,
				Controller: func() *bool {
					b := true
					return &b
				}(),
				BlockOwnerDeletion: func() *bool {
					b := true
					return &b
				}(),
			}},
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Kind:       StarRocksWarehouseKind,
				Name:       _defaultName,
				APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
			Metrics: []v2.MetricSpec{{
				Type: v2.PodsMetricSourceType,
				Object: &v2.ObjectMetricSource{
					DescribedObject: v2.CrossVersionObjectReference{
						Kind:       StarRocksWarehouseKind,
						Name:       _defaultName,
						APIVersion: srapi.SchemeBuilder.GroupVersion.String(),
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
				},
				Pods: &v2.PodsMetricSource{
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				Resource: &v2.ResourceMetricSource{
					Name: "test",
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
				ContainerResource: &v2.ContainerResourceMetricSource{
					Name: "test",
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
					Container: "test",
				},
				External: &v2.ExternalMetricSource{
					Metric: v2.MetricIdentifier{
						Name: "test",
						Selector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Target: v2.MetricTarget{
						Type:               v2.ValueMetricType,
						Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
						AverageUtilization: GetInt32Pointer(1),
					},
				},
			}},
		},
	}

	require.Equal(t, expectedHPA, BuildHPA(hpaParams, srapi.AutoScalerV2))
}
