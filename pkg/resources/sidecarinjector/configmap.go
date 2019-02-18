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
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, util.MergeLabels(sidecarInjectorLabels, labelSelector), r.Config),
		Data: map[string]string{
			"config": r.siConfig(),
		},
	}
}

func (r *Reconciler) siConfig() string {
	siConfig := map[string]string{
		"policy":   "enabled",
		"template": r.templateConfig(),
	}
	marshaledConfig, _ := yaml.Marshal(siConfig)
	// this is a static config, so we don't have to deal with errors
	return string(marshaledConfig)

}

func (r *Reconciler) templateConfig() string {
	return `initContainers:
- name: istio-init
  image: docker.io/istio/proxy_init:1.0.5
  args:
  - "-p"
  - [[ .MeshConfig.ProxyListenPort ]]
  - "-u"
  - 1337
  - "-m"
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode ]]
  - "-i"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` "` + r.includeIPRanges + `" ]]"
  - "-x"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` "` + r.excludeIPRanges + `" ]]"
  - "-b"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (includeInboundPorts .Spec.Containers) ]]"
  - "-d"
  - "[[ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` "" ) ]]"
  imagePullPolicy: IfNotPresent
  securityContext:
    capabilities:
      add:
      - NET_ADMIN
    privileged: true
  restartPolicy: Always
containers:
- name: istio-proxy
  image: "[[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/proxyImage` + "`" + ` "docker.io/istio/proxyv2:1.0.5" ]]"
  ports:
  - containerPort: 15090
    protocol: TCP
    name: http-envoy-prom
  args:
  - proxy
  - sidecar
  - --configPath
  - [[ .ProxyConfig.ConfigPath ]]
  - --binaryPath
  - [[ .ProxyConfig.BinaryPath ]]
  - --serviceCluster
  [[ if ne "" (index .ObjectMeta.Labels "app") -]]
  - [[ index .ObjectMeta.Labels "app" ]]
  [[ else -]]
  - "istio-proxy"
  [[ end -]]
  - --drainDuration
  - [[ formatDuration .ProxyConfig.DrainDuration ]]
  - --parentShutdownDuration
  - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
  - --discoveryAddress
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/discoveryAddress` + "`" + ` .ProxyConfig.DiscoveryAddress ]]
  - --discoveryRefreshDelay
  - [[ formatDuration .ProxyConfig.DiscoveryRefreshDelay ]]
  - --zipkinAddress
  - [[ .ProxyConfig.ZipkinAddress ]]
  - --connectTimeout
  - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
  - --proxyAdminPort
  - [[ .ProxyConfig.ProxyAdminPort ]]
  [[ if gt .ProxyConfig.Concurrency 0 -]]
  - --concurrency
  - [[ .ProxyConfig.Concurrency ]]
  [[ end -]]
  - --controlPlaneAuthPolicy
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/controlPlaneAuthPolicy` + "`" + ` .ProxyConfig.ControlPlaneAuthPolicy ]]
[[- if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) "0") ]]
  - --statusPort
  - [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ]]
  - --applicationPorts
  - [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/applicationPorts` + "`" + ` (applicationPorts .Spec.Containers) ]]
[[- end ]]
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
  - name: ISTIO_META_POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: ISTIO_META_INTERCEPTION_MODE
    value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
  [[ if .ObjectMeta.Annotations ]]
  - name: ISTIO_METAJSON_ANNOTATIONS
    value: |
           [[ toJson .ObjectMeta.Annotations ]]
  [[ end ]]
  [[ if .ObjectMeta.Labels ]]
  - name: ISTIO_METAJSON_LABELS
    value: |
           [[ toJson .ObjectMeta.Labels ]]
  [[ end ]]
  imagePullPolicy: IfNotPresent
  [[ if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ) "0") ]]
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "0" ]]
    initialDelaySeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/initialDelaySeconds` + "`" + ` "1" ]]
    periodSeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/periodSeconds` + "`" + ` "2" ]]
    failureThreshold: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/failureThreshold` + "`" + ` "30" ]]
  [[ end -]]
  securityContext:
    readOnlyRootFilesystem: true
    [[ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "TPROXY" -]]
    capabilities:
      add:
      - NET_ADMIN
    runAsGroup: 1337
    [[ else -]]
    runAsUser: 1337
    [[ end -]]
  restartPolicy: Always
  resources:
    [[ if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + `) -]]
    requests:
      cpu: [[ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyCPU` + "`" + ` ]]
      memory: [[ index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/proxyMemory` + "`" + ` ]]
  [[ else -]]
    requests:
      cpu: 10m
  [[ end -]]
  volumeMounts:
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
volumes:
- emptyDir:
    medium: Memory
  name: istio-envoy
- name: istio-certs
  secret:
    optional: true
    [[ if eq .Spec.ServiceAccountName "" -]]
    secretName: istio.default
    [[ else -]]
    secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
    [[ end -]]
`
}
