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

package k8sutils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-logr/logr"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

// ServiceEqual judges two services equal or not in some fields. developer can custom the function.
type ServiceEqual func(expect *corev1.Service, actual *corev1.Service) bool

// StatefulSetEqual judges two statefulset equal or not in some fields. developer can custom the function.
type StatefulSetEqual func(expect *appv1.StatefulSet, actual *appv1.StatefulSet) bool

func ApplyService(ctx context.Context, k8sClient client.Client, expectSvc *corev1.Service, equal ServiceEqual) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update k8s service", "name", expectSvc.Name)

	var actualSvc corev1.Service
	err := k8sClient.Get(ctx, types.NamespacedName{Name: expectSvc.Name, Namespace: expectSvc.Namespace}, &actualSvc)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sClient, expectSvc)
	} else if err != nil {
		return err
	}

	if equal(expectSvc, &actualSvc) {
		logger.Info("expectHash == actualHash, no need to update service resource")
		return nil
	}

	expectSvc.ResourceVersion = actualSvc.ResourceVersion
	return k8sClient.Patch(ctx, expectSvc, client.Merge)
}

func ApplyDeployment(ctx context.Context, k8sClient client.Client, deploy *appv1.Deployment) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update deployment", "name", deploy.Name)

	var actual appv1.Deployment
	err := k8sClient.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, &actual)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return CreateClientObject(ctx, k8sClient, deploy)
		}
		return err
	}

	// the hash value calculated from Deployment instance in k8s may will never equal to the hash value from
	// starrocks cluster. Because Deployment instance may be updated by k8s controller manager.
	// Every time you update the Deployment instance, a new reconcile will be triggered.
	var expectHash, actualHash string
	expectHash = hash.HashObject(deploy)
	if _, ok := actual.Annotations[srapi.ComponentResourceHash]; ok {
		actualHash = actual.Annotations[srapi.ComponentResourceHash]
	} else {
		actualHash = hash.HashObject(actual)
	}

	if expectHash == actualHash {
		logger.Info("expectHash == actualHash, no need to update deployment resource")
		return nil
	}

	deploy.ResourceVersion = actual.ResourceVersion
	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations[srapi.ComponentResourceHash] = expectHash
	return k8sClient.Patch(ctx, deploy, client.Merge)
}

func ApplyConfigMap(ctx context.Context, k8sClient client.Client, configmap *corev1.ConfigMap) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update configmap", "name", configmap.Name)

	var actual corev1.ConfigMap
	err := k8sClient.Get(ctx, types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, &actual)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return CreateClientObject(ctx, k8sClient, configmap)
		}
		return err
	}

	equal := func(configmap, actual *corev1.ConfigMap) bool {
		if len(configmap.Data) != len(actual.Data) {
			return false
		}
		for k, v := range configmap.Data {
			if actual.Data[k] != v {
				return false
			}
		}
		return true
	}

	// the hash value calculated from ConfigMap instance in k8s may will never equal to the hash value from
	// starrocks cluster. Because ConfigMap instance may be updated by k8s controller manager.
	if !equal(configmap, &actual) {
		return UpdateClientObject(ctx, k8sClient, configmap)
	}
	return nil
}

// ApplyStatefulSet when the object is not exist, create object. if exist and statefulset have been updated, patch the statefulset.
func ApplyStatefulSet(ctx context.Context, k8sClient client.Client, expect *appv1.StatefulSet,
	enableScaleTo1 bool, equal StatefulSetEqual) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update statefulset", "name", expect.Name)

	var actual appv1.StatefulSet
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: expect.Namespace, Name: expect.Name}, &actual)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sClient, expect)
	} else if err != nil {
		return err
	}

	if !enableScaleTo1 {
		if actual.Spec.Replicas != nil && *actual.Spec.Replicas > 1 {
			if expect.Spec.Replicas == nil || *expect.Spec.Replicas == 1 {
				return fmt.Errorf("the replicas of statefulset %s can not be scaled to 1", expect.Name)
			}
		}
	}

	// for compatible version <= v1.5, before v1.5 we use order policy to deploy pods.
	// and use `xx-domain-search` for internal service. we should exclude the interference.
	if actual.Spec.PodManagementPolicy == appv1.OrderedReadyPodManagement {
		expect.Spec.PodManagementPolicy = appv1.OrderedReadyPodManagement
	}
	expect.Spec.ServiceName = actual.Spec.ServiceName

	if equal(expect, &actual) {
		logger.Info("expectHash == actualHash, no need to update statefulset resource")
		return nil
	}

	expect.ResourceVersion = actual.ResourceVersion
	return k8sClient.Patch(ctx, expect, client.Merge)
}

func CreateClientObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	if err := k8sClient.Create(ctx, object); err != nil {
		return err
	}
	return nil
}

func UpdateClientObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	if err := k8sClient.Update(ctx, object); err != nil {
		return err
	}
	return nil
}

// DeleteStatefulset delete statefulset.
func DeleteStatefulset(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("delete statefulset from kubernetes", "name", name)

	var st appv1.StatefulSet
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &st); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &st)
}

// DeleteService delete service.
func DeleteService(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("delete service from kubernetes", "name", name)

	var svc corev1.Service
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &svc); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &svc)
}

// DeleteDeployment delete deployment.
func DeleteDeployment(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("delete deployment from kubernetes", "name", name)

	var deploy appv1.Deployment
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &deploy); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &deploy)
}

// DeleteConfigMap delete configmap.
func DeleteConfigMap(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("delete configmap from kubernetes", "name", name)

	var cm corev1.ConfigMap
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &cm); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &cm)
}

func DeleteAutoscaler(ctx context.Context, k8sClient client.Client, namespace, name string, version srapi.AutoScalerVersion) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("delete autoscaler from kubernetes", "name", name)

	hpaObject := version.CreateEmptyHPA(KUBE_MAJOR_VERSION, KUBE_MINOR_VERSION)
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, hpaObject); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, hpaObject)
}

func PodIsReady(status *corev1.PodStatus) bool {
	if status.ContainerStatuses == nil {
		return false
	}

	for _, cs := range status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}

	return true
}

// GetConfigMap get the configmap name=name, namespace=namespace.
func GetConfigMap(ctx context.Context, k8scient client.Client, namespace, name string) (*corev1.ConfigMap, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("fetch configmap from kubernetes", "name", name)

	var configMap corev1.ConfigMap
	if err := k8scient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &configMap); err != nil {
		return nil, err
	}

	return &configMap, nil
}

var (
	KUBE_MAJOR_VERSION string
	KUBE_MINOR_VERSION string
)

// GetKubernetesVersion get kubernetes version. It should not be executed concurrently.
// The global variable KUBE_MAJOR_VERSION and KUBE_MINOR_VERSION will be set.
func GetKubernetesVersion() error {
	var configPath string
	home := homedir.HomeDir()
	if home != "" {
		configPath = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	// create a discovery.DiscoveryClient object to query the metadata of the API server
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	// query the version information of the API server
	version, err := discoveryClient.ServerVersion()
	if err != nil {
		return err
	}

	KUBE_MAJOR_VERSION = version.Major
	// KUBE_MINOR_VERSION is the minor version of the kubernetes cluster, but in cloud provider, the minor version may
	// be like "28+" from alibaba cloud. So we need to remove the non-digit characters.
	KUBE_MINOR_VERSION = CleanMinorVersion(version.Minor)
	return nil
}

// cleanMinorVersion removes non-digit characters from a Kubernetes minor version string.
func CleanMinorVersion(version string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1 // Drop non-digit characters
	}, version)
}

// GetEnvVarValue returns the value of an environment variable. It handles both Value and ValueFrom cases.
// It assumes that the environment variable exists and is valid.
func GetEnvVarValue(ctx context.Context, k8sClient client.Client, namespace string, envVar corev1.EnvVar) (string, error) {
	if envVar.Value != "" {
		// If Value is not empty, return it directly
		return envVar.Value, nil
	} else if envVar.ValueFrom != nil {
		// If ValueFrom is not nil, handle different sources
		valueFrom := envVar.ValueFrom
		if valueFrom.ConfigMapKeyRef != nil {
			// If ConfigMapKeyRef is not nil, get the value from the configmap's key
			name := valueFrom.ConfigMapKeyRef.Name
			key := valueFrom.ConfigMapKeyRef.Key
			return GetValueFromConfigmap(ctx, k8sClient, namespace, name, key)
		} else if valueFrom.SecretKeyRef != nil {
			// If SecretKeyRef is not nil, get the value from the secret's key
			name := valueFrom.SecretKeyRef.Name
			key := valueFrom.SecretKeyRef.Key
			return GetValueFromSecret(ctx, k8sClient, namespace, name, key)
		}
	}
	return "", fmt.Errorf("invalid environment variable: %v", envVar)
}

// GetValueFromConfigmap returns the runtime value of a key in a configmap.
// It assumes that the configmap and the key exist and are valid.
func GetValueFromConfigmap(ctx context.Context, k8sClient client.Client, namespace string, name string, key string) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("fetch configmap from kubernetes", "name", name, "configmap-key", key)

	var configMap corev1.ConfigMap
	err := k8sClient.Get(ctx,
		types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, &configMap)
	if err != nil {
		return "", err
	}
	value, ok := configMap.Data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", key, name)
	}
	return value, nil
}

// GetValueFromSecret returns the value of a key in a secret.
// It assumes that the secret and the key exist and are valid.
func GetValueFromSecret(ctx context.Context, k8sClient client.Client, namespace string, name string, key string) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("fetch secret from kubernetes", "name", name, "secret-key", key)

	var secret corev1.Secret
	err := k8sClient.Get(ctx,
		types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, &secret)
	if err != nil {
		return "", err
	}
	value, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", key, name)
	}
	return string(value), nil
}

func HasVolume(volumes []corev1.Volume, newVolumeName string) bool {
	for _, v := range volumes {
		if v.Name == newVolumeName {
			return true
		}
	}
	return false
}

func HasMountPath(mounts []corev1.VolumeMount, newMountPath string) bool {
	for _, v := range mounts {
		if v.Name == newMountPath {
			return true
		}
	}
	return false
}

func CheckVolumes(volumes []corev1.Volume, mounts []corev1.VolumeMount) error {
	// check mount path first
	mountPaths := make(map[string]bool)
	for i := range mounts {
		path := mounts[i].MountPath
		if mountPaths[path] {
			return fmt.Errorf("mount path %s is duplicated", path)
		} else {
			mountPaths[path] = true
		}
	}

	volumeNames := make(map[string]bool)
	for i := range volumes {
		name := volumes[i].Name
		if volumeNames[name] {
			return fmt.Errorf("volume name %s is duplicated", name)
		} else {
			volumeNames[name] = true
		}
	}
	return nil
}
