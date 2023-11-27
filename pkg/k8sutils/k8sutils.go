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

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/constant"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceEqual judges two services equal or not in some fields. developer can custom the function.
type ServiceEqual func(svc1 *corev1.Service, svc2 *corev1.Service) bool

// StatefulSetEqual judges two statefulset equal or not in some fields. developer can custom the function.
type StatefulSetEqual func(st1 *appv1.StatefulSet, st2 *appv1.StatefulSet) bool

func ApplyService(ctx context.Context, k8sclient client.Client, svc *corev1.Service, equal ServiceEqual) error {
	// As stated in the RetryOnConflict's documentation, the returned error shouldn't be wrapped.
	var esvc corev1.Service
	err := k8sclient.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, &esvc)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sclient, svc)
	} else if err != nil {
		return err
	}

	if equal(svc, &esvc) {
		klog.Info("Apply service Name, Ports, Selector, ServiceType, Labels have not change ", "namespace ",
			svc.Namespace, " name ", svc.Name)
		return nil
	}

	svc.ResourceVersion = esvc.ResourceVersion
	return UpdateClientObject(ctx, k8sclient, svc)
}

func ApplyDeployment(ctx context.Context, k8sClient client.Client, deploy *appv1.Deployment) error {
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
		return nil
	}

	deploy.ResourceVersion = actual.ResourceVersion
	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations[srapi.ComponentResourceHash] = expectHash
	return UpdateClientObject(ctx, k8sClient, deploy)
}

func ApplyConfigMap(ctx context.Context, k8sClient client.Client, configmap *corev1.ConfigMap) error {
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
func ApplyStatefulSet(ctx context.Context, k8sClient client.Client, st *appv1.StatefulSet, equal StatefulSetEqual) error {
	var est appv1.StatefulSet
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est)
	if err != nil && apierrors.IsNotFound(err) {
		return CreateClientObject(ctx, k8sClient, st)
	} else if err != nil {
		return err
	}
	// for compatible version <= v1.5, before v1.5 we use order policy to deploy pods.
	// and use `xx-domain-search` for internal service. we should exclude the interference.
	if est.Spec.PodManagementPolicy == appv1.OrderedReadyPodManagement {
		st.Spec.PodManagementPolicy = appv1.OrderedReadyPodManagement
	}
	st.Spec.ServiceName = est.Spec.ServiceName

	// if have restart annotation we should exclude it impacts on hash.
	if equal(st, &est) {
		klog.Infof("ApplyStatefulSet Sync exist statefulset name=%s, namespace=%s, equals to new statefulset.", est.Name, est.Namespace)
		return nil
	}

	st.ResourceVersion = est.ResourceVersion
	return UpdateClientObject(ctx, k8sClient, st)
}

func CreateClientObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	klog.Infof("Creating k8s resource namespace=%s, name=%s, kind=%s", object.GetNamespace(), object.GetName(),
		object.GetObjectKind().GroupVersionKind().Kind)
	if err := k8sClient.Create(ctx, object); err != nil {
		return err
	}
	return nil
}

func UpdateClientObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	klog.Info("Updating resource service ", "namespace ", object.GetNamespace(), " name ", object.GetName(),
		" kind ", object.GetObjectKind())
	if err := k8sClient.Update(ctx, object); err != nil {
		return err
	}
	return nil
}

// PatchClientObject patch object when the object exist. if not return error.
func PatchClientObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	klog.V(constant.LOG_LEVEL).Infof("patch resource namespace=%s,name=%s,kind=%s.",
		object.GetNamespace(), object.GetName(), object.GetObjectKind())
	if err := k8sClient.Patch(ctx, object, client.Merge); err != nil {
		return err
	}

	return nil
}

// DeleteStatefulset delete statefulset.
func DeleteStatefulset(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	var st appv1.StatefulSet
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &st); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &st)
}

// DeleteService delete service.
func DeleteService(ctx context.Context, k8sclient client.Client, namespace, name string) error {
	var svc corev1.Service
	if err := k8sclient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &svc); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sclient.Delete(ctx, &svc)
}

// DeleteDeployment delete deployment.
func DeleteDeployment(ctx context.Context, k8sclient client.Client, namespace, name string) error {
	var deploy appv1.Deployment
	if err := k8sclient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &deploy); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sclient.Delete(ctx, &deploy)
}

// DeleteConfigMap delete configmap.
func DeleteConfigMap(ctx context.Context, k8sClient client.Client, namespace, name string) error {
	var cm corev1.ConfigMap
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &cm); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, &cm)
}

// DeleteAutoscaler as version type delete response autoscaler.
func DeleteAutoscaler(ctx context.Context, k8sClient client.Client,
	namespace, name string, autoscalerVersion srapi.AutoScalerVersion) error {
	var autoscaler client.Object
	switch autoscalerVersion {
	case srapi.AutoScalerV1:
		autoscaler = &v1.HorizontalPodAutoscaler{}
	case srapi.AutoScalerV2:
		autoscaler = &v2.HorizontalPodAutoscaler{}
	case srapi.AutoScalerV2Beta2:
		autoscaler = &v2beta2.HorizontalPodAutoscaler{}
	default:
		return fmt.Errorf("the autoscaler type %s is not supported", autoscalerVersion)
	}

	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, autoscaler); apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return k8sClient.Delete(ctx, autoscaler)
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
	KUBE_MINOR_VERSION = version.Minor
	return nil
}

// GetEnvVarValue returns the value of an environment variable. It handles both Value and ValueFrom cases.
// It assumes that the environment variable exists and is valid.
func GetEnvVarValue(k8sClient client.Client, namespace string, envVar corev1.EnvVar) (string, error) {
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
			return getValueFromConfigmap(k8sClient, namespace, name, key)
		} else if valueFrom.SecretKeyRef != nil {
			// If SecretKeyRef is not nil, get the value from the secret's key
			name := valueFrom.SecretKeyRef.Name
			key := valueFrom.SecretKeyRef.Key
			return getValueFromSecret(k8sClient, namespace, name, key)
		}
	}
	return "", fmt.Errorf("invalid environment variable: %v", envVar)
}

// getValueFromConfigmap returns the runtime value of a key in a configmap.
// It assumes that the configmap and the key exist and are valid.
func getValueFromConfigmap(k8sClient client.Client, namespace string, name string, key string) (string, error) {
	var configMap corev1.ConfigMap
	err := k8sClient.Get(context.Background(),
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

// getValueFromSecret returns the value of a key in a secret.
// It assumes that the secret and the key exist and are valid.
func getValueFromSecret(k8sClient client.Client, namespace string, name string, key string) (string, error) {
	var secret corev1.Secret
	err := k8sClient.Get(context.Background(),
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
