{{ valueIf (dict "key" "type" "value" (.GetSpec.GetType | toString)) }}
{{ valueIf (dict "key" "injectionTemplate" "value" .Properties.InjectionTemplate) }}
{{ valueIf (dict "key" "revision" "value" .Properties.Revision) }}
{{ valueIf (dict "key" "runAsRoot" "value" .GetSpec.GetRunAsRoot) }}

{{ define "global" }}
{{ valueIf (dict "key" "imagePullPolicy" "value" (default $.Properties.GetIstioControlPlane.GetSpec.GetContainerImageConfiguration.GetImagePullPolicy .GetSpec.GetDeployment.GetImagePullPolicy)) }}
{{ toYamlIf (dict "key" "imagePullSecrets" "value" (default $.Properties.GetIstioControlPlane.GetSpec.GetContainerImageConfiguration.GetImagePullSecrets .GetSpec.GetDeployment.GetImagePullSecrets)) }}
{{ end }}

{{- $x := (include "global" .) | reformatYaml }}
{{- if ne $x "" }}
global:
{{ $x | indent 2 }}
{{- end }}

deployment:
  name: {{ .Name | quote }}
{{ valueIf (dict "key" "enablePrometheusMerge" "value" .Properties.EnablePrometheusMerge) | indent 2 }}
{{- with .GetSpec.GetDeployment }}
{{ toYamlIf (dict "value" .GetDeploymentStrategy "key" "deploymentStrategy") | indent 2 }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") | indent 2 }}
{{ toYamlIf (dict "value" .GetEnv "key" "env") | indent 2 }}
{{ toYamlIf (dict "value" .GetAffinity "key" "affinity") | indent 2 }}
{{ toYamlIf (dict "value" .GetNodeSelector "key" "nodeSelector") | indent 2 }}
{{ valueIf (dict "value" .GetPriorityClassName "key" "priorityClassName") | indent 2 }}
{{ toYamlIf (dict "value" .GetReplicas "key" "replicas") | indent 2 }}
{{ toYamlIf (dict "value" .GetResources "key" "resources") | indent 2 }}
{{ toYamlIf (dict "value" .GetSecurityContext "key" "securityContext") | indent 2 }}
{{ toYamlIf (dict "value" .GetTolerations "key" "tolerations") | indent 2 }}
{{ toYamlIf (dict "value" .GetVolumeMounts "key" "volumeMounts") | indent 2 }}
{{ toYamlIf (dict "value" .GetVolumes "key" "volumes") | indent 2 }}
{{ toYamlIf (dict "value" .GetPodDisruptionBudget "key" "podDisruptionBudget") | indent 2 }}
{{ toYamlIf (dict "value" .GetPodMetadata "key" "podMetadata") | indent 2 }}
{{ valueIf (dict "value" .GetImage "key" "image") | indent 2 }}
{{- end }}

{{- with .GetSpec.GetService }}
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

{{- if $.Properties.GenerateExternalService }}
{{- with $.Status.GetGatewayAddress }}
externalService:
{{ toYamlIf (dict "value" . "key" "addresses") | indent 2 }}
{{- end }}
{{- end }}
