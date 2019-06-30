/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sidecarinjector

import (
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/gateways"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, util.MergeLabels(sidecarInjectorLabels, labelSelector), r.Config),
		Data: map[string]string{
			"config": r.siConfig(),
			"values": r.getValues(),
		},
	}
}

func (r *Reconciler) getValues() string {
	values := map[string]interface{}{
		"sidecarInjectorWebhook": map[string]interface{}{
			"rewriteAppHTTPProbe": r.Config.Spec.SidecarInjector.RewriteAppHTTPProbe,
		},
		"global": map[string]interface{}{
			"trustDomain":     "cluster.local",
			"imagePullPolicy": r.Config.Spec.ImagePullPolicy,
			"network":         r.Config.Spec.GetNetworkName(),
			"podDNSSearchNamespaces": []string{
				"global",
				"total",
				"{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global",
			},
			"proxy_init": map[string]interface{}{
				"image": r.Config.Spec.ProxyInit.Image,
			},
			"tracer": map[string]interface{}{
				"lightstep": map[string]interface{}{
					"CACertPath": r.Config.Spec.Tracing.Lightstep.CacertPath,
				},
			},
			"sds": map[string]interface{}{
				"customTokenDirectory": r.Config.Spec.SDS.CustomTokenDirectory,
				"useTrustworthyJwt":    r.Config.Spec.SDS.UseTrustworthyJwt,
				"enabled":              r.Config.Spec.SDS.Enabled,
			},
			"proxy": map[string]interface{}{
				"image":                        r.Config.Spec.Proxy.Image,
				"statusPort":                   15020,
				"tracer":                       r.Config.Spec.Tracing.Tracer,
				"clusterDomain":                "cluster.local",
				"logLevel":                     r.Config.Spec.Proxy.LogLevel,
				"componentLogLevel":            r.Config.Spec.Proxy.ComponentLogLevel,
				"dnsRefreshRate":               r.Config.Spec.Proxy.DNSRefreshRate,
				"enableCoreDump":               r.Config.Spec.Proxy.EnableCoreDump,
				"includeIPRanges":              r.Config.Spec.IncludeIPRanges,
				"excludeIPRanges":              r.Config.Spec.ExcludeIPRanges,
				"excludeInboundPorts":          "",
				"excludeOutboundPorts":         "",
				"privileged":                   r.Config.Spec.Proxy.Privileged,
				"readinessFailureThreshold":    30,
				"readinessInitialDelaySeconds": 1,
				"readinessPeriodSeconds":       2,
				"resources":                    r.Config.Spec.Proxy.Resources,
				"envoyMetricsService": map[string]interface{}{
					"enabled": false,
				},
				"envoyStatsd": map[string]interface{}{
					"enabled": false,
				},
			},
		},
	}

	j, _ := json.Marshal(&values)

	return string(j)
}

func (r *Reconciler) siConfig() string {
	autoInjection := "disabled"
	if util.PointerToBool(r.Config.Spec.SidecarInjector.AutoInjectionPolicyEnabled) {
		autoInjection = "enabled"
	}
	siConfig := map[string]string{
		"policy":   autoInjection,
		"template": r.templateConfig(),
	}
	marshaledConfig, _ := yaml.Marshal(siConfig)
	// this is a static config, so we don't have to deal with errors
	return string(marshaledConfig)

}

func (r *Reconciler) dnsConfig() string {
	if !util.PointerToBool(r.Config.Spec.MultiMesh) {
		return ""
	}
	return `
dnsConfig:
  searches:
  - global
  - "{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global"
`
}

func (r *Reconciler) templateConfig() string {
	return `rewriteAppHTTPProbe: {{ valueOrDefault .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe false }}
{{- if .Values.global.podDNSSearchNamespaces }}
dnsConfig:
  searches:
    {{- range .Values.global.podDNSSearchNamespaces }}
    - {{ render . }}
    {{- end }}
{{- end }}
initContainers:
{{ if ne (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "NONE" }}
` + r.proxyInitContainer() + `
{{ end -}}
` + r.coreDumpContainer() + `
containers:
- name: istio-proxy
  image: "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/proxyImage` + "`" + ` .Values.global.proxy.image }}"
  ports:
  - containerPort: 15090
    protocol: TCP
    name: http-envoy-prom
  args:
  - proxy
  - sidecar
  - --domain
  - $(POD_NAMESPACE).svc.{{ .Values.global.proxy.clusterDomain }}
  - --configPath
  - "{{ .ProxyConfig.ConfigPath }}"
  - --binaryPath
  - "{{ .ProxyConfig.BinaryPath }}"
  - --serviceCluster
  {{ if ne "" (index .ObjectMeta.Labels ` + "`" + `app` + "`" + `) -}}
  - "{{ index .ObjectMeta.Labels "app" }}.$(POD_NAMESPACE)"
  {{ else -}}
  - "{{ valueOrDefault .DeploymentMeta.Name ` + "`" + `istio-proxy` + "`" + ` }}.{{ valueOrDefault .DeploymentMeta.Namespace ` + "`" + `default` + "`" + ` }}"
  {{ end -}}
  - --drainDuration
  - "{{ formatDuration .ProxyConfig.DrainDuration }}"
  - --parentShutdownDuration
  - "{{ formatDuration .ProxyConfig.ParentShutdownDuration }}"
  - --discoveryAddress
  - "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/discoveryAddress` + "`" + ` .ProxyConfig.DiscoveryAddress }}"
` + r.tracingProxyArgs() + `
{{- if .Values.global.proxy.logLevel }}
  - --proxyLogLevel={{ .Values.global.proxy.logLevel }}
{{- end}}
{{- if .Values.global.proxy.componentLogLevel }}
  - --proxyComponentLogLevel={{ .Values.global.proxy.componentLogLevel }}
{{- end}}
  - --dnsRefreshRate
  - {{ .Values.global.proxy.dnsRefreshRate }}
  - --connectTimeout
  - "{{ formatDuration .ProxyConfig.ConnectTimeout }}"
  {{- if .Values.global.proxy.envoyStatsd.enabled }}
  - --statsdUdpAddress
  - "{{ .ProxyConfig.StatsdUdpAddress }}"
{{- end }}
{{- if .Values.global.proxy.envoyMetricsService.enabled }}
  - --envoyMetricsServiceAddress
  - "{{ .ProxyConfig.EnvoyMetricsServiceAddress }}"
{{- end }}
  - --proxyAdminPort
  - "{{ .ProxyConfig.ProxyAdminPort }}"
  {{ if gt .ProxyConfig.Concurrency 0 -}}
  - --concurrency
  - "{{ .ProxyConfig.Concurrency }}"
  {{ end -}}
  - --controlPlaneAuthPolicy
  - "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/controlPlaneAuthPolicy` + "`" + ` .ProxyConfig.ControlPlaneAuthPolicy }}"
{{- if (ne (annotation .ObjectMeta "status.sidecar.istio.io/port" .Values.global.proxy.statusPort) "0") }}
  - --statusPort
  - "{{ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort }}"
  - --applicationPorts
  - "{{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/applicationPorts` + "`" + ` (applicationPorts .Spec.Containers) }}"
{{- end }}
{{- if .Values.global.trustDomain }}
  - --trust-domain={{ .Values.global.trustDomain }}
{{- end }}
  env:
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: POD_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: INSTANCE_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP
{{ if eq .Values.global.proxy.tracer "datadog" }}
  - name: HOST_IP
    valueFrom:
      fieldRef:
        fieldPath: status.hostIP
{{ end }}
  - name: ISTIO_META_POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: ISTIO_META_CONFIG_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: ISTIO_META_INTERCEPTION_MODE
    value: "{{ or (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + `) .ProxyConfig.InterceptionMode.String }}"
  - name: ISTIO_META_INCLUDE_INBOUND_PORTS
    value: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (applicationPorts .Spec.Containers) }}"
  {{- if .Values.global.network }}
  - name: ISTIO_META_NETWORK
    value: "{{ .Values.global.network }}"
  {{- end }}
  {{ if .ObjectMeta.Annotations }}
  - name: ISTIO_METAJSON_ANNOTATIONS
    value: |
           {{ toJSON .ObjectMeta.Annotations }}
  {{ end }}
  {{ if .ObjectMeta.Labels }}
  - name: ISTIO_METAJSON_LABELS
    value: |
           {{ toJSON .ObjectMeta.Labels }}
  {{ end }}
  {{- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
  - name: ISTIO_BOOTSTRAP_OVERRIDE
    value: "/etc/istio/custom-bootstrap/custom_bootstrap.json"
  {{- end }}
  {{- if .Values.global.sds.customTokenDirectory }}
  - name: ISTIO_META_SDS_TOKEN_PATH
    value: "{{ .Values.global.sds.customTokenDirectory -}}/sdstoken"
  {{- end }}
  imagePullPolicy: {{ .Values.global.imagePullPolicy }}
  {{ if ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort) ` + "`" + `0` + "`" + ` }}
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: {{ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort }}
    initialDelaySeconds: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/initialDelaySeconds` + "`" + ` .Values.global.proxy.readinessInitialDelaySeconds }}
    periodSeconds: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/periodSeconds` + "`" + ` .Values.global.proxy.readinessPeriodSeconds }}
    failureThreshold: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/failureThreshold` + "`" + ` .Values.global.proxy.readinessFailureThreshold }}
  {{ end -}}
  securityContext:
    {{- if .Values.global.proxy.privileged }}
    privileged: true
    {{- end }}
    {{- if ne .Values.global.proxy.enableCoreDump true }}
    readOnlyRootFilesystem: true
    {{- end }}
    {{ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) ` + "`" + `TPROXY` + "`" + ` -}}
    capabilities:
      add:
      - NET_ADMIN
    runAsGroup: 1337
    {{ else -}}
    {{ if and .Values.global.sds.enabled .Values.global.sds.useTrustworthyJwt }}
    runAsGroup: 1337
    {{- end }}
    runAsUser: 1337
    {{- end }}
  resources:
  {{ if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) -}}
    requests:
      {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) -}}
      cpu: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + ` }}"
      {{ end}}
      {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) -}}
      memory: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + ` }}"
      {{ end }}
  {{ else -}}
{{- if .Values.global.proxy.resources }}
    {{ toYaml .Values.global.proxy.resources | indent 4 }}
{{- end }}
  {{  end -}}
  volumeMounts:
  {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
  - mountPath: /etc/istio/custom-bootstrap
    name: custom-bootstrap-volume
  {{- end }}
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  {{- if .Values.global.sds.enabled }}
  - mountPath: /var/run/sds
    name: sds-uds-path
    readOnly: true
  {{- if .Values.global.sds.useTrustworthyJwt }}
  - mountPath: /var/run/secrets/tokens
    name: istio-token
  {{- end }}
  {{- if .Values.global.sds.customTokenDirectory }}
  - mountPath: "{{ .Values.global.sds.customTokenDirectory -}}"
    name: custom-sds-token
    readOnly: true
  {{- end }}
  {{- else }}
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
  {{- end }}
  {{- if and (eq .Values.global.proxy.tracer "lightstep") .Values.global.tracer.lightstep.cacertPath }}
  - mountPath: {{ directory .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}
    name: lightstep-certs
    readOnly: true
  {{- end }}
    {{- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + ` }}
    {{ range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + `) }}
  - name: "{{  $index }}"
    {{ toYaml $value | indent 4 }}
    {{ end }}
    {{- end }}
volumes:
{{- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
- name: custom-bootstrap-volume
  configMap:
    name: {{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + ` "" }}
{{- end }}
- emptyDir:
    medium: Memory
  name: istio-envoy
{{- if .Values.global.sds.enabled }}
- name: sds-uds-path
  hostPath:
    path: /var/run/sds
{{- if .Values.global.sds.customTokenDirectory }}
- name: custom-sds-token
  secret:
    secretName: sdstokensecret
{{- end }}
{{- if .Values.global.sds.useTrustworthyJwt }}
- name: istio-token
  projected:
    sources:
    - serviceAccountToken:
        path: istio-token
        expirationSeconds: 43200
        audience: {{ .Values.global.trustDomain }}
{{- end }}
{{- else }}
- name: istio-certs
  secret:
    optional: true
    {{ if eq .Spec.ServiceAccountName "" }}
    secretName: istio.default
    {{ else -}}
    secretName: {{  printf "istio.%s" .Spec.ServiceAccountName }}
    {{  end -}}
  {{- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + ` }}
  {{range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + `) }}
- name: "{{ $index }}"
  {{ toYaml $value | indent 2 }}
  {{ end }}
  {{ end }}
{{- end }}
{{- if and (eq .Values.global.proxy.tracer "lightstep") .Values.global.tracer.lightstep.cacertPath }}
- name: lightstep-certs
  secret:
    optional: true
    secretName: lightstep.cacert
{{- end }}
`
}

func (r *Reconciler) tracingProxyArgs() string {
	if !util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		return ""
	}

	return `{{- if eq .Values.global.proxy.tracer "lightstep" }}
  - --lightstepAddress
  - "{{ .ProxyConfig.GetTracing.GetLightstep.GetAddress }}"
  - --lightstepAccessToken
  - "{{ .ProxyConfig.GetTracing.GetLightstep.GetAccessToken }}"
  - --lightstepSecure={{ .ProxyConfig.GetTracing.GetLightstep.GetSecure }}
  - --lightstepCacertPath
  - "{{ .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}"
{{- else if eq .Values.global.proxy.tracer "zipkin" }}
  - --zipkinAddress
  - "{{ .ProxyConfig.GetTracing.GetZipkin.GetAddress }}"
{{- else if eq .Values.global.proxy.tracer "datadog" }}
  - --datadogAgentAddress
  - "{{ .ProxyConfig.GetTracing.GetDatadog.GetAddress }}"
{{- end }}
`
}

func (r *Reconciler) coreDumpContainer() string {
	if !r.Config.Spec.Proxy.EnableCoreDump {
		return ""
	}

	coreDumpContainerYAML, err := yaml.Marshal([]apiv1.Container{
		gateways.GetCoreDumpContainer(r.Config),
	})
	if err != nil {
		return ""
	}

	return string(coreDumpContainerYAML)
}

func (r *Reconciler) proxyInitContainer() string {
	if util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled) {
		return ""
	}

	return `- name: istio-init
  image: "{{ .Values.global.proxy_init.image }}"
  args:
  - "-p"
  - "15001"
  - "-u"
  - 1337
  - "-m"
  - "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode }}"
  - "-i"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` .Values.global.proxy.includeIPRanges }}"
  - "-x"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` .Values.global.proxy.excludeIPRanges }}"
  - "-b"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (includeInboundPorts .Spec.Containers) }}"
  - "-d"
  - "{{ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` .Values.global.proxy.excludeInboundPorts) }}"
  {{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + `) (ne .Values.global.proxy.excludeOutboundPorts "") -}}
  - "-o"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + ` .Values.global.proxy.excludeOutboundPorts }}"
  {{ end -}}
  {{ if (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + `) -}}
  - "-k"
  - "{{ index .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + ` }}"
  {{ end -}}
  imagePullPolicy: "{{ .Values.global.imagePullPolicy }}"
` + r.getFormattedResources(r.Config.Spec.SidecarInjector.Init.Resources, 2) + `
  securityContext:
    runAsUser: 0
    runAsNonRoot: false
    capabilities:
      add:
      - NET_ADMIN
    {{- if .Values.global.proxy.privileged }}
    privileged: true
    {{- end }}
  restartPolicy: Always
  env:
  {{- if contains "*" (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` "") }}
  - name: INBOUND_CAPTURE_PORT
    value: 15006
  {{- end }}
  `
}

func (r *Reconciler) getFormattedResources(resources *apiv1.ResourceRequirements, indentSize int) string {
	type Resources struct {
		Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
	}

	requirements := templates.GetResourcesRequirementsOrDefault(
		resources,
		r.Config.Spec.DefaultResources,
	)

	requirementsYAML, err := yaml.Marshal(Resources{
		Resources: requirements,
	})
	if err != nil {
		return ""
	}

	return indentWithSpaces(string(requirementsYAML), indentSize)
}

func indentWithSpaces(v string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}
