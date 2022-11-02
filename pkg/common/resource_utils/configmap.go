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
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"strconv"
)

//the fe ports key
const (
	HTTP_PORT     = "http_port"
	RPC_PORT      = "rpc_port"
	QUERY_PORT    = "query_port"
	EDIT_LOG_PORT = "edit_log_port"
)

//the cn or be ports key
const (
	THRIFT_PORT            = "thrift_port"
	BE_PORT                = "be_port"
	WEBSERVER_PORT         = "webserver_port"
	HEARTBEAT_SERVICE_PORT = "heartbeat_service_port"
	BRPC_PORT              = "brpc_port"
)

func ResolveConfigMap(configMap *corev1.ConfigMap, key string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	data := configMap.Data
	if _, ok := data[key]; !ok {
		return res, nil
	}

	value, _ := data[key]
	klog.Info("the resolve message ", value)
	viper.SetConfigType("properties")
	viper.ReadConfig(bytes.NewBuffer([]byte(value)))

	return viper.AllSettings(), nil
}

//getPort get ports from config file.
func GetPort(config map[string]interface{}, key string) int32 {
	if key == HTTP_PORT {
		if v, ok := config[HTTP_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 8030
	}

	if key == RPC_PORT {
		if v, ok := config[RPC_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}
		return 9020
	}

	if key == QUERY_PORT {
		if v, ok := config[QUERY_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 9030
	}

	if key == EDIT_LOG_PORT {
		if v, ok := config[EDIT_LOG_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 9010
	}

	if key == THRIFT_PORT {
		if v, ok := config[THRIFT_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 9060
	}

	if key == BE_PORT {
		if v, ok := config[BE_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 9060
	}

	if key == WEBSERVER_PORT {
		if v, ok := config[WEBSERVER_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 8040
	}

	if key == HEARTBEAT_SERVICE_PORT {
		if v, ok := config[HEARTBEAT_SERVICE_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 9050
	}

	if key == BRPC_PORT {
		if v, ok := config[BRPC_PORT]; ok {
			if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil {
				return int32(port)
			}
		}

		return 8060
	}

	return 0
}
