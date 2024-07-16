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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	srobject "github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
)

const (
	_logPath         = "/opt/starrocks/cn/log"
	_logName         = "cn-log"
	_cnConfigPath    = "/etc/starrocks/cn/conf"
	_envCnConfigPath = "CONFIGMAP_MOUNT_PATH"

	_cnConfDirPath = "/opt/starrocks/cn/conf"
	_cnConfigKey   = "cn.conf"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(ctx context.Context, object srobject.StarRocksObject,
	cnSpec *srapi.StarRocksCnSpec, config map[string]interface{}) (*corev1.PodTemplateSpec, error) {
	vols, volumeMounts := pod.MountStorageVolumes(cnSpec)

	if !k8sutils.HasVolume(vols, _logName) && !k8sutils.HasMountPath(volumeMounts, _logPath) {
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _logName, _logPath, "")
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, cnSpec.ConfigMapInfo, _cnConfigPath)
	vols, volumeMounts = pod.MountConfigMaps(cnSpec, vols, volumeMounts, cnSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, cnSpec.Secrets)
	if err := k8sutils.CheckVolumes(vols, volumeMounts); err != nil {
		return nil, err
	}

	feExternalServiceName := service.ExternalServiceName(object.ClusterName, (*srapi.StarRocksFeSpec)(nil))
	envs := pod.Envs(cnSpec, config, feExternalServiceName, object.Namespace, cnSpec.CnEnvVars)
	webServerPort := rutils.GetPort(config, rutils.WEBSERVER_PORT)
	if object.Kind == srobject.StarRocksWarehouseKind {
		url := fmt.Sprintf("http://%v.%v:%v/api/v2/feature", feExternalServiceName, object.Namespace, rutils.GetPort(config, rutils.HTTP_PORT))
		if cc.addEnvForWarehouse || cc.addWarehouseEnv(ctx, url) {
			envs = append(envs, corev1.EnvVar{
				Name: "KUBE_STARROCKS_MULTI_WAREHOUSE",
				// the cn_entrypoint.sh in container will use this env to create warehouse. Because of '-' character
				// is not allowed in Warehouse SQL, so we replace it with '_'.
				Value: strings.ReplaceAll(object.Name, "-", "_"),
			})
		} else {
			return nil, GetFeFeatureInfoError
		}
	}
	cnContainer := corev1.Container{
		Name:            srapi.DEFAULT_CN,
		Image:           cnSpec.Image,
		Command:         pod.ContainerCommand(cnSpec),
		Args:            pod.ContainerArgs(cnSpec),
		Ports:           pod.Ports(cnSpec, config),
		Env:             envs,
		Resources:       cnSpec.ResourceRequirements,
		ImagePullPolicy: cnSpec.GetImagePullPolicy(),
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(cnSpec.GetStartupProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(cnSpec.GetLivenessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(cnSpec.GetReadinessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle(cnSpec.GetLifecycle(), "/opt/starrocks/cn_prestop.sh"),
		SecurityContext: pod.ContainerSecurityContext(cnSpec),
	}

	if cnSpec.ConfigMapInfo.ConfigMapName != "" && cnSpec.ConfigMapInfo.ResolveKey != "" {
		cnContainer.Env = append(cnContainer.Env, corev1.EnvVar{
			Name:  _envCnConfigPath,
			Value: _cnConfigPath,
		})
	}

	podSpec := pod.Spec(cnSpec, cnContainer, vols)
	annotations := pod.Annotations(cnSpec)
	podSpec.SecurityContext = pod.PodSecurityContext(cnSpec)
	metaName := object.Name + "-" + srapi.DEFAULT_CN
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			// Name should not define in here, but it is used to compute the value of srapi.ComponentResourceHash
			Name:        metaName,
			Annotations: annotations,
			Namespace:   object.Namespace,
			Labels:      pod.Labels(object.AliasName, cnSpec),
		},
		Spec: podSpec,
	}, nil
}

// addWarehouseEnv add env to cn pod if FE support multi-warehouse
// call FE /api/v2/feature to make sure FE support multi-warehouse
// the response is like:
//
//	{
//	 "features": [
//	   {
//	     "name": "Feature Name",
//	     "description": "Feature Description",
//	     "link": "https://github.com/starrocksdb/starrocks/issues/new"
//	   }
//	 ],
//	 "version": "feature/add-api-feature-interface",
//	 "status": "OK"
//	}
func (cc *CnController) addWarehouseEnv(ctx context.Context, url string) bool {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("call FE to get features information")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error(err, "failed to create request")
		return false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err, "failed to get features information from FE")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Info("FE return status code is not 200", "statusCode", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "failed to read response body")
		return false
	}

	result := struct {
		Features []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Link        string `json:"link"`
		} `json:"features"`
		Version string `json:"version"`
		Status  string `json:"status"`
	}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.Error(err, "failed to unmarshal response body")
		return false
	}

	for _, feature := range result.Features {
		if feature.Name == "multi-warehouse" {
			logger.Info("FE support multi-warehouse")
			return true
		}
	}
	logger.Info("FE does not support multi-warehouse")
	return false
}
