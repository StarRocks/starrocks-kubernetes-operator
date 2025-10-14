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
	"strconv"
)

// the fe ports key
const (
	HTTP_PORT         = "http_port"
	RPC_PORT          = "rpc_port"
	QUERY_PORT        = "query_port"
	EDIT_LOG_PORT     = "edit_log_port"
	ARROW_FLIGHT_PORT = "arrow_flight_port"
)

// the cn or be ports key
const (
	// THRIFT_PORT is the old name in CN.
	// From StarRocks 3.1, both CN and BE use the same port name be_port.
	THRIFT_PORT = "thrift_port"
	BE_PORT     = "be_port"

	// WEBSERVER_PORT and BE_HTTP_PORT
	// From StarRocks 3.0, the name of HTTP port is changed to be_http_port in BE and CN.
	WEBSERVER_PORT = "webserver_port"
	BE_HTTP_PORT   = "be_http_port"

	// HEARTBEAT_SERVICE_PORT and BRPC_PORT
	// both BE and CN have the these ports
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
	BE_HTTP_PORT:           8040,
	HEARTBEAT_SERVICE_PORT: 9050,
	BRPC_PORT:              8060,
}

// GetPort get ports from config file.
func GetPort(config map[string]interface{}, key string) int32 {
	if v, ok := config[key]; ok {
		if port, err := strconv.ParseInt(v.(string), 10, 32); err == nil && port != 0 {
			return int32(port)
		}
	}

	switch key {
	case THRIFT_PORT:
		// from StarRocks 3.1, the name of thrift_port is changed to be_port.
		// If both be_port and thrift_port are set, the thrift_port will be used in StarRocks.
		// see https://github.com/StarRocks/starrocks/pull/31747
		return GetPort(config, BE_PORT)
	case WEBSERVER_PORT:
		// If both webserver_port and be_http_port are set, the be_http_port will be used in StarRocks.
		return GetPort(config, BE_HTTP_PORT)
	}

	return DefMap[key]
}
