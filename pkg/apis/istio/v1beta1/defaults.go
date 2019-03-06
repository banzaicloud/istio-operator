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

const (
	defaultImageHub             = "gcr.io/istio-release"
	defaultImageVersion         = "release-1.1-latest-daily"
	defaultPilotImage           = defaultImageHub + "/" + "pilot" + ":" + defaultImageVersion
	defaultCitadelImage         = defaultImageHub + "/" + "citadel" + ":" + defaultImageVersion
	defaultGalleyImage          = defaultImageHub + "/" + "galley" + ":" + defaultImageVersion
	defaultMixerImage           = defaultImageHub + "/" + "mixer" + ":" + defaultImageVersion
	defaultSidecarInjectorImage = defaultImageHub + "/" + "sidecar_injector" + ":" + defaultImageVersion
	defaultProxyImage           = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyInitImage       = defaultImageHub + "/" + "proxy_init" + ":" + defaultImageVersion
	defaultIncludeIPRanges      = "*"
	defaultReplicaCount         = 1
	defaultMinReplicas          = 1
	defaultMaxReplicas          = 5
	defaultTraceSampling        = 1.0
)

func SetDefaults(config *Istio) {
	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// Pilot config
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
	if config.Spec.Citadel.Image == "" {
		config.Spec.Citadel.Image = defaultCitadelImage
	}
	if config.Spec.Citadel.ReplicaCount == 0 {
		config.Spec.Citadel.ReplicaCount = defaultReplicaCount
	}
	// Galley config
	if config.Spec.Galley.Image == "" {
		config.Spec.Galley.Image = defaultGalleyImage
	}
	if config.Spec.Galley.ReplicaCount == 0 {
		config.Spec.Galley.ReplicaCount = defaultReplicaCount
	}
	// Gateways config
	if config.Spec.Gateways.IngressConfig.ReplicaCount == 0 {
		config.Spec.Gateways.IngressConfig.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Gateways.IngressConfig.MinReplicas == 0 {
		config.Spec.Gateways.IngressConfig.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Gateways.IngressConfig.MaxReplicas == 0 {
		config.Spec.Gateways.IngressConfig.MaxReplicas = defaultMaxReplicas
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
	// Mixer config
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
	if config.Spec.SidecarInjector.Image == "" {
		config.Spec.SidecarInjector.Image = defaultSidecarInjectorImage
	}
	if config.Spec.SidecarInjector.ReplicaCount == 0 {
		config.Spec.SidecarInjector.ReplicaCount = defaultReplicaCount
	}
	// Proxy config
	if config.Spec.Proxy.Image == "" {
		config.Spec.Proxy.Image = defaultProxyImage
	}
	// Proxy Init config
	if config.Spec.ProxyInit.Image == "" {
		config.Spec.ProxyInit.Image = defaultProxyInitImage
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
