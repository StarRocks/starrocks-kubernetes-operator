{{/*
The default service account name to use for the operator.
*/}}
{{- define "operator.serviceAccountName" -}}
{{- default "starrocks" .Values.global.rbac.serviceAccountName }}
{{- end }}

{{- define "operator.namespace" -}}
{{- default .Release.Namespace .Values.starrocksOperator.namespaceOverride }}
{{- end }}

{{- define "kube-starrocks.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end }}
