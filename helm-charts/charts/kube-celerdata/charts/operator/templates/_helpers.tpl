{{/*
The default service account name to use for the operator.
*/}}
{{- define "operator.serviceAccountName" -}}
{{- default "celerdata" .Values.global.rbac.serviceAccount.name }}
{{- end }}

{{- define "operator.namespace" -}}
{{- default .Release.Namespace .Values.celerDataOperator.namespaceOverride }}
{{- end }}

{{- define "operator.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end }}
