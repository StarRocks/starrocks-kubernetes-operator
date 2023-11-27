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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	srobject "github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/object"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/pod"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	_logPath         = "/opt/starrocks/cn/log"
	_logName         = "cn-log"
	_cnConfigPath    = "/etc/starrocks/cn/conf"
	_envCnConfigPath = "CONFIGMAP_MOUNT_PATH"
)

// buildPodTemplate construct the podTemplate for deploy cn.
func (cc *CnController) buildPodTemplate(object srobject.StarRocksObject,
	cnSpec *srapi.StarRocksCnSpec, config map[string]interface{}) (*corev1.PodTemplateSpec, error) {
	vols, volumeMounts, vexist := pod.MountStorageVolumes(cnSpec)
	// add default volume about log
	if _, ok := vexist[_logPath]; !ok {
		vols, volumeMounts = pod.MountEmptyDirVolume(vols, volumeMounts, _logName, _logPath, "")
	}

	// mount configmap, secrets to pod if needed
	vols, volumeMounts = pod.MountConfigMapInfo(vols, volumeMounts, cnSpec.ConfigMapInfo, _cnConfigPath)
	vols, volumeMounts = pod.MountConfigMaps(vols, volumeMounts, cnSpec.ConfigMaps)
	vols, volumeMounts = pod.MountSecrets(vols, volumeMounts, cnSpec.Secrets)

	feExternalServiceName := service.ExternalServiceName(object.ClusterName, (*srapi.StarRocksFeSpec)(nil))
	envs := pod.Envs(cnSpec, config, feExternalServiceName, object.Namespace, cnSpec.CnEnvVars)
	webServerPort := rutils.GetPort(config, rutils.WEBSERVER_PORT)
	if object.Kind == srobject.StarRocksWarehouseKind {
		if cc.addWarehouseEnv(feExternalServiceName,
			strconv.FormatInt(int64(rutils.GetPort(config, rutils.HTTP_PORT)), 10)) {
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
		Command:         []string{"/opt/starrocks/cn_entrypoint.sh"},
		Args:            []string{"$(FE_SERVICE_NAME)"},
		Ports:           pod.Ports(cnSpec, config),
		Env:             envs,
		Resources:       cnSpec.ResourceRequirements,
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    volumeMounts,
		StartupProbe:    pod.StartupProbe(cnSpec.GetStartupProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		LivenessProbe:   pod.LivenessProbe(cnSpec.GetLivenessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		ReadinessProbe:  pod.ReadinessProbe(cnSpec.GetReadinessProbeFailureSeconds(), webServerPort, pod.HEALTH_API_PATH),
		Lifecycle:       pod.LifeCycle("/opt/starrocks/cn_prestop.sh"),
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
func (cc *CnController) addWarehouseEnv(feExternalServiceName string, feHTTPPort string) bool {
	klog.Infof("call FE to get features information")
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v2/feature", feExternalServiceName, feHTTPPort))
	if err != nil {
		klog.Errorf("failed to get features information from FE, err: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		klog.Infof("FE return status code: %d", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("failed to read response body, err: %v", err)
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
		klog.Errorf("failed to unmarshal response body, err: %v", err)
		return false
	}

	for _, feature := range result.Features {
		if feature.Name == "multi-warehouse" {
			klog.Infof("FE support multi-warehouse")
			return true
		}
	}
	klog.Infof("FE does not support multi-warehouse")
	return false
}
