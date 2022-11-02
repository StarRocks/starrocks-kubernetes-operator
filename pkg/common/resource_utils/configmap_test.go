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
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestResolveConfigMap(t *testing.T) {
	configMap := corev1.ConfigMap{
		Data: map[string]string{
			"fe.conf": "# Licensed to the Apache Software Foundation (ASF) under one\n# or more contributor license agreements.  See the NOTICE file\n# distributed with this work for additional information\n# regarding copyright ownership.  The ASF licenses this file\n# to you under the Apache License, Version 2.0 (the\n# \"License\"); you may not use this file except in compliance\n# with the License.  You may obtain a copy of the License at\n#\n#   http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing,\n# software distributed under the License is distributed on an\n# \"AS IS\" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY\n# KIND, either express or implied.  See the License for the\n# specific language governing permissions and limitations\n# under the License.\n\n#####################################################################\n## The uppercase properties are read and exported by bin/start_fe.sh.\n## To see all Frontend configurations,\n## see fe/src/com/starrocks/common/Config.java\n\n# the output dir of stderr/stdout/gc\nLOG_DIR = ${STARROCKS_HOME}/log\n\nDATE = \"$(date +%Y%m%d-%H%M%S)\"\nJAVA_OPTS=\"-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseMembar -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+PrintGCDateStamps -XX:+PrintGCDetails -XX:+UseConcMarkSweepGC -XX:+UseParNewGC -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xloggc:${LOG_DIR}/fe.gc.log.$DATE\"\n\n# For jdk 9+, this JAVA_OPTS will be used as default JVM options\nJAVA_OPTS_FOR_JDK_9=\"-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time\"\n\n##\n## the lowercase properties are read by main program.\n##\n\n# DEBUG, INFO, WARN, ERROR, FATAL\nsys_log_level = INFO\n\n# store metadata, create it if it is not exist.\n# Default value is ${STARROCKS_HOME}/meta\n# meta_dir = ${STARROCKS_HOME}/meta\n\nhttp_port = 8030\nrpc_port = 9020\nquery_port = 9030\nedit_log_port = 9010\nmysql_service_nio_enabled = true\n\n# Enable jaeger tracing by setting jaeger_grpc_endpoint\n# jaeger_grpc_endpoint = http://localhost:14250\n\n# Choose one if there are more than one ip except loopback address. \n# Note that there should at most one ip match this list.\n# If no ip match this rule, will choose one randomly.\n# use CIDR format, e.g. 10.10.10.0/24\n# Default value is empty.\n# priority_networks = 10.10.10.0/24;192.168.0.0/16\n\n# Advanced configurations \n# log_roll_size_mb = 1024\n# sys_log_dir = ${STARROCKS_HOME}/log\n# sys_log_roll_num = 10\n# sys_log_verbose_modules = \n# audit_log_dir = ${STARROCKS_HOME}/log\n# audit_log_modules = slow_query, query\n# audit_log_roll_num = 10\n# meta_delay_toleration_second = 10\n# qe_max_connection = 1024\n# max_conn_per_user = 100\n# qe_query_timeout_second = 300\n# qe_slow_log_ms = 5000\n",
		},
	}
	res, err := ResolveConfigMap(&configMap, "fe.conf")
	require.NoError(t, err)

	_, ok := res["http_port"]
	require.Equal(t, true, ok)
}
