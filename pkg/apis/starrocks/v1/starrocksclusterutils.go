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

package v1

// GetExternalServiceName generate the name of service that access the fe.
func GetExternalServiceName(clusterName string, spec SpecInterface) string {
	if spec.GetServiceName() != "" {
		return spec.GetServiceName()
	}
	// for compatible version <=1.3
	switch spec.(type) {
	case *StarRocksFeSpec:
		return clusterName + "-" + DEFAULT_FE + "-service"
	case *StarRocksBeSpec:
		return clusterName + "-" + DEFAULT_BE + "-service"
	case *StarRocksCnSpec:
		return clusterName + "-" + DEFAULT_CN + "-service"
	}
	return ""
}
