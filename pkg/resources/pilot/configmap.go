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

package pilot

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
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh":         r.meshConfig(),
			"meshNetworks": r.meshNetworks(),
		},
	}
}

func (r *Reconciler) meshConfig() string {
	defaultConfig := map[string]interface{}{
		"connectTimeout":         "10s",
		"configPath":             "/etc/istio/proxy",
		"binaryPath":             "/usr/local/bin/envoy",
		"serviceCluster":         "istio-proxy",
		"drainDuration":          "45s",
		"parentShutdownDuration": "1m0s",
		"proxyAdminPort":         15000,
		"concurrency":            0,
		"controlPlaneAuthPolicy": templates.ControlPlaneAuthPolicy(util.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"discoveryAddress":       fmt.Sprintf("istio-pilot.%s:%s", r.Config.Namespace, r.discoveryPort()),
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyStatsD.Enabled) {
		defaultConfig["statsdUdpAddress"] = fmt.Sprintf("%s:%d", r.Config.Spec.Proxy.EnvoyStatsD.Host, r.Config.Spec.Proxy.EnvoyStatsD.Port)
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyMetricsService.Enabled) {
		metricsService := map[string]interface{}{
			"address": fmt.Sprintf("%s:%d", r.Config.Spec.Proxy.EnvoyMetricsService.Host, r.Config.Spec.Proxy.EnvoyMetricsService.Port),
		}
		if r.Config.Spec.Proxy.EnvoyMetricsService.TLSSettings != nil {
			metricsService["tlsSettings"] = r.Config.Spec.Proxy.EnvoyMetricsService.TLSSettings
		}
		if r.Config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive != nil {
			metricsService["tcpKeepalive"] = r.Config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive
		}
		defaultConfig["envoyAccessLogService"] = metricsService
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		accessLogService := map[string]interface{}{
			"address": fmt.Sprintf("%s:%d", r.Config.Spec.Proxy.EnvoyAccessLogService.Host, r.Config.Spec.Proxy.EnvoyAccessLogService.Port),
		}
		if r.Config.Spec.Proxy.EnvoyAccessLogService.TLSSettings != nil {
			accessLogService["tlsSettings"] = r.Config.Spec.Proxy.EnvoyAccessLogService.TLSSettings
		}
		if r.Config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive != nil {
			accessLogService["tcpKeepalive"] = r.Config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive
		}
		defaultConfig["envoyAccessLogService"] = accessLogService
	}

	if util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		switch r.Config.Spec.Tracing.Tracer {
		case istiov1beta1.TracerTypeZipkin:
			defaultConfig["tracing"] = map[string]interface{}{
				"zipkin": map[string]interface{}{
					"address": r.Config.Spec.Tracing.Zipkin.Address,
				},
			}
		case istiov1beta1.TracerTypeLightstep:
			lightStep := map[string]interface{}{
				"address":     r.Config.Spec.Tracing.Lightstep.Address,
				"accessToken": r.Config.Spec.Tracing.Lightstep.AccessToken,
				"secure":      r.Config.Spec.Tracing.Lightstep.Secure,
			}
			if r.Config.Spec.Tracing.Lightstep.Secure {
				lightStep["cacertPath"] = r.Config.Spec.Tracing.Lightstep.CacertPath
			}
			defaultConfig["tracing"] = map[string]interface{}{
				"lightstep": lightStep,
			}
		case istiov1beta1.TracerTypeDatadog:
			defaultConfig["tracing"] = map[string]interface{}{
				"datadog": map[string]interface{}{
					"address": r.Config.Spec.Tracing.Datadog.Address,
				},
			}
		case istiov1beta1.TracerTypeStackdriver:
			defaultConfig["tracing"] = map[string]interface{}{
				"stackdriver": r.Config.Spec.Tracing.Strackdriver,
			}
		}
	}

	meshConfig := map[string]interface{}{
		"disablePolicyChecks":     !util.PointerToBool(r.Config.Spec.Policy.ChecksEnabled),
		"disableMixerHttpReports": false,
		"enableTracing":           r.Config.Spec.Tracing.Enabled,
		"accessLogFile":           r.Config.Spec.Proxy.AccessLogFile,
		"accessLogFormat":         r.Config.Spec.Proxy.AccessLogFormat,
		"accessLogEncoding":       r.Config.Spec.Proxy.AccessLogEncoding,
		"policyCheckFailOpen":     false,
		"ingressService":          "istio-ingressgateway",
		"ingressClass":            "istio",
		"ingressControllerMode":   2,
		"trustDomain":             r.Config.Spec.TrustDomain,
		"trustDomainAliases":      r.Config.Spec.TrustDomainAliases,
		"enableAutoMtls":          util.PointerToBool(r.Config.Spec.AutoMTLS),
		"outboundTrafficPolicy": map[string]interface{}{
			"mode": r.Config.Spec.OutboundTrafficPolicy.Mode,
		},
		"defaultConfig":               defaultConfig,
		"rootNamespace":               r.Config.Namespace,
		"connectTimeout":              "10s",
		"localityLbSetting":           r.getLocalityLBConfiguration(),
		"enableEnvoyAccessLogService": util.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled),
		"protocolDetectionTimeout":    r.Config.Spec.Proxy.ProtocolDetectionTimeout,
		"dnsRefreshRate":              r.Config.Spec.Proxy.DNSRefreshRate,
		"certificates":                r.Config.Spec.Certificates,
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		meshConfig["sdsUdsPath"] = "unix:/etc/istio/proxy/SDS"
	} else {
		meshConfig["sdsUdsPath"] = ""
		meshConfig["enableSdsTokenMount"] = false
		meshConfig["sdsUseK8sSaJwt"] = false
	}

	if util.PointerToBool(r.Config.Spec.Policy.Enabled) {
		meshConfig["mixerCheckServer"] = r.mixerServer("policy")
	}

	if util.PointerToBool(r.Config.Spec.Telemetry.Enabled) {
		meshConfig["mixerReportServer"] = r.mixerServer("telemetry")
		meshConfig["reportBatchMaxEntries"] = r.Config.Spec.Telemetry.ReportBatchMaxEntries
		meshConfig["reportBatchMaxTime"] = r.Config.Spec.Telemetry.ReportBatchMaxTime

		if util.PointerToBool(r.Config.Spec.Telemetry.SessionAffinityEnabled) {
			meshConfig["sidecarToTelemetrySessionAffinity"] = util.PointerToBool(r.Config.Spec.Telemetry.SessionAffinityEnabled)
		}
	}

	if util.PointerToBool(r.Config.Spec.UseMCP) {
		meshConfig["configSources"] = []map[string]interface{}{
			r.defaultConfigSource(),
		}
	}

	marshaledConfig, _ := yaml.Marshal(meshConfig)
	return string(marshaledConfig)
}

func (r *Reconciler) getLocalityLBConfiguration() *istiov1beta1.LocalityLBConfiguration {
	var localityLbConfiguration *istiov1beta1.LocalityLBConfiguration

	if r.Config.Spec.LocalityLB == nil || !util.PointerToBool(r.Config.Spec.LocalityLB.Enabled) {
		return localityLbConfiguration
	}

	if r.Config.Spec.LocalityLB != nil {
		localityLbConfiguration = r.Config.Spec.LocalityLB.DeepCopy()
		localityLbConfiguration.Enabled = nil
		if localityLbConfiguration.Distribute != nil && localityLbConfiguration.Failover != nil {
			localityLbConfiguration.Failover = nil
		}
	}

	return localityLbConfiguration
}

func (r *Reconciler) meshNetworks() string {
	marshaledConfig, _ := yaml.Marshal(r.Config.Spec.GetMeshNetworks())
	return string(marshaledConfig)
}

func (r *Reconciler) mixerServer(mixerType string) string {
	if r.remote {
		return fmt.Sprintf("istio-%s.%s:%s", mixerType, r.Config.Namespace, "15004")
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain, "15004")
	}
	return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain, "9091")
}

func (r *Reconciler) defaultConfigSource() map[string]interface{} {
	cs := map[string]interface{}{
		"address": fmt.Sprintf("istio-galley.%s.svc:9901", r.Config.Namespace),
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		cs["tlsSettings"] = map[string]interface{}{
			"mode": "ISTIO_MUTUAL",
		}
	}
	return cs
}

func (r *Reconciler) discoveryPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15011"
	}
	return "15010"
}
