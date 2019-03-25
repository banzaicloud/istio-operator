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

	"github.com/banzaicloud/istio-operator/pkg/util"
	apiv1 "k8s.io/api/core/v1"
)

const (
	defaultImageHub                  = "docker.io/istio"
	defaultImageVersion              = "1.1.0"
	defaultPilotImage                = defaultImageHub + "/" + "pilot" + ":" + defaultImageVersion
	defaultCitadelImage              = defaultImageHub + "/" + "citadel" + ":" + defaultImageVersion
	defaultGalleyImage               = defaultImageHub + "/" + "galley" + ":" + defaultImageVersion
	defaultMixerImage                = defaultImageHub + "/" + "mixer" + ":" + defaultImageVersion
	defaultSidecarInjectorImage      = defaultImageHub + "/" + "sidecar_injector" + ":" + defaultImageVersion
	defaultNodeAgentImage            = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultSDSImage                  = defaultImageHub + "/" + "node-agent-k8s" + ":" + defaultImageVersion
	defaultProxyImage                = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyInitImage            = defaultImageHub + "/" + "proxy_init" + ":" + defaultImageVersion
	defaultInitCNIImage              = "gcr.io/istio-release/install-cni:master-latest-daily"
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
)

func SetDefaults(config *Istio) {
	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// Pilot config
	if config.Spec.Pilot.Enabled == nil {
		config.Spec.Pilot.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Pilot.Image == "" {
		config.Spec.Pilot.Image = defaultPilotImage
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
	// Citadel config
	if config.Spec.Citadel.Enabled == nil {
		config.Spec.Citadel.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Citadel.Image == "" {
		config.Spec.Citadel.Image = defaultCitadelImage
	}
	if config.Spec.Citadel.ReplicaCount == 0 {
		config.Spec.Citadel.ReplicaCount = defaultReplicaCount
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
	if config.Spec.Gateways.K8sIngress.Enabled == nil {
		config.Spec.Gateways.K8sIngress.Enabled = util.BoolPointer(false)
	}
	// Mixer config
	if config.Spec.Mixer.Enabled == nil {
		config.Spec.Mixer.Enabled = util.BoolPointer(true)
	}
	if config.Spec.Mixer.Image == "" {
		config.Spec.Mixer.Image = defaultMixerImage
	}
	if config.Spec.Mixer.ReplicaCount == 0 {
		config.Spec.Mixer.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Mixer.MinReplicas == 0 {
		config.Spec.Mixer.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Mixer.MaxReplicas == 0 {
		config.Spec.Mixer.MaxReplicas = defaultMaxReplicas
	}
	// SidecarInjector config
	if config.Spec.SidecarInjector.Enabled == nil {
		config.Spec.SidecarInjector.Enabled = util.BoolPointer(true)
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
	// SDS config
	if config.Spec.SDS.Enabled == nil {
		config.Spec.SDS.Enabled = util.BoolPointer(false)
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
	if config.Spec.Tracing.Zipkin.Address == "" {
		config.Spec.Tracing.Zipkin.Address = fmt.Sprintf(defaultZipkinAddress, config.Namespace)
	}

	if config.Spec.ImagePullPolicy == "" {
		config.Spec.ImagePullPolicy = defaultImagePullPolicy
	}
}

func SetRemoteIstioDefaults(remoteconfig *RemoteIstio) {
	if remoteconfig.Spec.IncludeIPRanges == "" {
		remoteconfig.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// Citadel config
	if remoteconfig.Spec.Citadel.Image == "" {
		remoteconfig.Spec.Citadel.Image = defaultCitadelImage
	}
	if remoteconfig.Spec.Citadel.ReplicaCount == 0 {
		remoteconfig.Spec.Citadel.ReplicaCount = defaultReplicaCount
	}
	// SidecarInjector config
	if remoteconfig.Spec.SidecarInjector.Image == "" {
		remoteconfig.Spec.SidecarInjector.Image = defaultSidecarInjectorImage
	}
	if remoteconfig.Spec.SidecarInjector.ReplicaCount == 0 {
		remoteconfig.Spec.SidecarInjector.ReplicaCount = defaultReplicaCount
	}
}
