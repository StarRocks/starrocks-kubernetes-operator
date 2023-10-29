{{/*
The default service account name to use for the operator.
*/}}
{{- define "operator.serviceAccountName" -}}
{{- default "starrocks" .Values.global.rbac.serviceAccount.name }}
{{- end }}

{{- define "operator.namespace" -}}
{{- default .Release.Namespace .Values.starrocksOperator.namespaceOverride }}
{{- end }}

{{- define "operator.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end }}
