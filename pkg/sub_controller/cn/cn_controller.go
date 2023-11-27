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

package cn

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/constant"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/statefulset"
	subc "github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/sub_controller/fe"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/autoscaling/v1"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CnController struct {
	k8sClient client.Client
}

func New(k8sClient client.Client) *CnController {
	return &CnController{
		k8sClient: k8sClient,
	}
}

func (cc *CnController) GetControllerName() string {
	return "cnController"
}

var SpecMissingError = errors.New("spec.template or spec.starRocksCluster is missing")
var StarRocksClusterMissingError = errors.New("custom resource StarRocksCluster is missing")
var FeNotReadyError = errors.New("component fe is not ready")
var StarRocksClusterRunModeError = errors.New("StarRocks Cluster should run in shared_data mode")
var GetFeFeatureInfoError = errors.New("failed to invoke FE /api/v2/feature or FE does not support multi-warehouse feature")

func (cc *CnController) SyncWarehouse(ctx context.Context, warehouse *srapi.StarRocksWarehouse) error {
	template := warehouse.Spec.Template
	if warehouse.Spec.StarRocksCluster == "" || template == nil {
		return SpecMissingError
	}

	klog.Infof("CnController get StarRocksCluster %s/%s to sync warehouse %s/%s",
		warehouse.Namespace, warehouse.Spec.StarRocksCluster, warehouse.Namespace, warehouse.Name)
	_, err := cc.getStarRocksCluster(warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return StarRocksClusterMissingError
		}
		return err
	}

	feconfig, err := cc.getFeConfig(ctx, warehouse.Namespace, warehouse.Spec.StarRocksCluster)
	if err != nil {
		return err
	}
	if val := feconfig["run_mode"]; val == nil || !strings.Contains(val.(string), "shared_data") {
		return StarRocksClusterRunModeError
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, warehouse.Namespace, warehouse.Spec.StarRocksCluster) {
		return FeNotReadyError
	}
	return cc.SyncCnSpec(ctx, object.NewFromWarehouse(warehouse), template.ToCnSpec())
}

func (cc *CnController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksCnSpec == nil {
		if err := cc.ClearResources(ctx, src); err != nil {
			klog.Errorf("cnController sync namespace=%s, name=%s, err=%s", src.Namespace, src.Name, err.Error())
		}
		return nil
	}

	if !fe.CheckFEReady(ctx, cc.k8sClient, src.Namespace, src.Name) {
		return nil
	}

	return cc.SyncCnSpec(ctx, object.NewFromCluster(src), src.Spec.StarRocksCnSpec)
}

func (cc *CnController) SyncCnSpec(ctx context.Context, object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec) error {
	if err := cc.mutating(cnSpec); err != nil {
		return err
	}

	if err := cc.validating(cnSpec); err != nil {
		return err
	}

	klog.Infof("CnController get the query port from fe ConfigMap to resolve port, namespace=%s, name=%s",
		object.Namespace, object.Name)
	config, err := cc.GetConfig(ctx, &cnSpec.ConfigMapInfo, object.Namespace)
	if err != nil {
		return err
	}
	feconfig, err := cc.getFeConfig(ctx, object.Namespace, object.ClusterName)
	if err != nil {
		return err
	}
	config[rutils.QUERY_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.QUERY_PORT)), 10)
	config[rutils.HTTP_PORT] = strconv.FormatInt(int64(rutils.GetPort(feconfig, rutils.HTTP_PORT)), 10)

	klog.Infof("CnController build and apply statefulset for cn, namespace=%s, name=%s", object.Namespace, object.AliasName)
	podTemplateSpec, err := cc.buildPodTemplate(object, cnSpec, config)
	if err != nil {
		return err
	}
	sts := statefulset.MakeStatefulset(object, cnSpec, *podTemplateSpec)
	if err = cc.applyStatefulset(ctx, &sts); err != nil {
		return err
	}

	klog.Infof("CnController build external and internal service for cn, namespace=%s, name=%s",
		object.Namespace, object.AliasName)
	externalsvc := rutils.BuildExternalService(object, cnSpec, config,
		load.Selector(object.AliasName, cnSpec), load.Labels(object.AliasName, cnSpec))
	searchServiceName := service.SearchServiceName(object.AliasName, cnSpec)
	internalService := service.MakeSearchService(searchServiceName, &externalsvc, []corev1.ServicePort{
		{
			Name:       "heartbeat",
			Port:       rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT),
			TargetPort: intstr.FromInt(int(rutils.GetPort(config, rutils.HEARTBEAT_SERVICE_PORT))),
		},
	})

	klog.Infof("CnController apply external and internal service for cn, namespace=%s, name=%s",
		object.Namespace, object.AliasName)
	if err := k8sutils.ApplyService(ctx, cc.k8sClient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		return err
	}
	if err := k8sutils.ApplyService(ctx, cc.k8sClient, internalService, func(new *corev1.Service, esvc *corev1.Service) bool {
		// for compatible v1.5, we use `cn-domain-search` for internal communicating.
		internalService.Name = sts.Spec.ServiceName
		return rutils.ServiceDeepEqual(new, esvc)
	}); err != nil {
		return err
	}

	klog.Infof("CnController build and apply HPA for cn, namespace=%s, name=%s", object.Namespace, object.AliasName)
	if cnSpec.AutoScalingPolicy != nil {
		return cc.deployAutoScaler(ctx, object, cnSpec, *cnSpec.AutoScalingPolicy, &sts)
	}
	return nil
}

func (cc *CnController) applyStatefulset(ctx context.Context, st *appv1.StatefulSet) error {
	// create or update the status. create statefulset return, must ensure the
	var est appv1.StatefulSet
	if err := cc.k8sClient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est); apierrors.IsNotFound(err) {
		return k8sutils.CreateClientObject(ctx, cc.k8sClient, st)
	} else if err != nil {
		klog.Errorf("CnController Sync create statefulset name=%s, namespace=%s error=%s", st.Name, st.Namespace, err.Error())
		return err
	}
	// if the spec is changed, update the status of cn on src.
	var excludeReplica bool
	// if replicas =0 and not the first time, exclude the hash for autoscaler
	if st.Spec.Replicas == nil {
		if _, ok := est.Annotations[srapi.ComponentReplicasEmpty]; !ok {
			excludeReplica = true
		}
	}

	// for compatible version <= v1.5, use `cn-domain-search` for internal service. we should exclude the interference.
	st.Spec.ServiceName = est.Spec.ServiceName

	if !rutils.StatefulSetDeepEqual(st, &est, excludeReplica) {
		// if the replicas not zero, represent user have cancel autoscaler.
		if st.Spec.Replicas != nil {
			if _, ok := est.Annotations[srapi.ComponentReplicasEmpty]; ok {
				rutils.MergeStatefulSets(st, est) // ResourceVersion will be set
				delete(st.Annotations, srapi.ComponentReplicasEmpty)
				return k8sutils.UpdateClientObject(ctx, cc.k8sClient, st)
			}
		}
		st.ResourceVersion = est.ResourceVersion
		return k8sutils.UpdateClientObject(ctx, cc.k8sClient, st)
	}

	return nil
}

// UpdateWarehouseStatus updates the status of StarRocksWarehouse.
func (cc *CnController) UpdateWarehouseStatus(warehouse *srapi.StarRocksWarehouse) error {
	template := warehouse.Spec.Template
	if template == nil {
		warehouse.Status.WarehouseComponentStatus = nil
		return nil
	}

	status := warehouse.Status.WarehouseComponentStatus
	status.Phase = srapi.ComponentReconciling
	return cc.UpdateStatus(object.NewFromWarehouse(warehouse), template.ToCnSpec(), status)
}

// UpdateClusterStatus update the status of StarRocksCluster.
func (cc *CnController) UpdateClusterStatus(src *srapi.StarRocksCluster) error {
	cnSpec := src.Spec.StarRocksCnSpec
	if cnSpec == nil {
		src.Status.StarRocksCnStatus = nil
		return nil
	}

	if src.Status.StarRocksCnStatus == nil {
		src.Status.StarRocksCnStatus = &srapi.StarRocksCnStatus{
			StarRocksComponentStatus: srapi.StarRocksComponentStatus{
				Phase: srapi.ComponentReconciling,
			},
		}
	}
	cs := src.Status.StarRocksCnStatus
	cs.Phase = srapi.ComponentReconciling

	return cc.UpdateStatus(object.NewFromCluster(src), cnSpec, cs)
}

func (cc *CnController) UpdateStatus(object object.StarRocksObject,
	cnSpec *srapi.StarRocksCnSpec, cnStatus *srapi.StarRocksCnStatus) error {
	var st appv1.StatefulSet
	statefulSetName := load.Name(object.AliasName, cnSpec)
	namespacedName := types.NamespacedName{Namespace: object.Namespace, Name: statefulSetName}
	if err := cc.k8sClient.Get(context.Background(), namespacedName, &st); apierrors.IsNotFound(err) {
		klog.Infof("CnController UpdateStatus the statefulset name=%s is not found.\n", statefulSetName)
		return nil
	}

	if cnSpec.AutoScalingPolicy != nil {
		cnStatus.HorizontalScaler.Name = cc.generateAutoScalerName(object.AliasName, cnSpec)
		cnStatus.HorizontalScaler.Version = cnSpec.AutoScalingPolicy.Version.Complete(k8sutils.KUBE_MAJOR_VERSION,
			k8sutils.KUBE_MINOR_VERSION)
	} else {
		cnStatus.HorizontalScaler = srapi.HorizontalScaler{}
	}

	cnStatus.ServiceName = service.ExternalServiceName(object.AliasName, cnSpec)
	cnStatus.ResourceNames = rutils.MergeSlices(cnStatus.ResourceNames, []string{statefulSetName})

	if err := subc.UpdateStatus(&cnStatus.StarRocksComponentStatus, cc.k8sClient,
		object.Namespace, load.Name(object.AliasName, cnSpec), pod.Labels(object.AliasName, cnSpec), subc.StatefulSetLoadType); err != nil {
		return err
	}

	return nil
}

func (cc *CnController) ClearWarehouse(ctx context.Context, namespace string, name string) error {
	executor, err := NewSQLExecutor(cc.k8sClient, namespace, object.GetAliasName(name))
	if err != nil {
		klog.Infof("CnController ClearWarehouse NewSQLExecutor error=%s", err.Error())
		return err
	}

	err = executor.Execute(ctx, fmt.Sprintf("DROP WAREHOUSE %s", strings.ReplaceAll(name, "-", "_")))
	if err != nil {
		klog.Infof("CnController failed DROP WAREHOUSE <%v>, error=%s", strings.ReplaceAll(name, "-", "_"), err.Error())
		// we do not return error here, because we want to delete the statefulset anyway.
	}

	// Remove the finalizer from cn statefulset
	var sts appv1.StatefulSet
	if err = cc.k8sClient.Get(context.Background(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      load.Name(object.GetAliasName(name), (*srapi.StarRocksCnSpec)(nil)),
		},
		&sts); err != nil {
		return err
	}
	sts.Finalizers = nil
	if err = k8sutils.UpdateClientObject(context.Background(), cc.k8sClient, &sts); err != nil {
		return err
	}

	// return err
	return err
}

// Deploy autoscaler
func (cc *CnController) deployAutoScaler(ctx context.Context, object object.StarRocksObject, cnSpec *srapi.StarRocksCnSpec,
	policy srapi.AutoScalingPolicy, target *appv1.StatefulSet) error {
	labels := rutils.Labels{}
	labels.AddLabel(target.Labels)
	labels.Add(srapi.ComponentLabelKey, "autoscaler")
	autoscalerParams := &rutils.PodAutoscalerParams{
		Namespace:       target.Namespace,
		Name:            cc.generateAutoScalerName(object.AliasName, cnSpec),
		Labels:          labels,
		AutoscalerType:  cnSpec.AutoScalingPolicy.Version, // cnSpec.AutoScalingPolicy can not be nil
		TargetName:      target.Name,
		OwnerReferences: target.OwnerReferences,
		ScalerPolicy:    &policy,
	}

	autoScaler := rutils.BuildHorizontalPodAutoscaler(autoscalerParams)
	autoScaler.SetAnnotations(make(map[string]string))
	var clientObject client.Object
	t := autoscalerParams.AutoscalerType.Complete(k8sutils.KUBE_MAJOR_VERSION, k8sutils.KUBE_MINOR_VERSION)
	switch t {
	case srapi.AutoScalerV1:
		clientObject = &v1.HorizontalPodAutoscaler{}
	case srapi.AutoScalerV2:
		clientObject = &v2.HorizontalPodAutoscaler{}
	case srapi.AutoScalerV2Beta2:
		clientObject = &v2beta2.HorizontalPodAutoscaler{}
	}
	if err := cc.k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: autoscalerParams.Namespace,
			Name:      autoscalerParams.Name,
		},
		clientObject,
	); err != nil {
		if apierrors.IsNotFound(err) {
			return cc.k8sClient.Create(ctx, autoScaler)
		}
		return err
	}

	var expectHash, actualHash string
	expectHash = hash.HashObject(autoScaler)
	if v, ok := clientObject.GetAnnotations()[srapi.ComponentResourceHash]; ok {
		actualHash = v
	} else {
		actualHash = hash.HashObject(clientObject)
	}

	if expectHash == actualHash {
		klog.Infof("cnController deployAutoscaler not need update, namespace=%s,name=%s,version=%s",
			autoScaler.GetNamespace(), autoScaler.GetName(), t)
		return nil
	}
	autoScaler.GetAnnotations()[srapi.ComponentResourceHash] = expectHash
	return cc.k8sClient.Update(ctx, autoScaler)
}

// deleteAutoScaler delete the autoscaler.
func (cc *CnController) deleteAutoScaler(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Status.StarRocksCnStatus == nil {
		return nil
	}

	if src.Status.StarRocksCnStatus.HorizontalScaler.Name == "" {
		klog.V(constant.LOG_LEVEL).Infof("cnController not need delete the autoScaler, namespace=%s, src name=%s.", src.Namespace, src.Name)
		return nil
	}

	autoScalerName := src.Status.StarRocksCnStatus.HorizontalScaler.Name
	version := src.Status.StarRocksCnStatus.HorizontalScaler.Version
	if err := k8sutils.DeleteAutoscaler(ctx, cc.k8sClient, src.Namespace, autoScalerName, version); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController sync deploy or delete failed, namespace=%s, autoscaler name=%s, autoscaler version=%s",
			src.GetNamespace(), autoScalerName, version)
		return err
	}

	src.Status.StarRocksCnStatus.HorizontalScaler = srapi.HorizontalScaler{}
	return nil
}

// ClearResources clear the deployed resource about cn. statefulset, services, hpa.
func (cc *CnController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksCnSpec != nil {
		return nil
	}

	cnSpec := src.Spec.StarRocksCnSpec
	statefulSetName := load.Name(src.Name, cnSpec)
	err := k8sutils.DeleteStatefulset(ctx, cc.k8sClient, src.Namespace, statefulSetName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete statefulset failed, namespace=%s,name=%s, error=%s.",
			src.Namespace, statefulSetName, err.Error())
		return err
	}

	searchServiceName := service.SearchServiceName(src.Name, cnSpec)
	err = k8sutils.DeleteService(ctx, cc.k8sClient, src.Namespace, searchServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete search service, namespace=%s,name=%s,error=%s.",
			src.Namespace, searchServiceName, err.Error())
		return err
	}
	externalServiceName := service.ExternalServiceName(src.Name, cnSpec)
	err = k8sutils.DeleteService(ctx, cc.k8sClient, src.Namespace, externalServiceName)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("cnController ClearResources delete external service, namespace=%s, name=%s,error=%s.",
			src.Namespace, externalServiceName, err.Error())
		return err
	}

	if err := cc.deleteAutoScaler(ctx, src); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (cc *CnController) GetConfig(ctx context.Context,
	configMapInfo *srapi.ConfigMapInfo, namespace string) (map[string]interface{}, error) {
	klog.Infof("CnController get configMap from %s/%s", namespace, configMapInfo.ConfigMapName)
	configMap, err := k8sutils.GetConfigMap(ctx, cc.k8sClient, namespace, configMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.Infof("ConfigMap for conf is missing, namespace=%s", namespace)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}

	res, err := rutils.ResolveConfigMap(configMap, configMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) getFeConfig(ctx context.Context,
	clusterNamespace string, clusterName string) (map[string]interface{}, error) {
	src, err := cc.getStarRocksCluster(clusterNamespace, clusterName)
	if err != nil {
		return nil, err
	}
	feconfigMapInfo := &src.Spec.StarRocksFeSpec.ConfigMapInfo

	feconfigMap, err := k8sutils.GetConfigMap(ctx, cc.k8sClient, clusterNamespace, feconfigMapInfo.ConfigMapName)
	if err != nil && apierrors.IsNotFound(err) {
		klog.V(constant.LOG_LEVEL).Info("CnController getFeConfig fe config is not exist namespace ", clusterNamespace,
			" configmapName ", feconfigMapInfo.ConfigMapName)
		return make(map[string]interface{}), nil
	} else if err != nil {
		return make(map[string]interface{}), err
	}
	res, err := rutils.ResolveConfigMap(feconfigMap, feconfigMapInfo.ResolveKey)
	return res, err
}

func (cc *CnController) mutating(cnSpec *srapi.StarRocksCnSpec) error {
	// Mutating because of the autoscaling policy.
	// When the HPA policy with a fixed replica count is set: every time the starrockscluster CR is
	// applied, the replica count of the StatefulSet object in K8S will be reset to the value
	// specified by the 'Replicas' field, erasing the value previously set by HPA.
	policy := cnSpec.AutoScalingPolicy
	if policy != nil {
		cnSpec.Replicas = nil
	}
	return nil
}

func (cc *CnController) validating(cnSpec *srapi.StarRocksCnSpec) error {
	// validating the auto scaling policy
	policy := cnSpec.AutoScalingPolicy
	if policy != nil {
		minReplicas := int32(1) // default value
		if policy.MinReplicas != nil {
			minReplicas = *policy.MinReplicas
			if minReplicas < 1 {
				return fmt.Errorf("the min replicas must not be smaller than 1")
			}
		}

		maxReplicas := policy.MaxReplicas
		if maxReplicas < minReplicas {
			return fmt.Errorf("the MaxReplicas must not be smaller than MinReplicas")
		}
	}
	return nil
}

// getStarRocksCluster get the StarRocksCluster object by namespace and name.
func (cc *CnController) getStarRocksCluster(namespace, name string) (*srapi.StarRocksCluster, error) {
	src := &srapi.StarRocksCluster{}
	err := cc.k8sClient.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, src)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func (cc *CnController) generateAutoScalerName(srcName string, cnSpec srapi.SpecInterface) string {
	return load.Name(srcName, cnSpec) + "-autoscaler"
}
