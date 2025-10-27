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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
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
type StatefulSetEqual func(expect *appsv1.StatefulSet, actual *appsv1.StatefulSet) (string, bool)

const (
	LastAppliedConfigAnnotation = "starrocks.kubernetes.operator/last-applied-configuration"
)

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

func ApplyDeployment(ctx context.Context, k8sClient client.Client, deploy *appsv1.Deployment) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update deployment", "name", deploy.Name)

	var actual appsv1.Deployment
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
func ApplyStatefulSet(ctx context.Context, k8sClient client.Client, expect *appsv1.StatefulSet,
	enableScaleTo1 bool, equal StatefulSetEqual) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create or update statefulset", "name", expect.Name)

	var actual appsv1.StatefulSet
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: expect.Namespace, Name: expect.Name}, &actual)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sClient, expect)
	} else if err != nil {
		return err
	}

	// When user delete the statefulset, we should remove the finalizers.
	if actual.DeletionTimestamp != nil && actual.Finalizers != nil {
		actual.Finalizers = nil
		return k8sClient.Update(ctx, &actual)
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
	if actual.Spec.PodManagementPolicy == appsv1.OrderedReadyPodManagement {
		expect.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
	}
	expect.Spec.ServiceName = actual.Spec.ServiceName

	newHashValue, b := equal(expect, &actual)
	if b {
		logger.Info("expectHash == actualHash, no need to update statefulset resource")
		return nil
	}
	expect.Annotations[srapi.ComponentResourceHash] = newHashValue
	expect.ResourceVersion = actual.ResourceVersion

	// Use Client-Side Three-Way Merge

	// 1. Get the last applied configuration
	// If there is no last applied configuration, we let it be an empty JSON object, which means all fields in the
	// actual statefulset are user-modified, but not operator-modified.
	var lastAppliedBytes []byte
	if lastAppliedConfig, ok := actual.Annotations[LastAppliedConfigAnnotation]; ok {
		lastAppliedBytes = []byte(lastAppliedConfig)
	}

	// 2. Marshal the expected state, and actual state
	expectBytes, err := json.Marshal(expect)
	if err != nil {
		return fmt.Errorf("failed to marshal expected state: %w", err)
	}
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		return fmt.Errorf("failed to marshal actual state: %w", err)
	}

	// 3. calculate the strategic merge patch
	schema, err := strategicpatch.NewPatchMetaFromStruct(&appsv1.StatefulSet{})
	if err != nil {
		return fmt.Errorf("failed to create patch meta: %w", err)
	}
	// if overwrite is true, the fields in expectBytes will overwrite the fields in actualBytes
	patchBytes, err := strategicpatch.CreateThreeWayMergePatch(lastAppliedBytes, expectBytes, actualBytes, schema, true)
	if err != nil {
		return fmt.Errorf("failed to create merge patch: %w", err)
	}
	if string(patchBytes) == "{}" {
		logger.Info("no changes detected, skipping update")
		return nil
	}

	// 4. apply patch. Note: we need to use RawPatch and StrategicMergePatchType here
	if err := k8sClient.Patch(ctx, &actual, client.RawPatch(types.StrategicMergePatchType, patchBytes)); err != nil {
		return fmt.Errorf("failed to patch statefulset: %w", err)
	}

	// 5. update annotation only
	retry.RetryOnConflict(retry.DefaultRetry, func() error {
		k8sClient.Get(ctx, types.NamespacedName{Namespace: expect.Namespace, Name: expect.Name}, &actual)
		if actual.Annotations == nil {
			actual.Annotations = make(map[string]string)
		}
		actual.Annotations[LastAppliedConfigAnnotation] = string(expectBytes)
		return k8sClient.Update(ctx, &actual)
	})

	return nil
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

	var st appsv1.StatefulSet
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

	var deploy appsv1.Deployment
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
		logger := logr.FromContextOrDiscard(ctx)
		logger.Info("error when get HPA object", "error", err)
		// If we mistakenly determine the type of HPA, we will receive an error message similar
		// to "no matches for kind 'HorizontalPodAutoscaler' in version 'autoscaling/v2beta2'".
		// This error cannot be identified through apierrors, and using string comparison to
		// determine if it is this error is not a good approach.
		// Therefore, our temporary solution is to always switch to another version of HPA for deletion.
		wrongVersion := version
		if wrongVersion == srapi.AutoScalerV2Beta2 {
			err = k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace},
				srapi.AutoScalerV2.CreateEmptyHPA(KUBE_MAJOR_VERSION, KUBE_MINOR_VERSION))
		} else { // HPA v2 exists in higher version of k8s
			err = k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace},
				srapi.AutoScalerV2Beta2.CreateEmptyHPA(KUBE_MAJOR_VERSION, KUBE_MINOR_VERSION))
		}
		if apierrors.IsNotFound(err) {
			return nil
		} else if err != nil {
			logger.Error(err, "error again when get HPA object")
			return err
		}
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

// CleanMinorVersion removes non-digit characters from a Kubernetes minor version string.
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

// HasVolume is used to decide whether operator should create a default volume for a component.
func HasVolume(volumes []corev1.Volume, defaultVolumeName string) bool {
	for _, v := range volumes {
		if v.Name == defaultVolumeName {
			return true
		}

		// The defaultVolumeName is like be-data, be-log, fe-meta, fe-log, cn-log.
		// If a user deploy StarRocks by helm chart with multiple volumes, their volume names may be like: be0-data, be1-data...,
		// and their mount paths may be like: /opt/starrocks/be/storage0, /opt/starrocks/be/storage1...
		// Considering this situation, we will only check if they have the same suffix, e.g. -data, -log
		subStrings1 := strings.Split(defaultVolumeName, "-")
		suffixFromDefaultVolumeName := subStrings1[len(subStrings1)-1]
		subStrings2 := strings.Split(v.Name, "-")
		suffixFromVolumeName := subStrings2[len(subStrings2)-1]
		if len(subStrings1) > 1 && len(subStrings2) > 1 && suffixFromDefaultVolumeName == suffixFromVolumeName {
			return true
		}
	}
	return false
}

// HasMountPath is used to decide whether operator should create a default volume for a component.
func HasMountPath(mounts []corev1.VolumeMount, defaultMountPath string) bool {
	for _, v := range mounts {
		// The defaultVolumeName is like be-data, be-log, fe-meta, fe-log, cn-log.
		// If a user deploy StarRocks by helm chart with multiple volumes, their volume names may be like: be0-data, be1-data...,
		// and their mount paths may be like: /opt/starrocks/be/storage0, /opt/starrocks/be/storage1...
		// Considering this situation, we need to check if the defaultMountPath is a prefix of the mount path.
		if strings.Contains(v.MountPath, defaultMountPath) {
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

// GetConfig get the config of component.
// First, It tries to read the config from the ConfigMapInfo, which has the configMap name and key.
// Second, if the ConfigMapInfo is empty, it will try to read the config from the ConfigMaps.
// Last, If the fe ConfigMapInfo is empty and the configMaps is nil, it will return an empty map.
func GetConfig(ctx context.Context, k8sClient client.Client,
	configMapInfo srapi.ConfigMapInfo,
	configMaps []srapi.ConfigMapReference, expectMountPath, expectKey string,
	namespace string) (map[string]interface{}, error) {
	if configMapInfo.ConfigMapName != "" || configMapInfo.ResolveKey != "" {
		if configMapInfo.ConfigMapName == "" || configMapInfo.ResolveKey == "" {
			return make(map[string]interface{}), nil
		}
		configMap, err := GetConfigMap(ctx, k8sClient, namespace, configMapInfo.ConfigMapName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return make(map[string]interface{}), nil
			}
			return nil, err
		}

		res, err := ResolveConfigMap(configMap, configMapInfo.ResolveKey)
		return res, err
	}
	return getConfigFromConfigMaps(ctx, k8sClient, configMaps, expectMountPath, expectKey, namespace)
}

func ResolveConfigMap(configMap *corev1.ConfigMap, key string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	data := configMap.Data
	if _, ok := data[key]; !ok {
		return res, nil
	}
	value := data[key]

	// We use a new viper instance, not the global one, in order to avoid concurrency problems: concurrent map iteration
	// and map write,
	v := viper.New()
	v.SetConfigType("properties")
	if err := v.ReadConfig(bytes.NewBuffer([]byte(value))); err != nil {
		return nil, err
	}
	return v.AllSettings(), nil
}

// getConfigFromConfigMaps try to read the config from the configMaps. The strategy is to match
// the mountPath with expectMountPath.
//   - if subpath is empty, the mount path should equal to expectMountPath. And it will use expectKey as the key.
//   - if subpath is not empty, it should equal to expectKey, and the mount path should be expectMountPath/expectKey.
func getConfigFromConfigMaps(ctx context.Context, k8sClient client.Client,
	configMaps []srapi.ConfigMapReference, expectMountPath, expectKey string,
	namespace string) (map[string]interface{}, error) {
	configMapName := ""
	for i := range configMaps {
		subPath := configMaps[i].SubPath
		if subPath == "" {
			if configMaps[i].MountPath == expectMountPath {
				configMapName = configMaps[i].Name
				// don't break here, we need to use the ConfigMapReference with the subPath first.
			}
		} else {
			if configMaps[i].MountPath == filepath.Join(expectMountPath, expectKey) && expectKey == subPath {
				configMapName = configMaps[i].Name
				break
			}
		}
	}
	if configMapName == "" {
		return make(map[string]interface{}), nil
	}

	configMap, err := GetConfigMap(ctx, k8sClient, namespace, configMapName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	res, err := ResolveConfigMap(configMap, expectKey)
	return res, err
}

// BuildNodeSelectorPatch builds a JSON patch that sets keys to new values
// and sets keys that exist in actual but missing in expect to null.
func BuildNodeSelectorPatch(expectSel, actualSel map[string]string) ([]byte, error) {
	if (expectSel == nil && actualSel == nil) || reflect.DeepEqual(expectSel, actualSel) {
		return nil, nil
	}

	// map that will be marshaled: missing keys -> nil (json null)
	nodeMap := map[string]interface{}{}

	// set expected keys to their values
	for k, v := range expectSel {
		nodeMap[k] = v
	}

	// for keys present in actual but not in expect, set to nil -> will become null in JSON
	for k := range actualSel {
		if _, ok := expectSel[k]; !ok {
			nodeMap[k] = nil
		}
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"nodeSelector": nodeMap,
				},
			},
		},
	}

	b, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	return b, nil
}
