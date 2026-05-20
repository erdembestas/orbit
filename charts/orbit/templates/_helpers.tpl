{{- define "orbit.name" -}}
orbit
{{- end -}}

{{- define "orbit.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "orbit.commonLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ include "orbit.chart" . }}
{{- end -}}

{{- define "orbit.apiServiceAccountName" -}}
{{- default "orbit-api" .Values.api.serviceAccount.name -}}
{{- end -}}

{{- define "orbit.agentServiceAccountName" -}}
{{- default "orbit-agent" .Values.agent.serviceAccount.name -}}
{{- end -}}

{{- define "orbit.controllerServiceAccountName" -}}
{{- default "orbit-controller" .Values.controller.serviceAccount.name -}}
{{- end -}}

{{- define "orbit.agentReadonlyName" -}}
{{- printf "%s-%s-orbit-agent-readonly" .Release.Name .Release.Namespace -}}
{{- end -}}

{{- define "orbit.controllerReadonlyName" -}}
{{- printf "%s-%s-orbit-controller-readonly" .Release.Name .Release.Namespace -}}
{{- end -}}
