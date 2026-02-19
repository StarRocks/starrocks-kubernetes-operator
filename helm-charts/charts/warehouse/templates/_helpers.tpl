{{- define "celerdatawarehouse.name" -}}
{{ .Release.Name }}
{{- end }}

{{- define "celerdatawarehouse.namespace" -}}
{{ .Release.Namespace }}
{{- end }}

{{- define "celerdatawarehouse.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "celerdatawarehouse.configmap.name" -}}
{{- print (include "celerdatawarehouse.name" .) "-cm" }}
{{- end }}

{{- define "celerdatawarehouse.config" -}}
cn.conf: |
{{- if .Values.spec.config }}
{{- .Values.spec.config | nindent 2 }}
{{- end }}
{{- end }}

{{/*
celerdatawarehouse.config.hash is used to calculate the hash value of the cn.conf, and due to the length limit, only
the first 8 digits are taken, which will be used as the annotations for pods.
*/}}
{{- define "celerdatawarehouse.config.hash" }}
  {{- if .Values.spec.config }}
    {{- $hash := toJson .Values.spec.config | sha256sum | trunc 8 }}
    {{- printf "%s" $hash }}
  {{- else }}
    {{- printf "no-config" }}
  {{- end }}
{{- end }}

{{- define "celerdatawarehouse.webserver.port" -}}
{{- include "celerdatawarehouse.get.webserver.port" .Values.spec }}
{{- end }}

{{- define "celerdatawarehouse.get.webserver.port" -}}
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

{{- define "celerdatacluster.cn.data.suffix" -}}
{{- print "-data" }}
{{- end }}

{{- define "celerdatacluster.cn.data.path" -}}
{{- print "/opt/starrocks/cn/storage" }}
{{- end }}

{{- define "celerdatacluster.cn.log.suffix" -}}
{{- print "-log" }}
{{- end }}

{{- define "celerdatacluster.cn.log.path" -}}
{{- print "/opt/starrocks/cn/log" }}
{{- end }}
