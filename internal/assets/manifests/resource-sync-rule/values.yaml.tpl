{{ valueIf (dict "key" "revision" "value" .Name) }}
{{- if .GetSpec.GetMode }}
mode: {{ .GetSpec.GetMode | toString }}
{{- end }}
{{ valueIf (dict "key" "distribution" "value" .GetSpec.GetDistribution) }}
