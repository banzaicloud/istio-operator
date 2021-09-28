{{ valueIf (dict "key" "revision" "value" .Name) }}
{{ with .GetSpec.GetMeshExpansion }}
{{ valueIf (dict "key" "exposeIstiod" "value" .GetIstiod.GetExpose) }}
{{ valueIf (dict "key" "exposeWebhook" "value" .GetWebhook.GetExpose) }}
{{ valueIf (dict "key" "exposeClusterServices" "value" .GetClusterServices.GetExpose) }}
{{ end }}
{{ with .GetSpec.GetMeshExpansion.GetGateway }}
{{ valueIf (dict "key" "runAsRoot" "value" .GetRunAsRoot) }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") }}
{{ toYamlIf (dict "value" .GetDeployment "key" "deployment") }}
{{ toYamlIf (dict "value" .GetService "key" "service") }}
{{ end }}
{{- if .GetSpec.GetMode }}
mode: {{ .GetSpec.GetMode | toString }}
{{- end }}
{{ valueIf (dict "key" "distribution" "value" .GetSpec.GetDistribution) }}
{{ valueIf (dict "key" "network" "value" .GetSpec.GetNetworkName) }}
