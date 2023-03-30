{{- define "taint" }}
{{ valueIf (dict "key" "enabled" "value" .GetEnabled ) }}
{{ valueIf (dict "key" "image" "value" .GetContainer.GetImage) }}
{{ toYamlIf (dict "value" .GetContainer.GetResources "key" "resources") }}
{{ toYamlIf (dict "value" .GetContainer.GetEnv "key" "env") }}
{{ toYamlIf (dict "value" .GetContainer.GetVolumeMounts "key" "volumeMounts") }}
{{ toYamlIf (dict "value" .GetContainer.GetSecurityContext "key" "securityContext") }}
{{- end }}

{{- define "repair" }}
{{ valueIf (dict "key" "enabled" "value" .GetEnabled ) }}
{{ valueIf (dict "key" "labelPods" "value" .GetLabelPods ) }}
{{ valueIf (dict "key" "deletePods" "value" .GetDeletePods ) }}
{{ valueIf (dict "key" "initContainerName" "value" .GetInitContainerName ) }}
{{ valueIf (dict "key" "brokenPodLabelKey" "value" .GetBrokenPodLabelKey ) }}
{{ valueIf (dict "key" "brokenPodLabelValue" "value" .GetBrokenPodLabelValue ) }}
{{- end }}

{{- define "cni" }}
{{- with .GetSpec.GetProxyInit.GetCni }}

{{ valueIf (dict "key" "enabled" "value" .GetEnabled ) }}
{{ valueIf (dict "key" "chained" "value" .GetChained ) }}
{{ valueIf (dict "key" "cniBinDir" "value" .GetBinDir ) }}
{{ valueIf (dict "key" "cniConfDir" "value" .GetConfDir ) }}
{{ valueIf (dict "key" "cniConfFileName" "value" .GetConfFileName ) }}
{{ valueIf (dict "key" "logLevel" "value" .GetLogLevel ) }}
{{ valueIf (dict "key" "psp_cluster_role" "value" .GetPspClusterRoleName ) }}
{{ toYamlIf (dict "value" .GetExcludeNamespaces "key" "excludeNamespaces") }}
{{ toYamlIf (dict "value" .GetIncludeNamespaces "key" "includeNamespaces") }}
{{ toYamlIf (dict "value" .GetResourceQuotas "key" "resourceQuotas") }}
{{ with .GetDaemonset }}
{{ valueIf (dict "key" "image" "value" .GetImage ) }}
{{ toYamlIf (dict "value" .GetMetadata "key" "metadata") }}
{{ toYamlIf (dict "value" .GetPodMetadata "key" "podMetadata") }}
{{ toYamlIf (dict "value" .GetDeploymentStrategy "key" "deploymentStrategy") }}
{{ toYamlIf (dict "value" .GetEnv "key" "env") }}
{{ toYamlIf (dict "value" .GetNodeSelector "key" "nodeSelector") }}
{{ toYamlIf (dict "value" .GetAffinity "key" "affinity") }}
{{ toYamlIf (dict "value" .GetTolerations "key" "tolerations") }}
{{ toYamlIf (dict "value" .GetVolumes "key" "volumes") }}
{{ toYamlIf (dict "value" .GetVolumeMounts "key" "volumeMounts") }}
{{ toYamlIf (dict "value" .GetResources "key" "resources") }}
{{ toYamlIf (dict "value" .GetSecurityContext "key" "securityContext") }}
{{ valueIf (dict "key" "priorityClassName" "value" .GetPriorityClassName ) }}
{{ end }}

{{- $x := (include "taint" .GetTaint) | reformatYaml }}
{{- if ne $x "" }}
taint:
{{ $x | indent 2 }}
{{- end }}

{{- $x := (include "repair" .GetRepair) | reformatYaml }}
{{- if ne $x "" }}
repair:
{{ $x | indent 2 }}
{{- end }}

{{- end }}
{{- end }}

{{- define "global" }}
{{ valueIf (dict "key" "hub" "value" .GetSpec.GetContainerImageConfiguration.GetHub) }}
{{ valueIf (dict "key" "tag" "value" .GetSpec.GetContainerImageConfiguration.GetTag) }}
{{ valueIf (dict "key" "tag" "value" .GetSpec.GetContainerImageConfiguration.GetTag) }}
{{- with .GetSpec.GetProxyInit.GetCni }}
{{ valueIf (dict "key" "imagePullPolicy" "value" (default $.GetSpec.GetContainerImageConfiguration.GetImagePullPolicy .GetDaemonset.GetImagePullPolicy) ) }}
{{ toYamlIf (dict "value" (default $.GetSpec.GetContainerImageConfiguration.GetImagePullSecrets .GetDaemonset.GetImagePullSecrets) "key" "imagePullSecrets") }}
{{ end }}
  ambient:
{{ valueIf (dict "key" "enabled" "value" .GetSpec.GetAmbientTopology) }}
{{ end }}

{{- $x := (include "cni" .) | reformatYaml }}
{{- if ne $x "" }}
cni:
{{ $x | indent 2 }}
{{- end }}

{{- $x := (include "global" .) | reformatYaml }}
{{- if ne $x "" }}
global:
{{ $x | indent 2 }}
{{- end }}

{{ valueIf (dict "key" "revision" "value" .Name) }}
