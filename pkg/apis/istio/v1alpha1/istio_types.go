package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// Global section
	// Version of Istio to be deployed
	Version string `json:"version,omitempty"`
	// MTLS signals if mTLS should be enabled in the cluster
	MTLS bool `json:"mtls,omitempty"`
	// Proxy related config
	Proxy *ProxySpec `json:"proxy,omitempty"`
	// Gateway related config
	Gateway *GatewaySpec `json:gateway,omitempty`
	// Galley related config
	Galley *GalleySpec `json:galley,omitempty`
	// Mixer related config
	Mixer *MixerSpec `json:mixer,omitempty`
	// Pilot related config
	Pilot *PilotSpec `json:pilot,omitempty`
	// Citadel related config
	Citadel *CitadelSpec `json:citadel,omitempty`
}

// ProxySpec defines configurations related to Envoy sidecars
type ProxySpec struct {
	// Image name of the proxy
	Image string `json:image,omitempty`
	// IncludeIPRanges is the egress capture whitelist
	IncludeIPRanges string `json:includeIpRanges,omitempty`
	// AutoInject controls automatic sidecar injection
	AutoInject bool `json:autoInject,omitempty`
	// AutoInjectNamespaces is a list of namespaces where auto sidecar injection is enabled
	AutoInjectNamespaces []string `json:autoInjectNamespaces,omitempty`
}

// GatewaySpec defines Istio gateway related configurations
type GatewaySpec struct {
	// Enabled when set to true, a pair of Ingress and Egress gateways will be created for the mesh
	Enabled bool `json:enabled,omitempty`
}

// GalleySpec defines Galley related configurations
type GalleySpec struct {
	// Enabled switches Galley on/off
	Enabled bool `json:enabled,omitempty`
}

// MixerSpec defines Mixer related configurations
type MixerSpec struct {
	// Enabled switches Mixer on/off
	Enabled bool `json:enabled,omitempty`
}

// PilotSpec defines Pilot related configurations
type PilotSpec struct {
	// Enabled switches Pilot on/off
	Enabled bool `json:enabled,omitempty`
}

// CitadelSpec defines the security related configurations
type CitadelSpec struct {
	// Enabled switches Citadel on/off
	Enabled bool `json:enabled,omitempty`
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	// Galley describes the Istio Galley component
	Galley *ComponentStatus `json:"galley"`
	// Mixer describes the Istio Mixer component
	Mixer *ComponentStatus `json:"mixer"`
	// Pilot describes the Istio Pilot component
	Pilot *ComponentStatus `json:"pilot"`
	// Citadel describes the Istio Citadel component
	Citadel *ComponentStatus `json:"citadel"`
	// Gateway describes the Istio ingress and egress gateways
	Gateway *GatewayStatus `json:"gateway"`
}

// ComponentStatus describes an Istio component's replicas
type ComponentStatus struct {
	// Total number of non-terminated pods
	Replicas int32 `json:"replicas"`
	// Total number of available pods
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

// Gateway describes the Istio gateways in the mesh
type GatewayStatus struct {
	// Ingress is the ingress gateway's status
	Ingress *ComponentStatus `json:"ingress"`
	// Egress is the egress gateway's status
	Egress *ComponentStatus `json:"egress"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Istio is the Schema for the istios API
// +k8s:openapi-gen=true
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
	Items []Istio   `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Istio{}, &IstioList{})
}
