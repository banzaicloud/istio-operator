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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultImageVersion         = "1.0.5"
	defaultPilotImage           = "istio/pilot" + ":" + defaultImageVersion
	defaultCitadelImage         = "istio/citadel" + ":" + defaultImageVersion
	defaultGalleyImage          = "istio/galley" + ":" + defaultImageVersion
	defaultMixerImage           = "istio/mixer" + ":" + defaultImageVersion
	defaultSidecarInjectorImage = "istio/sidecar_injector" + ":" + defaultImageVersion
	defaultProxyImage           = "istio/proxyv2" + ":" + defaultImageVersion
	defaultIncludeIPRanges      = "*"
	defaultReplicaCount         = 1
	defaultMinReplicas          = 1
	defaultMaxReplicas          = 5
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
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	Image        string `json:"image,omitempty"`
	ReplicaCount int32  `json:"replicaCount,omitempty"`
	MinReplicas  int32  `json:"minReplicas,omitempty"`
	MaxReplicas  int32  `json:"maxReplicas,omitempty"`
}

// CitadelConfiguration defines config options for Citadel
type CitadelConfiguration struct {
	Image        string `json:"image,omitempty"`
	ReplicaCount int32  `json:"replicaCount,omitempty"`
}

// GalleyConfiguration defines config options for Galley
type GalleyConfiguration struct {
	Image        string `json:"image,omitempty"`
	ReplicaCount int32  `json:"replicaCount,omitempty"`
}

// GatewaysConfiguration defines config options for Gateways
type GatewaysConfiguration struct {
	IngressConfig GatewayConfiguration `json:"ingress,omitempty"`
	EgressConfig  GatewayConfiguration `json:"egress,omitempty"`
}

type GatewayConfiguration struct {
	ReplicaCount       int32             `json:"replicaCount,omitempty"`
	MinReplicas        int32             `json:"minReplicas,omitempty"`
	MaxReplicas        int32             `json:"maxReplicas,omitempty"`
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	ServiceLabels      map[string]string `json:"serviceLabels,omitempty"`
}

// MixerConfiguration defines config options for Mixer
type MixerConfiguration struct {
	Image        string `json:"image,omitempty"`
	ReplicaCount int32  `json:"replicaCount,omitempty"`
	MinReplicas  int32  `json:"minReplicas,omitempty"`
	MaxReplicas  int32  `json:"maxReplicas,omitempty"`
}

// SidecarInjectorConfiguration defines config options for SidecarInjector
type SidecarInjectorConfiguration struct {
	Image        string `json:"image,omitempty"`
	ReplicaCount int32  `json:"replicaCount,omitempty"`
}

// ProxyConfiguration defines config options for Proxy
type ProxyConfiguration struct {
	Image string `json:"image,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// MTLS enables or disables global mTLS
	MTLS bool `json:"mtls"`

	// IncludeIPRanges the range where to capture egress traffic
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`

	// ExcludeIPRanges the range where not to capture egress traffic
	ExcludeIPRanges string `json:"excludeIPRanges,omitempty"`

	// List of namespaces to label with sidecar auto injection enabled
	AutoInjectionNamespaces []string `json:"autoInjectionNamespaces,omitempty"`

	// ControlPlaneSecurityEnabled control plane services are communicating through mTLS
	ControlPlaneSecurityEnabled bool `json:"controlPlaneSecurityEnabled,omitempty"`

	// Pilot configuration options
	Pilot PilotConfiguration `json:"pilot,omitempty"`

	// Citadel configuration options
	Citadel CitadelConfiguration `json:"citadel,omitempty"`

	// Galley configuration options
	Galley GalleyConfiguration `json:"galley,omitempty"`

	// Gateways configuration options
	Gateways GatewaysConfiguration `json:"gateways,omitempty"`

	// Mixer configuration options
	Mixer MixerConfiguration `json:"mixer,omitempty"`

	// SidecarInjector configuration options
	SidecarInjector SidecarInjectorConfiguration `json:"sidecarInjector,omitempty"`

	// Proxy configuration options
	Proxy ProxyConfiguration `json:"proxy,omitempty"`
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	Status       ConfigState
	ErrorMessage string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Istio is the Schema for the istios API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Istio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioSpec   `json:"spec,omitempty"`
	Status IstioStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioList contains a list of Istio
type IstioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Istio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Istio{}, &IstioList{})
}
