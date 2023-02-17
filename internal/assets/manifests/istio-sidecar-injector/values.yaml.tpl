{{ valueIf (dict "key" "revision" "value" .Name) }}

{{ define "global" }}
{{ valueIf (dict "key" "distribution" "value" .GetSpec.GetDistribution) }}
{{ valueIf (dict "key" "hub" "value" .GetSpec.GetContainerImageConfiguration.GetHub) }}
{{ valueIf (dict "key" "tag" "value" .GetSpec.GetContainerImageConfiguration.GetTag) }}
{{ valueIf (dict "key" "imagePullPolicy" "value" (default .GetSpec.GetContainerImageConfiguration.GetImagePullPolicy .GetSpec.GetSidecarInjector.GetDeployment.GetImagePullPolicy)) }}
{{ toYamlIf (dict "key" "imagePullSecrets" "value" (default .GetSpec.GetContainerImageConfiguration.GetImagePullSecrets .GetSpec.GetSidecarInjector.GetDeployment.GetImagePullSecrets)) }}
{{- if .GetSpec.GetJwtPolicy }}
jwtPolicy: {{ .GetSpec.GetJwtPolicy | toString | lower | replace "_" "-" }}
{{- end }}
{{ end }}

{{- $x := (include "global" .) | reformatYaml }}
{{- if ne $x "" }}
global:
{{ $x | indent 2 }}
{{- end }}

{{ define "deployment" }}
{{ with .GetSpec.GetSidecarInjector.GetDeployment }}
{{ valueIf (dict "key" "image" "value" .GetImage) }}
{{ toYamlIf (dict "value" .GetDeploymentStrategy "key" "deploymentStrategy") }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") }}
{{ toYamlIf (dict "value" .GetEnv "key" "env") }}
{{ toYamlIf (dict "value" .GetAffinity "key" "affinity") }}
{{ toYamlIf (dict "value" .GetNodeSelector "key" "nodeSelector") }}
{{ valueIf (dict "value" .GetPriorityClassName "key" "priorityClassName") }}
{{ toYamlIf (dict "value" .GetReplicas "key" "replicas") }}
{{ toYamlIf (dict "value" .GetResources "key" "resources") }}
{{ toYamlIf (dict "value" .GetSecurityContext "key" "securityContext") }}
{{ toYamlIf (dict "value" .GetTolerations "key" "tolerations") }}
{{ toYamlIf (dict "value" .GetTopologySpreadConstraints "key" "topologySpreadConstraints") }}
{{ toYamlIf (dict "value" .GetVolumeMounts "key" "volumeMounts") }}
{{ toYamlIf (dict "value" .GetVolumes "key" "volumes") }}
{{ toYamlIf (dict "value" .GetPodDisruptionBudget "key" "podDisruptionBudget") }}
{{ toYamlIf (dict "value" .GetPodMetadata "key" "podMetadata") }}
{{ toYamlIf (dict "value" .GetLivenessProbe "key" "livenessProbe") }}
{{ toYamlIf (dict "value" .GetReadinessProbe "key" "readinessProbe") }}
{{ end }}
{{ end }}


{{- $x := (include "deployment" .) | reformatYaml }}
{{- if ne $x "" }}
deployment:
{{ $x | indent 2 }}
{{- end }}

{{- with .GetSpec.GetSidecarInjector.GetService }}
service:
  type: {{ .GetType }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") | indent 2 }}
{{ toYamlIf (dict "value" .GetPorts "key" "ports") | indent 2 }}
{{ toYamlIf (dict "value" .GetSelector "key" "selector") | indent 2 }}
{{ valueIf (dict "value" .GetClusterIP "key" "clusterIP") | indent 2 }}
{{ toYamlIf (dict "value" .GetExternalIPs "key" "externalIPs") | indent 2 }}
{{ valueIf (dict "value" .GetSessionAffinity "key" "sessionAffinity") | indent 2 }}
{{ valueIf (dict "value" .GetLoadBalancerIP "key" "loadBalancerIP") | indent 2 }}
{{ toYamlIf (dict "value" .GetLoadBalancerSourceRanges "key" "loadBalancerSourceRanges") | indent 2 }}
{{ valueIf (dict "value" .GetExternalName "key" "externalName") | indent 2 }}
{{ valueIf (dict "value" .GetExternalTrafficPolicy "key" "externalTrafficPolicy") | indent 2 }}
{{ valueIf (dict "value" .GetHealthCheckNodePort "key" "healthCheckNodePort") | indent 2 }}
{{ valueIf (dict "value" .GetPublishNotReadyAddresses "key" "publishNotReadyAddresses") | indent 2 }}
{{ toYamlIf (dict "value" .GetSessionAffinityConfig "key" "sessionAffinityConfig") | indent 2 }}
{{ valueIf (dict "value" .GetIpFamily "key" "ipFamily") | indent 2 }}
{{- end }}
