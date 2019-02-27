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
	defaultIncludeIPRanges = "*"
	defaultReplicaCount    = 1
	defaultMinReplicas     = 1
	defaultMaxReplicas     = 5
)

func SetDefaults(config *Istio) {
	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}

	// Pilot config
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
	if config.Spec.Citadel.ReplicaCount == 0 {
		config.Spec.Citadel.ReplicaCount = defaultReplicaCount
	}

	// Galley config
	if config.Spec.Galley.ReplicaCount == 0 {
		config.Spec.Galley.ReplicaCount = defaultReplicaCount
	}

	// Gateways config
	if config.Spec.Gateways.ReplicaCount == 0 {
		config.Spec.Gateways.ReplicaCount = defaultReplicaCount
	}
	if config.Spec.Gateways.MinReplicas == 0 {
		config.Spec.Gateways.MinReplicas = defaultMinReplicas
	}
	if config.Spec.Gateways.MaxReplicas == 0 {
		config.Spec.Gateways.MaxReplicas = defaultMaxReplicas
	}

	// Mixer config
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
	if config.Spec.SidecarInjector.ReplicaCount == 0 {
		config.Spec.SidecarInjector.ReplicaCount = defaultReplicaCount
	}
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
	MinReplicas  int32 `json:"minReplicas,omitempty"`
	MaxReplicas  int32 `json:"maxReplicas,omitempty"`
}

// CitadelConfiguration defines config options for Citadel
type CitadelConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
}

// GalleyConfiguration defines config options for Galley
type GalleyConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
}

// GatewaysConfiguration defines config options for Gateways
type GatewaysConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
	MinReplicas  int32 `json:"minReplicas,omitempty"`
	MaxReplicas  int32 `json:"maxReplicas,omitempty"`
}

// MixerConfiguration defines config options for Mixer
type MixerConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
	MinReplicas  int32 `json:"minReplicas,omitempty"`
	MaxReplicas  int32 `json:"maxReplicas,omitempty"`
}

// SidecarInjectorConfiguration defines config options for SidecarInjector
type SidecarInjectorConfiguration struct {
	ReplicaCount int32 `json:"replicaCount,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	MTLS            bool   `json:"mtls"`
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`
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
