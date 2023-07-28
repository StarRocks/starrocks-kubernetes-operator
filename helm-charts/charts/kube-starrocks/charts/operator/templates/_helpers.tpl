{{- define "operator.serviceAccountName" -}}
{{- print "starrocks" }}
{{- end }}

{{- define "operator.namespace" -}}
{{- default .Release.Namespace .Values.starrocksOperator.namespaceOverride }}
{{- end }}

{{- define "kube-starrocks.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end }}