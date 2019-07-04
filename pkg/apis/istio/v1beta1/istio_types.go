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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const supportedIstioMinorVersionRegex = "^1.2"

// IstioVersion stores the intended Istio version
type IstioVersion string

// SDSConfiguration defines Secret Discovery Service config options
type SDSConfiguration struct {
	// If set to true, mTLS certificates for the sidecars will be
	// distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates.
	Enabled *bool `json:"enabled,omitempty"`
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

	CustomTokenDirectory string `json:"customTokenDirectory,omitempty"`
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	Enabled       *bool                        `json:"enabled,omitempty"`
	Image         string                       `json:"image,omitempty"`
	Sidecar       *bool                        `json:"sidecar,omitempty"`
	ReplicaCount  int32                        `json:"replicaCount,omitempty"`
	MinReplicas   int32                        `json:"minReplicas,omitempty"`
	MaxReplicas   int32                        `json:"maxReplicas,omitempty"`
	TraceSampling float32                      `json:"traceSampling,omitempty"`
	Resources     *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector  map[string]string            `json:"nodeSelector,omitempty"`
	Affinity      *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations   []corev1.Toleration          `json:"tolerations,omitempty"`
}

// CitadelConfiguration defines config options for Citadel
type CitadelConfiguration struct {
	Enabled      *bool  `json:"enabled,omitempty"`
	Image        string `json:"image,omitempty"`
	CASecretName string `json:"caSecretName,omitempty"`
	// Enable health checking on the Citadel CSR signing API. https://istio.io/docs/tasks/security/health-check/
	HealthCheck *bool `json:"healthCheck,omitempty"`
	// For the workloads running in Kubernetes, the lifetime of their Istio certificates is controlled by the workload-cert-ttl flag on Citadel. The default value is 90 days. This value should be no greater than max-workload-cert-ttl of Citadel.
	WorkloadCertTTL string `json:"workloadCertTTL,omitempty"`
	// Citadel uses a flag max-workload-cert-ttl to control the maximum lifetime for Istio certificates issued to workloads. The default value is 90 days. If workload-cert-ttl on Citadel or node agent is greater than max-workload-cert-ttl, Citadel will fail issuing the certificate.
	MaxWorkloadCertTTL string                       `json:"maxWorkloadCertTTL,omitempty"`
	Resources          *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector       map[string]string            `json:"nodeSelector,omitempty"`
	Affinity           *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations        []corev1.Toleration          `json:"tolerations,omitempty"`
}

// GalleyConfiguration defines config options for Galley
type GalleyConfiguration struct {
	Enabled      *bool                        `json:"enabled,omitempty"`
	Image        string                       `json:"image,omitempty"`
	ReplicaCount int32                        `json:"replicaCount,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector map[string]string            `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration          `json:"tolerations,omitempty"`
}

// GatewaysConfiguration defines config options for Gateways
type GatewaysConfiguration struct {
	Enabled       *bool                   `json:"enabled,omitempty"`
	IngressConfig GatewayConfiguration    `json:"ingress,omitempty"`
	EgressConfig  GatewayConfiguration    `json:"egress,omitempty"`
	K8sIngress    K8sIngressConfiguration `json:"k8singress,omitempty"`
}

type GatewaySDSConfiguration struct {
	Enabled   *bool                        `json:"enabled,omitempty"`
	Image     string                       `json:"image,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type GatewayConfiguration struct {
	Enabled      *bool `json:"enabled,omitempty"`
	ReplicaCount int32 `json:"replicaCount,omitempty"`
	MinReplicas  int32 `json:"minReplicas,omitempty"`
	MaxReplicas  int32 `json:"maxReplicas,omitempty"`
	// +kubebuilder:validation:Enum=ClusterIP,NodePort,LoadBalancer
	ServiceType          corev1.ServiceType           `json:"serviceType,omitempty"`
	LoadBalancerIP       string                       `json:"loadBalancerIP,omitempty"`
	ServiceAnnotations   map[string]string            `json:"serviceAnnotations,omitempty"`
	ServiceLabels        map[string]string            `json:"serviceLabels,omitempty"`
	SDS                  GatewaySDSConfiguration      `json:"sds,omitempty"`
	Resources            *corev1.ResourceRequirements `json:"resources,omitempty"`
	Ports                []corev1.ServicePort         `json:"ports,omitempty"`
	ApplicationPorts     string                       `json:"applicationPorts,omitempty"`
	RequestedNetworkView string                       `json:"requestedNetworkView,omitempty"`
	NodeSelector         map[string]string            `json:"nodeSelector,omitempty"`
	Affinity             *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations          []corev1.Toleration          `json:"tolerations,omitempty"`
}

type K8sIngressConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// MixerConfiguration defines config options for Mixer
type MixerConfiguration struct {
	Enabled      *bool                        `json:"enabled,omitempty"`
	Image        string                       `json:"image,omitempty"`
	ReplicaCount int32                        `json:"replicaCount,omitempty"`
	MinReplicas  int32                        `json:"minReplicas,omitempty"`
	MaxReplicas  int32                        `json:"maxReplicas,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector map[string]string            `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration          `json:"tolerations,omitempty"`
	// Turn it on if you use mixer that supports multi cluster telemetry
	MultiClusterSupport *bool `json:"multiClusterSupport,omitempty"`
}

// InitCNIConfiguration defines config for the sidecar proxy init CNI plugin
type InitCNIConfiguration struct {
	// If true, the privileged initContainer istio-init is not needed to perform the traffic redirect
	// settings for the istio-proxy
	Enabled *bool  `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
	// Must be the same as the environment’s --cni-bin-dir setting (kubelet parameter)
	BinDir string `json:"binDir,omitempty"`
	// Must be the same as the environment’s --cni-conf-dir setting (kubelet parameter)
	ConfDir string `json:"confDir,omitempty"`
	// List of namespaces to exclude from Istio pod check
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
	// Logging level for CNI binary
	LogLevel string           `json:"logLevel,omitempty"`
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

// SidecarInjectorInitConfiguration defines options for init containers in the sidecar
type SidecarInjectorInitConfiguration struct {
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// SidecarInjectorConfiguration defines config options for SidecarInjector
type SidecarInjectorConfiguration struct {
	Enabled              *bool                            `json:"enabled,omitempty"`
	Image                string                           `json:"image,omitempty"`
	ReplicaCount         int32                            `json:"replicaCount,omitempty"`
	Resources            *corev1.ResourceRequirements     `json:"resources,omitempty"`
	Init                 SidecarInjectorInitConfiguration `json:"init,omitempty"`
	InitCNIConfiguration InitCNIConfiguration             `json:"initCNIConfiguration,omitempty"`
	// If true, sidecar injector will rewrite PodSpec for liveness
	// health check to redirect request to sidecar. This makes liveness check work
	// even when mTLS is enabled.
	RewriteAppHTTPProbe bool `json:"rewriteAppHTTPProbe,omitempty"`
	// This controls the 'policy' in the sidecar injector
	AutoInjectionPolicyEnabled *bool `json:"autoInjectionPolicyEnabled,omitempty"`
	// This controls whether the webhook looks for namespaces for injection enabled or disabled
	EnableNamespacesByDefault *bool `json:"enableNamespacesByDefault,omitempty"`
	// NeverInjectSelector: Refuses the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	// Takes precedence over AlwaysInjectSelector.
	NeverInjectSelector []metav1.LabelSelector `json:"neverInjectSelector,omitempty"`
	// AlwaysInjectSelector: Forces the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	AlwaysInjectSelector []metav1.LabelSelector `json:"alwaysInjectSelector,omitempty"`
	NodeSelector         map[string]string      `json:"nodeSelector,omitempty"`
	Affinity             *corev1.Affinity       `json:"affinity,omitempty"`
	Tolerations          []corev1.Toleration    `json:"tolerations,omitempty"`
}

// NodeAgentConfiguration defines config options for NodeAgent
type NodeAgentConfiguration struct {
	Enabled      *bool                        `json:"enabled,omitempty"`
	Image        string                       `json:"image,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector map[string]string            `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration          `json:"tolerations,omitempty"`
}

// ProxyConfiguration defines config options for Proxy
type ProxyConfiguration struct {
	Image string `json:"image,omitempty"`
	// If set to true, istio-proxy container will have privileged securityContext
	Privileged bool `json:"privileged,omitempty"`
	// If set, newly injected sidecars will have core dumps enabled.
	EnableCoreDump bool `json:"enableCoreDump,omitempty"`
	// Log level for proxy, applies to gateways and sidecars. If left empty, "warning" is used.
	// Expected values are: trace|debug|info|warning|error|critical|off
	// +kubebuilder:validation:Enum=trace,debug,info,warning,error,critical,off
	LogLevel string `json:"logLevel,omitempty"`
	// Per Component log level for proxy, applies to gateways and sidecars. If a component level is
	// not set, then the "LogLevel" will be used. If left empty, "misc:error" is used.
	ComponentLogLevel string `json:"componentLogLevel,omitempty"`
	// Configure the DNS refresh rate for Envoy cluster of type STRICT_DNS
	// This must be given it terms of seconds. For example, 300s is valid but 5m is invalid.
	// +kubebuilder:validation:Pattern=^[0-9]{1,5}s$
	DNSRefreshRate string `json:"dnsRefreshRate,omitempty"`

	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ProxyInitConfiguration defines config options for Proxy Init containers
type ProxyInitConfiguration struct {
	Image string `json:"image,omitempty"`
}

// PDBConfiguration holds Pod Disruption Budget related config options
type PDBConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
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

// Configuration for Envoy to send trace data to Lightstep
type LightstepConfiguration struct {
	// the <host>:<port> of the satellite pool
	// +kubebuilder:validation:Pattern=^[^\:]+:[0-9]{1,5}$
	Address string `json:"address,omitempty"`
	// required for sending data to the pool
	AccessToken string `json:"accessToken,omitempty"`
	// specifies whether data should be sent with TLS
	Secure bool `json:"secure,omitempty"`
	// the path to the file containing the cacert to use when verifying TLS. If secure is true, this is
	// required. If a value is specified then a secret called "lightstep.cacert" must be created in the destination
	// namespace with the key matching the base of the provided cacertPath and the value being the cacert itself.
	CacertPath string `json:"cacertPath,omitempty"`
}

// Configuration for Envoy to send trace data to Datadog
type DatadogConfiugration struct {
	// Host:Port for submitting traces to the Datadog agent.
	// +kubebuilder:validation:Pattern=^[^\:]+:[0-9]{1,5}$
	Address string `json:"address,omitempty"`
}

type TracerType string

const (
	TracerTypeZipkin    TracerType = "zipkin"
	TracerTypeLightstep TracerType = "lightstep"
	TracerTypeDatadog   TracerType = "datadog"
)

type TracingConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Enum=zipkin,lightstep,datadog
	Tracer    TracerType             `json:"tracer,omitempty"`
	Zipkin    ZipkinConfiguration    `json:"zipkin,omitempty"`
	Lightstep LightstepConfiguration `json:"lightstep,omitempty"`
	Datadog   DatadogConfiugration   `json:"datadog,omitempty"`
}

type IstioCoreDNS struct {
	Enabled      *bool                        `json:"enabled,omitempty"`
	Image        string                       `json:"image,omitempty"`
	PluginImage  string                       `json:"pluginImage,omitempty"`
	ReplicaCount int32                        `json:"replicaCount,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector map[string]string            `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration          `json:"tolerations,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// Contains the intended Istio version
	// +kubebuilder:validation:Pattern=^1.2
	Version IstioVersion `json:"version"`

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

	// DefaultResources are applied for all Istio components by default, can be overridden for each component
	DefaultResources *corev1.ResourceRequirements `json:"defaultResources,omitempty"`

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
	UseMCP *bool `json:"useMCP,omitempty"`

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

	// ImagePullPolicy describes a policy for if/when to pull a container image
	// +kubebuilder:validation:Enum=Always,Never,IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// If set to true, the pilot and citadel mtls will be exposed on the
	// ingress gateway also the remote istios will be connected through gateways
	MeshExpansion *bool `json:"meshExpansion,omitempty"`

	// Set to true to connect two or more meshes via their respective
	// ingressgateway services when workloads in each cluster cannot directly
	// talk to one another. All meshes should be using Istio mTLS and must
	// have a shared root CA for this model to work.
	MultiMesh *bool `json:"multiMesh,omitempty"`

	// Istio CoreDNS provides DNS resolution for services in multi mesh setups
	IstioCoreDNS IstioCoreDNS `json:"istioCoreDNS,omitempty"`

	networkName  string
	meshNetworks *MeshNetworks
}

type MeshNetworkEndpoint struct {
	FromCIDR     string `json:"fromCidr,omitempty"`
	FromRegistry string `json:"fromRegistry,omitempty"`
}

type MeshNetworkGateway struct {
	Address string `json:"address"`
	Port    uint   `json:"port"`
}

type MeshNetwork struct {
	Endpoints []MeshNetworkEndpoint `json:"endpoints,omitempty"`
	Gateways  []MeshNetworkGateway  `json:"gateways,omitempty"`
}

type MeshNetworks struct {
	Networks map[string]MeshNetwork `json:"networks"`
}

func (s *IstioSpec) SetMeshNetworks(networks *MeshNetworks) *IstioSpec {
	s.meshNetworks = networks
	return s
}

func (s *IstioSpec) GetMeshNetworks() *MeshNetworks {
	return s.meshNetworks
}

func (s *IstioSpec) GetMeshNetworksHash() string {
	hash := ""
	j, err := json.Marshal(s.meshNetworks)
	if err != nil {
		return hash
	}

	hash = fmt.Sprintf("%x", md5.Sum(j))

	return hash
}

func (s *IstioSpec) SetNetworkName(networkName string) *IstioSpec {
	s.networkName = networkName
	return s
}

func (s *IstioSpec) GetNetworkName() string {
	return s.networkName
}

func (s IstioSpec) GetDefaultConfigVisibility() string {
	if s.DefaultConfigVisibility == "" || s.DefaultConfigVisibility == "." {
		return s.DefaultConfigVisibility
	}
	return "*"
}

func (v IstioVersion) IsSupported() bool {
	re, _ := regexp.Compile(supportedIstioMinorVersionRegex)

	return re.Match([]byte(v))
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	Status         ConfigState
	GatewayAddress []string
	ErrorMessage   string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Istio is the Schema for the istios API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Gateways",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateways of the resource"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
