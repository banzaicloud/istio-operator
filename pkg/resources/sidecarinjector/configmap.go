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
	"strconv"

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
	return `rewriteAppHTTPProbe: false
initContainers:
[[ if ne (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "NONE" ]]
- name: istio-init
  image: ` + r.Config.Spec.ProxyInit.Image + `
  args:
  - "-p"
  - [[ .MeshConfig.ProxyListenPort ]]
  - "-u"
  - 1337
  - "-m"
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode ]]
  - "-i"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeOutboundIPRanges` + "`" + ` "` + r.Config.Spec.IncludeIPRanges + `" ]]"
  - "-x"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeOutboundIPRanges` + "`" + ` "` + r.Config.Spec.ExcludeIPRanges + `" ]]"
  - "-b"
  - "[[ annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/includeInboundPorts` + "`" + ` (includeInboundPorts .Spec.Containers) ]]"
  - "-d"
  - "[[ excludeInboundPort (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` "15020" ) (annotation .ObjectMeta ` + "`" + `traffic.sidecar.istio.io/excludeInboundPorts` + "`" + ` "" ) ]]"
  [[ if (isset .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + `) -]]
  - "-k"
  - "[[ index .ObjectMeta.Annotations ` + "`" + `traffic.sidecar.istio.io/kubevirtInterfaces` + "`" + ` ]]"
  [[ end -]]
  imagePullPolicy: IfNotPresent
  resources:
    requests:
      cpu: 10m
      memory: 10Mi
    limits:
      cpu: 100m
      memory: 50Mi
  securityContext:
    capabilities:
      add:
      - NET_ADMIN
    privileged: ` + strconv.FormatBool(r.Config.Spec.Proxy.Privileged) + `
  restartPolicy: Always
` + r.coreDumpContainer() + `
[[ end -]]
containers:
- name: istio-proxy
  image: "[[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/proxyImage` + "` \"" + r.Config.Spec.Proxy.Image + `" ]]"
  ports:
  - containerPort: 15090
    protocol: TCP
    name: http-envoy-prom
  args:
  - proxy
  - sidecar
  - --domain
  - $(POD_NAMESPACE).svc.cluster.local
  - --configPath
  - [[ .ProxyConfig.ConfigPath ]]
  - --binaryPath
  - [[ .ProxyConfig.BinaryPath ]]
  - --serviceCluster
  [[ if ne "" (index .ObjectMeta.Labels "app") -]]
  - [[ index .ObjectMeta.Labels "app" ]].$(POD_NAMESPACE)
  [[ else -]]
  - [[ valueOrDefault .DeploymentMeta.Name "istio-proxy" ]].[[ valueOrDefault .DeploymentMeta.Namespace "default" ]]
  [[ end -]]
  - --drainDuration
  - [[ formatDuration .ProxyConfig.DrainDuration ]]
  - --parentShutdownDuration
  - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
  - --discoveryAddress
  - [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/discoveryAddress` + "`" + ` .ProxyConfig.DiscoveryAddress ]]
  - --zipkinAddress
  - [[ .ProxyConfig.GetTracing.GetZipkin.GetAddress ]]
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
[[- if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` 15020 ) "0") ]]
  - --statusPort
  - [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` 15020 ]]
  - --applicationPorts
  - "[[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/applicationPorts` + "`" + ` (applicationPorts .Spec.Containers) ]]"
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
  - name: ISTIO_META_CONFIG_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: ISTIO_META_INTERCEPTION_MODE
    value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
  [[ if .ObjectMeta.Annotations ]]
  - name: ISTIO_METAJSON_ANNOTATIONS
    value: |
           [[ toJSON .ObjectMeta.Annotations ]]
  [[ end ]]
  [[ if .ObjectMeta.Labels ]]
  - name: ISTIO_METAJSON_LABELS
    value: |
           [[ toJSON .ObjectMeta.Labels ]]
  [[ end ]]
  [[- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) ]]
  - name: ISTIO_BOOTSTRAP_OVERRIDE
    value: "/etc/istio/custom-bootstrap/custom_bootstrap.json"
  [[- end ]]
  imagePullPolicy: IfNotPresent
  [[ if (ne (annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` 15020 ) "0") ]]
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: [[ annotation .ObjectMeta ` + "`" + `status.sidecar.istio.io/port` + "`" + ` 15020 ]]
    initialDelaySeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/initialDelaySeconds` + "`" + ` "1" ]]
    periodSeconds: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/periodSeconds` + "`" + ` "2" ]]
    failureThreshold: [[ annotation .ObjectMeta ` + "`" + `readiness.status.sidecar.istio.io/failureThreshold` + "`" + ` "30" ]]
  [[ end -]]
  securityContext:
    privileged: ` + strconv.FormatBool(r.Config.Spec.Proxy.Privileged) + `
    readOnlyRootFilesystem: ` + strconv.FormatBool(!r.Config.Spec.Proxy.EnableCoreDump) + `
    [[ if eq (annotation .ObjectMeta ` + "`" + `sidecar.istio.io/interceptionMode` + "`" + ` .ProxyConfig.InterceptionMode) "TPROXY" -]]
    capabilities:
      add:
      - NET_ADMIN
    runAsGroup: 1337
    [[ else -]] 
    ` + r.runAsGroup() + `
    runAsUser: 1337
    [[- end ]]
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
  [[- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) ]]
  - mountPath: /etc/istio/custom-bootstrap
    name: custom-bootstrap-volume
  [[- end ]]
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  ` + r.volumeMounts() + `
    [[- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + ` ]]
    [[ range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolumeMount` + "`" + `) ]]
  - name: "[[ $index ]]"
    [[ toYaml $value | indent 4 ]]
    [[ end ]]
    [[- end ]]
volumes:
[[- if (isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/bootstrapOverride` + "`" + `) ]]
- name: custom-bootstrap-volume
  configMap:
    name: [[ annotation .ObjectMeta ` + "`" + `sidecar.istio.io/bootstrapOverride` + "` ``" + ` ]]
[[- end ]]
- emptyDir:
    medium: Memory
  name: istio-envoy
` + r.volumes()
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

func (r *Reconciler) runAsGroup() string {
	if r.Config.Spec.SDS.Enabled && r.Config.Spec.SDS.UseTrustworthyJwt {
		return "runAsGroup: 1337"
	}
	return ""
}

func (r *Reconciler) volumeMounts() string {
	if !r.Config.Spec.SDS.Enabled {
		return `- mountPath: /etc/certs/
    name: istio-certs
    readOnly: true`
	}
	vms := `- mountPath: /var/run/sds
    name: sds-uds-path`
	if r.Config.Spec.SDS.UseTrustworthyJwt {
		vms = vms + `
  - mountPath: /var/run/secrets/tokens
    name: istio-token`
	}
	return vms
}

func (r *Reconciler) volumes() string {
	if !r.Config.Spec.SDS.Enabled {
		return `- name: istio-certs
  secret:
    optional: true
    [[ if eq .Spec.ServiceAccountName "" -]]
    secretName: istio.default
    [[ else -]]
    secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
    [[ end -]]
  [[- if isset .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + ` ]]
  [[ range $index, $value := fromJSON (index .ObjectMeta.Annotations ` + "`" + `sidecar.istio.io/userVolume` + "`" + `) ]]
- name: "[[ $index ]]"
  [[ toYaml $value | indent 2 ]]
  [[ end ]]
  [[ end ]]`
	}
	volumes := `- name: sds-uds-path
  hostPath:
    path: /var/run/sds`
	if r.Config.Spec.SDS.UseTrustworthyJwt {
		volumes = volumes + `
- name: istio-token
  projected:
    sources:
    - serviceAccountToken:
        path: istio-token
        expirationSeconds: 43200
        audience: ""`
	}
	return volumes
}
