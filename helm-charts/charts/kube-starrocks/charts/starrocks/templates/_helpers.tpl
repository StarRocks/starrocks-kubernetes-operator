{{/*
Common labels
*/}}
{{- define "starrockscluster.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
initpassword secret name
*/}}

{{- define "starrockscluster.initpassword.secret.name" -}}
{{ default (print (include "starrockscluster.name" .) "-credential") .Values.initPassword.passwordSecret }}
{{- end }}

{{/*
starrockscluster
*/}}

{{- define "starrockscluster.name" -}}
{{ default (default .Chart.Name .Values.nameOverride) .Values.starrocksCluster.name }}
{{- end }}

{{- define "starrockscluster.namespace" -}}
{{ default .Release.Namespace .Values.starrocksCluster.namespace }}
{{- end }}

{{- define "starrockscluster.fe.name" -}}
{{- print (include "starrockscluster.name" .) "-fe" }}
{{- end }}

{{- define "starrockscluster.cn.name" -}}
{{- print (include "starrockscluster.name" .) "-cn" }}
{{- end }}

{{- define "starrockscluster.be.name" -}}
{{- print (include "starrockscluster.name" .) "-be" }}
{{- end }}

{{- define "starrockscluster.be.configmap.name" -}}
{{- print (include "starrockscluster.be.name" .) "-cm" }}
{{- end }}

{{- define "starrockscluster.fe.configmap.name" -}}
{{- print (include "starrockscluster.fe.name" .) "-cm" }}
{{- end }}

{{- define "starrockscluster.cn.configmap.name" -}}
{{- print (include "starrockscluster.cn.name" .) "-cm" }}
{{- end }}

{{- define "starrockscluster.fe.config" -}}
fe.conf: |
{{- if .Values.starrocksFESpec.config }}
{{ .Values.starrocksFESpec.config | indent 2 }}
{{- end }}
{{- end }}

{{- define "starrockscluster.cn.config" -}}
cn.conf: |
{{- if .Values.starrocksCnSpec.config | indent 2 }}
{{ .Values.starrocksCnSpec.config | indent 2 }}
{{- end }}
{{- end }}

{{- define "starrocksclster.be.config" -}}
be.conf: |
{{- if .Values.starrocksBeSpec.config | indent 2 }}
{{ .Values.starrocksBeSpec.config | indent 2 }}
{{- end }}
{{- end }}

{{- define "starrockscluster.fe.meta.suffix" -}}
{{- print "-meta" }}
{{- end }}

{{- define "starrockscluster.fe.meta.path" -}}
{{- print "/opt/starrocks/fe/meta" }}
{{- end }}

{{- define "starrockscluster.fe.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "starrockscluster.fe.log.path" -}}
{{- print "/opt/starrocks/fe/log" }}
{{- end }}

{{- define "starrockscluster.be.data.suffix" -}}
{{- print "-data" }}
{{- end }}

{{- define "starrockscluster.be.data.path" -}}
{{- print "/opt/starrocks/be/storage" }}
{{- end }}

{{- define "starrockscluster.be.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "starrockscluster.be.log.path" -}}
{{- print "/opt/starrocks/be/log" }}
{{- end }}

{{- define "starrockscluster.cn.data.suffix" -}}
{{- print "-data" }}
{{- end }}

{{- define "starrockscluster.cn.data.path" -}}
{{- print "/opt/starrocks/cn/storage" }}
{{- end }}

{{- define "starrockscluster.cn.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "starrockscluster.cn.log.path" -}}
{{- print "/opt/starrocks/cn/log" }}
{{- end }}

{{/*
starrockscluster.fe.config.hash is used to calculate the hash value of the fe.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "starrockscluster.fe.config.hash" }}
  {{- if .Values.starrocksFESpec.config }}
    {{- $hash := toJson .Values.starrocksFESpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}


{{/*
starrockscluster.be.config.hash is used to calculate the hash value of the be.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "starrockscluster.be.config.hash" }}
  {{- if .Values.starrocksBeSpec.config }}
    {{- $hash := toJson .Values.starrocksBeSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{/*
starrockscluster.cn.config.hash is used to calculate the hash value of the cn.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "starrockscluster.cn.config.hash" }}
  {{- if .Values.starrocksCnSpec.config }}
    {{- $hash := toJson .Values.starrocksCnSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "starrockscluster.fe.http.port" -}}
{{- $config := index .Values.starrocksFESpec.config  -}}
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

{{- define "starrockscluster.be.webserver.port" -}}
{{- include "starrockscluster.webserver.port" .Values.starrocksBeSpec }}
{{- end }}

{{- define "starrockscluster.cn.webserver.port" -}}
{{- include "starrockscluster.webserver.port" .Values.starrocksCnSpec }}
{{- end }}

{{- define "starrockscluster.webserver.port" -}}
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
