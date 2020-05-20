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

package base

import (
	"fmt"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func (r *Reconciler) configMap() runtime.Object {
	meshConfig := MeshConfig(r.Config, r.remote)
	marshaledConfig, _ := yaml.Marshal(meshConfig)

	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(IstioConfigMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh":         string(marshaledConfig),
			"meshNetworks": meshNetworks(r.Config),
		},
	}
}

func MeshConfig(config *istiov1beta1.Istio, remote bool) map[string]interface{} {
	defaultConfig := map[string]interface{}{
		"connectTimeout":         "10s",
		"configPath":             "./etc/istio/proxy",
		"binaryPath":             "/usr/local/bin/envoy",
		"serviceCluster":         "istio-proxy",
		"drainDuration":          "45s",
		"parentShutdownDuration": "1m0s",
		"proxyAdminPort":         15000,
		"concurrency":            0,
		"controlPlaneAuthPolicy": templates.ControlPlaneAuthPolicy(util.PointerToBool(config.Spec.Istiod.Enabled), config.Spec.ControlPlaneSecurityEnabled),
		"discoveryAddress":       config.GetDiscoveryAddress(),
	}

	if util.PointerToBool(config.Spec.Proxy.EnvoyStatsD.Enabled) {
		defaultConfig["statsdUdpAddress"] = fmt.Sprintf("%s:%d", config.Spec.Proxy.EnvoyStatsD.Host, config.Spec.Proxy.EnvoyStatsD.Port)
	}

	if util.PointerToBool(config.Spec.Proxy.EnvoyMetricsService.Enabled) {
		defaultConfig["envoyAccessLogService"] = config.Spec.Proxy.EnvoyMetricsService.GetData()
	}

	if util.PointerToBool(config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		defaultConfig["envoyAccessLogService"] = config.Spec.Proxy.EnvoyAccessLogService.GetData()
	}

	if util.PointerToBool(config.Spec.Tracing.Enabled) {
		switch config.Spec.Tracing.Tracer {
		case istiov1beta1.TracerTypeZipkin:
			defaultConfig["tracing"] = map[string]interface{}{
				"zipkin": map[string]interface{}{
					"address": config.Spec.Tracing.Zipkin.Address,
				},
			}
		case istiov1beta1.TracerTypeLightstep:
			lightStep := map[string]interface{}{
				"address":     config.Spec.Tracing.Lightstep.Address,
				"accessToken": config.Spec.Tracing.Lightstep.AccessToken,
			}
			defaultConfig["tracing"] = map[string]interface{}{
				"lightstep": lightStep,
			}
		case istiov1beta1.TracerTypeDatadog:
			defaultConfig["tracing"] = map[string]interface{}{
				"datadog": map[string]interface{}{
					"address": config.Spec.Tracing.Datadog.Address,
				},
			}
		case istiov1beta1.TracerTypeStackdriver:
			defaultConfig["tracing"] = map[string]interface{}{
				"stackdriver": config.Spec.Tracing.Strackdriver,
			}
		}
	}

	meshConfig := map[string]interface{}{
		"disablePolicyChecks":     !util.PointerToBool(config.Spec.Policy.ChecksEnabled),
		"disableMixerHttpReports": false,
		"enableTracing":           config.Spec.Tracing.Enabled,
		"accessLogFile":           config.Spec.Proxy.AccessLogFile,
		"accessLogFormat":         config.Spec.Proxy.AccessLogFormat,
		"accessLogEncoding":       config.Spec.Proxy.AccessLogEncoding,
		"policyCheckFailOpen":     false,
		"ingressService":          "istio-ingressgateway",
		"ingressClass":            "istio",
		"ingressControllerMode":   "STRICT",
		"trustDomain":             config.Spec.TrustDomain,
		"trustDomainAliases":      config.Spec.TrustDomainAliases,
		"enableAutoMtls":          util.PointerToBool(config.Spec.AutoMTLS),
		"outboundTrafficPolicy": map[string]interface{}{
			"mode": config.Spec.OutboundTrafficPolicy.Mode,
		},
		"defaultConfig":               defaultConfig,
		"rootNamespace":               config.Namespace,
		"connectTimeout":              "10s",
		"localityLbSetting":           getLocalityLBConfiguration(config),
		"enableEnvoyAccessLogService": util.PointerToBool(config.Spec.Proxy.EnvoyAccessLogService.Enabled),
		"protocolDetectionTimeout":    config.Spec.Proxy.ProtocolDetectionTimeout,
		"dnsRefreshRate":              config.Spec.Proxy.DNSRefreshRate,
	}

	if len(config.Spec.Certificates) != 0 {
		meshConfig["certificates"] = config.Spec.Certificates
	}

	meshConfig["sdsUdsPath"] = "unix:/etc/istio/proxy/SDS"

	if util.PointerToBool(config.Spec.Policy.Enabled) {
		meshConfig["mixerCheckServer"] = mixerServer(config, "policy", remote)
	}

	if util.PointerToBool(config.Spec.Telemetry.Enabled) {
		meshConfig["mixerReportServer"] = mixerServer(config, "telemetry", remote)
		meshConfig["reportBatchMaxEntries"] = config.Spec.Telemetry.ReportBatchMaxEntries
		meshConfig["reportBatchMaxTime"] = config.Spec.Telemetry.ReportBatchMaxTime

		if util.PointerToBool(config.Spec.Telemetry.SessionAffinityEnabled) {
			meshConfig["sidecarToTelemetrySessionAffinity"] = util.PointerToBool(config.Spec.Telemetry.SessionAffinityEnabled)
		}
	}

	return meshConfig
}

func getLocalityLBConfiguration(config *istiov1beta1.Istio) *istiov1beta1.LocalityLBConfiguration {
	var localityLbConfiguration *istiov1beta1.LocalityLBConfiguration

	if config.Spec.LocalityLB == nil || !util.PointerToBool(config.Spec.LocalityLB.Enabled) {
		return localityLbConfiguration
	}

	if config.Spec.LocalityLB != nil {
		localityLbConfiguration = config.Spec.LocalityLB.DeepCopy()
		localityLbConfiguration.Enabled = nil
		if localityLbConfiguration.Distribute != nil && localityLbConfiguration.Failover != nil {
			localityLbConfiguration.Failover = nil
		}
	}

	return localityLbConfiguration
}

func meshNetworks(config *istiov1beta1.Istio) string {
	marshaledConfig, _ := yaml.Marshal(config.Spec.GetMeshNetworks())
	return string(marshaledConfig)
}

func mixerServer(config *istiov1beta1.Istio, mixerType string, remote bool) string {
	if remote {
		return fmt.Sprintf("istio-%s.%s:%s", mixerType, config.Namespace, "15004")
	}
	if config.Spec.ControlPlaneSecurityEnabled {
		return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, config.Namespace, config.Spec.Proxy.ClusterDomain, "15004")
	}
	return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, config.Namespace, config.Spec.Proxy.ClusterDomain, "9091")
}
