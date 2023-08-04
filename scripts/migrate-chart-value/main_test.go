package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

const OPERATOR_YAML = `
global:
  rbac:
    create: true
timeZone: Asia/Shanghai
nameOverride: "kube-starrocks"
starrocksOperator:
  enabled: true
  annotations: {}
  namespaceOverride: ""
  image:
    repository: starrocks/operator
    tag: v1.7.1
  imagePullPolicy: Always
  replicaCount: 1
  resources:
    limits:
      cpu: 500m
      memory: 200Mi
    requests:
      cpu: 500m
      memory: 200Mi
  nodeSelector: {}
  tolerations: []
`
const STARROCKS_YAML = `
nameOverride: "kube-starrocks"
initPassword:
  enabled: false
  password: ""
  passwordSecret: ""
timeZone: Asia/Shanghai
datadog:
  log:
    enabled: false
  metrics:
    enabled: false
starrocksCluster:
  name: ""
  namespace: ""
  annotations: {}
  enabledCn: false
starrocksFESpec:
  replicas: 1
  image:
    repository: starrocks/fe-ubuntu
    tag: 3.0-latest
  annotations: {}
  runAsNonRoot: false
  service:
    type: "ClusterIP"
    loadbalancerIP: ""
    annotations: {}
  imagePullSecrets: []
  serviceAccount: ""
  nodeSelector: {}
  podLabels: {}
  hostAliases: []
  schedulerName: ""
  feEnvVars: []
  affinity: {}
  tolerations: []
  resources:
    requests:
      cpu: 4
      memory: 4Gi
    limits:
      cpu: 8
      memory: 8Gi
  storageSpec:
    name: ""
    storageClassName: ""
    storageSize: 1Gi
    logStorageSize: 1Gi
  config: |
    LOG_DIR = ${STARROCKS_HOME}/log
    DATE = "$(date +%Y%m%d-%H%M%S)"
    JAVA_OPTS="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseMembar -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+PrintGCDateStamps -XX:+PrintGCDetails -XX:+UseConcMarkSweepGC -XX:+UseParNewGC -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xloggc:${LOG_DIR}/fe.gc.log.$DATE"
    JAVA_OPTS_FOR_JDK_9="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time"
    http_port = 8030
    rpc_port = 9020
    query_port = 9030
    edit_log_port = 9010
    mysql_service_nio_enabled = true
    sys_log_level = INFO
  secrets: []
  configMaps: []
starrocksCnSpec:
  image:
    repository: starrocks/cn-ubuntu
    tag: 3.0-latest
  serviceAccount: ""
  annotations: {}
  runAsNonRoot: false
  service:
    type: "ClusterIP"
    loadbalancerIP: ""
    annotations: {}
  imagePullSecrets: []
  nodeSelector: {}
  podLabels: {}
  hostAliases: []
  schedulerName: ""
  cnEnvVars: []
  affinity: {}
  tolerations: []
  autoScalingPolicy: {}
  resources:
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 4
      memory: 8Gi
  config: |
    sys_log_level = INFO
    thrift_port = 9060
    webserver_port = 8040
    heartbeat_service_port = 9050
    brpc_port = 8060
  secrets: []
  configMaps: []
starrocksBeSpec:
  replicas: 1
  image:
    repository: starrocks/be-ubuntu
    tag: 3.0-latest
  serviceAccount: ""
  annotations: {}
  runAsNonRoot: false
  service:
    type: "ClusterIP"
    loadbalancerIP: ""
    annotations: {}
  imagePullSecrets: []
  nodeSelector: {}
  podLabels: {}
  hostAliases: []
  schedulerName: ""
  beEnvVars: []
  affinity: {}
  tolerations: []
  resources:
    requests:
      cpu: 4
      memory: 4Gi
    limits:
      cpu: 8
      memory: 8Gi
  storageSpec:
    name: ""
    storageClassName: ""
    storageSize: 1Ti
    logStorageSize: 1Gi
  config: |
    be_port = 9060
    webserver_port = 8040
    heartbeat_service_port = 9050
    brpc_port = 8060
    sys_log_level = INFO
    default_rowset_type = beta
  secrets: []
  configMaps: []
secrets: []
configMaps: []
`
const V1_8_0_YAML = `
operator:
  global:
    rbac:
      create: true
  starrocksOperator:
    enabled: true
    annotations: {}
    namespaceOverride: ""
    image:
      repository: starrocks/operator
      tag: v1.7.1
    imagePullPolicy: Always
    replicaCount: 1
    resources:
      limits:
        cpu: 500m
        memory: 200Mi
      requests:
        cpu: 500m
        memory: 200Mi
    nodeSelector: {}
    tolerations: []
starrocks:
  nameOverride: "kube-starrocks"
  initPassword:
    enabled: false
    password: ""
    passwordSecret: ""
  timeZone: Asia/Shanghai
  datadog:
    log:
      enabled: false
    metrics:
      enabled: false
  starrocksCluster:
    name: ""
    namespace: ""
    annotations: {}
    enabledCn: false
  starrocksFESpec:
    replicas: 1
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.0-latest
    annotations: {}
    runAsNonRoot: false
    service:
      type: "ClusterIP"
      loadbalancerIP: ""
      annotations: {}
    imagePullSecrets: []
    serviceAccount: ""
    nodeSelector: {}
    podLabels: {}
    hostAliases: []
    schedulerName: ""
    feEnvVars: []
    affinity: {}
    tolerations: []
    resources:
      requests:
        cpu: 4
        memory: 4Gi
      limits:
        cpu: 8
        memory: 8Gi
    storageSpec:
      name: ""
      storageClassName: ""
      storageSize: 1Gi
      logStorageSize: 1Gi
    config: |
      LOG_DIR = ${STARROCKS_HOME}/log
      DATE = "$(date +%Y%m%d-%H%M%S)"
      JAVA_OPTS="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseMembar -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+PrintGCDateStamps -XX:+PrintGCDetails -XX:+UseConcMarkSweepGC -XX:+UseParNewGC -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xloggc:${LOG_DIR}/fe.gc.log.$DATE"
      JAVA_OPTS_FOR_JDK_9="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:SurvivorRatio=8 -XX:MaxTenuringThreshold=7 -XX:+CMSClassUnloadingEnabled -XX:-CMSParallelRemarkEnabled -XX:CMSInitiatingOccupancyFraction=80 -XX:SoftRefLRUPolicyMSPerMB=0 -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time"
      http_port = 8030
      rpc_port = 9020
      query_port = 9030
      edit_log_port = 9010
      mysql_service_nio_enabled = true
      sys_log_level = INFO
    secrets: []
    configMaps: []
  starrocksCnSpec:
    image:
      repository: starrocks/cn-ubuntu
      tag: 3.0-latest
    serviceAccount: ""
    annotations: {}
    runAsNonRoot: false
    service:
      type: "ClusterIP"
      loadbalancerIP: ""
      annotations: {}
    imagePullSecrets: []
    nodeSelector: {}
    podLabels: {}
    hostAliases: []
    schedulerName: ""
    cnEnvVars: []
    affinity: {}
    tolerations: []
    autoScalingPolicy: {}
    resources:
      limits:
        cpu: 8
        memory: 8Gi
      requests:
        cpu: 4
        memory: 8Gi
    config: |
      sys_log_level = INFO
      thrift_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
    secrets: []
    configMaps: []
  starrocksBeSpec:
    replicas: 1
    image:
      repository: starrocks/be-ubuntu
      tag: 3.0-latest
    serviceAccount: ""
    annotations: {}
    runAsNonRoot: false
    service:
      type: "ClusterIP"
      loadbalancerIP: ""
      annotations: {}
    imagePullSecrets: []
    nodeSelector: {}
    podLabels: {}
    hostAliases: []
    schedulerName: ""
    beEnvVars: []
    affinity: {}
    tolerations: []
    resources:
      requests:
        cpu: 4
        memory: 4Gi
      limits:
        cpu: 8
        memory: 8Gi
    storageSpec:
      name: ""
      storageClassName: ""
      storageSize: 1Ti
      logStorageSize: 1Gi
    config: |
      be_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
      sys_log_level = INFO
      default_rowset_type = beta
    secrets: []
    configMaps: []
  secrets: []
  configMaps: []
`

var V1_7_1_YAML = func() string {
	// remove the line has timeZone or nameOverride
	val := strings.ReplaceAll(OPERATOR_YAML, "timeZone: Asia/Shanghai\n", "")
	val = strings.ReplaceAll(val, "nameOverride: \"kube-starrocks\"\n", "")
	// remove whitespace line
	val = strings.ReplaceAll(val, "\n\n", "\n")
	return fmt.Sprintf("%v\n%v", val, STARROCKS_YAML)
}()

func addHeader2(data string, header string) string {
	fields := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(data), &fields)
	if err != nil {
		panic(err)
	}
	output, err := AddHeader(fields, header)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func TestDo(t *testing.T) {
	type args struct {
		reader       io.Reader
		chartVersion string
	}
	tests := []struct {
		name    string
		args    args
		wantW1  string
		wantErr bool
	}{
		{
			name: "equalValues-v1.7.1",
			args: args{
				reader:       strings.NewReader(V1_7_1_YAML),
				chartVersion: "v1.7.1",
			},
			wantW1:  "",
			wantErr: false,
		},
		{
			name: "equalValues-v1.8.0",
			args: args{
				reader:       strings.NewReader(V1_8_0_YAML),
				chartVersion: "v1.8.0",
			},
			wantW1:  "",
			wantErr: false,
		},
		{
			name: "upgradeValues-v1.8.0",
			args: args{
				reader:       strings.NewReader(V1_7_1_YAML),
				chartVersion: "v1.8.0",
			},
			wantW1: fmt.Sprintf("%v\n%v", marshal(addHeader2(OPERATOR_YAML, "operator")),
				marshal(addHeader2(STARROCKS_YAML, "starrocks"))),
			wantErr: false,
		},
		{
			name: "downgradeValues-v1.7.1",
			args: args{
				reader:       strings.NewReader(V1_8_0_YAML),
				chartVersion: "v1.7.1",
			},
			wantW1: func() string {
				v := map[string]interface{}{}
				if err := yaml.Unmarshal([]byte(OPERATOR_YAML), &v); err != nil {
					panic(err)
				}
				delete(v, "timeZone")
				delete(v, "nameOverride")
				data, err := yaml.Marshal(v)
				if err != nil {
					panic(err)
				}
				return fmt.Sprintf("%v\n%v", marshal(string(data)), marshal(STARROCKS_YAML))
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w1 := &bytes.Buffer{}
			err := Do(tt.args.reader, tt.args.chartVersion, w1)
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW1 := w1.String(); gotW1 != tt.wantW1 {
				t.Errorf("Do() gotW1 = %v, want %v", gotW1, tt.wantW1)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	type args struct {
		keys   []string
		header string
		s      map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "do not have keys",
			args: args{
				s: map[string]interface{}{
					"targetChartVersion": "v1.7.1",
				},
				keys:   []string{""},
				header: "operator",
			},
			wantW:   "operator:",
			wantErr: false,
		},
		{
			name: "do not have keys or header",
			args: args{
				s: map[string]interface{}{
					"targetChartVersion": "v1.7.1",
				},
				keys:   []string{""},
				header: "",
			},
			wantW:   "",
			wantErr: false,
		},
		{
			name: "has keys and header",
			args: args{
				s: map[string]interface{}{
					"targetChartVersion": "v1.7.1",
				},
				keys:   []string{"targetChartVersion"},
				header: "operator",
			},
			wantW:   "operator:\n  targetChartVersion: v1.7.1",
			wantErr: false,
		},
		{
			name: "has keys but does not have header",
			args: args{
				s: map[string]interface{}{
					"targetChartVersion": "v1.7.1",
				},
				keys:   []string{"targetChartVersion"},
				header: "",
			},
			wantW:   "targetChartVersion: v1.7.1",
			wantErr: false,
		},
		{
			name: "do not match keys",
			args: args{
				s: map[string]interface{}{
					"targetChartVersion": "v1.7.1",
				},
				keys:   []string{"version"},
				header: "operator",
			},
			wantW:   "operator:",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := Write(w, tt.args.s, tt.args.keys, tt.args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr\n%v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); strings.Trim(gotW, "\n ") != strings.Trim(tt.wantW, "\n ") {
				t.Errorf("Write() gotW =\n%v, want\n%v", gotW, tt.wantW)
				fmt.Println(gotW)
				fmt.Println(tt.wantW)
			}
		})
	}
}

func marshal(s string) string {
	v := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		panic(err)
	}
	b, err := yaml.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
