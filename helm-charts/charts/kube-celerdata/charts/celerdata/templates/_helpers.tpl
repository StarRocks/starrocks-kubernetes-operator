{{/*
Common labels
*/}}
{{- define "celerdatacluster.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
initpassword secret name
*/}}

{{- define "celerdatacluster.initpassword.secret.name" -}}
{{ default (print (include "celerdatacluster.name" .) "-credential") .Values.initPassword.passwordSecret }}
{{- end }}

{{/*
celerdatacluster
*/}}

{{- define "celerdatacluster.name" -}}
{{ default (default .Chart.Name .Values.nameOverride) .Values.celerDataCluster.name }}
{{- end }}

{{- define "celerdatacluster.namespace" -}}
{{ default .Release.Namespace .Values.celerDataCluster.namespace }}
{{- end }}

{{- define "celerdatacluster.fe.name" -}}
{{- print (include "celerdatacluster.name" .) "-fe" }}
{{- end }}

{{- define "celerdatacluster.cn.name" -}}
{{- print (include "celerdatacluster.name" .) "-cn" }}
{{- end }}

{{- define "celerdatacluster.be.name" -}}
{{- print (include "celerdatacluster.name" .) "-be" }}
{{- end }}

{{- define "celerdatacluster.be.configmap.name" -}}
{{- print (include "celerdatacluster.be.name" .) "-cm" }}
{{- end }}

{{- define "celerdatacluster.fe.configmap.name" -}}
{{- print (include "celerdatacluster.fe.name" .) "-cm" }}
{{- end }}

{{- define "celerdatacluster.cn.configmap.name" -}}
{{- print (include "celerdatacluster.cn.name" .) "-cm" }}
{{- end }}

{{- define "celerdatacluster.fe.config" -}}
fe.conf: |
{{- if and .Values.celerDataFeSpec.configyaml (kindIs "map" .Values.celerDataFeSpec.configyaml) }}
  {{- range $key, $value := .Values.celerDataFeSpec.configyaml }}
    {{ $key }} = {{ $value }}
  {{- end }}
{{- else if .Values.celerDataFeSpec.configyaml }}
  {{ fail "configyaml must be a map" }}
{{- else }}
  {{- .Values.celerDataFeSpec.config | nindent 2 }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.cn.config" -}}
cn.conf: |
{{- if and .Values.celerDataCnSpec.configyaml (kindIs "map" .Values.celerDataCnSpec.configyaml) }}
  {{- range $key, $value := .Values.celerDataCnSpec.configyaml }}
    {{ $key }} = {{ $value }}
  {{- end }}
{{- else if .Values.celerDataCnSpec.configyaml }}
  {{ fail "configyaml must be a map" }}
{{- else }}
  {{- .Values.celerDataCnSpec.config | nindent 2 }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.be.config" -}}
be.conf: |
{{- if and .Values.celerDataBeSpec.configyaml (kindIs "map" .Values.celerDataBeSpec.configyaml) }}
  {{- range $key, $value := .Values.celerDataBeSpec.configyaml }}
    {{ $key }} = {{ $value }}
  {{- end }}
{{- else if .Values.celerDataBeSpec.configyaml }}
  {{ fail "configyaml must be a map" }}
{{- else }}
  {{- .Values.celerDataBeSpec.config | nindent 2 }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.fe.meta.suffix" -}}
{{- print "-meta" }}
{{- end }}

{{- define "celerdatacluster.fe.meta.path" -}}
{{- if .Values.celerDataFeSpec.storageSpec.storageMountPath }}
{{- print .Values.celerDataFeSpec.storageSpec.storageMountPath }}
{{- else }}
{{- print "/opt/starrocks/fe/meta" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.fe.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "celerdatacluster.fe.log.path" -}}
{{- if .Values.celerDataFeSpec.storageSpec.logMountPath }}
{{- print .Values.celerDataFeSpec.storageSpec.logMountPath }}
{{- else }}
{{- print "/opt/starrocks/fe/log" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.be.data.suffix" -}}
{{- print "-data" }}
{{- end }}

{{- define "celerdatacluster.be.data.path" -}}
{{- if .Values.celerDataBeSpec.storageSpec.storageMountPath }}
{{- print .Values.celerDataBeSpec.storageSpec.storageMountPath }}
{{- else }}
{{- print "/opt/starrocks/be/storage" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.be.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "celerdatacluster.be.log.path" -}}
{{- if .Values.celerDataBeSpec.storageSpec.logMountPath }}
{{- print .Values.celerDataBeSpec.storageSpec.logMountPath }}
{{- else }}
{{- print "/opt/starrocks/be/log" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.be.spill.suffix" -}}
{{- print "-spill" }}
{{- end }}

{{- define "celerdatacluster.be.spill.path" -}}
{{- if .Values.celerDataBeSpec.storageSpec.spillMountPath }}
{{- print .Values.celerDataBeSpec.storageSpec.spillMountPath }}
{{- else }}
{{- print "/opt/starrocks/be/spill" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.cn.data.suffix" -}}
{{- print "-data" }}
{{- end }}

{{- define "celerdatacluster.cn.data.path" -}}
{{- if .Values.celerDataCnSpec.storageSpec.storageMountPath }}
{{- print .Values.celerDataCnSpec.storageSpec.storageMountPath }}
{{- else }}
{{- print "/opt/starrocks/cn/storage" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.cn.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "celerdatacluster.cn.log.path" -}}
{{- if .Values.celerDataCnSpec.storageSpec.logMountPath }}
{{- print .Values.celerDataCnSpec.storageSpec.logMountPath }}
{{- else }}
{{- print "/opt/starrocks/cn/log" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.cn.spill.suffix" -}}
{{- print "-spill" }}
{{- end }}

{{- define "celerdatacluster.cn.spill.path" -}}
{{- if .Values.celerDataCnSpec.storageSpec.spillMountPath }}
{{- print .Values.celerDataCnSpec.storageSpec.spillMountPath }}
{{- else }}
{{- print "/opt/starrocks/cn/spill" }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.entrypoint.script.name" -}}
{{- print "entrypoint.sh" }}
{{- end }}

{{- define "celerdatacluster.entrypoint.mount.path" -}}
{{- print "/etc/celerdata" }}
{{- end }}

{{- define "celerdatacluster.fe.entrypoint.script.configmap.name" -}}
{{- print (include "celerdatacluster.name" .) "-fe-entrypoint-script" }}
{{- end }}

{{- define "celerdatacluster.be.entrypoint.script.configmap.name" -}}
{{- print (include "celerdatacluster.name" .) "-be-entrypoint-script" }}
{{- end }}

{{- define "celerdatacluster.cn.entrypoint.script.configmap.name" -}}
{{- print (include "celerdatacluster.name" .) "-cn-entrypoint-script" }}
{{- end }}

{{/*
Define a function to handle resource limits for fe
*/}}
{{- define "celerdatacluster.fe.resources" -}}
requests:
  {{- toYaml .Values.celerDataFeSpec.resources.requests | nindent 2 }}
limits:
{{- range $key, $value := .Values.celerDataFeSpec.resources.limits }}
  {{- if ne (toString $value) "unlimited" }}
  {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Define a function to handle resource limits for be
*/}}
{{- define "celerdatacluster.be.resources" -}}
requests:
  {{- toYaml .Values.celerDataBeSpec.resources.requests | nindent 2 }}
limits:
{{- range $key, $value := .Values.celerDataBeSpec.resources.limits }}
  {{- if ne (toString $value) "unlimited" }}
  {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Define a function to handle resource limits for cn
*/}}
{{- define "celerdatacluster.cn.resources" -}}
requests:
  {{- toYaml .Values.celerDataCnSpec.resources.requests | nindent 2 }}
limits:
{{- range $key, $value := .Values.celerDataCnSpec.resources.limits }}
  {{- if ne (toString $value) "unlimited" }}
  {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
celerdatacluster.fe.config.hash is used to calculate the hash value of the fe.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "celerdatacluster.fe.config.hash" }}
  {{- if and .Values.celerDataFeSpec.configyaml (kindIs "map" .Values.celerDataFeSpec.configyaml) }}
    {{- $hash := toJson .Values.celerDataFeSpec.configyaml | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else if .Values.celerDataFeSpec.configyaml }}
    {{ fail "configyaml must be a map" }}
  {{- else if .Values.celerDataFeSpec.config }}
    {{- $hash := toJson .Values.celerDataFeSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}


{{/*
celerdatacluster.be.config.hash is used to calculate the hash value of the be.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "celerdatacluster.be.config.hash" }}
  {{- if and .Values.celerDataBeSpec.configyaml (kindIs "map" .Values.celerDataBeSpec.configyaml) }}
    {{- $hash := toJson .Values.celerDataBeSpec.configyaml | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else if .Values.celerDataBeSpec.configyaml }}
    {{ fail "configyaml must be a map" }}
  {{- else if .Values.celerDataBeSpec.config }}
    {{- $hash := toJson .Values.celerDataBeSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{/*
celerdatacluster.cn.config.hash is used to calculate the hash value of the cn.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "celerdatacluster.cn.config.hash" }}
  {{- if and .Values.celerDataCnSpec.configyaml (kindIs "map" .Values.celerDataCnSpec.configyaml) }}
    {{- $hash := toJson .Values.celerDataCnSpec.configyaml | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else if .Values.celerDataCnSpec.configyaml }}
    {{ fail "configyaml must be a map" }}
  {{- else if .Values.celerDataCnSpec.config }}
    {{- $hash := toJson .Values.celerDataCnSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "celerdatacluster.fe.query.port" -}}
{{- $config := index .Values.celerDataFeSpec.config  -}}
{{- $configMap := dict -}}
{{- range $line := splitList "\n" $config -}}
{{- $pair := splitList "=" $line -}}
{{- if eq (len $pair) 2 -}}
{{- $_ := set $configMap (trim (index $pair 0)) (trim (index $pair 1)) -}}
{{- end -}}
{{- end -}}
{{- if (index $configMap "query_port") -}}
{{- print (index $configMap "query_port") }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.fe.entrypoint.script.hash" }}
  {{- if .Values.celerDataFeSpec.entrypoint }}
    {{- $hash := toJson .Values.celerDataFeSpec.entrypoint.script | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "celerdatacluster.be.entrypoint.script.hash" }}
  {{- if .Values.celerDataBeSpec.entrypoint }}
    {{- $hash := toJson .Values.celerDataBeSpec.entrypoint.script | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "celerdatacluster.cn.entrypoint.script.hash" }}
  {{- if .Values.celerDataCnSpec.entrypoint }}
    {{- $hash := toJson .Values.celerDataCnSpec.entrypoint.script | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "celerdatacluster.fe.http.port" -}}
{{- $config := index .Values.celerDataFeSpec.config  -}}
{{- $configMap := dict -}}
{{- range $line := splitList "\n" $config -}}
{{- $pair := splitList "=" $line -}}
{{- if eq (len $pair) 2 -}}
{{- $_ := set $configMap (trim (index $pair 0)) (trim (index $pair 1)) -}}
{{- end -}}
{{- end -}}
{{- if (index $configMap "http_port") -}}
{{- print (index $configMap "http_port") }}
{{- end }}
{{- end }}

{{- define "celerdatacluster.be.webserver.port" -}}
{{- include "celerdatacluster.webserver.port" .Values.celerDataBeSpec }}
{{- end }}

{{- define "celerdatacluster.cn.webserver.port" -}}
{{- include "celerdatacluster.webserver.port" .Values.celerDataCnSpec }}
{{- end }}

{{- define "celerdatacluster.webserver.port" -}}
{{- $config := index .config  -}}
{{- $configMap := dict -}}
{{- range $line := splitList "\n" $config -}}
{{- $pair := splitList "=" $line -}}
{{- if eq (len $pair) 2 -}}
{{- $_ := set $configMap (trim (index $pair 0)) (trim (index $pair 1)) -}}
{{- end -}}
{{- end -}}
{{- if (index $configMap "webserver_port") -}}
{{- print (index $configMap "webserver_port") }}
{{- end }}
{{- end }}

{{/*
Get the value of the schedulerName field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.schedulerName" -}}
{{- if .Values.celerDataFeSpec.schedulerName -}}
{{- .Values.celerDataFeSpec.schedulerName -}}
{{- else if .Values.celerDataCluster.componentValues.schedulerName -}}
{{- .Values.celerDataCluster.componentValues.schedulerName -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the schedulerName field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.schedulerName" -}}
{{- if .Values.celerDataBeSpec.schedulerName -}}
{{- .Values.celerDataBeSpec.schedulerName -}}
{{- else if .Values.celerDataCluster.componentValues.schedulerName -}}
{{- .Values.celerDataCluster.componentValues.schedulerName -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the schedulerName field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.schedulerName" -}}
{{- if .Values.celerDataCnSpec.schedulerName -}}
{{- .Values.celerDataCnSpec.schedulerName -}}
{{- else if .Values.celerDataCluster.componentValues.schedulerName -}}
{{- .Values.celerDataCluster.componentValues.schedulerName -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the serviceAccount field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.serviceAccount" -}}
{{- if .Values.celerDataFeSpec.serviceAccount -}}
{{- .Values.celerDataFeSpec.serviceAccount -}}
{{- else if .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the serviceAccount field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.serviceAccount" -}}
{{- if .Values.celerDataBeSpec.serviceAccount -}}
{{- .Values.celerDataBeSpec.serviceAccount -}}
{{- else if .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the serviceAccount field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.serviceAccount" -}}
{{- if .Values.celerDataCnSpec.serviceAccount -}}
{{- .Values.celerDataCnSpec.serviceAccount -}}
{{- else if .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- .Values.celerDataCluster.componentValues.serviceAccount -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the imagePullSecrets field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.imagePullSecrets" -}}
{{- if .Values.celerDataFeSpec.imagePullSecrets -}}
{{- toYaml .Values.celerDataFeSpec.imagePullSecrets -}}
{{- else if .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- toYaml .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the imagePullSecrets field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.imagePullSecrets" -}}
{{- if .Values.celerDataBeSpec.imagePullSecrets -}}
{{- toYaml .Values.celerDataBeSpec.imagePullSecrets -}}
{{- else if .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- toYaml .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the imagePullSecrets field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.imagePullSecrets" -}}
{{- if .Values.celerDataCnSpec.imagePullSecrets -}}
{{- toYaml .Values.celerDataCnSpec.imagePullSecrets -}}
{{- else if .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- toYaml .Values.celerDataCluster.componentValues.imagePullSecrets -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the tolerations field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.tolerations" -}}
{{- if .Values.celerDataFeSpec.tolerations -}}
{{- toYaml .Values.celerDataFeSpec.tolerations -}}
{{- else if .Values.celerDataCluster.componentValues.tolerations -}}
{{- toYaml .Values.celerDataCluster.componentValues.tolerations -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the tolerations field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.tolerations" -}}
{{- if .Values.celerDataBeSpec.tolerations -}}
{{- toYaml .Values.celerDataBeSpec.tolerations -}}
{{- else if .Values.celerDataCluster.componentValues.tolerations -}}
{{- toYaml .Values.celerDataCluster.componentValues.tolerations -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the tolerations field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.tolerations" -}}
{{- if .Values.celerDataCnSpec.tolerations -}}
{{- toYaml .Values.celerDataCnSpec.tolerations -}}
{{- else if .Values.celerDataCluster.componentValues.tolerations -}}
{{- toYaml .Values.celerDataCluster.componentValues.tolerations -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the nodeSelector field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.nodeSelector" -}}
{{- if .Values.celerDataFeSpec.nodeSelector -}}
{{- toYaml .Values.celerDataFeSpec.nodeSelector -}}
{{- else if .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- toYaml .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the nodeSelector field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.nodeSelector" -}}
{{- if .Values.celerDataBeSpec.nodeSelector -}}
{{- toYaml .Values.celerDataBeSpec.nodeSelector -}}
{{- else if .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- toYaml .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the nodeSelector field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.nodeSelector" -}}
{{- if .Values.celerDataCnSpec.nodeSelector -}}
{{- toYaml .Values.celerDataCnSpec.nodeSelector -}}
{{- else if .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- toYaml .Values.celerDataCluster.componentValues.nodeSelector -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the affinity field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.affinity" -}}
{{- if .Values.celerDataFeSpec.affinity -}}
{{- toYaml .Values.celerDataFeSpec.affinity -}}
{{- else if .Values.celerDataCluster.componentValues.affinity -}}
{{- toYaml .Values.celerDataCluster.componentValues.affinity -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the affinity field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.affinity" -}}
{{- if .Values.celerDataBeSpec.affinity -}}
{{- toYaml .Values.celerDataBeSpec.affinity -}}
{{- else if .Values.celerDataCluster.componentValues.affinity -}}
{{- toYaml .Values.celerDataCluster.componentValues.affinity -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the affinity field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.affinity" -}}
{{- if .Values.celerDataCnSpec.affinity -}}
{{- toYaml .Values.celerDataCnSpec.affinity -}}
{{- else if .Values.celerDataCluster.componentValues.affinity -}}
{{- toYaml .Values.celerDataCluster.componentValues.affinity -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the topologySpreadConstraints field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.topologySpreadConstraints" -}}
{{- if .Values.celerDataFeSpec.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataFeSpec.topologySpreadConstraints -}}
{{- else if .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the topologySpreadConstraints field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.topologySpreadConstraints" -}}
{{- if .Values.celerDataBeSpec.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataBeSpec.topologySpreadConstraints -}}
{{- else if .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the topologySpreadConstraints field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.topologySpreadConstraints" -}}
{{- if .Values.celerDataCnSpec.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataCnSpec.topologySpreadConstraints -}}
{{- else if .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- toYaml .Values.celerDataCluster.componentValues.topologySpreadConstraints -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of the runAsNonRoot field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.runAsNonRoot" -}}
{{- if .Values.celerDataFeSpec.runAsNonRoot -}}
{{- .Values.celerDataFeSpec.runAsNonRoot -}}
{{- else if .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Get the value of the runAsNonRoot field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.runAsNonRoot" -}}
{{- if .Values.celerDataBeSpec.runAsNonRoot -}}
{{- .Values.celerDataBeSpec.runAsNonRoot -}}
{{- else if .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Get the value of the runAsNonRoot field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.runAsNonRoot" -}}
{{- if .Values.celerDataCnSpec.runAsNonRoot -}}
{{- .Values.celerDataCnSpec.runAsNonRoot -}}
{{- else if .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- .Values.celerDataCluster.componentValues.runAsNonRoot -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Get the value of hostAliases field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.hostAliases" -}}
{{- if .Values.celerDataFeSpec.hostAliases -}}
{{- toYaml .Values.celerDataFeSpec.hostAliases -}}
{{- else if .Values.celerDataCluster.componentValues.hostAliases -}}
{{- toYaml .Values.celerDataCluster.componentValues.hostAliases -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of hostAliases field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.hostAliases" -}}
{{- if .Values.celerDataBeSpec.hostAliases -}}
{{- toYaml .Values.celerDataBeSpec.hostAliases -}}
{{- else if .Values.celerDataCluster.componentValues.hostAliases -}}
{{- toYaml .Values.celerDataCluster.componentValues.hostAliases -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of hostAliases field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.hostAliases" -}}
{{- if .Values.celerDataCnSpec.hostAliases -}}
{{- toYaml .Values.celerDataCnSpec.hostAliases -}}
{{- else if .Values.celerDataCluster.componentValues.hostAliases -}}
{{- toYaml .Values.celerDataCluster.componentValues.hostAliases -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of tag field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.image.tag" -}}
{{- if and .Values.celerDataFeSpec.image.tag (ne (toString .Values.celerDataFeSpec.image.tag) "") -}}
{{- .Values.celerDataFeSpec.image.tag -}}
{{- else -}}
{{- .Values.celerDataCluster.componentValues.image.tag -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of tag field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.image.tag" -}}
{{- if and .Values.celerDataBeSpec.image.tag (ne (toString .Values.celerDataBeSpec.image.tag) "") -}}
{{- .Values.celerDataBeSpec.image.tag -}}
{{- else -}}
{{- .Values.celerDataCluster.componentValues.image.tag -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of tag field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.image.tag" -}}
{{- if and .Values.celerDataCnSpec.image.tag (ne (toString .Values.celerDataCnSpec.image.tag) "") -}}
{{- .Values.celerDataCnSpec.image.tag -}}
{{- else -}}
{{- .Values.celerDataCluster.componentValues.image.tag -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of podLabels field in the celerDataFeSpec
*/}}
{{- define "celerdatacluster.fe.podLabels" -}}
{{- if .Values.celerDataFeSpec.podLabels -}}
{{- toYaml .Values.celerDataFeSpec.podLabels -}}
{{- else if .Values.celerDataCluster.componentValues.podLabels -}}
{{- toYaml .Values.celerDataCluster.componentValues.podLabels -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of podLabels field in the celerDataBeSpec
*/}}
{{- define "celerdatacluster.be.podLabels" -}}
{{- if .Values.celerDataBeSpec.podLabels -}}
{{- toYaml .Values.celerDataBeSpec.podLabels -}}
{{- else if .Values.celerDataCluster.componentValues.podLabels -}}
{{- toYaml .Values.celerDataCluster.componentValues.podLabels -}}
{{- end -}}
{{- end -}}

{{/*
Get the value of podLabels field in the celerDataCnSpec
*/}}
{{- define "celerdatacluster.cn.podLabels" -}}
{{- if .Values.celerDataCnSpec.podLabels -}}
{{- toYaml .Values.celerDataCnSpec.podLabels -}}
{{- else if .Values.celerDataCluster.componentValues.podLabels -}}
{{- toYaml .Values.celerDataCluster.componentValues.podLabels -}}
{{- end -}}
{{- end -}}

{{/*
Build the Datadog log annotation value for a given component.
Arguments (passed as a dict via "include"):
  .root = root context (has .Values)
  .component = "fe", "be", or "cn"
  .multilinePattern = regex pattern string for multi_line rule
  .multilineName = name for the multi_line rule
Usage: include "celerdatacluster.datadog.log.annotation" (dict "root" . "component" "fe" "multilinePattern" "..." "multilineName" "...")
*/}}
{{- define "celerdatacluster.datadog.log.annotation" -}}
{{- $root := .root -}}
{{- $component := .component -}}
{{- $multilinePattern := .multilinePattern -}}
{{- $multilineName := .multilineName -}}
{{- $logConfig := $root.Values.datadog.log.logConfig -}}
{{- $base := dict "service" "celerdata" "source" $component -}}
{{- if eq (kindOf $logConfig) "map" -}}
  {{- $base = merge $base $logConfig -}}
{{- else if eq (kindOf $logConfig) "string" -}}
  {{- if ne (trimAll " {}" $logConfig) "" -}}
    {{- $extra := fromJson $logConfig -}}
    {{- $base = merge $base $extra -}}
  {{- end -}}
{{- end -}}
{{- if $root.Values.datadog.log.enableMultilineLogParsing -}}
  {{- if not (hasKey $base "log_processing_rules") -}}
    {{- $rule := dict "name" $multilineName "pattern" $multilinePattern "type" "multi_line" -}}
    {{- $_ := set $base "log_processing_rules" (list $rule) -}}
  {{- end -}}
{{- end -}}
{{- list $base | toJson | squote -}}
{{- end -}}
