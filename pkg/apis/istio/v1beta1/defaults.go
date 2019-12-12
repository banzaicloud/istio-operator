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

package v1beta1

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	defaultImageHub                  = "docker.io/istio"
	defaultImageVersion              = "1.3.6"
	defaultPilotImage                = defaultImageHub + "/" + "pilot" + ":" + defaultImageVersion
	defaultCitadelImage              = defaultImageHub + "/" + "citadel" + ":" + defaultImageVersion
	defaultGalleyImage               = defaultImageHub + "/" + "galley" + ":" + defaultImageVersion
	defaultMixerImage                = defaultImageHub + "/" + "mixer" + ":" + defaultImageVersion
	defaultSidecarInjectorImage      = defaultImageHub + "/" + "sidecar_injector" + ":" + defaultImageVersion
	defaultNodeAgentImage            = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultSDSImage                  = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultProxyImage                = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyInitImage            = defaultImageHub + "/" + "proxy_init" + ":" + defaultImageVersion
	defaultInitCNIImage              = defaultImageHub + "/" + "install-cni:" + defaultImageVersion
	defaultCoreDNSImage              = "coredns/coredns:1.1.2"
	defaultCoreDNSPluginImage        = defaultImageHub + "/coredns-plugin:0.2-istio-1.1"
	defaultIncludeIPRanges           = "*"
	defaultReplicaCount              = 1
	defaultMinReplicas               = 1
	defaultMaxReplicas               = 5
	defaultTraceSampling             = 1.0
	defaultIngressGatewayServiceType = apiv1.ServiceTypeLoadBalancer
	defaultEgressGatewayServiceType  = apiv1.ServiceTypeClusterIP
	outboundTrafficPolicyAllowAny    = "ALLOW_ANY"
	defaultZipkinAddress             = "zipkin.%s:9411"
	defaultInitCNIBinDir             = "/opt/cni/bin"
	defaultInitCNIConfDir            = "/etc/cni/net.d"
	defaultInitCNILogLevel           = "info"
	defaultImagePullPolicy           = "IfNotPresent"
	defaultMeshExpansion             = false
	defaultEnvoyAccessLogFile        = "/dev/stdout"
	defaultEnvoyAccessLogFormat      = ""
	defaultEnvoyAccessLogEncoding    = "TEXT"
)

var defaultResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU: resource.MustParse("10m"),
	},
}

var defaultProxyResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("100m"),
		apiv1.ResourceMemory: resource.MustParse("128Mi"),
	},
	Limits: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("2000m"),
		apiv1.ResourceMemory: resource.MustParse("1024Mi"),
	},
}

var defaultInitResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("10m"),
		apiv1.ResourceMemory: resource.MustParse("10Mi"),
	},
	Limits: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("100m"),
		apiv1.ResourceMemory: resource.MustParse("50Mi"),
	},
}

var defaultIngressGatewayPorts = []apiv1.ServicePort{
	{Port: 15020, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15020), Name: "status-port", NodePort: 31460},
	{Port: 80, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80), Name: "http2", NodePort: 31380},
	{Port: 443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443), Name: "https", NodePort: 31390},
	{Port: 15443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15443), Name: "tls", NodePort: 31450},
}

var defaultEgressGatewayPorts = []apiv1.ServicePort{
	{Port: 80, Name: "http2", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80)},
	{Port: 443, Name: "https", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443)},
	{Port: 15443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15443), Name: "tls"},
}

func SetDefaults(config *Istio) {
	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	if config.Spec.Proxy.Resources == nil {
		if config.Spec.DefaultResources == nil {
			config.Spec.Proxy.Resources = defaultProxyResources
		} else {
			config.Spec.Proxy.Resources = defaultResources
		}
	}
	if config.Spec.DefaultResources == nil {
		config.Spec.DefaultResources = defaultResources
	}
	// Pilot config
	if config.Spec.Pilot.Enabled == nil {
		config.Spec.Pilot.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Pilot.Image == "" {
		config.Spec.Pilot.Image = defaultPilotImage
	}
	if config.Spec.Pilot.Sidecar == nil {
		config.Spec.Pilot.Sidecar = util.BoolPointer(true)
	}
	if config.Spec.Pilot.ReplicaCount == 0 {
		config.Spec.Pilot.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Pilot.MinReplicas == 0 {
		config.Spec.Pilot.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Pilot.MaxReplicas == 0 {
		config.Spec.Pilot.MaxReplicas = defaultMaxReplicas
	}
	if config.Spec.Pilot.TraceSampling == 0 {
		config.Spec.Pilot.TraceSampling = defaultTraceSampling
	}
	if config.Spec.Pilot.EnableProtocolSniffing == nil {
		config.Spec.Pilot.EnableProtocolSniffing = util.BoolPointer(true)
	}
	if config.Spec.Pilot.EnableProtocolSniffingOutbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingOutbound = config.Spec.Pilot.EnableProtocolSniffing
	}
	if config.Spec.Pilot.EnableProtocolSniffingInbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingInbound = util.BoolPointer(false)
	}
	// Citadel config
	if config.Spec.Citadel.Enabled == nil {
		config.Spec.Citadel.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Citadel.Image == "" {
		config.Spec.Citadel.Image = defaultCitadelImage
	}
	if config.Spec.Citadel.EnableNamespacesByDefault == nil {
		config.Spec.Citadel.EnableNamespacesByDefault = util.BoolPointer(true)
	}
	// Galley config
	if config.Spec.Galley.Enabled == nil {
		config.Spec.Galley.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Galley.Image == "" {
		config.Spec.Galley.Image = defaultGalleyImage
	}
	if config.Spec.Galley.ReplicaCount == 0 {
		config.Spec.Galley.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Galley.ConfigValidation == nil {
		config.Spec.Galley.ConfigValidation = util.BoolPointer(true)
	}
	// Gateways config
	if config.Spec.Gateways.Enabled == nil {
		config.Spec.Gateways.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Gateways.IngressConfig.Enabled == nil {
		config.Spec.Gateways.IngressConfig.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Gateways.IngressConfig.ReplicaCount == 0 {
		config.Spec.Gateways.IngressConfig.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Gateways.IngressConfig.MinReplicas == 0 {
		config.Spec.Gateways.IngressConfig.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Gateways.IngressConfig.MaxReplicas == 0 {
		config.Spec.Gateways.IngressConfig.MaxReplicas = defaultMaxReplicas
	}
	if config.Spec.Gateways.IngressConfig.SDS.Enabled == nil {
		config.Spec.Gateways.IngressConfig.SDS.Enabled = util.BoolPointer(false)
	}
	if len(config.Spec.Gateways.IngressConfig.Ports) == 0 {
		config.Spec.Gateways.IngressConfig.Ports = defaultIngressGatewayPorts
	}
	if config.Spec.Gateways.EgressConfig.Enabled == nil {
		config.Spec.Gateways.EgressConfig.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Gateways.EgressConfig.ReplicaCount == 0 {
		config.Spec.Gateways.EgressConfig.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Gateways.EgressConfig.MinReplicas == 0 {
		config.Spec.Gateways.EgressConfig.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Gateways.EgressConfig.MaxReplicas == 0 {
		config.Spec.Gateways.EgressConfig.MaxReplicas = defaultMaxReplicas
	}
	if config.Spec.Gateways.IngressConfig.ServiceType == "" {
		config.Spec.Gateways.IngressConfig.ServiceType = defaultIngressGatewayServiceType
	}
	if config.Spec.Gateways.EgressConfig.ServiceType == "" {
		config.Spec.Gateways.EgressConfig.ServiceType = defaultEgressGatewayServiceType
	}
	if config.Spec.Gateways.EgressConfig.SDS.Enabled == nil {
		config.Spec.Gateways.EgressConfig.SDS.Enabled = util.BoolPointer(false)
	}
	if len(config.Spec.Gateways.EgressConfig.Ports) == 0 {
		config.Spec.Gateways.EgressConfig.Ports = defaultEgressGatewayPorts
	}
	if config.Spec.Gateways.K8sIngress.Enabled == nil {
		config.Spec.Gateways.K8sIngress.Enabled = util.BoolPointer(false)
	}
	// Mixer config
	if config.Spec.Mixer.Enabled == nil {
		config.Spec.Mixer.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Mixer.Image == nil {
		config.Spec.Mixer.Image = util.StrPointer(defaultMixerImage)
	}
	if config.Spec.Mixer.ReplicaCount == nil {
		config.Spec.Mixer.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Mixer.MinReplicas == nil {
		config.Spec.Mixer.MinReplicas = util.IntPointer(defaultMinReplicas)
	}
	if config.Spec.Mixer.MaxReplicas == nil {
		config.Spec.Mixer.MaxReplicas = util.IntPointer(defaultMaxReplicas)
	}
	if config.Spec.Mixer.ReportBatchMaxEntries == nil {
		config.Spec.Mixer.ReportBatchMaxEntries = util.IntPointer(100)
	}
	if config.Spec.Mixer.ReportBatchMaxTime == nil {
		config.Spec.Mixer.ReportBatchMaxTime = util.StrPointer("1s")
	}
	if config.Spec.Mixer.SessionAffinityEnabled == nil {
		config.Spec.Mixer.SessionAffinityEnabled = util.BoolPointer(false)
	}
	if config.Spec.Mixer.StdioAdapterEnabled == nil {
		config.Spec.Mixer.StdioAdapterEnabled = util.BoolPointer(false)
	}
	// SidecarInjector config
	if config.Spec.SidecarInjector.Enabled == nil {
		config.Spec.SidecarInjector.Enabled = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.AutoInjectionPolicyEnabled == nil {
		config.Spec.SidecarInjector.AutoInjectionPolicyEnabled = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.Image == "" {
		config.Spec.SidecarInjector.Image = defaultSidecarInjectorImage
	}
	if config.Spec.SidecarInjector.ReplicaCount == 0 {
		config.Spec.SidecarInjector.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Enabled == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Enabled = util.BoolPointer(false)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Image == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.Image = defaultInitCNIImage
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.BinDir == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.BinDir = defaultInitCNIBinDir
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir = defaultInitCNIConfDir
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.ExcludeNamespaces == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.ExcludeNamespaces = []string{config.Namespace}
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.LogLevel == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.LogLevel = defaultInitCNILogLevel
	}
	if config.Spec.SidecarInjector.Init.Resources == nil {
		config.Spec.SidecarInjector.Init.Resources = defaultInitResources
	}
	// SDS config
	if config.Spec.SDS.Enabled == nil {
		config.Spec.SDS.Enabled = util.BoolPointer(false)
	}
	if config.Spec.SDS.TokenAudience == "" {
		config.Spec.SDS.TokenAudience = "istio-ca"
	}
	// NodeAgent config
	if config.Spec.NodeAgent.Enabled == nil {
		config.Spec.NodeAgent.Enabled = util.BoolPointer(false)
	}
	if config.Spec.NodeAgent.Image == "" {
		config.Spec.NodeAgent.Image = defaultNodeAgentImage
	}
	if config.Spec.Gateways.IngressConfig.SDS.Image == "" {
		config.Spec.Gateways.IngressConfig.SDS.Image = defaultSDSImage
	}
	if config.Spec.Gateways.EgressConfig.SDS.Image == "" {
		config.Spec.Gateways.EgressConfig.SDS.Image = defaultSDSImage
	}
	// Proxy config
	if config.Spec.Proxy.Image == "" {
		config.Spec.Proxy.Image = defaultProxyImage
	}
	// Proxy Init config
	if config.Spec.ProxyInit.Image == "" {
		config.Spec.ProxyInit.Image = defaultProxyInitImage
	}
	if config.Spec.Proxy.AccessLogFile == nil {
		config.Spec.Proxy.AccessLogFile = util.StrPointer(defaultEnvoyAccessLogFile)
	}
	if config.Spec.Proxy.AccessLogFormat == nil {
		config.Spec.Proxy.AccessLogFormat = util.StrPointer(defaultEnvoyAccessLogFormat)
	}
	if config.Spec.Proxy.AccessLogEncoding == nil {
		config.Spec.Proxy.AccessLogEncoding = util.StrPointer(defaultEnvoyAccessLogEncoding)
	}
	if config.Spec.Proxy.ComponentLogLevel == "" {
		config.Spec.Proxy.ComponentLogLevel = "misc:error"
	}
	if config.Spec.Proxy.LogLevel == "" {
		config.Spec.Proxy.LogLevel = "warning"
	}
	if config.Spec.Proxy.DNSRefreshRate == "" {
		config.Spec.Proxy.DNSRefreshRate = "300s"
	}
	if config.Spec.Proxy.EnvoyStatsD.Enabled == nil {
		config.Spec.Proxy.EnvoyStatsD.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyMetricsService.Enabled == nil {
		config.Spec.Proxy.EnvoyMetricsService.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyAccessLogService.Enabled == nil {
		config.Spec.Proxy.EnvoyAccessLogService.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.ProtocolDetectionTimeout == nil {
		config.Spec.Proxy.ProtocolDetectionTimeout = util.StrPointer("100ms")
	}
	if config.Spec.Proxy.ClusterDomain == "" {
		config.Spec.Proxy.ClusterDomain = "cluster.local"
	}
	// PDB config
	if config.Spec.DefaultPodDisruptionBudget.Enabled == nil {
		config.Spec.DefaultPodDisruptionBudget.Enabled = util.BoolPointer(false)
	}
	// Outbound traffic policy config
	if config.Spec.OutboundTrafficPolicy.Mode == "" {
		config.Spec.OutboundTrafficPolicy.Mode = outboundTrafficPolicyAllowAny
	}
	// Tracing config
	if config.Spec.Tracing.Enabled == nil {
		config.Spec.Tracing.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Tracing.Tracer == "" {
		config.Spec.Tracing.Tracer = TracerTypeZipkin
	}
	if config.Spec.Tracing.Datadog.Address == "" {
		config.Spec.Tracing.Datadog.Address = "$(HOST_IP):8126"
	}
	if config.Spec.Tracing.Zipkin.Address == "" {
		config.Spec.Tracing.Zipkin.Address = fmt.Sprintf(defaultZipkinAddress, config.Namespace)
	}

	// Policy
	if config.Spec.Policy.ChecksEnabled == nil {
		config.Spec.Policy.ChecksEnabled = util.BoolPointer(false)
	}
	if config.Spec.Policy.Enabled == nil {
		config.Spec.Policy.Enabled = config.Spec.Mixer.Enabled
	}
	if config.Spec.Policy.Image == nil {
		config.Spec.Policy.Image = config.Spec.Mixer.Image
	}
	if config.Spec.Policy.ReplicaCount == nil {
		config.Spec.Policy.ReplicaCount = config.Spec.Mixer.ReplicaCount
	}
	if config.Spec.Policy.MinReplicas == nil {
		config.Spec.Policy.MinReplicas = config.Spec.Mixer.MinReplicas
	}
	if config.Spec.Policy.MaxReplicas == nil {
		config.Spec.Policy.MaxReplicas = config.Spec.Mixer.MaxReplicas
	}
	if config.Spec.Policy.Resources == nil {
		config.Spec.Policy.Resources = config.Spec.Mixer.Resources
	}
	if config.Spec.Policy.NodeSelector == nil {
		config.Spec.Policy.NodeSelector = config.Spec.Mixer.NodeSelector
	}
	if config.Spec.Policy.Affinity == nil {
		config.Spec.Policy.Affinity = config.Spec.Mixer.Affinity
	}
	if config.Spec.Policy.Tolerations == nil {
		config.Spec.Policy.Tolerations = config.Spec.Mixer.Tolerations
	}
	if config.Spec.Policy.SessionAffinityEnabled == nil {
		config.Spec.Policy.SessionAffinityEnabled = config.Spec.Mixer.SessionAffinityEnabled
	}

	// Telemetry
	if config.Spec.Telemetry.Enabled == nil {
		config.Spec.Telemetry.Enabled = config.Spec.Mixer.Enabled
	}
	if config.Spec.Telemetry.Image == nil {
		config.Spec.Telemetry.Image = config.Spec.Mixer.Image
	}
	if config.Spec.Telemetry.ReplicaCount == nil {
		config.Spec.Telemetry.ReplicaCount = config.Spec.Mixer.ReplicaCount
	}
	if config.Spec.Telemetry.MinReplicas == nil {
		config.Spec.Telemetry.MinReplicas = config.Spec.Mixer.MinReplicas
	}
	if config.Spec.Telemetry.MaxReplicas == nil {
		config.Spec.Telemetry.MaxReplicas = config.Spec.Mixer.MaxReplicas
	}
	if config.Spec.Telemetry.Resources == nil {
		config.Spec.Telemetry.Resources = config.Spec.Mixer.Resources
	}
	if config.Spec.Telemetry.NodeSelector == nil {
		config.Spec.Telemetry.NodeSelector = config.Spec.Mixer.NodeSelector
	}
	if config.Spec.Telemetry.Affinity == nil {
		config.Spec.Telemetry.Affinity = config.Spec.Mixer.Affinity
	}
	if config.Spec.Telemetry.Tolerations == nil {
		config.Spec.Telemetry.Tolerations = config.Spec.Mixer.Tolerations
	}
	if config.Spec.Telemetry.ReportBatchMaxEntries == nil {
		config.Spec.Telemetry.ReportBatchMaxEntries = config.Spec.Mixer.ReportBatchMaxEntries
	}
	if config.Spec.Telemetry.ReportBatchMaxTime == nil {
		config.Spec.Telemetry.ReportBatchMaxTime = config.Spec.Mixer.ReportBatchMaxTime
	}

	// Multi mesh support
	if config.Spec.MultiMesh == nil {
		config.Spec.MultiMesh = util.BoolPointer(false)
	}

	// Istio CoreDNS for multi mesh support
	if config.Spec.IstioCoreDNS.Enabled == nil {
		config.Spec.IstioCoreDNS.Enabled = util.BoolPointer(false)
	}
	if config.Spec.IstioCoreDNS.Image == "" {
		config.Spec.IstioCoreDNS.Image = defaultCoreDNSImage
	}
	if config.Spec.IstioCoreDNS.PluginImage == "" {
		config.Spec.IstioCoreDNS.PluginImage = defaultCoreDNSPluginImage
	}
	if config.Spec.IstioCoreDNS.ReplicaCount == 0 {
		config.Spec.IstioCoreDNS.ReplicaCount = defaultReplicaCount
	}

	if config.Spec.ImagePullPolicy == "" {
		config.Spec.ImagePullPolicy = defaultImagePullPolicy
	}

	if config.Spec.MeshExpansion == nil {
		config.Spec.MeshExpansion = util.BoolPointer(defaultMeshExpansion)
	}

	if config.Spec.UseMCP == nil {
		config.Spec.UseMCP = util.BoolPointer(true)
	}

	if config.Spec.MixerlessTelemetry == nil {
		config.Spec.MixerlessTelemetry = &MixerlessTelemetryConfiguration{
			Enabled: util.BoolPointer(false),
		}
	}

	if config.Spec.TrustDomain == "" {
		config.Spec.TrustDomain = "cluster.local"
	}
}

func SetRemoteIstioDefaults(remoteconfig *RemoteIstio) {
	if remoteconfig.Spec.IncludeIPRanges == "" {
		remoteconfig.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// SidecarInjector config
	if remoteconfig.Spec.SidecarInjector.ReplicaCount == 0 {
		remoteconfig.Spec.SidecarInjector.ReplicaCount = defaultReplicaCount
	}
}
