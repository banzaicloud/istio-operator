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
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/gateways"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) configMapInjector() runtime.Object {
	labels := util.MergeStringMaps(sidecarInjectorLabels, labelSelector)
	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		labels = nil
	}
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMetaWithRevision(configMapNameInjector, labels, r.Config),
		Data: map[string]string{
			"config": r.siConfig(),
			"values": r.getValues(),
		},
	}
}

func (r *Reconciler) getValues() string {
	podDNSSearchNamespaces := make([]string, 0)
	proxyInitContainerName := "istio-init"
	if util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled) {
		proxyInitContainerName = "istio-validation"
	}

	values := map[string]interface{}{
		"sidecarInjectorWebhook": map[string]interface{}{
			"rewriteAppHTTPProbe": util.PointerToBool(r.Config.Spec.SidecarInjector.RewriteAppHTTPProbe),
			"httpProxyEnvs": map[string]interface{}{
				"httpProxy":  r.Config.Spec.HTTPProxyEnvs.HTTPProxy,
				"httpsProxy": r.Config.Spec.HTTPProxyEnvs.HTTPSProxy,
				"noProxy":    r.Config.Spec.HTTPProxyEnvs.NoProxy,
			},
		},
		"global": map[string]interface{}{
			"revision": r.Config.NamespacedRevision(),
			"mtls": map[string]interface{}{
				"auto": util.PointerToBool(r.Config.Spec.AutoMTLS),
			},
			"istiod": map[string]interface{}{
				"enabled":          util.PointerToBool(r.Config.Spec.Istiod.Enabled),
				"caRootConfigName": r.Config.WithRevisionIf("istio-ca-root-cert", util.PointerToBool(r.Config.Spec.Istiod.MultiControlPlaneSupport)),
			},
			"caAddress":         r.Config.GetCAAddress(),
			"jwtPolicy":         r.Config.Spec.JWTPolicy,
			"pilotCertProvider": r.Config.Spec.Pilot.CertProvider,
			"trustDomain":       r.Config.Spec.TrustDomain,
			"imagePullPolicy":   r.Config.Spec.ImagePullPolicy,
			"imagePullSecrets": func() (names []string) {
				for _, name := range r.Config.Spec.ImagePullSecrets {
					names = append(names, name.Name)
				}
				return names
			}(),
			"network":                r.Config.Spec.NetworkName,
			"podDNSSearchNamespaces": podDNSSearchNamespaces,
			"proxy_init": map[string]interface{}{
				"cniEnabled":    util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled),
				"cniChained":    util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Chained),
				"containerName": proxyInitContainerName,
				"image":         r.Config.Spec.ProxyInit.Image,
			},
			"sds": map[string]interface{}{
				"customTokenDirectory": r.Config.Spec.SDS.CustomTokenDirectory,
				"enabled":              r.Config.Spec.SDS.Enabled,
				"token": map[string]interface{}{
					"aud": r.Config.Spec.SDS.TokenAudience,
				},
			},
			"multicluster": map[string]interface{}{
				"clusterName": r.Config.Spec.ClusterName,
			},
			"mountMtlsCerts": util.PointerToBool(r.Config.Spec.MountMtlsCerts),
			"meshID":         r.Config.Spec.MeshID,
			"proxy": map[string]interface{}{
				"image":                           r.Config.Spec.Proxy.Image,
				"statusPort":                      v1beta1.ProxyStatusPort,
				"tracer":                          r.Config.Spec.Tracing.Tracer,
				"clusterDomain":                   r.Config.Spec.Proxy.ClusterDomain,
				"logLevel":                        r.Config.Spec.Proxy.LogLevel,
				"componentLogLevel":               r.Config.Spec.Proxy.ComponentLogLevel,
				"dnsRefreshRate":                  r.Config.Spec.Proxy.DNSRefreshRate,
				"enableCoreDump":                  r.Config.Spec.Proxy.EnableCoreDump,
				"holdApplicationUntilProxyStarts": util.PointerToBool(r.Config.Spec.Proxy.HoldApplicationUntilProxyStarts),
				"includeIPRanges":                 r.Config.Spec.IncludeIPRanges,
				"excludeIPRanges":                 r.Config.Spec.ExcludeIPRanges,
				"excludeInboundPorts":             "",
				"excludeOutboundPorts":            "",
				"privileged":                      r.Config.Spec.Proxy.Privileged,
				"readinessFailureThreshold":       30,
				"readinessInitialDelaySeconds":    1,
				"readinessPeriodSeconds":          2,
				"resources":                       r.Config.Spec.Proxy.Resources,
				"envoyMetricsService": map[string]interface{}{
					"enabled": r.Config.Spec.Proxy.EnvoyMetricsService.Enabled,
				},
				"envoyAccessLogService": map[string]interface{}{
					"enabled": r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled,
				},
				"envoyStatsd": map[string]interface{}{
					"enabled": r.Config.Spec.Proxy.EnvoyStatsD.Enabled,
				},
				"lifecycle": r.Config.Spec.Proxy.Lifecycle,
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
	siConfig := map[string]interface{}{
		"policy":   autoInjection,
		"template": r.templateConfig(),

		"defaultTemplates": []string{"sidecar"},
		"templates":        r.templates(),
	}

	if len(r.Config.Spec.SidecarInjector.AlwaysInjectSelector) > 0 {
		siConfig["alwaysInjectSelector"] = r.Config.Spec.SidecarInjector.AlwaysInjectSelector
	}
	if len(r.Config.Spec.SidecarInjector.NeverInjectSelector) > 0 {
		siConfig["neverInjectSelector"] = r.Config.Spec.SidecarInjector.NeverInjectSelector
	}
	if len(r.Config.Spec.SidecarInjector.InjectedAnnotations) > 0 {
		siConfig["injectedAnnotations"] = r.Config.Spec.SidecarInjector.InjectedAnnotations
	}

	marshaledConfig, _ := yaml.Marshal(siConfig)
	// this is a static config, so we don't have to deal with errors
	return string(marshaledConfig)

}

func (r *Reconciler) templates() map[string]string {
	return map[string]string{"sidecar": `
{{- $containers := list }}
{{- range $index, $container := .Spec.Containers }}{{ if not (eq $container.Name "istio-proxy") }}{{ $containers = append $containers $container.Name }}{{end}}{{- end}}
metadata:
  labels:
    security.istio.io/tlsMode: {{ index .ObjectMeta.Labels ` + "`" + `security.istio.io/tlsMode` + "`" + ` | default "istio" | quote }}
    service.istio.io/canonical-name: {{ index .ObjectMeta.Labels ` + "`" + `service.istio.io/canonical-name` + "`" + ` | default (index .ObjectMeta.Labels ` + "`" + `app.kubernetes.io/name` + "`" + `) | default (index .ObjectMeta.Labels ` + "`" + `app` + "`" + `) | default .DeploymentMeta.Name | quote }}
    service.istio.io/canonical-revision: {{ index .ObjectMeta.Labels ` + "`" + `service.istio.io/canonical-revision` + "`" + ` | default (index .ObjectMeta.Labels ` + "`" + `app.kubernetes.io/version` + "`" + `) | default (index .ObjectMeta.Labels ` + "`" + `version` + "`" + `) | default "latest" | quote }}
    istio.io/rev: {{ .Revision | default "default" | quote }}
  annotations:
    {{- if eq (len $containers) 1 }}
    kubectl.kubernetes.io/default-logs-container: "{{ index $containers 0 }}"
    {{ end }}
{{ if .Values.global.proxy_init.cniEnabled -}}
    {{- if not .Values.global.proxy_init.cniChained }}
    k8s.v1.cni.cncf.io/networks: '{{ appendMultusNetwork (index .ObjectMeta.Annotations ` + "`" + `k8s.v1.cni.cncf.io/networks` + "`" + `) ` + "`" + `istio-cni` + "`" + ` }}'
    {{- end }}
    sidecar.istio.io/interceptionMode: "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode }}"
    {{ with annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` .Values.global.proxy.includeIPRanges }}
    traffic.sidecar.istio.io/includeOutboundIPRanges: "{{.}}"
    {{ end }}
    {{ with annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` .Values.global.proxy.excludeIPRanges }}
    traffic.sidecar.istio.io/excludeOutboundIPRanges: "{{.}}"
    {{ end }}
    traffic.sidecar.istio.io/includeInboundPorts: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` ` + "`" + `*` + "`" + ` }}"
    traffic.sidecar.istio.io/excludeInboundPorts: "{{ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` .Values.global.proxy.excludeInboundPorts) }}"
    {{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts` + "`" + `) (ne (valueOrDefault .Values.global.proxy.includeOutboundPorts "") "") }}
    traffic.sidecar.istio.io/includeOutboundPorts: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts` + "`" + ` .Values.global.proxy.includeOutboundPorts }}"
    {{- end }}
    {{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + `) (ne .Values.global.proxy.excludeOutboundPorts "") }}
    traffic.sidecar.istio.io/excludeOutboundPorts: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + ` .Values.global.proxy.excludeOutboundPorts }}"
    {{- end }}
    {{ with index .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + ` }}traffic.sidecar.istio.io/kubevirtInterfaces: "{{.}}",{{ end }}
{{- end }}
spec:
` + indentWithSpaces(r.templatePodSpec(2), 2) + `
`}
}

func (r *Reconciler) templatePodSpec(indentSize int) string {
	return `initContainers:
{{ if ne (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "NONE" }}
` + r.proxyInitContainer(indentSize) + `
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
  - --serviceCluster
  {{ if ne "" (index .ObjectMeta.Labels ` + "`" + `app` + "`" + `) -}}
  - "{{ index .ObjectMeta.Labels "app" }}.$(POD_NAMESPACE)"
  {{ else -}}
  - "{{ valueOrDefault .DeploymentMeta.Name ` + "`" + `istio-proxy` + "`" + ` }}.{{ valueOrDefault .DeploymentMeta.Namespace ` + "`" + `default` + "`" + ` }}"
  {{ end -}}
{{- if .Values.global.proxy.logLevel }}
  - --proxyLogLevel={{ .Values.global.proxy.logLevel }}
{{- end}}
{{- if .Values.global.proxy.componentLogLevel }}
  - --proxyComponentLogLevel={{ .Values.global.proxy.componentLogLevel }}
{{- end}}
{{- if .Values.global.istiod.enabled }}
  {{- if gt .ProxyConfig.Concurrency.GetValue 0 }}
  - --concurrency
  - "{{ .ProxyConfig.Concurrency.GetValue }}"
  {{- end -}}
{{- end }}
` + r.injectedAddtionalContainerArgs() + `
{{- if .Values.global.proxy.lifecycle }}
  lifecycle:
    {{ toYaml .Values.global.proxy.lifecycle | indent ` + fmt.Sprintf("%d", (4+indentSize)) + ` }}
{{- else if .Values.global.proxy.holdApplicationUntilProxyStarts}}
  lifecycle:
    postStart:
      exec:
        command:
        - pilot-agent
        - wait
{{- end }}
  env:
{{- if .Values.global.istiod.enabled }}
  - name: JWT_POLICY
    value: {{ .Values.global.jwtPolicy }}
  - name: PILOT_CERT_PROVIDER
    value: {{ .Values.global.pilotCertProvider }}
  - name: CA_ADDR
    value: {{ .Values.global.caAddress }}
{{- end }}
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
  - name: SERVICE_ACCOUNT
    valueFrom:
      fieldRef:
        fieldPath: spec.serviceAccountName
  {{- if .Values.global.mtls.auto }}
  - name: ISTIO_AUTO_MTLS_ENABLED
    value: "true"
  {{- end }}
  - name: HOST_IP
    valueFrom:
      fieldRef:
        fieldPath: status.hostIP
{{- if eq .Values.global.proxy.tracer "datadog" }}
{{- if isset .ObjectMeta.Annotations ` + "`" + `apm.datadoghq.com/env` + "`" + ` }}
{{- range $key, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `apm.datadoghq.com/env` + "`" + `) }}
  - name: {{ $key }}
    value: "{{ $value }}"
{{- end }}
{{- end }}
{{- end }}
  - name: CANONICAL_SERVICE
    valueFrom:
      fieldRef:
        fieldPath: metadata.labels['service.istio.io/canonical-name']
  - name: CANONICAL_REVISION
    valueFrom:
      fieldRef:
        fieldPath: metadata.labels['service.istio.io/canonical-revision']
  - name: PROXY_CONFIG
    value: |
           {{ protoToJSON .ProxyConfig }}
  - name: ISTIO_META_POD_PORTS
    value: |-
      [
      {{- $first := true }}
      {{- range $index1, $c := .Spec.Containers }}
        {{- range $index2, $p := $c.Ports }}
          {{- if (structToJSON $p) }}
          {{if not $first}},{{end}}{{ structToJSON $p }}
          {{- $first = false }}
          {{- end }}
        {{- end}}
      {{- end}}
      ]
  - name: ISTIO_META_APP_CONTAINERS
    value: "{{- range $index, $container := .Spec.Containers }}{{- if ne $index 0}},{{- end}}{{ $container.Name }}{{- end}}"
  - name: ISTIO_META_CLUSTER_ID
    value: "{{ valueOrDefault .Values.global.multicluster.clusterName ` + "`" + `Kubernetes` + "`" + `}}"
  - name: ISTIO_META_INTERCEPTION_MODE
    value: "{{ or (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + `) .ProxyConfig.InterceptionMode.String }}"
  {{- if .Values.global.network }}
  - name: ISTIO_META_NETWORK
    value: "{{ .Values.global.network }}"
  {{- end }}
  {{ if .ObjectMeta.Annotations }}
  - name: ISTIO_METAJSON_ANNOTATIONS
    value: |
           {{ toJSON .ObjectMeta.Annotations }}
  {{ end }}
  {{- if not .Values.global.istiod.enabled }}
  {{ if .ObjectMeta.Labels }}
  - name: ISTIO_METAJSON_LABELS
    value: |
           {{ toJSON .ObjectMeta.Labels }}
  {{ end }}
  {{- end }}
  {{- if .DeploymentMeta.Name }}
  - name: ISTIO_META_WORKLOAD_NAME
    value: "{{ .DeploymentMeta.Name }}"
  {{ end }}
  {{- if and .TypeMeta.APIVersion .DeploymentMeta.Name }}
  - name: ISTIO_META_OWNER
    value: kubernetes://apis/{{ .TypeMeta.APIVersion }}/namespaces/{{ valueOrDefault .DeploymentMeta.Namespace ` + "`" + `default` + "`" + ` }}/{{ toLower .TypeMeta.Kind}}s/{{ .DeploymentMeta.Name }}
   {{- end}}
  - name: ISTIO_META_REVISION
    value: "{{ .Values.global.revision }}"
  {{- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
  - name: ISTIO_BOOTSTRAP_OVERRIDE
    value: "/etc/istio/custom-bootstrap/custom_bootstrap.json"
  {{- end }}
  {{- if not .Values.global.istiod.enabled }}
  - name: SDS_ENABLED
    value: {{ $.Values.global.sds.enabled }}
  {{- if .Values.global.sds.customTokenDirectory }}
  - name: ISTIO_META_SDS_TOKEN_PATH
    value: "{{ .Values.global.sds.customTokenDirectory -}}/sdstoken"
  {{- end }}
  {{- end }}
  {{- if .Values.global.meshID }}
  - name: ISTIO_META_MESH_ID
    value: "{{ .Values.global.meshID }}"
  {{- else if .Values.global.trustDomain }}
  - name: ISTIO_META_MESH_ID
    value: "{{ .Values.global.trustDomain }}"
  {{- end }}
  - name: TRUST_DOMAIN
    value: "{{ .Values.global.trustDomain }}"
{{- if .Values.global.proxy.envoyAccessLogService.enabled }}
  - name: ISTIO_META_ALS_ENABLED
    value: "true"
{{- end }}
{{- range $key, $value := .ProxyConfig.ProxyMetadata }}
  - name: {{ $key }}
    value: "{{ $value }}"
{{- end }}
` + r.trustBundleManagerEnvVar() + `
` + r.injectedAddtionalEnvVars() + `
  imagePullPolicy: {{ .Values.global.imagePullPolicy }}
  {{ if ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` (valueOrDefault .Values.global.proxy.statusPort 0 )) ` + "`" + `0` + "`" + ` }}
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: 15021
    initialDelaySeconds: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/initialDelaySeconds` + "`" + ` .Values.global.proxy.readinessInitialDelaySeconds }}
    periodSeconds: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/periodSeconds` + "`" + ` .Values.global.proxy.readinessPeriodSeconds }}
    timeoutSeconds: 3
    failureThreshold: {{ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/failureThreshold` + "`" + ` .Values.global.proxy.readinessFailureThreshold }}
  {{ end -}}
  securityContext:
    allowPrivilegeEscalation: {{ .Values.global.proxy.privileged }}
    capabilities:
      {{ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) ` + "`" + `TPROXY` + "`" + ` -}}
      add:
      - NET_ADMIN
      {{- end }}
      drop:
      - ALL
    privileged: {{ .Values.global.proxy.privileged }}
    readOnlyRootFilesystem: {{ not .Values.global.proxy.enableCoreDump }}
    runAsGroup: 1337
    {{- if .Values.global.istiod.enabled }}
    fsGroup: 1337
    {{- end }}
    {{ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) ` + "`" + `TPROXY` + "`" + ` -}}
    runAsNonRoot: false
    runAsUser: 0
    {{- else }}
    runAsNonRoot: true
    runAsUser: 1337
    {{- end }}
  {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) }}
  resources:
    {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) }}
      requests:
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) -}}
        cpu: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + ` }}"
        {{ end }}
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) -}}
        memory: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + ` }}"
        {{ end }}
    {{- end }}
    {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) }}
      limits:
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) -}}
        cpu: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + ` }}"
        {{ end }}
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) -}}
        memory: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + ` }}"
        {{ end }}
    {{- end }}
  {{- else }}
` + r.getFormattedResources(r.Config.Spec.Proxy.Resources, 2) + `
  {{- end }}
  volumeMounts:
  {{- if .Values.global.istiod.enabled }}
  {{- if eq .Values.global.pilotCertProvider "istiod" }}
  - mountPath: /var/run/secrets/istio
    name: istiod-ca-cert
  {{- end }}
  {{- end }}
  - mountPath: /var/lib/istio/data
    name: istio-data
  {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
  - mountPath: /etc/istio/custom-bootstrap
    name: custom-bootstrap-volume
  {{- end }}
  # SDS channel between istioagent and Envoy
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  {{- if .Values.global.istiod.enabled }}
  {{- if eq .Values.global.jwtPolicy "third-party-jwt" }}
  - mountPath: /var/run/secrets/tokens
    name: istio-token
  {{- end }}
  {{- if .Values.global.mountMtlsCerts }}
  # Use the key and cert mounted to /etc/certs/ for the in-cluster mTLS communications.
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
  {{- end }}
  - name: istio-podinfo
    mountPath: /etc/istio/pod
  {{- else }}
  {{- if .Values.global.sds.enabled }}
  - mountPath: /var/run/sds
    name: sds-uds-path
    readOnly: true
  - mountPath: /var/run/secrets/tokens
    name: istio-token
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
  {{- end }}
  {{- if and (eq .Values.global.proxy.tracer "lightstep") .ProxyConfig.GetTracing.GetTlsSettings }}
  - mountPath: {{ directory .ProxyConfig.GetTracing.GetTlsSettings.GetCaCertificates }}
    name: lightstep-certs
    readOnly: true
  {{- end }}
    {{- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + ` }}
    {{ range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + `) }}
  - name: "{{  $index }}"
    {{ toYaml $value | indent ` + fmt.Sprintf("%d", (4+indentSize)) + ` }}
    {{ end }}
    {{- end }}
volumes:
{{- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) }}
- name: custom-bootstrap-volume
  configMap:
    name: {{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + ` "" }}
{{- end }}
# SDS channel between istioagent and Envoy
- emptyDir:
    medium: Memory
  name: istio-envoy
{{- if .Values.global.istiod.enabled }}
- name: istio-data
  emptyDir: {}
- name: istio-podinfo
  downwardAPI:
    items:
      - path: "labels"
        fieldRef:
          fieldPath: metadata.labels
      - path: "annotations"
        fieldRef:
          fieldPath: metadata.annotations
      - path: "cpu-limit"
        resourceFieldRef:
          containerName: istio-proxy
          resource: limits.cpu
          divisor: 1m
      - path: "cpu-request"
        resourceFieldRef:
          containerName: istio-proxy
          resource: requests.cpu
          divisor: 1m
{{- if eq .Values.global.jwtPolicy "third-party-jwt" }}
- name: istio-token
  projected:
    sources:
    - serviceAccountToken:
        path: istio-token
        expirationSeconds: 43200
        audience: {{ .Values.global.sds.token.aud }}
{{- end }}
{{- if eq .Values.global.pilotCertProvider "istiod" }}
- name: istiod-ca-cert
  configMap:
    name: {{ .Values.global.istiod.caRootConfigName }}
{{- end }}
{{- if .Values.global.mountMtlsCerts }}
# Use the key and cert mounted to /etc/certs/ for the in-cluster mTLS communications.
- name: istio-certs
  secret:
    optional: true
    {{ if eq .Spec.ServiceAccountName "" }}
    secretName: istio.default
    {{ else -}}
    secretName: {{  printf "istio.%s" .Spec.ServiceAccountName }}
    {{  end -}}
{{- end }}
{{- else }}
{{- if .Values.global.sds.enabled }}
- name: sds-uds-path
  hostPath:
    path: /var/run/sds
- name: istio-token
  projected:
    sources:
      - serviceAccountToken:
          path: istio-token
          expirationSeconds: 43200
          audience: {{ .Values.global.sds.token.aud }}
{{- if .Values.global.sds.customTokenDirectory }}
- name: custom-sds-token
  secret:
    secretName: sdstokensecret
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
{{- end }}
{{- end }}
  {{- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + ` }}
  {{range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + `) }}
- name: "{{ $index }}"
  {{ toYaml $value | indent ` + fmt.Sprintf("%d", (2+indentSize)) + ` }}
  {{ end }}
  {{ end }}
{{- if and (eq .Values.global.proxy.tracer "lightstep") .ProxyConfig.GetTracing.GetTlsSettings }}
- name: lightstep-certs
  secret:
    optional: true
    secretName: lightstep.cacert
{{- end }}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
  {{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
  {{- end }}
{{- end }}
{{- if eq (env "ENABLE_LEGACY_FSGROUP_INJECTION" "true") "true" }}
securityContext:
  fsGroup: 1337
{{- end }}
`
}

func (r *Reconciler) templateConfig() string {
	return `rewriteAppHTTPProbe: {{ valueOrDefault .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe true }}
{{- if .Values.global.podDNSSearchNamespaces }}
dnsConfig:
  searches:
    {{- range .Values.global.podDNSSearchNamespaces }}
    - {{ render . }}
    {{- end }}
{{- end }}
httpProxyEnvs:
  httpProxy: {{ .Values.sidecarInjectorWebhook.httpProxyEnvs.httpProxy }}
  httpsProxy: {{ .Values.sidecarInjectorWebhook.httpProxyEnvs.httpsProxy }}
  noProxy: {{ .Values.sidecarInjectorWebhook.httpProxyEnvs.noProxy }}
` + r.templatePodSpec(0) + `
podRedirectAnnot:
{{- if and (.Values.global.proxy_init.cniEnabled) (not .Values.global.proxy_init.cniChained) }}
k8s.v1.cni.cncf.io/networks: '{{ appendMultusNetwork (index .ObjectMeta.Annotations ` + "`" + `k8s.v1.cni.cncf.io/networks` + "`" + `) ` + "`" + `istio-cni` + "`" + ` }}'
{{- end }}
   sidecar.istio.io/interceptionMode: "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode }}"
   traffic.sidecar.istio.io/includeOutboundIPRanges: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` .Values.global.proxy.includeIPRanges }}"
   traffic.sidecar.istio.io/excludeOutboundIPRanges: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` .Values.global.proxy.excludeIPRanges }}"
   traffic.sidecar.istio.io/includeInboundPorts: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (includeInboundPorts .Spec.Containers) }}"
   traffic.sidecar.istio.io/excludeInboundPorts: "{{ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` .Values.global.proxy.excludeInboundPorts) }}"
{{ if or (isset .ObjectMeta.Annotations  ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts ` + "`" + `) (ne (valueOrDefault .Values.global.proxy.includeOutboundPorts "") "") }}
    traffic.sidecar.istio.io/includeOutboundPorts: "{{ annotation .ObjectMeta  ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts ` + "`" + ` .Values.global.proxy.includeOutboundPorts }}"
{{- end }}
{{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + `) (ne .Values.global.proxy.excludeOutboundPorts "") }}
   traffic.sidecar.istio.io/excludeOutboundPorts: "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + ` .Values.global.proxy.excludeOutboundPorts }}"
{{- end }}
   traffic.sidecar.istio.io/kubevirtInterfaces: "{{ index .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + ` }}"
`
}

func (r *Reconciler) coreDumpContainer() string {
	if !util.PointerToBool(r.Config.Spec.Proxy.EnableCoreDump) {
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

func (r *Reconciler) injectedAddtionalContainerArgs() string {
	if len(r.Config.Spec.SidecarInjector.InjectedContainerAdditionalArgs) == 0 {
		return ""
	}

	additionalArgs, err := yaml.Marshal(r.Config.Spec.SidecarInjector.InjectedContainerAdditionalArgs)
	if err != nil {
		return ""
	}

	return indentWithSpaces(string(additionalArgs), 2)
}

func (r *Reconciler) injectedAddtionalEnvVars() string {
	if len(r.Config.Spec.SidecarInjector.InjectedContainerAdditionalEnvVars) == 0 {
		return ""
	}

	additionalEnvVars, err := yaml.Marshal(r.Config.Spec.SidecarInjector.InjectedContainerAdditionalEnvVars)
	if err != nil {
		return ""
	}

	return indentWithSpaces(string(additionalEnvVars), 2)
}

func (r *Reconciler) trustBundleManagerEnvVar() string {
	if !util.PointerToBool(r.Config.Spec.Pilot.SPIFFE.OperatorEndpoints.Enabled) {
		return ""
	}

	env, err := yaml.Marshal([]apiv1.EnvVar{
		{
			Name:  "TRUSTBUNDLE_MANAGER",
			Value: "true",
		},
	})
	if err != nil {
		return ""
	}

	return indentWithSpaces(string(env), 2)
}

func (r *Reconciler) proxyInitContainer(indentSize int) string {
	return `- name: {{ .Values.global.proxy_init.containerName }}
  image: "{{ .Values.global.proxy_init.image }}"
  args:
  - istio-iptables
  - "-p"
  - "15001"
  - "-z"
  - "15006"
  - "-u"
  - "1337"
  - "-m"
  - "{{ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode }}"
  - "-i"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` .Values.global.proxy.includeIPRanges }}"
  - "-x"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` .Values.global.proxy.excludeIPRanges }}"
  - "-b"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "` `*`" + ` }}"
  - "-d"
  {{- if excludeInboundPort (annotation .ObjectMeta  ` + "`" + `status.sidecar.istio.io/port ` + "`" + ` .Values.global.proxy.statusPort) (annotation .ObjectMeta  ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts ` + "`" + ` .Values.global.proxy.excludeInboundPorts) }}
  - "15090,15021,{{ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` .Values.global.proxy.statusPort) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` .Values.global.proxy.excludeInboundPorts) }}"
  {{- else }}
  - "15090,15021"
  {{- end }}
  {{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts` + "`" + `) (ne (valueOrDefault .Values.global.proxy.includeOutboundPorts "") "") -}}
  - "-q"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundPorts` + "`" + ` .Values.global.proxy.includeOutboundPorts }}"
  {{ end -}}
  {{ if or (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + `) (ne (valueOrDefault .Values.global.proxy.excludeOutboundPorts "") "") -}}
  - "-o"
  - "{{ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundPorts` + "`" + ` .Values.global.proxy.excludeOutboundPorts }}"
  {{ end -}}
  {{ if (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + `) -}}
  - "-k"
  - "{{ index .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + ` }}"
  {{ end -}}
  {{ if .Values.global.proxy_init.cniEnabled -}}
  - "--run-validation"
  - "--skip-rule-apply"
  {{ end -}}
  imagePullPolicy: "{{ valueOrDefault .Values.global.imagePullPolicy  ` + "`" + `Always ` + "`" + ` }}"
  env:
` + r.injectedAddtionalEnvVars() + `
  {{- if .ProxyConfig.ProxyMetadata }}
  {{- range $key, $value := .ProxyConfig.ProxyMetadata }}
  - name: {{ $key }}
    value: "{{ $value }}"
  {{- end }}
  {{- end }}
  securityContext:
  {{- if not .Values.global.proxy_init.cniEnabled }}
    readOnlyRootFilesystem: false
    runAsUser: 0
    runAsGroup: 0
    runAsNonRoot: false
  {{- else }}
    readOnlyRootFilesystem: true
    runAsGroup: 1337
    runAsUser: 1337
    runAsNonRoot: true
  {{- end }}
    allowPrivilegeEscalation: {{ .Values.global.proxy.privileged }}
    privileged: {{ .Values.global.proxy.privileged }}
    capabilities:
      {{- if not .Values.global.proxy_init.cniEnabled }}
      add:
      - NET_ADMIN
      - NET_RAW
      {{- end }}
      drop:
      - ALL
  restartPolicy: Always
  {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) }}
  resources:
    {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) }}
      requests:
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) -}}
        cpu: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + ` }}"
        {{ end }}
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + `) -}}
        memory: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + ` }}"
        {{ end }}
    {{- end }}
    {{- if or (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) }}
      limits:
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + `) -}}
        cpu: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPULimit` + "`" + ` }}"
        {{ end }}
        {{ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + `) -}}
        memory: "{{ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemoryLimit` + "`" + ` }}"
        {{ end }}
    {{- end }}
  {{- else }}
` + r.getFormattedResources(r.Config.Spec.SidecarInjector.Init.Resources, 2) + `
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
	lines := make([]string, 0)
	for _, l := range strings.Split(v, "\n") {
		if len(l) > 0 {
			l = pad + l
		}
		lines = append(lines, l)
	}
	return strings.Join(lines, "\n")
}
