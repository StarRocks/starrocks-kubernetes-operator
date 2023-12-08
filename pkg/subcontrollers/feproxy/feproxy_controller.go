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

package feproxy

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/log"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/deployment"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

type FeProxyController struct {
	k8sClient client.Client
}

var _ subcontrollers.ClusterSubController = &FeProxyController{}

// New construct a FeController.
func New(k8sClient client.Client) *FeProxyController {
	return &FeProxyController{
		k8sClient: k8sClient,
	}
}

func (controller *FeProxyController) GetControllerName() string {
	return "feProxyController"
}

func (controller *FeProxyController) SyncCluster(ctx context.Context, src *srapi.StarRocksCluster) error {
	feProxySpec := src.Spec.StarRocksFeProxySpec
	logger := logr.FromContextOrDiscard(ctx).WithName(controller.GetControllerName()).WithValues(log.ActionKey, log.ActionSyncCluster)
	ctx = logr.NewContext(ctx, logger)

	if feProxySpec == nil {
		logger.Info("src.Spec.StarRocksFeProxySpec == nil, clear fe proxy resource")
		if err := controller.ClearResources(ctx, src); err != nil {
			logger.Error(err, "clear fe proxy resource failed", "StarRocksCluster", src)
			return err
		}
		return nil
	}

	if !fe.CheckFEReady(ctx, controller.k8sClient, src.Namespace, src.Name) {
		logger.Info("FE is not ready, stop sync fe proxy")
		return nil
	}

	err := controller.SyncConfigMap(ctx, src)
	if err != nil {
		logger.Error(err, "sync fe proxy configmap failed", "StarRocksCluster", src)
		return err
	}

	podTemplate := controller.buildPodTemplate(src)
	expectDeployment := deployment.MakeDeployment(src, feProxySpec, podTemplate)
	err = k8sutils.ApplyDeployment(ctx, controller.k8sClient, expectDeployment)
	if err != nil {
		logger.Error(err, "sync fe proxy deployment failed", "StarRocksCluster", src)
		return err
	}

	object := object.NewFromCluster(src)
	externalsvc := rutils.BuildExternalService(object, feProxySpec, nil,
		load.Selector(src.Name, feProxySpec), load.Labels(src.Name, feProxySpec))
	if err := k8sutils.ApplyService(ctx, controller.k8sClient, &externalsvc, rutils.ServiceDeepEqual); err != nil {
		return err
	}

	return nil
}

// UpdateClusterStatus update the all resource status about fe.
func (controller *FeProxyController) UpdateClusterStatus(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(controller.GetControllerName()).
		WithValues(log.ActionKey, log.ActionUpdateClusterStatus)
	ctx = logr.NewContext(ctx, logger)

	feProxySpec := src.Spec.StarRocksFeProxySpec
	if feProxySpec == nil {
		src.Status.StarRocksFeProxyStatus = nil
		return nil
	}

	status := &srapi.StarRocksFeProxyStatus{
		StarRocksComponentStatus: srapi.StarRocksComponentStatus{
			Phase: srapi.ComponentReconciling,
		},
	}
	if src.Status.StarRocksFeProxyStatus != nil {
		status = src.Status.StarRocksFeProxyStatus.DeepCopy()
	}
	src.Status.StarRocksFeProxyStatus = status

	// TODO(yandongxiao): delete it
	var actual appsv1.Deployment
	deploymentName := load.Name(src.Name, feProxySpec)
	err := controller.k8sClient.Get(ctx, types.NamespacedName{
		Namespace: src.Namespace,
		Name:      deploymentName,
	}, &actual)
	if err != nil {
		logger.Error(err, "get fe proxy deployment failed", "StarRocksCluster", src)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	status.ServiceName = service.ExternalServiceName(src.Name, feProxySpec)
	if err := subcontrollers.UpdateStatus(&status.StarRocksComponentStatus, controller.k8sClient,
		src.Namespace, load.Name(src.Name, feProxySpec), pod.Labels(src.Name, feProxySpec), subcontrollers.DeploymentLoadType); err != nil {
		logger.Error(err, "update fe proxy status failed", "StarRocksCluster", src)
		return err
	}

	return nil
}

// ClearResources clear resource about fe.
func (controller *FeProxyController) ClearResources(ctx context.Context, src *srapi.StarRocksCluster) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(controller.GetControllerName()).
		WithValues(log.ActionKey, log.ActionClearResources)
	ctx = logr.NewContext(ctx, logger)

	if src.Spec.StarRocksFeProxySpec != nil {
		return nil
	}

	feProxySpec := src.Spec.StarRocksFeProxySpec
	loadName := load.Name(src.Name, feProxySpec)
	if err := k8sutils.DeleteDeployment(ctx, controller.k8sClient, src.Namespace, loadName); err != nil {
		logger.Error(err, "delete fe proxy deployment failed", "StarRocksCluster", src)
		return err
	}

	externalServiceName := service.ExternalServiceName(src.Name, feProxySpec)
	if err := k8sutils.DeleteService(ctx, controller.k8sClient, src.Namespace, externalServiceName); err != nil {
		logger.Error(err, "delete fe proxy service failed", "StarRocksCluster", src)
		return err
	}

	configMapName := load.Name(src.Name, feProxySpec)
	if err := k8sutils.DeleteConfigMap(ctx, controller.k8sClient, src.Namespace, configMapName); err != nil {
		logger.Error(err, "delete fe proxy configmap failed", "StarRocksCluster", src)
		return err
	}

	return nil
}

func (controller *FeProxyController) buildPodTemplate(src *srapi.StarRocksCluster) corev1.PodTemplateSpec {
	feProxySpec := src.Spec.StarRocksFeProxySpec
	vols, volumeMounts, _ := pod.MountStorageVolumes(feProxySpec)

	vols, volumeMounts = pod.MountConfigMaps(vols, volumeMounts, []srapi.ConfigMapReference{
		{
			Name:      load.Name(src.Name, feProxySpec),
			MountPath: "/etc/nginx",
		},
	})

	var port int32 = 8080
	image := "nginx:1.24.0"
	if feProxySpec.Image != "" && !strings.HasPrefix(feProxySpec.Image, ":") {
		image = feProxySpec.Image
	}
	container := corev1.Container{
		Name:            "nginx",
		Image:           image,
		Ports:           pod.Ports(feProxySpec, nil),
		Resources:       feProxySpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		LivenessProbe:   pod.LivenessProbe(feProxySpec.GetLivenessProbeFailureSeconds(), port, "/nginx/health"),
		ReadinessProbe:  pod.ReadinessProbe(feProxySpec.GetReadinessProbeFailureSeconds(), port, "/nginx/health"),
		SecurityContext: pod.ContainerSecurityContext(feProxySpec),
	}

	// nginx container will run as nginx user, not allowed to change
	var userID int64 = 101
	var groupID int64 = 101
	runAsNonRoot := true
	container.SecurityContext = &corev1.SecurityContext{
		RunAsUser:                &userID,
		RunAsGroup:               &groupID,
		RunAsNonRoot:             &runAsNonRoot,
		AllowPrivilegeEscalation: func() *bool { b := false; return &b }(),
		// nginx will write content to some file specified by client_body_temp_path
		ReadOnlyRootFilesystem: func() *bool { b := false; return &b }(),
	}

	podSpec := pod.Spec(feProxySpec, container, vols)
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: feProxySpec.GetAnnotations(),
			Labels:      pod.Labels(src.Name, feProxySpec),
		},
		Spec: podSpec,
	}
}
