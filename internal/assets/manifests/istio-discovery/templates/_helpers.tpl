{{- define "revision" -}}
{{- default "default" (.Values.revision | replace "." "-") -}}
{{- end -}}

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

{{- define "replicas" }}
{{- $count := default 1 .Values.pilot.replicas.count | int }}
{{- $min := default 1 .Values.pilot.replicas.min | int }}
{{- $max := default 5 .Values.pilot.replicas.max | int }}
{{- $cpuUtilization := default 80 .Values.pilot.replicas.targetCPUUtilizationPercentage | int }}
{{- $autoscalingEnabled := false }}
{{- if and .Values.pilot.replicas.count (not .Values.pilot.replicas.min) (not .Values.pilot.replicas.max) }}
{{- $min = 0 }}
{{- end }}
{{- if and $min (gt $min $count) }}
{{- $count = $min }}
{{- end }}
{{- if and (gt $min 0) (gt $max $min) }}
{{- $autoscalingEnabled = true }}
{{- end }}
{{- if and $autoscalingEnabled (gt $count $max) }}
{{- $count = $max }}
{{- end }}
autoscalingEnabled: {{ $autoscalingEnabled }}
count: {{ $count }}
min: {{ $min }}
max: {{ $max }}
targetCPUUtilizationPercentage: {{ $cpuUtilization }}
{{- end }}
