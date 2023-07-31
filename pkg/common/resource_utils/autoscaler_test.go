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

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/stretchr/testify/require"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildHorizontalPodAutoscalerV1(t *testing.T) {
	labels := Labels{}
	labels["cluster"] = "test"
	labels["namespace"] = "default"
	pap := &PodAutoscalerParams{
		AutoscalerType: srapi.AutoScalerV1,
		Namespace:      "default",
		Name:           "test",
		Labels:         labels,
		TargetName:     "test-statefulset",
		OwnerReferences: []metav1.OwnerReference{{
			Kind: "StarRocksCluster",
			Name: "test-starrockscluster",
		}},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
		},
	}

	ls := make(map[string]string)
	ls["cluster"] = "test"
	ls["namespace"] = "default"
	ha := &v1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pap.Name,
			Namespace: pap.Namespace,
			Labels:    ls,
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "StarRocksCluster",
				Name: "test-starrockscluster",
			}},
		},
		Spec: v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v1.CrossVersionObjectReference{
				Name:       "test-statefulset",
				Kind:       StatefulSetKind,
				APIVersion: appv1.SchemeGroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
		},
	}
	require.Equal(t, buildAutoscalerV1(pap), ha)
}

func TestBuildHorizontalPodAutoscalerV2beta2(t *testing.T) {
	labels := Labels{}
	labels["cluster"] = "test"
	labels["namespace"] = "default"
	pap := &PodAutoscalerParams{
		AutoscalerType: srapi.AutoScalerV1,
		Namespace:      "default",
		Name:           "test",
		Labels:         labels,
		TargetName:     "test-statefulset",
		OwnerReferences: []metav1.OwnerReference{{
			Kind: "StarRocksCluster",
			Name: "test-starrockscluster",
		}},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
			HPAPolicy: &srapi.HPAPolicy{
				Metrics: []srapi.MetricSpec{{
					Type: srapi.PodsMetricSourceType,
					Object: &srapi.ObjectMetricSource{
						DescribedObject: srapi.CrossVersionObjectReference{
							Kind:       "statefulset",
							Name:       "test-statefulset",
							APIVersion: "apps/v2beta2",
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
					},
					Pods: &srapi.PodsMetricSource{
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					Resource: &srapi.ResourceMetricSource{
						Name: "test",
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					ContainerResource: &srapi.ContainerResourceMetricSource{
						Name: "test",
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Container: "test",
					},
					External: &srapi.ExternalMetricSource{
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
				}},
			},
		},
	}

	ha := &v2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       AutoscalerKind,
			APIVersion: v2beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "StarRocksCluster",
				Name: "test-starrockscluster",
			}},
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				Name:       "test-statefulset",
				Kind:       StatefulSetKind,
				APIVersion: appv1.SchemeGroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
			Metrics: []v2beta2.MetricSpec{{
				Type: v2beta2.PodsMetricSourceType,
				Object: &v2beta2.ObjectMetricSource{
					DescribedObject: v2beta2.CrossVersionObjectReference{
						Kind:       "statefulset",
						Name:       "test-statefulset",
						APIVersion: "apps/v2beta2",
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

	require.Equal(t, ha, buildAutoscalerV2beta2(pap))
}

func TestBuildHorizontalPodAutoscalerV2(t *testing.T) {
	labels := Labels{}
	labels["cluster"] = "test"
	labels["namespace"] = "default"
	pap := &PodAutoscalerParams{
		AutoscalerType: srapi.AutoScalerV1,
		Namespace:      "default",
		Name:           "test",
		Labels:         labels,
		TargetName:     "test-statefulset",
		OwnerReferences: []metav1.OwnerReference{{
			Kind: "StarRocksCluster",
			Name: "test-starrockscluster",
		}},
		ScalerPolicy: &srapi.AutoScalingPolicy{
			Version:     srapi.AutoScalerV1,
			MinReplicas: GetInt32Pointer(1),
			MaxReplicas: 10,
			HPAPolicy: &srapi.HPAPolicy{
				Metrics: []srapi.MetricSpec{{
					Type: srapi.PodsMetricSourceType,
					Object: &srapi.ObjectMetricSource{
						DescribedObject: srapi.CrossVersionObjectReference{
							Kind:       "statefulset",
							Name:       "test-statefulset",
							APIVersion: "apps/v2beta2",
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
					},
					Pods: &srapi.PodsMetricSource{
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					Resource: &srapi.ResourceMetricSource{
						Name: "test",
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
					ContainerResource: &srapi.ContainerResourceMetricSource{
						Name: "test",
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
						Container: "test",
					},
					External: &srapi.ExternalMetricSource{
						Metric: srapi.MetricIdentifier{
							Name: "test",
							Selector: &metav1.LabelSelector{
								MatchLabels: make(map[string]string),
							},
						},
						Target: srapi.MetricTarget{
							Type:               srapi.ValueMetricType,
							Value:              resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
							AverageUtilization: GetInt32Pointer(1),
						},
					},
				}},
			},
		},
	}

	ha := &v2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       AutoscalerKind,
			APIVersion: v2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "StarRocksCluster",
				Name: "test-starrockscluster",
			}},
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Name:       "test-statefulset",
				Kind:       StatefulSetKind,
				APIVersion: appv1.SchemeGroupVersion.String(),
			},
			MaxReplicas: 10,
			MinReplicas: GetInt32Pointer(1),
			Metrics: []v2.MetricSpec{{
				Type: v2.PodsMetricSourceType,
				Object: &v2.ObjectMetricSource{
					DescribedObject: v2.CrossVersionObjectReference{
						Kind:       "statefulset",
						Name:       "test-statefulset",
						APIVersion: "apps/v2beta2",
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

	require.Equal(t, ha, buildAutoscalerV2(pap))
}
