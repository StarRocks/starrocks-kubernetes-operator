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

package resource_utils

import (
	"bytes"
	"strconv"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

// the fe ports key
const (
	HTTP_PORT     = "http_port"
	RPC_PORT      = "rpc_port"
	QUERY_PORT    = "query_port"
	EDIT_LOG_PORT = "edit_log_port"
)

// the cn or be ports key
const (
	THRIFT_PORT            = "thrift_port"
	BE_PORT                = "be_port"
	WEBSERVER_PORT         = "webserver_port"
	HEARTBEAT_SERVICE_PORT = "heartbeat_service_port"
	BRPC_PORT              = "brpc_port"
)

// the fe proxy ports key
const (
	FE_PROXY_HTTP_PORT      = 8080
	FE_PORXY_HTTP_PORT_NAME = "http-port"
)

// DefMap the default port about abilities.
var DefMap = map[string]int32{
	HTTP_PORT:              8030,
	RPC_PORT:               9020,
	QUERY_PORT:             9030,
	EDIT_LOG_PORT:          9010,
	THRIFT_PORT:            9060,
	BE_PORT:                9060,
	WEBSERVER_PORT:         8040,
	HEARTBEAT_SERVICE_PORT: 9050,
	BRPC_PORT:              8060,
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

// GetPort get ports from config file.
func GetPort(config map[string]interface{}, key string) int32 {
	if v, ok := config[key]; ok {
		if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
			return int32(port)
		}
	}
	return DefMap[key]
}
