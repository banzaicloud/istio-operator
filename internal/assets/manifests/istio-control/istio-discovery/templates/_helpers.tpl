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
