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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	banzaiImageHub                         = "ghcr.io/banzaicloud"
	banzaiImageVersion                     = "1.7.6-bzc"
	defaultImageHub                        = "docker.io/istio"
	defaultImageVersion                    = "1.7.6"
	defaultLogLevel                        = "default:info"
	defaultMeshPolicy                      = PERMISSIVE
	defaultPilotImage                      = defaultImageHub + "/" + "pilot" + ":" + defaultImageVersion
	defaultCitadelImage                    = defaultImageHub + "/" + "citadel" + ":" + defaultImageVersion
	defaultGalleyImage                     = defaultImageHub + "/" + "galley" + ":" + defaultImageVersion
	defaultMixerImage                      = defaultImageHub + "/" + "mixer" + ":" + defaultImageVersion
	defaultSidecarInjectorImage            = banzaiImageHub + "/" + "istio-sidecar-injector" + ":" + banzaiImageVersion
	defaultNodeAgentImage                  = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultSDSImage                        = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultProxyImage                      = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyInitImage                  = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyCoreDumpImage              = "busybox"
	defaultProxyCoreDumpDirectory          = "/var/lib/istio/data"
	defaultInitCNIImage                    = defaultImageHub + "/" + "install-cni:" + defaultImageVersion
	defaultCoreDNSImage                    = "coredns/coredns:1.6.2"
	defaultCoreDNSPluginImage              = defaultImageHub + "/coredns-plugin:0.2-istio-1.1"
	defaultIncludeIPRanges                 = "*"
	defaultReplicaCount                    = 1
	defaultMinReplicas                     = 1
	defaultMaxReplicas                     = 5
	defaultTraceSampling                   = 1.0
	defaultIngressGatewayServiceType       = apiv1.ServiceTypeLoadBalancer
	defaultEgressGatewayServiceType        = apiv1.ServiceTypeClusterIP
	defaultMeshExpansionGatewayServiceType = apiv1.ServiceTypeLoadBalancer
	outboundTrafficPolicyAllowAny          = "ALLOW_ANY"
	defaultZipkinAddress                   = "zipkin.%s:9411"
	defaultInitCNIBinDir                   = "/opt/cni/bin"
	defaultInitCNIConfDir                  = "/etc/cni/net.d"
	defaultInitCNILogLevel                 = "info"
	defaultInitCNIContainerName            = "istio-validation"
	defaultInitCNIBrokenPodLabelKey        = "cni.istio.io/uninitialized"
	defaultInitCNIBrokenPodLabelValue      = "true"
	defaultImagePullPolicy                 = "IfNotPresent"
	defaultEnvoyAccessLogFile              = "/dev/stdout"
	defaultEnvoyAccessLogFormat            = ""
	defaultEnvoyAccessLogEncoding          = "TEXT"
	defaultClusterName                     = "Kubernetes"
	defaultNetworkName                     = "local-network"
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

var defaultSecurityContext = &apiv1.SecurityContext{
	RunAsUser:                util.Int64Pointer(1337),
	RunAsGroup:               util.Int64Pointer(1337),
	RunAsNonRoot:             util.BoolPointer(true),
	Privileged:               util.BoolPointer(false),
	AllowPrivilegeEscalation: util.BoolPointer(false),
	Capabilities: &apiv1.Capabilities{
		Drop: []apiv1.Capability{"ALL"},
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

const ProxyStatusPort = 15020
const PortStatusPortNumber = 15021
const PortStatusPortName = "status-port"

var defaultIngressGatewayPorts = []ServicePort{}
var defaultEgressGatewayPorts = []ServicePort{}
var defaultMeshExpansionGatewayPorts = []ServicePort{}

// SetDefaults used to support generic defaulter interface
func (config *Istio) SetDefaults() {
	SetDefaults(config)
}

func SetDefaults(config *Istio) {
	// MeshPolicy config
	if config.Spec.MeshPolicy.MTLSMode == "" {
		if util.PointerToBool(config.Spec.MTLS) {
			config.Spec.MeshPolicy.MTLSMode = STRICT
		} else {
			config.Spec.MeshPolicy.MTLSMode = defaultMeshPolicy
		}
	}

	if config.Spec.ClusterName == "" {
		config.Spec.ClusterName = defaultClusterName
	}

	if config.Spec.NetworkName == "" {
		config.Spec.NetworkName = defaultNetworkName
	}

	if config.Spec.AutoMTLS == nil {
		config.Spec.AutoMTLS = util.BoolPointer(true)
	}

	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	if config.Spec.MountMtlsCerts == nil {
		config.Spec.MountMtlsCerts = util.BoolPointer(false)
	}
	if config.Spec.Logging.Level == nil {
		config.Spec.Logging.Level = util.StrPointer(defaultLogLevel)
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

	// Istiod config
	if config.Spec.Istiod.Enabled == nil {
		config.Spec.Istiod.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Istiod.EnableAnalysis == nil {
		config.Spec.Istiod.EnableAnalysis = util.BoolPointer(false)
	}
	if config.Spec.Istiod.EnableStatus == nil {
		config.Spec.Istiod.EnableStatus = util.BoolPointer(false)
	}

	// Pilot config
	if config.Spec.Pilot.Enabled == nil {
		config.Spec.Pilot.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Pilot.Image == nil {
		config.Spec.Pilot.Image = util.StrPointer(defaultPilotImage)
	}
	if config.Spec.Pilot.Sidecar == nil {
		config.Spec.Pilot.Sidecar = util.BoolPointer(true)
	}
	if config.Spec.Pilot.ReplicaCount == nil {
		config.Spec.Pilot.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Pilot.MinReplicas == nil {
		config.Spec.Pilot.MinReplicas = util.IntPointer(defaultMinReplicas)
	}
	if config.Spec.Pilot.MaxReplicas == nil {
		config.Spec.Pilot.MaxReplicas = util.IntPointer(defaultMaxReplicas)
	}
	if config.Spec.Pilot.TraceSampling == 0 {
		config.Spec.Pilot.TraceSampling = defaultTraceSampling
	}
	if config.Spec.Pilot.EnableProtocolSniffingOutbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingOutbound = util.BoolPointer(true)
	}
	if config.Spec.Pilot.EnableProtocolSniffingInbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingInbound = util.BoolPointer(true)
	}
	if config.Spec.Pilot.CertProvider == "" {
		config.Spec.Pilot.CertProvider = PilotCertProviderTypeIstiod
	}
	if config.Spec.Pilot.SecurityContext == nil {
		config.Spec.Pilot.SecurityContext = defaultSecurityContext
	}
	// Citadel config
	if config.Spec.Citadel.Enabled == nil {
		config.Spec.Citadel.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Citadel.Image == nil {
		config.Spec.Citadel.Image = util.StrPointer(defaultCitadelImage)
	}
	if config.Spec.Citadel.EnableNamespacesByDefault == nil {
		config.Spec.Citadel.EnableNamespacesByDefault = util.BoolPointer(true)
	}
	// Galley config
	if config.Spec.Galley.Enabled == nil {
		config.Spec.Galley.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Galley.Image == nil {
		config.Spec.Galley.Image = util.StrPointer(defaultGalleyImage)
	}
	if config.Spec.Galley.ReplicaCount == nil {
		config.Spec.Galley.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Galley.ConfigValidation == nil {
		config.Spec.Galley.ConfigValidation = util.BoolPointer(true)
	}
	if config.Spec.Galley.EnableServiceDiscovery == nil {
		config.Spec.Galley.EnableServiceDiscovery = util.BoolPointer(false)
	}
	if config.Spec.Galley.EnableAnalysis == nil {
		config.Spec.Galley.EnableAnalysis = util.BoolPointer(false)
	}
	// Gateways config
	if config.Spec.Gateways.Enabled == nil {
		config.Spec.Gateways.Enabled = util.BoolPointer(true)
	}
	ingress := &config.Spec.Gateways.Ingress
	ingress.MeshGatewayConfiguration.SetDefaults()
	if ingress.ServiceType == "" {
		ingress.ServiceType = defaultIngressGatewayServiceType
	}
	if len(ingress.Ports) == 0 {
		ingress.Ports = defaultIngressGatewayPorts
	}
	if ingress.CreateOnly == nil {
		ingress.CreateOnly = util.BoolPointer(false)
	}
	if ingress.Enabled == nil {
		ingress.Enabled = util.BoolPointer(false)
	}
	egress := &config.Spec.Gateways.Egress
	egress.MeshGatewayConfiguration.SetDefaults()
	if egress.ServiceType == "" {
		egress.ServiceType = defaultEgressGatewayServiceType
	}
	if len(egress.Ports) == 0 {
		egress.Ports = defaultEgressGatewayPorts
	}
	if egress.CreateOnly == nil {
		egress.CreateOnly = util.BoolPointer(false)
	}
	if egress.Enabled == nil {
		egress.Enabled = util.BoolPointer(false)
	}
	mexpgw := &config.Spec.Gateways.MeshExpansion
	mexpgw.MeshGatewayConfiguration.SetDefaults()
	if mexpgw.ServiceType == "" {
		mexpgw.ServiceType = defaultMeshExpansionGatewayServiceType
	}
	if len(mexpgw.Ports) == 0 {
		mexpgw.Ports = defaultMeshExpansionGatewayPorts
	}
	if mexpgw.CreateOnly == nil {
		mexpgw.CreateOnly = util.BoolPointer(false)
	}
	if mexpgw.Enabled == nil {
		mexpgw.Enabled = config.Spec.MeshExpansion
	}
	if config.Spec.Gateways.K8sIngress.Enabled == nil {
		config.Spec.Gateways.K8sIngress.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Gateways.K8sIngress.EnableHttps == nil {
		config.Spec.Gateways.K8sIngress.EnableHttps = util.BoolPointer(false)
	}
	// Mixer config
	if config.Spec.Mixer.Enabled == nil {
		config.Spec.Mixer.Enabled = util.BoolPointer(false)
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
	if config.Spec.Mixer.SecurityContext == nil {
		config.Spec.Mixer.SecurityContext = defaultSecurityContext
	}
	// SidecarInjector config
	if config.Spec.SidecarInjector.Enabled == nil {
		config.Spec.SidecarInjector.Enabled = util.BoolPointer(false)
	}
	if config.Spec.SidecarInjector.AutoInjectionPolicyEnabled == nil {
		config.Spec.SidecarInjector.AutoInjectionPolicyEnabled = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.Image == nil {
		config.Spec.SidecarInjector.Image = util.StrPointer(defaultSidecarInjectorImage)
	}
	if config.Spec.SidecarInjector.ReplicaCount == nil {
		config.Spec.SidecarInjector.ReplicaCount = util.IntPointer(defaultReplicaCount)
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
	if config.Spec.SidecarInjector.InitCNIConfiguration.Chained == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Chained = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.RewriteAppHTTPProbe == nil {
		config.Spec.SidecarInjector.RewriteAppHTTPProbe = util.BoolPointer(true)
	}
	// Wasm Config
	if config.Spec.ProxyWasm.Enabled == nil {
		config.Spec.ProxyWasm.Enabled = util.BoolPointer(false)
	}
	// CNI repair config
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Enabled == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Enabled = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Hub == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Hub = util.StrPointer("")
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Tag == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Tag = util.StrPointer("")
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.LabelPods == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.LabelPods = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.DeletePods == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.DeletePods = util.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.InitContainerName == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.InitContainerName = util.StrPointer(defaultInitCNIContainerName)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelKey == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelKey = util.StrPointer(defaultInitCNIBrokenPodLabelKey)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelValue == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelValue = util.StrPointer(defaultInitCNIBrokenPodLabelValue)
	}
	if config.Spec.SidecarInjector.Init.Resources == nil {
		config.Spec.SidecarInjector.Init.Resources = defaultInitResources
	}
	if config.Spec.SidecarInjector.SecurityContext == nil {
		config.Spec.SidecarInjector.SecurityContext = defaultSecurityContext
	}
	// SDS config
	if config.Spec.SDS.Enabled == nil {
		config.Spec.SDS.Enabled = util.BoolPointer(false)
	}
	if config.Spec.SDS.TokenAudience == "" {
		config.Spec.SDS.TokenAudience = "istio-ca"
	}
	if config.Spec.SDS.UdsPath == "" {
		config.Spec.SDS.UdsPath = "unix:/var/run/sds/uds_path"
	}
	// NodeAgent config
	if config.Spec.NodeAgent.Enabled == nil {
		config.Spec.NodeAgent.Enabled = util.BoolPointer(false)
	}
	if config.Spec.NodeAgent.Image == nil {
		config.Spec.NodeAgent.Image = util.StrPointer(defaultNodeAgentImage)
	}

	if config.Spec.Gateways.Ingress.SDS.Image == "" {
		config.Spec.Gateways.Ingress.SDS.Image = defaultSDSImage
	}
	if config.Spec.Gateways.Egress.SDS.Image == "" {
		config.Spec.Gateways.Egress.SDS.Image = defaultSDSImage
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
	if config.Spec.Proxy.HoldApplicationUntilProxyStarts == nil {
		config.Spec.Proxy.HoldApplicationUntilProxyStarts = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyStatsD.Enabled == nil {
		config.Spec.Proxy.EnvoyStatsD.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyMetricsService.Enabled == nil {
		config.Spec.Proxy.EnvoyMetricsService.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyMetricsService.TLSSettings == nil {
		config.Spec.Proxy.EnvoyMetricsService.TLSSettings = &TLSSettings{
			Mode: "DISABLE",
		}
	}
	if config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive == nil {
		config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive = &TCPKeepalive{
			Probes:   3,
			Time:     "10s",
			Interval: "10s",
		}
	}
	if config.Spec.Proxy.EnvoyAccessLogService.Enabled == nil {
		config.Spec.Proxy.EnvoyAccessLogService.Enabled = util.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyAccessLogService.TLSSettings == nil {
		config.Spec.Proxy.EnvoyAccessLogService.TLSSettings = &TLSSettings{
			Mode: "DISABLE",
		}
	}
	if config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive == nil {
		config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive = &TCPKeepalive{
			Probes:   3,
			Time:     "10s",
			Interval: "10s",
		}
	}
	if config.Spec.Proxy.ProtocolDetectionTimeout == nil {
		config.Spec.Proxy.ProtocolDetectionTimeout = util.StrPointer("5s")
	}
	if config.Spec.Proxy.ClusterDomain == "" {
		config.Spec.Proxy.ClusterDomain = "cluster.local"
	}
	if config.Spec.Proxy.EnableCoreDump == nil {
		config.Spec.Proxy.EnableCoreDump = util.BoolPointer(false)
	}
	if config.Spec.Proxy.CoreDumpImage == "" {
		config.Spec.Proxy.CoreDumpImage = defaultProxyCoreDumpImage
	}
	if config.Spec.Proxy.CoreDumpDirectory == "" {
		config.Spec.Proxy.CoreDumpDirectory = defaultProxyCoreDumpDirectory
	}
	if config.Spec.Proxy.SecurityContext == nil {
		config.Spec.Proxy.SecurityContext = defaultSecurityContext
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
	if config.Spec.Tracing.Zipkin.Address == "" {
		config.Spec.Tracing.Zipkin.Address = fmt.Sprintf(defaultZipkinAddress, config.Namespace)
	}
	if config.Spec.Tracing.Tracer == TracerTypeDatadog {
		if config.Spec.Tracing.Datadog.Address == "" {
			config.Spec.Tracing.Datadog.Address = "$(HOST_IP):8126"
		}
	}
	if config.Spec.Tracing.Tracer == TracerTypeStackdriver {
		if config.Spec.Tracing.Strackdriver.Debug == nil {
			config.Spec.Tracing.Strackdriver.Debug = util.BoolPointer(false)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes = util.IntPointer(200)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations = util.IntPointer(200)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents = util.IntPointer(200)
		}
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
	if config.Spec.Policy.SecurityContext == nil {
		config.Spec.Policy.SecurityContext = defaultSecurityContext
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
	if config.Spec.Telemetry.SessionAffinityEnabled == nil {
		config.Spec.Telemetry.SessionAffinityEnabled = config.Spec.Mixer.SessionAffinityEnabled
	}
	if config.Spec.Telemetry.SecurityContext == nil {
		config.Spec.Telemetry.SecurityContext = defaultSecurityContext
	}

	// Multi mesh support
	if config.Spec.MultiMesh == nil {
		config.Spec.MultiMesh = util.BoolPointer(false)
	}

	// Istio CoreDNS for multi mesh support
	if config.Spec.IstioCoreDNS.Enabled == nil {
		config.Spec.IstioCoreDNS.Enabled = util.BoolPointer(false)
	}
	if config.Spec.IstioCoreDNS.Image == nil {
		config.Spec.IstioCoreDNS.Image = util.StrPointer(defaultCoreDNSImage)
	}
	if config.Spec.IstioCoreDNS.PluginImage == "" {
		config.Spec.IstioCoreDNS.PluginImage = defaultCoreDNSPluginImage
	}
	if config.Spec.IstioCoreDNS.ReplicaCount == nil {
		config.Spec.IstioCoreDNS.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if config.Spec.IstioCoreDNS.MinReplicas == nil {
		config.Spec.IstioCoreDNS.MinReplicas = util.IntPointer(defaultMinReplicas)
	}
	if config.Spec.IstioCoreDNS.MaxReplicas == nil {
		config.Spec.IstioCoreDNS.MaxReplicas = util.IntPointer(defaultMaxReplicas)
	}
	if config.Spec.IstioCoreDNS.SecurityContext == nil {
		config.Spec.IstioCoreDNS.SecurityContext = defaultSecurityContext
	}

	if config.Spec.ImagePullPolicy == "" {
		config.Spec.ImagePullPolicy = defaultImagePullPolicy
	}

	if config.Spec.MeshExpansion == nil {
		config.Spec.MeshExpansion = util.BoolPointer(false)
	}

	if config.Spec.UseMCP == nil {
		config.Spec.UseMCP = util.BoolPointer(false)
	}

	if config.Spec.MixerlessTelemetry == nil {
		config.Spec.MixerlessTelemetry = &MixerlessTelemetryConfiguration{
			Enabled: util.BoolPointer(true),
		}
	}

	if config.Spec.TrustDomain == "" {
		config.Spec.TrustDomain = "cluster.local"
	}

	if config.Spec.Proxy.UseMetadataExchangeFilter == nil {
		config.Spec.Proxy.UseMetadataExchangeFilter = util.BoolPointer(false)
	}

	if config.Spec.JWTPolicy == "" {
		config.Spec.JWTPolicy = JWTPolicyFirstPartyJWT
	}

	if config.Spec.ControlPlaneAuthPolicy == "" {
		config.Spec.ControlPlaneAuthPolicy = ControlPlaneAuthPolicyMTLS
	}

	if config.Spec.ImagePullSecrets == nil {
		config.Spec.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
	}
}

func SetRemoteIstioDefaults(remoteconfig *RemoteIstio) {
	if remoteconfig.Spec.IncludeIPRanges == "" {
		remoteconfig.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// SidecarInjector config
	if remoteconfig.Spec.SidecarInjector.ReplicaCount == nil {
		remoteconfig.Spec.SidecarInjector.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if remoteconfig.Spec.Proxy.UseMetadataExchangeFilter == nil {
		remoteconfig.Spec.Proxy.UseMetadataExchangeFilter = util.BoolPointer(false)
	}
}
