# template for pilot values
{{- define "pilot" }}
{{- if and .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMin .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMax }}
autoscaleEnabled: {{ and (gt (.GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMin | int) 0) (gt (.GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMax | int) (.GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMin | int)) }}
{{- end }}
{{ valueIf (dict "key" "autoscaleMin" "value" .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMin) }}
{{ valueIf (dict "key" "autoscaleMax" "value" .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetMax) }}
{{ valueIf (dict "key" "replicaCount" "value" .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetCount) }}
{{ valueIf (dict "key" "image" "value" .GetSpec.GetIstiod.GetDeployment.GetImage) }}
{{ valueIf (dict "key" "traceSampling" "value" .GetSpec.GetIstiod.GetTraceSampling) }}

{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetResources "key" "resources") }}
env:
  - name: INJECTION_WEBHOOK_CONFIG_NAME
    value: istio-sidecar-injector-cp-v111x-istio-system
  - name: ISTIOD_CUSTOM_HOST
    value: istiod-cp-v111x.istio-system.svc
  - name: PILOT_ENABLE_STATUS
{{ if .GetSpec.GetIstiod.GetEnableStatus }}
    value: "{{ .GetSpec.GetIstiod.GetEnableStatus }}"
{{ else }}
    value: "false"
{{ end }}
  - name: VALIDATION_WEBHOOK_CONFIG_NAME
    value: istiod-cp-v111x-istio-system
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetEnv) | indent 2 }}

{{ valueIf (dict "key" "enableProtocolSniffingForOutbound" "value" .GetSpec.GetIstiod.GetEnableProtocolSniffingOutbound) }}
{{ valueIf (dict "key" "enableProtocolSniffingForInbound" "value" .GetSpec.GetIstiod.GetEnableProtocolSniffingInbound) }}

{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetVolumes "key" "volumes") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetVolumeMounts "key" "volumeMounts") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetNodeSelector "key" "nodeSelector") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetAffinity "key" "affinity") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetTolerations "key" "tolerations") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetPodMetadata.GetAnnotations "key" "podAnnotations") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetPodMetadata.GetLabels "key" "podLabels") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetSecurityContext "key" "securityContext") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetPodSecurityContext "key" "podSecurityContext") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetMetadata.GetLabels "key" "deploymentLabels") }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetMetadata.GetAnnotations "key" "deploymentAnnotations") }}
{{- if .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetTargetCPUUtilizationPercentage }}
cpu:
  targetAverageUtilization: {{ .GetSpec.GetIstiod.GetDeployment.GetReplicas.GetTargetCPUUtilizationPercentage }}
{{- end }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetDeploymentStrategy "key" "deploymentStrategy") }}
{{- end }}

# template for proxy values
{{- define "proxy" }}
{{ valueIf (dict "key" "image" "value" .GetSpec.GetProxy.GetImage) }}
{{ valueIf (dict "key" "clusterDomain" "value" .GetSpec.GetProxy.GetClusterDomain) }}
{{ valueIf (dict "key" "componentLogLevel" "value" .GetSpec.GetProxy.GetComponentLogLevel) }}
{{ valueIf (dict "key" "enableCoreDump" "value" .GetSpec.GetProxy.GetEnableCoreDump) }}
{{ valueIf (dict "key" "excludeInboundPorts" "value" .GetSpec.GetProxy.GetExcludeInboundPorts) }}
{{ valueIf (dict "key" "includeIPRanges" "value" .GetSpec.GetProxy.GetIncludeIPRanges) }}
{{ valueIf (dict "key" "excludeIPRanges" "value" .GetSpec.GetProxy.GetExcludeIPRanges) }}
{{ valueIf (dict "key" "excludeOutboundPorts" "value" .GetSpec.GetProxy.GetExcludeOutboundPorts) }}
{{ if .GetSpec.GetProxy.GetLogLevel }}
logLevel: {{ .GetSpec.GetProxy.GetLogLevel | toString | lower }}
{{ end }}
{{ valueIf (dict "key" "privileged" "value" .GetSpec.GetProxy.GetPrivileged) }}
{{ valueIf (dict "key" "holdApplicationUntilProxyStarts" "value" .GetSpec.GetProxy.GetHoldApplicationUntilProxyStarts) }}
{{ toYamlIf (dict "value" .GetSpec.GetProxy.GetResources "key" "resources") }}
{{ toYamlIf (dict "value" .GetSpec.GetProxy.GetLifecycle "key" "lifecycle") }}
{{- end }}

# template for proxy init values
{{- define "proxyInit" }}
{{- valueIf (dict "key" "image" "value" .GetSpec.GetProxyInit.GetImage) }}
{{- toYamlIf (dict "value" .GetSpec.GetProxyInit.GetResources "key" "resources") }}
{{- end }}

{{ valueIf (dict "key" "revision" "value" .Name) }}

{{- $x := (include "pilot" .) | reformatYaml }}
{{- if ne $x "" }}
pilot:
{{ $x | indent 2 }}
{{- end }}

{{- if .GetSpec.GetHttpProxyEnvs }}
sidecarInjectorWebhook:
  # Supported only in Cisco provided istio-proxy images
{{ toYamlIf (dict "value" .GetSpec.GetHttpProxyEnvs "key" "httpProxyEnvs") | indent 2 }}
{{- end }}

{{- if or .GetSpec.GetTelemetryV2.GetEnabled .GetSpec.GetProxyWasm.GetEnabled }}
telemetry:
  v2:
    # For Null VM case now.
    # This also enables metadata exchange.
    {{ valueIf (dict "key" "enabled" "value" .GetSpec.GetTelemetryV2.GetEnabled) }}
    {{- if .GetSpec.GetProxyWasm.GetEnabled }}
    metadataExchange:
      # Indicates whether to enable WebAssembly runtime for metadata exchange filter.
      wasmEnabled: {{ .GetSpec.GetProxyWasm.GetEnabled }}
    # Indicate if prometheus stats filter is enabled or not
    prometheus:
      # Indicates whether to enable WebAssembly runtime for stats filter.
      wasmEnabled: {{ .GetSpec.GetProxyWasm.GetEnabled }}
    {{- end }}
{{- end }}

{{- define "global" }}
{{ valueIf (dict "key" "distribution" "value" .GetSpec.GetDistribution) }}
{{- if .GetSpec.GetMode }}
mode: {{ .GetSpec.GetMode | toString }}
{{- end }}
{{- if .GetSpec.GetIstiod.GetEnableAnalysis }}
istiod:
  enableAnalysis: {{ .GetSpec.GetIstiod.GetEnableAnalysis }}
{{- end }}
{{ toYamlIf (dict "value" .GetSpec.GetLogging "key" "logging")}}
{{ valueIf (dict "key" "oneNamespace" "value" .GetSpec.GetWatchOneNamespace) }}
{{ valueIf (dict "key" "imagePullPolicy" "value" .GetSpec.GetIstiod.GetDeployment.GetImagePullPolicy) }}
{{ valueIf (dict "key" "priorityClassName" "value" .GetSpec.GetIstiod.GetDeployment.GetPriorityClassName) }}
{{ toYamlIf (dict "value" .GetSpec.GetIstiod.GetDeployment.GetImagePullSecrets "key" "imagePullSecrets") }}

{{ valueIf (dict "key" "hub" "value" .GetSpec.GetContainerImageConfiguration.GetHub) }}
{{ valueIf (dict "key" "tag" "value" .GetSpec.GetContainerImageConfiguration.GetTag) }}
{{ valueIf (dict "key" "imagePullPolicy" "value" (default $.GetSpec.GetContainerImageConfiguration.GetImagePullPolicy .GetSpec.GetIstiod.GetDeployment.GetImagePullPolicy) ) }}
{{ toYamlIf (dict "value" (default $.GetSpec.GetContainerImageConfiguration.GetImagePullSecrets .GetSpec.GetIstiod.GetDeployment.GetImagePullSecrets) "key" "imagePullSecrets") }}

{{- if or .GetSpec.GetIstiod.GetDeployment.GetPodDisruptionBudget.GetMinAvailable .GetSpec.GetIstiod.GetDeployment.GetPodDisruptionBudget.GetMaxUnavailable }}
defaultPodDisruptionBudget:
  enabled: true
  {{ valueIf (dict "key" "minAvailable" "value" .GetSpec.GetIstiod.GetDeployment.GetPodDisruptionBudget.GetMinAvailable) }}
  {{ valueIf (dict "key" "maxUnavailable" "value" .GetSpec.GetIstiod.GetDeployment.GetPodDisruptionBudget.GetMaxUnavailable) }}
{{- end }}

{{- $x := (include "proxy" .) | reformatYaml }}
{{- if ne $x "" }}
proxy:
{{ $x | indent 2 }}
{{- end }}

{{- $x := (include "proxyInit" .) | reformatYaml }}
{{- if ne $x "" }}
proxy_init:
{{ $x | indent 2 }}
{{- end }}

  ##############################################################################################
  # The following values are found in other charts. To effectively modify these values, make   #
  # make sure they are consistent across your Istio helm charts                                #
  ##############################################################################################

{{ valueIf (dict "key" "caAddress" "value" .GetSpec.GetCaAddress) }}
{{ valueIf (dict "key" "externalIstiod" "value" .GetSpec.GetIstiod.GetExternalIstiod.GetEnabled) }}
{{- if .GetSpec.GetJwtPolicy }}
jwtPolicy: {{ .GetSpec.GetJwtPolicy | toString | lower | replace "_" "-" }}
{{- end }}
{{ valueIf (dict "key" "meshID" "value" .GetSpec.GetMeshID) }}
{{ valueIf (dict "key" "mountMtlsCerts" "value" .GetSpec.GetMountMtlsCerts) }}
{{ valueIf (dict "key" "network" "value" .GetSpec.GetNetworkName) }}
{{- if .GetSpec.GetIstiod.GetCertProvider }}
pilotCertProvider: {{ .GetSpec.GetIstiod.GetCertProvider | toString | lower }}
{{- end }}
{{- if .GetSpec.GetSds.GetTokenAudience }}
sds:
  token:
    aud: {{ .GetSpec.GetSds.GetTokenAudience }}
{{- end }}
{{ if .GetSpec.GetClusterID }}
multiCluster:
{{ valueIf (dict "key" "clusterName" "value" .GetSpec.GetClusterID) | indent 2 }}
{{ end }}
{{- end }}

{{- $x := (include "global" .) | reformatYaml }}
{{- if ne $x "" }}
global:
{{ $x | indent 2 }}
{{- end }}

{{- define "mesh" }}
rootNamespace: {{ .Namespace }}
{{- end }}

{{- $mesh := mergeOverwrite (.Properties.GetMesh.GetSpec.GetConfig | toJsonPB | fromYaml) (.GetSpec.GetMeshConfig | toJsonPB | fromYaml) (include "mesh" . | fromYaml) }}

meshConfig:
{{ toYaml $mesh | indent 2}}

{{- if or .GetSpec.GetProxyInit.GetCni.GetEnabled .GetSpec.GetProxyInit.GetCni.GetChained }}
istio_cni:
{{ valueIf (dict "key" "enabled" "value" .GetSpec.GetProxyInit.GetCni.GetEnabled) | indent 2 }}
{{ valueIf (dict "key" "chained" "value" .GetSpec.GetProxyInit.GetCni.GetChained) | indent 2 }}
{{- end }}
