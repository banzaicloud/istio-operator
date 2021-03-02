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
		ObjectMeta: templates.ObjectMetaWithRevision(IstioConfigMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh":         string(marshaledConfig),
			"meshNetworks": meshNetworks(r.Config),
		},
	}
}

func MeshConfig(config *istiov1beta1.Istio, remote bool) map[string]interface{} {
	defaultConfig := map[string]interface{}{
		"configPath":             "./etc/istio/proxy",
		"binaryPath":             "/usr/local/bin/envoy",
		"serviceCluster":         "istio-proxy",
		"drainDuration":          "45s",
		"parentShutdownDuration": "1m0s",
		"proxyAdminPort":         15000,
		"concurrency":            0,
		"controlPlaneAuthPolicy": config.GetControlPlaneAuthPolicy(),
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
				"zipkin": config.Spec.Tracing.Zipkin.GetData(),
			}
			if config.Spec.Tracing.Zipkin.TLSSettings != nil {
				defaultConfig["tracing"].(map[string]interface{})["tlsSettings"] = config.Spec.Tracing.Zipkin.TLSSettings
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
		if config.Spec.Tracing.CustomTags != nil && len(config.Spec.Tracing.CustomTags) > 0 {
			defaultConfig["tracing"].(map[string]interface{})["custom_tags"] = config.Spec.Tracing.CustomTags
		}
	}

	meshConfig := map[string]interface{}{
		"connectTimeout":              "10s",
		"protocolDetectionTimeout":    config.Spec.Proxy.ProtocolDetectionTimeout,
		"ingressClass":                "istio",
		"ingressService":              "istio-ingressgateway",
		"ingressControllerMode":       "STRICT",
		"enableTracing":               config.Spec.Tracing.Enabled,
		"accessLogFile":               config.Spec.Proxy.AccessLogFile,
		"accessLogFormat":             config.Spec.Proxy.AccessLogFormat,
		"accessLogEncoding":           config.Spec.Proxy.AccessLogEncoding,
		"enableEnvoyAccessLogService": util.PointerToBool(config.Spec.Proxy.EnvoyAccessLogService.Enabled),
		"defaultConfig":               defaultConfig,
		"outboundTrafficPolicy": map[string]interface{}{
			"mode": config.Spec.OutboundTrafficPolicy.Mode,
		},
		"enableAutoMtls":     util.PointerToBool(config.Spec.AutoMTLS),
		"trustDomain":        config.Spec.TrustDomain,
		"trustDomainAliases": config.Spec.TrustDomainAliases,
		"rootNamespace":      config.Namespace,
		"localityLbSetting":  getLocalityLBConfiguration(config),
		"dnsRefreshRate":     config.Spec.Proxy.DNSRefreshRate,
	}

	if config.Spec.MeshID != "" {
		defaultConfig["meshId"] = config.Spec.MeshID
	} else {
		defaultConfig["meshId"] = config.Spec.TrustDomain
	}

	if len(config.Spec.Certificates) != 0 {
		meshConfig["certificates"] = config.Spec.Certificates
	}

	if util.PointerToBool(config.Spec.Policy.Enabled) {
		meshConfig["mixerCheckServer"] = mixerServerWithRevision(config, "policy", remote)
	}

	if util.PointerToBool(config.Spec.Telemetry.Enabled) {
		meshConfig["mixerReportServer"] = mixerServerWithRevision(config, "telemetry", remote)
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

func mixerServerWithRevision(config *istiov1beta1.Istio, mixerType string, remote bool) string {
	return mixerServer(config, config.WithRevision(mixerType), remote)
}

func mixerServer(config *istiov1beta1.Istio, mixerType string, remote bool) string {
	if remote {
		return fmt.Sprintf("istio-%s.%s:%s", mixerType, config.Namespace, "15004")
	}
	return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, config.Namespace, config.Spec.Proxy.ClusterDomain, "15004")
}
