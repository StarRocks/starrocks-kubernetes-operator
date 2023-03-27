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

package v1alpha1

//GetFeExternalServiceName generate the name of service that access the fe.
func GetFeExternalServiceName(src *StarRocksCluster) string {
	if src.Spec.StarRocksFeSpec != nil && src.Spec.StarRocksFeSpec.Service != nil && src.Spec.StarRocksFeSpec.Service.Name != "" {
		return src.Spec.StarRocksFeSpec.Service.Name
	}

	//for compatible version <=1.3
	return src.Name + "-" + DEFAULT_FE + "-service"
}

//GetCnExternalServiceName generate the name of service that access the cn
func GetCnExternalServiceName(src *StarRocksCluster) string {
	if src.Spec.StarRocksCnSpec != nil && src.Spec.StarRocksCnSpec.Service != nil && src.Spec.StarRocksCnSpec.Service.Name != "" {
		return src.Spec.StarRocksCnSpec.Service.Name
	}

	return src.Name + "-" + DEFAULT_CN + "-service"
}

//GetBeExternalServiceName generate the name of service that access the be.
func GetBeExternalServiceName(src *StarRocksCluster) string {
	if src.Spec.StarRocksBeSpec != nil && src.Spec.StarRocksBeSpec.Service != nil && src.Spec.StarRocksBeSpec.Service.Name != "" {
		return src.Spec.StarRocksBeSpec.Service.Name
	}

	return src.Name + "-" + DEFAULT_BE + "-service"
}

func BeStatefulSetName(src *StarRocksCluster) string {
	stname := src.Name + "-" + DEFAULT_BE
	/*if src.Spec.StarRocksBeSpec.Name != "" {
		stname = src.Spec.StarRocksBeSpec.Name
	}*/
	return stname
}

func CnStatefulSetName(src *StarRocksCluster) string {
	stname := src.Name + "-" + DEFAULT_CN
	/*	if src.Spec.StarRocksCnSpec.Name != "" {
		stname = src.Spec.StarRocksCnSpec.Name
	}*/
	return stname
}

func FeStatefulSetName(src *StarRocksCluster) string {
	stname := src.Name + "-" + DEFAULT_FE

	return stname
}
