type: {{ .GetSpec.GetType }}

# The injection template to use for the gateway. If not set, no injection will be performed.
injectionTemplate: {{ .Properties.InjectionTemplate | quote }}

# Revision is set as 'version' label and part of the resource names when installing multiple control planes.
revision: {{ .Properties.Revision | quote }}

{{- if .GetSpec.GetRunAsRoot }}
runAsRoot: {{ .GetSpec.GetRunAsRoot }}
{{- end }}

deployment:
  name: {{ .Name | quote }}
  enablePrometheusMerge: {{ .Properties.EnablePrometheusMerge }}
{{- with .GetSpec.GetDeployment }}
{{ toYamlIf (dict "value" .GetDeploymentStrategy "key" "deploymentStrategy") | indent 2 }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") | indent 2 }}
{{ toYamlIf (dict "value" .GetEnv "key" "env") | indent 2 }}
  imagePullPolicy: {{ .GetImagePullPolicy | quote }}
{{ toYamlIf (dict "value" .GetImagePullSecrets "key" "imagePullSecrets") | indent 2 }}
{{ toYamlIf (dict "value" .GetAffinity "key" "affinity") | indent 2 }}
{{ toYamlIf (dict "value" .GetNodeSelector "key" "nodeSelector") | indent 2 }}
  priorityClassName: {{ .GetPriorityClassName | quote }}
{{ toYamlIf (dict "value" .GetReplicas "key" "replicas") | indent 2 }}
{{ toYamlIf (dict "value" .GetResources "key" "resources") | indent 2 }}
{{ toYamlIf (dict "value" .GetSecurityContext "key" "securityContext") | indent 2 }}
{{ toYamlIf (dict "value" .GetTolerations "key" "tolerations") | indent 2 }}
{{ toYamlIf (dict "value" .GetVolumeMounts "key" "volumeMounts") | indent 2 }}
{{ toYamlIf (dict "value" .GetVolumes "key" "volumes") | indent 2 }}
{{ toYamlIf (dict "value" .GetPodDisruptionBudget "key" "podDisruptionBudget") | indent 2 }}
{{ toYamlIf (dict "value" .GetPodMetadata "key" "podMetadata") | indent 2 }}
{{- end }}

{{- with .GetSpec.GetService }}
service:
  type: {{ .GetType }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") | indent 2 }}
{{ toYamlIf (dict "value" .GetPorts "key" "ports") | indent 2 }}
{{ toYamlIf (dict "value" .GetSelector "key" "selector") | indent 2 }}
  clusterIP: {{ .GetClusterIP | quote }}
{{ toYamlIf (dict "value" .GetExternalIPs "key" "externalIPs") | indent 2 }}
  sessionAffinity: {{ .GetSessionAffinity | quote }}
  loadBalancerIP: {{ .GetLoadBalancerIP | quote }}
{{ toYamlIf (dict "value" .GetLoadBalancerSourceRanges "key" "loadBalancerSourceRanges") | indent 2 }}
  externalName: {{ .GetExternalName | quote }}
  externalTrafficPolicy: {{ .GetExternalTrafficPolicy | quote }}
  healthCheckNodePort: {{ .GetHealthCheckNodePort }}
  publishNotReadyAddresses: {{ .GetPublishNotReadyAddresses }}
{{ toYamlIf (dict "value" .GetSessionAffinityConfig "key" "sessionAffinityConfig") | indent 2 }}
  ipFamily: {{ .GetIpFamily | quote }}
{{- end }}
