package k8sutils

import (
	"context"
	"testing"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestDetectPVCExpansion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)

	tests := []struct {
		name                       string
		existingStatefulSet        *appv1.StatefulSet
		existingPVCs               []corev1.PersistentVolumeClaim
		newStorageVolumes          []v1.StorageVolume
		expectedNeedsExpansion     bool
		expectedNeedsRecreation    bool
		expectedOnlyPVCSizeChanged bool
		expectedPVCCount           int
		expectedValidationErrors   int
	}{
		{
			name: "no expansion needed - same size",
			existingStatefulSet: &appv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-fe",
					Namespace: "default",
				},
				Spec: appv1.StatefulSetSpec{
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
			newStorageVolumes: []v1.StorageVolume{
				{
					Name:             "fe-meta",
					StorageSize:      "10Gi",
					StorageClassName: nil, // Same as existing
				},
			},
			expectedNeedsExpansion:     false,
			expectedNeedsRecreation:    false,
			expectedOnlyPVCSizeChanged: false,
			expectedPVCCount:           0,
			expectedValidationErrors:   0,
		},
		{
			name: "expansion needed - size increase",
			existingStatefulSet: &appv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-fe",
					Namespace: "default",
				},
				Spec: appv1.StatefulSetSpec{
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
								StorageClassName: stringPtr("standard"), // Use expandable storage class
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
			existingPVCs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fe-meta-test-fe-0",
						Namespace: "default",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
			newStorageVolumes: []v1.StorageVolume{
				{
					Name:             "fe-meta",
					StorageSize:      "20Gi",
					StorageClassName: stringPtr("standard"), // Use expandable storage class
				},
			},
			expectedNeedsExpansion:     true,
			expectedNeedsRecreation:    false,
			expectedOnlyPVCSizeChanged: true, // Only PVC size changed
			expectedPVCCount:           1,
			expectedValidationErrors:   0,
		},
		{
			name: "validation error - size decrease",
			existingStatefulSet: &appv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-fe",
					Namespace: "default",
				},
				Spec: appv1.StatefulSetSpec{
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("20Gi"),
									},
								},
							},
						},
					},
				},
			},
			newStorageVolumes: []v1.StorageVolume{
				{
					Name:             "fe-meta",
					StorageSize:      "10Gi",
					StorageClassName: nil, // Same as existing
				},
			},
			expectedNeedsExpansion:     false,
			expectedNeedsRecreation:    false,
			expectedOnlyPVCSizeChanged: false,
			expectedPVCCount:           0,
			expectedValidationErrors:   1,
		},
		{
			name: "recreation needed - storage class change",
			existingStatefulSet: &appv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-fe",
					Namespace: "default",
				},
				Spec: appv1.StatefulSetSpec{
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
								StorageClassName: stringPtr("standard"),
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
			newStorageVolumes: []v1.StorageVolume{
				{
					Name:             "fe-meta",
					StorageClassName: stringPtr("premium"),
					StorageSize:      "10Gi",
				},
			},
			expectedNeedsExpansion:     false,
			expectedNeedsRecreation:    true,
			expectedOnlyPVCSizeChanged: false,
			expectedPVCCount:           0,
			expectedValidationErrors:   0,
		},
		{
			name: "storage class does not support expansion",
			existingStatefulSet: &appv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-fe",
					Namespace: "default",
				},
				Spec: appv1.StatefulSetSpec{
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "fe-meta"},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
								StorageClassName: stringPtr("no-expansion"),
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("10Gi"),
									},
								},
							},
						},
					},
				},
			},
			newStorageVolumes: []v1.StorageVolume{
				{
					Name:             "fe-meta",
					StorageClassName: stringPtr("no-expansion"),
					StorageSize:      "20Gi",
				},
			},
			expectedNeedsExpansion:     false,
			expectedNeedsRecreation:    false,
			expectedOnlyPVCSizeChanged: false,
			expectedPVCCount:           0,
			expectedValidationErrors:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}
			if tt.existingStatefulSet != nil {
				objects = append(objects, tt.existingStatefulSet)
			}
			for i := range tt.existingPVCs {
				objects = append(objects, &tt.existingPVCs[i])
			}

			// Add storage classes for testing
			objects = append(objects, &storagev1.StorageClass{
				ObjectMeta:           metav1.ObjectMeta{Name: "standard"},
				AllowVolumeExpansion: boolPtr(true),
			})
			objects = append(objects, &storagev1.StorageClass{
				ObjectMeta:           metav1.ObjectMeta{Name: "premium"},
				AllowVolumeExpansion: boolPtr(true),
			})
			objects = append(objects, &storagev1.StorageClass{
				ObjectMeta:           metav1.ObjectMeta{Name: "no-expansion"},
				AllowVolumeExpansion: boolPtr(false),
			})

			client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

			result, err := DetectPVCExpansion(context.TODO(), client, "default", "test-fe", tt.newStorageVolumes)
			if err != nil {
				t.Fatalf("DetectPVCExpansion() error = %v", err)
			}

			if result.NeedsExpansion != tt.expectedNeedsExpansion {
				t.Errorf("DetectPVCExpansion() NeedsExpansion = %v, want %v", result.NeedsExpansion, tt.expectedNeedsExpansion)
			}

			if result.NeedsStatefulSetRecreation != tt.expectedNeedsRecreation {
				t.Errorf("DetectPVCExpansion() NeedsStatefulSetRecreation = %v, want %v", result.NeedsStatefulSetRecreation, tt.expectedNeedsRecreation)
			}

			if result.OnlyPVCSizeChanged != tt.expectedOnlyPVCSizeChanged {
				t.Errorf("DetectPVCExpansion() OnlyPVCSizeChanged = %v, want %v", result.OnlyPVCSizeChanged, tt.expectedOnlyPVCSizeChanged)
			}

			if len(result.PVCsToExpand) != tt.expectedPVCCount {
				t.Errorf("DetectPVCExpansion() PVCsToExpand count = %v, want %v", len(result.PVCsToExpand), tt.expectedPVCCount)
			}

			if len(result.ValidationErrors) != tt.expectedValidationErrors {
				t.Errorf("DetectPVCExpansion() ValidationErrors count = %v, want %v", len(result.ValidationErrors), tt.expectedValidationErrors)
			}
		})
	}
}

func TestExpandPVCs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(pvc).Build()

	expansionInfos := []PVCExpansionInfo{
		{
			PVCName:     "test-pvc",
			Namespace:   "default",
			CurrentSize: resource.MustParse("10Gi"),
			NewSize:     resource.MustParse("20Gi"),
			VolumeName:  "test-volume",
		},
	}

	err := ExpandPVCs(context.TODO(), client, expansionInfos)
	if err != nil {
		t.Fatalf("ExpandPVCs() error = %v", err)
	}

	// Verify the PVC was updated
	var updatedPVC corev1.PersistentVolumeClaim
	err = client.Get(context.TODO(), types.NamespacedName{Name: "test-pvc", Namespace: "default"}, &updatedPVC)
	if err != nil {
		t.Fatalf("Failed to get updated PVC: %v", err)
	}

	expectedSize := resource.MustParse("20Gi")
	actualSize := updatedPVC.Spec.Resources.Requests[corev1.ResourceStorage]
	if !actualSize.Equal(expectedSize) {
		t.Errorf("PVC size not updated correctly: got %v, want %v", actualSize, expectedSize)
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestValidateStorageClassExpansion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = storagev1.AddToScheme(scheme)

	tests := []struct {
		name             string
		storageClassName *string
		storageClasses   []storagev1.StorageClass
		expectError      bool
		errorContains    string
	}{
		{
			name:             "storage class supports expansion",
			storageClassName: stringPtr("expandable"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "expandable"},
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectError: false,
		},
		{
			name:             "storage class does not support expansion",
			storageClassName: stringPtr("non-expandable"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "non-expandable"},
					AllowVolumeExpansion: boolPtr(false),
				},
			},
			expectError:   true,
			errorContains: "does not support volume expansion",
		},
		{
			name:             "storage class not found",
			storageClassName: stringPtr("missing"),
			storageClasses:   []storagev1.StorageClass{},
			expectError:      true,
			errorContains:    "not found",
		},
		{
			name:             "emptyDir storage class (special case)",
			storageClassName: stringPtr("emptyDir"),
			storageClasses:   []storagev1.StorageClass{},
			expectError:      false,
		},
		{
			name:             "hostPath storage class (special case)",
			storageClassName: stringPtr("hostPath"),
			storageClasses:   []storagev1.StorageClass{},
			expectError:      false,
		},
		{
			name:             "default storage class supports expansion",
			storageClassName: nil,
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-sc",
						Annotations: map[string]string{
							"storageclass.kubernetes.io/is-default-class": "true",
						},
					},
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectError: false,
		},
		{
			name:             "default storage class does not support expansion",
			storageClassName: nil,
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-sc",
						Annotations: map[string]string{
							"storageclass.kubernetes.io/is-default-class": "true",
						},
					},
					AllowVolumeExpansion: boolPtr(false),
				},
			},
			expectError:   true,
			errorContains: "does not support volume expansion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}
			for i := range tt.storageClasses {
				objects = append(objects, &tt.storageClasses[i])
			}

			client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

			err := ValidateStorageClassExpansion(context.TODO(), client, tt.storageClassName)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateStorageClassExpansion() expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateStorageClassExpansion() error = %v, expected to contain %v", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStorageClassExpansion() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCheckStorageClassRequiresDetachment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = storagev1.AddToScheme(scheme)

	tests := []struct {
		name               string
		storageClassName   *string
		storageClasses     []storagev1.StorageClass
		expectedDetachment bool
		expectError        bool
	}{
		{
			name:             "Unknown provisioner requires detachment (conservative default)",
			storageClassName: stringPtr("unknown-storage"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "unknown-storage"},
					Provisioner:          "unknown.csi.driver",
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectedDetachment: true,
			expectError:        false,
		},
		{
			name:             "Azure Disk CSI requires detachment (unknown provisioner)",
			storageClassName: stringPtr("azure-disk"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "azure-disk"},
					Provisioner:          "disk.csi.azure.com",
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectedDetachment: true,
			expectError:        false,
		},
		{
			name:             "GCE PD supports online expansion",
			storageClassName: stringPtr("gce-pd"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "gce-pd"},
					Provisioner:          "kubernetes.io/gce-pd",
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectedDetachment: false,
			expectError:        false,
		},
		{
			name:             "AWS EBS supports online expansion",
			storageClassName: stringPtr("aws-ebs"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "aws-ebs"},
					Provisioner:          "ebs.csi.aws.com",
					AllowVolumeExpansion: boolPtr(true),
				},
			},
			expectedDetachment: false,
			expectError:        false,
		},
		{
			name:             "Storage class with explicit offline expansion mode",
			storageClassName: stringPtr("offline-expansion"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "offline-expansion"},
					Provisioner:          "custom.csi.driver",
					AllowVolumeExpansion: boolPtr(true),
					Parameters: map[string]string{
						"expansion-mode": "offline",
					},
				},
			},
			expectedDetachment: true,
			expectError:        false,
		},
		{
			name:             "Storage class with explicit online expansion mode",
			storageClassName: stringPtr("online-expansion"),
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta:           metav1.ObjectMeta{Name: "online-expansion"},
					Provisioner:          "custom.csi.driver",
					AllowVolumeExpansion: boolPtr(true),
					Parameters: map[string]string{
						"expansion-mode": "online",
					},
				},
			},
			expectedDetachment: false,
			expectError:        false,
		},
		{
			name:               "emptyDir storage class (special case)",
			storageClassName:   stringPtr("emptyDir"),
			storageClasses:     []storagev1.StorageClass{},
			expectedDetachment: false,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}
			for i := range tt.storageClasses {
				objects = append(objects, &tt.storageClasses[i])
			}

			client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

			requiresDetachment, err := CheckStorageClassRequiresDetachment(context.TODO(), client, tt.storageClassName)

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckStorageClassRequiresDetachment() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("CheckStorageClassRequiresDetachment() unexpected error = %v", err)
				}
				if requiresDetachment != tt.expectedDetachment {
					t.Errorf("CheckStorageClassRequiresDetachment() = %v, want %v", requiresDetachment, tt.expectedDetachment)
				}
			}
		})
	}
}
