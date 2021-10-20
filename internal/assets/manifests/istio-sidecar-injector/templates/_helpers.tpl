{{- define "revision" -}}
{{- default "default" (.Values.revision | replace "." "-") -}}
{{- end -}}

{{- define "name-with-revision-by-distribution" -}}
{{- if eq .context.Values.global.distribution "cisco" -}}
{{ include "name-with-revision" ( dict "name" .name "context" .context) }}
{{- else -}}
{{ .name }}{{ if not (eq .context.Values.revision "") }}-{{ .context.Values.revision }}.{{ .context.Release.Namespace }}{{- end }}
{{- end -}}
{{- end -}}

{{- define "name-with-namespaced-revision-by-distribution" -}}
{{- if eq .context.Values.global.distribution "cisco" -}}
{{ include "name-with-namespaced-revision" ( dict "name" .name "context" .context) }}
{{- else -}}
{{ .name }}{{ if not (eq .context.Values.revision "") }}-{{ .context.Values.revision }}.{{ .context.Release.Namespace }}{{- end }}
{{- end -}}
{{- end -}}

{{- define "serviceHostnames" }}
{{- $servicename := include "name-with-revision" ( dict "name" "istio-sidecar-injector" "context" $) -}}
{{ $servicename }}.{{ .Release.Namespace }},{{ $servicename }}.{{ .Release.Namespace }}.svc,{{ $servicename }}.{{ .Release.Namespace }}.svc.{{ .Values.global.clusterDomain }}
{{- end }}

{{- define "generic.labels" }}
release: {{ .Release.Name }}
istio: "sidecar-injector"
app: "istio-sidecar-injector"
{{- if .Values.revision }}
istio.io/rev: {{ include "namespaced-revision" . }}
{{- end }}
{{- end }}

{{- define "labels" }}
{{- include "toYamlIf" (dict "value" .labels) }}
{{- end }}

{{- define "pod.labels" }}
{{- include "labels" (dict "context" . "labels" .Values.deployment.podMetadata.labels) }}
{{- include "generic.labels" . }}
{{- end }}

{{- define "namespaced-revision" -}}
{{- $revision := (include "revision" .) -}}
{{- if eq $revision "default" -}}
{{- printf "%s" $revision -}}
{{- else -}}
{{- printf "%s.%s" $revision .Release.Namespace -}}
{{- end -}}
{{- end -}}

{{- define "name-with-revision" -}}
{{- if .context.Values.revision -}}
{{- printf "%s-%s" .name (include "revision" .context) -}}
{{- else -}}
{{- .name -}}
{{- end -}}
{{- end -}}

{{- define "name-with-namespaced-revision" -}}
{{- if .context.Values.revision -}}
{{- printf "%s-%s" (include "name-with-revision" .) .context.Release.Namespace -}}
{{- else -}}
{{- .name -}}
{{- end -}}
{{- end -}}

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

{{- define "dockerImage" }}
{{- if contains "/" .image }}
image: "{{ .image }}"
{{- else }}
image: "{{ .hub }}/{{ .image }}:{{ .tag }}"
{{- end }}
{{- end }}
