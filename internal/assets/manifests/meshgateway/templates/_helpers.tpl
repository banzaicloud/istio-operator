{{- define "labels" }}
{{- include "toYamlIf" (dict "value" (merge .labels (dict "gateway-name" .context.Values.deployment.name "gateway-type" .context.Values.type))) }}
{{- end }}

{{- define "generic.labels" }}
release: {{ .Release.Name }}
{{- if .Values.revision }}
istio.io/rev: {{ .Values.revision }}
{{- end }}
{{- end }}

{{- define "deployment.labels" }}
{{- include "labels" (dict "context" . "labels" .Values.deployment.metadata.labels) }}
{{- include "generic.labels" . }}
{{- end }}

{{- define "pod.labels" }}
{{- include "labels" (dict "context" . "labels" .Values.deployment.podMetadata.labels) }}
{{- include "deployment.labels" . }}
{{- end }}

{{- define "service.labels" }}
{{- include "labels" (dict "context" . "labels" .Values.service.metadata.labels) }}
{{- end }}

{{- define "toYamlIf" }}
{{- if .value }}
{{- if .key }}
{{ .key }}:
{{- end }}
{{- if gt (.indent | int) 0 }}
{{ .value | toYaml | indent .indent }}
{{- else }}
{{ .value | toYaml }}
{{- end }}
{{- end }}
{{- end }}

{{- define "revision" -}}
{{- .Values.revision | replace "." "-" -}}
{{- end -}}
