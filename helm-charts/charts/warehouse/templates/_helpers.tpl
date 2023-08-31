{{- define "starrockswarehouse.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end }}

{{- define "starrockswarehouse.namespace" -}}
{{ .Release.Namespace }}
{{- end }}

{{- define "starrockswarehouse.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "starrockswarehouse.configmap.name" -}}
{{- print (include "starrockswarehouse.name" .) "-cm" }}
{{- end }}

{{- define "starrockswarehouse.config" -}}
cn.conf: |
{{- if .Values.starrocksCnSpec.config | indent 2 }}
{{ .Values.starrocksCnSpec.config | indent 2 }}
{{- end }}
{{- end }}

{{/*
starrockswarehouse.config.hash is used to calculate the hash value of the cn.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "starrockswarehouse.config.hash" }}
  {{- if .Values.starrocksCnSpec.config }}
    {{- $hash := toJson .Values.starrocksCnSpec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "starrockswarehouse.webserver.port" -}}
{{- include "starrockswarehouse.get.webserver.port" .Values.starrocksCnSpec }}
{{- end }}

{{- define "starrockswarehouse.get.webserver.port" -}}
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

{{- define "starrockswarehouse.initpassword.secret.name" -}}
{{ default (print (include "starrockswarehouse.name" .) "-credential") .Values.initPassword.passwordSecret }}
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
