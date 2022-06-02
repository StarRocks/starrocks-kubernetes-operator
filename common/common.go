/*
Copyright 2022 StarRocks.

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
package common

const (
	CnApiVersionV1ALPHA = "starrocks.com/v1alpha1"
	CnKind              = "ComputeNodeGroup"

	CnOfflineJobRole = "cn-manager-role"
)

const (
	// env keys
	EnvKeyFeAddrs = "FE_ADDRS"
	EnvKeyFeUsr   = "FE_USR"
	EnvKeyFePwd   = "FE_PWD"
	EnvKeyCnName  = "CN_NAME"
	EnvKeyCnNs    = "CN_NS"
	EnvKeyCnPort  = "CN_PORT"

	FeAddrsSeparator = ","
)

const (
	// finalizer name of cn resource
	CnFinalizerName = "cn.starrocks.finalizers"
)

const (
	// port of cn to keep-alive
	CnHeartBeatPort = "9050"
)
