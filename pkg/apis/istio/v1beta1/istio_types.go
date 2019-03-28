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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// SDSConfiguration defines Secret Discovery Service config options
type SDSConfiguration struct {
	// If set to true, mTLS certificates for the sidecars will be
	// distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates.
	Enabled bool `json:"enabled,omitempty"`
	// Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
	// key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.
	UdsPath string `json:"udsPath,omitempty"`
	// If set to true, Istio will inject volumes mount for k8s service account JWT,
	// so that K8s API server mounts k8s service account JWT to envoy container, which
	// will be used to generate key/cert eventually.
	// (prerequisite: https://kubernetes.io/docs/concepts/storage/volumes/#projected)
	UseTrustworthyJwt bool `json:"useTrustworthyJwt,omitempty"`
	// If set to true, envoy will fetch normal k8s service account JWT from '/var/run/secrets/kubernetes.io/serviceaccount/token'
	// (https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#accessing-the-api-from-a-pod)
	// and pass to sds server, which will be used to request key/cert eventually
	// this flag is ignored if UseTrustworthyJwt is set
	UseNormalJwt bool `json:"useNormalJwt,omitempty"`
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	Image         string  `json:"image,omitempty"`
	ReplicaCount  int32   `json:"replicaCount,omitempty"`
	MinReplicas   int32   `json:"minReplicas,omitempty"`
	MaxReplicas   int32   `json:"maxReplicas,omitempty"`
	TraceSampling float64 `json:"traceSampling,omitempty"`
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
	IngressConfig GatewayConfiguration    `json:"ingress,omitempty"`
	EgressConfig  GatewayConfiguration    `json:"egress,omitempty"`
	K8sIngress    K8sIngressConfiguration `json:"k8singress,omitempty"`
}

type GatewaySDSConfiguration struct {
	Enabled bool   `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
}

type GatewayConfiguration struct {
	ReplicaCount       int32                   `json:"replicaCount,omitempty"`
	MinReplicas        int32                   `json:"minReplicas,omitempty"`
	MaxReplicas        int32                   `json:"maxReplicas,omitempty"`
	ServiceAnnotations map[string]string       `json:"serviceAnnotations,omitempty"`
	ServiceLabels      map[string]string       `json:"serviceLabels,omitempty"`
	SDS                GatewaySDSConfiguration `json:"sds,omitempty"`
}

type K8sIngressConfiguration struct {
	Enabled bool `json:"enabled,omitempty"`
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
	// If true, sidecar injector will rewrite PodSpec for liveness
	// health check to redirect request to sidecar. This makes liveness check work
	// even when mTLS is enabled.
	RewriteAppHTTPProbe bool `json:"rewriteAppHTTPProbe,omitempty"`
}

// NodeAgentConfiguration defines config options for NodeAgent
type NodeAgentConfiguration struct {
	Enabled bool   `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
}

// ProxyConfiguration defines config options for Proxy
type ProxyConfiguration struct {
	Image string `json:"image,omitempty"`
	// If set to true, istio-proxy container will have privileged securityContext
	Privileged bool `json:"privileged,omitempty"`
	// If set, newly injected sidecars will have core dumps enabled.
	EnableCoreDump bool `json:"enableCoreDump,omitempty"`
}

// ProxyInitConfiguration defines config options for Proxy Init containers
type ProxyInitConfiguration struct {
	Image string `json:"image,omitempty"`
}

// PDBConfiguration holds Pod Disruption Budget related config options
type PDBConfiguration struct {
	Enabled bool `json:"enabled,omitempty"`
}

type OutboundTrafficPolicyConfiguration struct {
	// +kubebuilder:validation:Enum=ALLOW_ANY,REGISTRY_ONLY
	Mode string `json:"mode,omitempty"`
}

// Configuration for Envoy to send trace data to Zipkin/Jaeger.
type ZipkinConfiguration struct {
	// Host:Port for reporting trace data in zipkin format. If not specified, will default to zipkin service (port 9411) in the same namespace as the other istio components.
	// +kubebuilder:validation:Pattern=^[^\:]+:[0-9]{1,5}$
	Address string `json:"address,omitempty"`
}

type TracingConfiguration struct {
	Zipkin ZipkinConfiguration `json:"zipkin,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// ImageHub is the hub to be applied for all images, if otherwise
	// unspecified.  Example: docker.io
	ImageHub string `json:"imageHub,omitempty"`

	// ImageTag is the tag to be applied for all images, if otherwise
	// unspecified.  Example: 1.1.0
	ImageTag string `json:"imageTag,omitempty"`

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

	// If SDS is configured, mTLS certificates for the sidecars will be distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates
	SDS SDSConfiguration `json:"sds,omitempty"`

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

	// NodeAgent configuration options
	NodeAgent NodeAgentConfiguration `json:"nodeAgent,omitempty"`

	// Proxy configuration options
	Proxy ProxyConfiguration `json:"proxy,omitempty"`

	// Proxy Init configuration options
	ProxyInit ProxyInitConfiguration `json:"proxyInit,omitempty"`

	// Whether to restrict the applications namespace the controller manages
	WatchOneNamespace bool `json:"watchOneNamespace,omitempty"`

	// Use the Mesh Control Protocol (MCP) for configuring Mixer and Pilot. Requires galley.
	UseMCP bool `json:"useMCP,omitempty"`

	// Set the default set of namespaces to which services, service entries, virtual services, destination rules should be exported to
	DefaultConfigVisibility string `json:"defaultConfigVisibility,omitempty"`

	// Whether or not to establish watches for adapter-specific CRDs
	WatchAdapterCRDs bool `json:"watchAdapterCRDs,omitempty"`

	// Enable pod disruption budget for the control plane, which is used to ensure Istio control plane components are gradually upgraded or recovered
	DefaultPodDisruptionBudget PDBConfiguration `json:"defaultPodDisruptionBudget,omitempty"`

	// Set the default behavior of the sidecar for handling outbound traffic from the application (ALLOW_ANY or REGISTRY_ONLY)
	OutboundTrafficPolicy OutboundTrafficPolicyConfiguration `json:"outboundTrafficPolicy,omitempty"`

	// Configuration for each of the supported tracers
	Tracing TracingConfiguration `json:"tracing,omitempty"`
}

func (s IstioSpec) GetDefaultConfigVisibility() string {
	if s.DefaultConfigVisibility == "" || s.DefaultConfigVisibility == "." {
		return s.DefaultConfigVisibility
	}
	return "*"
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	Status       ConfigState                             `json:"status,omitempty"`
	ErrorMessage string                                  `json:"errorMessage,omitempty"`
	Resources    map[string][]*unstructured.Unstructured `json:"resources,omitempty"`
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
