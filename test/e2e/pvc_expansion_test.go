package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

var _ = Describe("PVC Expansion", func() {
	var (
		namespace string
		cluster   *srapi.StarRocksCluster
	)

	BeforeEach(func() {
		namespace = "pvc-expansion-test"
		createNamespace(namespace)

		cluster = &srapi.StarRocksCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: namespace,
			},
			Spec: srapi.StarRocksClusterSpec{
				StarRocksFeSpec: &srapi.StarRocksFeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Replicas: getInt32Ptr(1),
							Image:    "starrocks/fe-ubuntu:3.3-latest",
							StorageVolumes: []srapi.StorageVolume{
								{
									Name:        "fe-meta",
									StorageSize: "10Gi",
									MountPath:   "/opt/starrocks/fe/meta",
								},
								{
									Name:        "fe-log",
									StorageSize: "5Gi",
									MountPath:   "/opt/starrocks/fe/log",
								},
							},
						},
					},
				},
				StarRocksBeSpec: &srapi.StarRocksBeSpec{
					StarRocksComponentSpec: srapi.StarRocksComponentSpec{
						StarRocksLoadSpec: srapi.StarRocksLoadSpec{
							Replicas: getInt32Ptr(1),
							Image:    "starrocks/be-ubuntu:3.3-latest",
							StorageVolumes: []srapi.StorageVolume{
								{
									Name:        "be-data",
									StorageSize: "100Gi",
									MountPath:   "/opt/starrocks/be/storage",
								},
								{
									Name:        "be-log",
									StorageSize: "10Gi",
									MountPath:   "/opt/starrocks/be/log",
								},
							},
						},
					},
				},
			},
		}
	})

	AfterEach(func() {
		if cluster != nil {
			deleteCluster(cluster)
		}
		deleteNamespace(namespace)
	})

	Context("FE PVC Expansion", func() {
		It("should expand FE meta storage successfully", func() {
			By("Creating the initial cluster")
			createCluster(cluster)

			By("Waiting for FE StatefulSet to be ready")
			Eventually(func() bool {
				return isStatefulSetReady(namespace, "test-cluster-fe")
			}, 5*time.Minute, 10*time.Second).Should(BeTrue())

			By("Verifying initial PVC size")
			pvcName := "fe-meta-test-cluster-fe-0"
			initialSize := getPVCSize(namespace, pvcName)
			Expect(initialSize.String()).To(Equal("10Gi"))

			By("Updating FE meta storage size")
			updateCluster := cluster.DeepCopy()
			updateCluster.Spec.StarRocksFeSpec.StorageVolumes[0].StorageSize = "20Gi"
			updateClusterSpec(updateCluster)

			By("Waiting for PVC to be expanded")
			Eventually(func() string {
				return getPVCSize(namespace, pvcName).String()
			}, 2*time.Minute, 5*time.Second).Should(Equal("20Gi"))

			By("Verifying StatefulSet is still ready after expansion")
			Consistently(func() bool {
				return isStatefulSetReady(namespace, "test-cluster-fe")
			}, 30*time.Second, 5*time.Second).Should(BeTrue())
		})

		It("should reject storage size reduction", func() {
			By("Creating the initial cluster")
			createCluster(cluster)

			By("Waiting for FE StatefulSet to be ready")
			Eventually(func() bool {
				return isStatefulSetReady(namespace, "test-cluster-fe")
			}, 5*time.Minute, 10*time.Second).Should(BeTrue())

			By("Attempting to reduce FE meta storage size")
			updateCluster := cluster.DeepCopy()
			updateCluster.Spec.StarRocksFeSpec.StorageVolumes[0].StorageSize = "5Gi"

			By("Expecting the update to fail with validation error")
			err := k8sClient.Update(context.TODO(), updateCluster)
			// The validation should happen in the operator, so we check the cluster status
			Eventually(func() bool {
				var currentCluster srapi.StarRocksCluster
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      cluster.Name,
					Namespace: cluster.Namespace,
				}, &currentCluster)
				if err != nil {
					return false
				}
				// Check if there's an error condition in the status
				return currentCluster.Status.StarRocksFeStatus != nil &&
					currentCluster.Status.StarRocksFeStatus.Phase == srapi.ComponentFailed
			}, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	Context("BE PVC Expansion", func() {
		It("should expand BE data storage successfully", func() {
			By("Creating the initial cluster")
			createCluster(cluster)

			By("Waiting for BE StatefulSet to be ready")
			Eventually(func() bool {
				return isStatefulSetReady(namespace, "test-cluster-be")
			}, 5*time.Minute, 10*time.Second).Should(BeTrue())

			By("Verifying initial PVC size")
			pvcName := "be-data-test-cluster-be-0"
			initialSize := getPVCSize(namespace, pvcName)
			Expect(initialSize.String()).To(Equal("100Gi"))

			By("Updating BE data storage size")
			updateCluster := cluster.DeepCopy()
			updateCluster.Spec.StarRocksBeSpec.StorageVolumes[0].StorageSize = "200Gi"
			updateClusterSpec(updateCluster)

			By("Waiting for PVC to be expanded")
			Eventually(func() string {
				return getPVCSize(namespace, pvcName).String()
			}, 2*time.Minute, 5*time.Second).Should(Equal("200Gi"))

			By("Verifying StatefulSet is still ready after expansion")
			Consistently(func() bool {
				return isStatefulSetReady(namespace, "test-cluster-be")
			}, 30*time.Second, 5*time.Second).Should(BeTrue())
		})
	})
})

// Helper functions

func createCluster(cluster *srapi.StarRocksCluster) {
	Expect(k8sClient.Create(context.TODO(), cluster)).To(Succeed())
}

func deleteCluster(cluster *srapi.StarRocksCluster) {
	Expect(k8sClient.Delete(context.TODO(), cluster)).To(Succeed())
}

func updateClusterSpec(cluster *srapi.StarRocksCluster) {
	Expect(k8sClient.Update(context.TODO(), cluster)).To(Succeed())
}

func isStatefulSetReady(namespace, name string) bool {
	var sts appv1.StatefulSet
	err := k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &sts)
	if err != nil {
		return false
	}

	return sts.Status.ReadyReplicas == *sts.Spec.Replicas
}

func getPVCSize(namespace, name string) resource.Quantity {
	var pvc corev1.PersistentVolumeClaim
	err := k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &pvc)
	if err != nil {
		return resource.Quantity{}
	}

	return pvc.Spec.Resources.Requests[corev1.ResourceStorage]
}

func createNamespace(name string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
}

func deleteNamespace(name string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	Expect(k8sClient.Delete(context.TODO(), ns)).To(Succeed())
}

func getInt32Ptr(i int32) *int32 {
	return &i
}
