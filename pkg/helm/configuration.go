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

package helm

import (
	meshv1alpha1 "istio.io/api/mesh/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
)

// HelmValuesType is typedef for Helm .Values
type HelmValuesType map[string]interface{}

// IstioHelmValues defines the desired state of ControlPlane
// XXX: NOT ALL FIELDS ARE MAPPED AND SOME MAY BE OUT OF DATE
// XXX: while a good idea in theory, it may be best to just treat this as a map[string]interface{}
type IstioHelmValues struct {
	Global          *GlobalConfig          `json:"global,omitempty"`
	Galley          *GalleyConfig          `json:"galley,omitempty"`
	Gateways        *GatewaysConfig        `json:"gateways,omitempty"`
	Mixer           *MixerConfig           `json:"mixer,omitempty"`
	Pilot           *PilotConfig           `json:"pilot,omitempty"`
	Security        *SecurityConfig        `json:"security,omitempty"`
	SidecarInjector *SidecarInjectorConfig `json:"sidecarInjectorWebhook,omitempty"`
	Kiali           *KialiConfig           `json:"kiali,omitempty"`
}

// Globals

// GlobalConfig represents the available "global" settings used with the Istio
// Helm templates.
type GlobalConfig struct {
	// Arch is a map of architectures to apply for node affinity settings.
	// key is the arch type, value is preference: 0-never scheduled,
	// 1-least preferred, 2-no preference, 3-most preferred
	Arch map[string]int32 `json:"arch,omitempty"`
	// ConfigRootNamespace is the namespace to use for locating control plane
	// configuration, like sidecar configuration, DestinationRules applying to
	// the entire mesh, etc. Example, istio-config
	ConfigRootNamespace string `json:"configRootNamespace,omitempty"`
	// ConfigValidation determines whether or not the validationWebhook is
	// installed.  Defaults to true.
	ConfigValidation *bool `json:"configValidation,omitempty"`
	// ControlPlaneSecurityEnabled specifies whether or not TLS should be used
	// for communication within the control plane. Defaults to false.
	ControlPlaneSecurityEnabled *bool `json:"controlPlaneSecurityEnabled,omitempty"`
	// CreateRemoteSvcEndpoints specifies that external services should be
	// created for the control plane services.  Used with RemotePilotAddress,
	// RemotePolicyAddress, RemoteTelemetryAddress, and IstioRemot.
	//  Defaults to false
	CreateRemoteSvcEndpoints *bool `json:"createRemoteSvcEndpoints,omitempty"`
	// RemotePilotCreateSvcEndpoint specifies that an external service should be
	// created for the Pilot service.  Use with RemotPilotAddress.
	// Deprecated:  Use CreateRemoteSvcEndpoints.
	RemotePilotCreateSvcEndpoint *bool `json:"remotePilotCreateSvcEndpoint,omitempty"`
	// RemotePilotAddress is the address of an external Pilot service.  Used with
	// CreateRemoteSvcEndpoints.
	RemotePilotAddress string `json:"remotePilotAddress,omitempty"`
	// RemotePilotAddress is the address of an external Mixer Policy service.
	// Used with CreateRemoteSvcEndpoints.
	RemotePolicyAddress string `json:"remotePolicyAddress,omitempty"`
	// RemotePilotAddress is the address of an external Mixer Telemetry service.
	// Used with CreateRemoteSvcEndpoints.
	RemoteTelemetryAddress string `json:"remoteTelemetryAddress,omitempty"`
	// IstioRemote specifies whether or not the control plane is remote.  Used
	// with CreateRemoteSvcEndpoints and other Remote* fields.  Defaults to false
	IstioRemote *bool `json:"istioRemote,omitempty"`
	// DefaultConfigVisibilitySettings is the set of namespaces to which
	// services, service entries, virtual services, destination rules should be
	// exported.  Currently, only a single item may be specified:
	// * implies these objects are visible to all namespaces, enabling any sidecar to talk to any other sidecar.
	// . implies these objects are visible to only to sidecars in the same namespace, or if imported as a Sidecar.egress.host
	DefaultConfigVisibilitySettings []string `json:"defaultConfigVisibilitySettings,omitempty"`
	// DefaultNodeSelector is a set of key/value pairs to be used as the default
	// node selector for Istio pods.
	DefaultNodeSelector map[string]string `json:"defaultNodeSelector,omitempty"`

	// DefaultPodDisruptionBudget represents the PodDisruptionBudget to be
	// applied for Istio pods.
	DefaultPodDisruptionBudget *PodDisruptionBudget `json:"defaultPodDisruptionBudget,omitempty"`

	// DefaultResources are default resource requirements to be applied to Istio
	// pods.
	DefaultResources *corev1.ResourceRequirements `json:"defaultResources,omitempty"`
	// DisablePolicyChecks specifies whether or not Mixer policy checks should
	// be enabled.  Defaults to false.
	DisablePolicyChecks *bool `json:"disablePolicyChecks,omitempty"`
	// EnableTracing enables tracing for Istio.  Components relying on the istio
	// ConfigMap must be restarted, e.g. Pilot.  Defaults to true.
	EnableTracing *bool `json:"enableTracing,omitempty"`
	// ImageHub is the hub to be applied for all images, if otherwise
	// unspecified.  Example: docker.io
	ImageHub string `json:"hub,omitempty"`
	// ImageTag is the tag to be applied for all images, if otherwise
	// unspecified.  Example: 1.1.0
	ImageTag string `json:"tag,omitempty"`
	// ImagePullPolicy is the pull policy to be used for Istio images.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ImagePullSecrets is a list of references to secrets to use for pulling
	// images.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// IstioNamespace is the namespace in which the control plane is installed.
	// XXX: I don't think this is necessary within the operator.  The operator
	// should be deploying the control plane within the same namespace as the
	// CustomResource.
	IstioNamespace string `json:"istioNamespace,omitempty"`

	// KubernetesIngress represents the configuration for Kubernetes Ingress.
	KubernetesIngress *KubernetesIngressConfig `json:"k8sIngress,omitempty"`

	// KubernetesIngressSelector is the name of the Istio ingress service.
	// Defaults to istio-ingressgateway
	KubernetesIngressSelector string `json:"k8sIngressSelector,omitempty"`

	// MeshExpansion represents the configuration for mesh expansion.
	MeshExpansion *MeshExpansionConfig `json:"meshExpansion,omitempty"`

	// MeshNetworks configures the mesh networks to be used by the Split Horizon EDS
	MeshNetworks MeshNetworksType `json:"meshNetworks,omitempty"`
	// MonitoringPort provided by Istio components.  Defaults to 15014
	MonitoringPort *int32 `json:"monitoringPort,omitempty"`
	// MTLS configures mTLS for the mesh.
	MTLS *MTLSConfig `json:"mtls,omitempty"`

	// MultiCluster configures multi-cluster.
	MultiCluster *MultiClusterConfig `json:"multiCluster,omitempty"`

	// Network is the network endpoint to which sidecar endpoints should be
	// associated.  The ISTIO_META_NETWORK environment variable on the sidecars
	// is set to this value.
	Network string `json:"network,omitempty"`
	// OmitSidecarInjectorConfigMap specifies whether or not the sidecar
	// injector ConfigMap should be created.  This should always be false when
	// installing a control plane.  Defaults to false.
	OmitSidecarInjectorConfigMap *bool `json:"omitSidecarInjectorConfigMap,omitempty"`
	// OneNamespace specifies whether or not the Istio controllers watch a single
	// namespace or all namespaces.  Defaults to false.
	OneNamespace *bool `json:"oneNamespace,omitempty"`

	// OutboundTrafficPolicy for sidecars.
	OutboundTrafficPolicy *OutboundTrafficPolicyConfig `json:"outboundTrafficPolicy,omitempty"`

	// PodDNSSearchNamespaces is a list of DNS search suffixes to be applied to
	// sidecars.
	PodDNSSearchNamespaces []string `json:"podDNSSearchNamespaces,omitempty"`
	// PolicyCheckFailOpen determines whether or not traffic is allowed when
	// the Mixer policy service cannot be reached.  Default is false, which means
	// traffic is denied.
	PolicyCheckFailOpen *bool `json:"policyCheckFailOpen,omitempty"`
	// PriorityClassName is the priority class to use on Istio pods.
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// RemoteZipkinAddress is the address of an external Zipkin service.
	// Tracer.Zipkin.Address takes precedence over this.
	RemoteZipkinAddress string `json:"remoteZipkinAddress,omitempty"`
	// TrustDomain represents the trust root to be used with SPIFFE identity URLs.
	// Should default to cluster.local in Kubernetes environments.
	// TODO: verify the default
	TrustDomain string `json:"trustDomain,omitempty"`
	// UseMCP specifies whether or not Mesh Control Protocol should be used for
	// Mixer and Pilot.  Implies the use of Galley if true.  Defaults to true.
	UseMCP *bool `json:"useMCP,omitempty"`

	// Proxy configuration
	Proxy *ProxyConfig `json:"proxy,omitempty"`
	// ProxyInit configuration
	ProxyInit *ProxyInitConfig `json:"proxy_init,omitempty"`

	// SDS configuration
	SDS *SDSConfig `json:"sds,omitempty"`

	// Tracer configuration
	Tracer *ProxyTracerConfig `json:"tracer,omitempty"`
}

// PodDisruptionBudget to apply to Istio pods.  Simply adds an "enabled" field
// to the standard PodDisruptionBudget
type PodDisruptionBudget struct {
	// Enabled specifies whether or not a PodDisruptionBudget should be applied
	// for Istio pods.  Defaults to true.
	EnabledField                      `json:",inline"`
	policyv1beta1.PodDisruptionBudget `json:",inline"`
}

// KubernetesIngressConfig represents the configuration for Kubernetes Ingress.
type KubernetesIngressConfig struct {
	// Enabled specifies whether or not a Kubernetes Ingress should be created.
	// Defaults to false.
	EnabledField `json:",inline"`
	// EnableHTTPS specifies whether or not Kubernetes Ingress should expose
	// port 443.  Defaults to false.
	EnableHTTPS *bool `json:"enableHttps,omitempty"`
	// GatewayName represents the name of the Istio Gateway backing the ingress.
	// Defaults to ingress.
	GatewayName string `json:"gatewayName,omitempty"`
}

// MeshExpansionConfig represents the configuration for mesh expansion.
type MeshExpansionConfig struct {
	// Enabled specifies whether or not mesh expansion is enabled.  Defaults
	// to false.
	EnabledField `json:",inline"`
	// UseILB specifies whether or not the mesh is exposed through an ILB
	// Gateway.  If specified, gateways.istio-ilbgateway.enabled should also be
	// set to true.  Defaults to false.
	UseILB *bool `json:"useILB,omitempty"`
}

// MeshNetworksType typedef for MeshNetworks field
type MeshNetworksType map[string]meshv1alpha1.Network

// MTLSConfig configures mTLS for the mesh.
type MTLSConfig struct {
	// Enabled specifies whether or not mTLS is enabled.  Defaults to false.
	EnabledField `json:",inline"`
}

// MultiClusterConfig configures multi-cluster.
type MultiClusterConfig struct {
	// Enabled specifies whether or not multi-cluster is enabled.  If this is
	// enabled, gateways.istio-egressgateway.enabled and gateways.istio-ingressgateway.enabled
	// should also be set to true. Defaults to false.
	EnabledField `json:",inline"`
}

// OutboundTrafficPolicyMode is a type alias for OutboundTrafficPolicyMode
type OutboundTrafficPolicyMode string

const (
	// OutboundTrafficPolicyModeRegistryOnly only allows outbound traffic from
	// the sidecar to services in the registry.
	OutboundTrafficPolicyModeRegistryOnly OutboundTrafficPolicyMode = "REGISTRY_ONLY"
	// OutboundTrafficPolicyModeAllowAny allows outbound traffic from the sidecar
	// to any service, regardless of whether or not it is in the registry.
	OutboundTrafficPolicyModeAllowAny OutboundTrafficPolicyMode = "ALLOW_ANY"
)

// OutboundTrafficPolicyConfig for sidecars.
type OutboundTrafficPolicyConfig struct {
	// Mode is the outbound traffic policy mode for sidecars.  Defaults to
	// REGISTRY_ONLY.
	Mode OutboundTrafficPolicyMode `json:"mode,omitempty"`
}

// AccessLogEncodingType is a type def for AccessLogEncoding values
type AccessLogEncodingType string

const (
	// AccessLogeEncodingTypeTEXT represents TEXT
	AccessLogeEncodingTypeTEXT AccessLogEncodingType = "TEXT"
	// AccessLogeEncodingTypeJSON represents JSON
	AccessLogeEncodingTypeJSON AccessLogEncodingType = "JSON"
)

// ProxyConfig specifies how proxies are configured within Istio.
type ProxyConfig struct {
	// AccessLogEncoding represents the encoding for the access log.  May be one
	// of TEXT or JSON.  Defaults to TEXT
	AccessLogEncoding string `json:"accessLogEncoding,omitempty"`
	// AccessLogFile sets the location to which log messages are sent.  An empty
	// string disables logging.  Defaults to /dev/stdout
	AccessLogFile string `json:"accessLogFile,omitempty"`
	// AccessLogFormat is the format of the fields emitted in the log messages.
	// An empty value implies the default format.  Defaults to ""
	// TEXT example: "[%START_TIME%] %REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\n"
	// JSON example: '{"start_time": "%START_TIME%", "req_method": "%REQ(:METHOD)%", "path": "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%" "protocol": "%PROTOCOL%"}'
	AccessLogFormat string `json:"accessLogFormat,omitempty"`
	// AutoInject specifies whether or not automatic injection of sidecars
	// should be enabled.  Defaults to enabled
	AutoInject string `json:"autoInject,omitempty"`
	// ClusterDomain is the domain for the cluster.  Defaults to cluster.local
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// Concurrency controls the number of working threads used by the proxy container.
	// 0 specifies one thread per core.  Defaults to 0
	Concurrency *int32 `json:"concurrency,omitempty"`
	// EnableCoreDump specifies whether or not core dumps will be generated if
	// failures occur on the proxies.  Defaults to false
	EnableCoreDump *bool `json:"enableCoreDump,omitempty"`
	// ExcludeInboundPorts represent ports to be blacklisted for ingress.
	ExcludeInboundPorts string `json:"excludeInboundPorts,omitempty"`
	// ExcludeIPRanges represent IP ranges to be blacklisted for egress.
	ExcludeIPRanges string `json:"excludeIPRanges,omitempty"`
	// Image represents the name of the proxy image to use.  Defaults to proxyv2
	Image string `json:"image,omitempty"`
	// IncludeInboundPorts represents ports to be whitelisted for ingress.
	// XXX: doesn't appear to be referenced by any templates
	IncludeInboundPorts string `json:"includeInboundPorts,omitempty"`
	// IncludeIPRanges represents IP ranges to be whitelisted for egress.
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`
	// Privileged specifies whether or not the Istio proxy container runs in
	// privileged mode.  Defaults to false
	Privileged *bool `json:"privileged,omitempty"`
	// ReadinessFailureThreshold represents the failure threshold for the
	// readiness probe on the proxy container.  Defaults to 30
	ReadinessFailureThreshold *int32 `json:"readinessFailureThreshold,omitempty"`
	// ReadinessInitialDelaySeconds represents the initial delay for the
	// readiness probe on the proxy container.  Defaults to 1
	ReadinessInitialDelaySeconds *int32 `json:"readinessInitialDelaySeconds,omitempty"`
	// ReadinessPeriodSeconds represents the check interval for the readiness
	// probe on the proxy container.  Defaults to 2
	ReadinessPeriodSeconds *int32 `json:"readinessPeriodSeconds,omitempty"`
	// Resources specifies the resource requirements for the proxy containers.
	// Defaults to {requests: {cpu: 10m}}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// StatusPort is the port number to use for health checks.  Defaults to 15020
	StatusPort *int32 `json:"statusPort,omitempty"`
	// Tracer represents the type of tracer to use.  Defaults to zipkin
	Tracer ProxyTracerType `json:"tracer,omitempty"`

	// EnvoyStatsD represents the configuration to the statsd server for envoy
	// Deprecated
	EnvoyStatsD *EnvoyStatsDConfig `json:"envoyStatsd,omitempty"`
}

// ProxyInitConfig is the configuration for proxy_init
type ProxyInitConfig struct {
	// Image is the name of the proxy init image.  Defaults to proxy_init
	Image string `json:"image,omitempty"`
}

// EnvoyStatsDConfig represents the configuration of the Envoy statsd server.
type EnvoyStatsDConfig struct {
	// Enabled specifies whether a statsd server is used.  Defaults to false
	EnabledField `json:",inline"`
	// Host of the statsd server
	Host string `json:"host,omitempty"`
	// Port of the statsd server
	Port string `json:"port,omitempty"`
}

// SDSConfig represents the configuration for SDS
type SDSConfig struct {
	// Enabled specifies whether or not SDS is enabled.  Defaults to false.
	EnabledField `json:",inline"`
	// UDSPath is the Unix Domain Socket path where keys/certs are located.
	// Empty value uses secret mounts.  Defaults to empty.
	UDSPath string `json:"udsPath,omitempty"`
	// UseNormalJWT specifies that the kubernetes service account token should
	// be used when requesting keys/certs from SDS.  Defaults to false
	UseNormalJWT *bool `json:"useNormalJwt,omitempty"`
	// UseTrustworthyJWT specifies that the service account JWT be mounted into
	// the Envoy container so it can be used when generating keys/certs.
	// Defaults to false
	UseTrustworthyJWT *bool `json:"useTrustworthyJwt,omitempty"`
}

// ProxyTracerType is a custom type for specifying the type of tracer configured.
type ProxyTracerType string

const (
	// ZipkinTracerType for a Zipkin tracer
	ZipkinTracerType ProxyTracerType = "zipkin"
	// LightStepTracerType for a LightStep tracer
	LightStepTracerType ProxyTracerType = "lightstep"
)

// ProxyTracerConfig represents the configuration of a tracer
type ProxyTracerConfig struct {
	// Type of tracer.  Defaults to zipkin
	Type ProxyTracerType `json:"type,omitempty"`
	// LightStep configuration
	LightStep *ProxyTracerLightStepConfig `json:"lightstep,omitempty"`
	// Zipkin configuration
	Zipkin *ProxyTracerZipkinConfig `json:"zipkin,omitempty"`
}

// ProxyTracerLightStepConfig represents the configuration of the LightStep tracer
type ProxyTracerLightStepConfig struct {
	// AccessToken is the token used to access the LightStep server
	AccessToken string `json:"accessToken,omitempty"`
	// Address is the address of the LightStep server
	Address string `json:"address,omitempty"`
	// CACertPath is the path to the certs use when verifying TLS connection.
	CACertPath string `json:"cacertPath,omitempty"`
	// Secure specifies that TLS should be used when communicating with the
	// LightStep server.  Defaults to true.
	Secure *bool `json:"secure,omitempty"`
}

// ProxyTracerZipkinConfig represents the configuration of the Zipkin tracer
type ProxyTracerZipkinConfig struct {
	// Address of the Zipkin server.  Defaults to zipkin:9411
	Address string `json:"address,omitempty"`
}

// Galley component

// GalleyConfig is the configuration for the Galley component
type GalleyConfig struct {
	// Enabled specifies whether or not the galley templates should be processed
	// Defaults to true.
	CommonComponentConfig `json:",inline"`
	// Defaults: Image: galley, ReplicaCount: 1
	DeploymentFields `json:",inline"`
}

// Gateways component

// GatewaysConfig represents configuration specific to the gateways subchart
type GatewaysConfig struct {
	// Enabled specifies whether or not the gateways charts should be processed
	// Defaults to true.
	CommonComponentConfig `json:",inline"`

	// Gateways is a name->config map of GatewayConfig.  A gateway will be
	// configured for each map entry, assuming its enabled field is set to true.
	Gateways map[string]GatewayConfig
}

// GatewayConfig specifies the configuration for an Istio gateway
type GatewayConfig struct {
	// Enabled specifies whether or not this gateway should be configured.
	EnabledField `json:",inline"`
	// Defaults: AutoscaleEnabled: trueAutoscaleMin: 1, AutoscaleMax: 5, CPU: TargetAverageUtilization: 80
	DeploymentFields `json:",inline"`

	// Additional Deployment specific configuration

	// AdditionalContainers are any additional containers that should be on the
	// Pod for the gateway Deployment.
	AdditionalContainers []corev1.Container `json:"additionalContainers,omitempty"`
	// ConfigVolumes are ConfigMap volumes to be added to the gateway Deployment.
	// XXX: it doesn't appear that the deployment.yaml template configures mounts
	// for these, only the volumes themselves.
	ConfigVolumes []ConfigMapVolume `json:"configVolumes,omitempty"`
	// SDS controls SDS configuration for the gateway.
	SDS *SDSContainerConfig `json:"sds,omitempty"`
	// SecretVolumes are secret Volumes to be added to the gateway Deployment
	SecretVolumes []SecretVolume `json:"secretVolumes,omitempty"`

	// Service specific configuration

	// ExternalIPs to specify on the gateway Service.
	ExternalIPs []string `json:"externalIPs,omitempty"`
	// ExternalTrafficPolicy to set on the gateway Service.
	ExternalTrafficPolicy string `json:"externalTrafficPolicy,omitempty"`
	// LoadBalancerIP for the gateway Service.
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`
	// LoadBalancerSourceRanges for the gateway Service.
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty"`
	// MeshExpansionPorts are ports to be exposed by the gateway Service
	MeshExpansionPorts []corev1.ServicePort `json:"meshExpansionPorts,omitempty"`
	// ServiceAnnotations to be added to the gateway Service
	ServiceAnnotations AnnotationsType `json:"serviceAnnotations,omitempty"`

	// Shared items

	// Labels to add to the gateway Deployment, Pod and Service.  Should specify
	// a label with key "istio" similar to the gateway name, e.g. egressgateway
	// for istio-egressgateway.
	Labels map[string]string `json:"labels,omitempty"`
	// Namespace within which the gateway should be created.
	// XXX: I think we should use the namespace field on the custom resource,
	// i.e. the user should create a CR in the same namespace as the gateway to
	// be created.
	Namespace string `json:"namespace,omitempty"`
	// Ports to be exposed by the gateway Service and Deployment.  The
	// containerPort will be created matching the port value.
	Ports []corev1.ServicePort `json:"ports,omitempty"`
}

// SDSContainerConfig is used to configure an SDS container on a Pod.
type SDSContainerConfig struct {
	// Enabled specifies whether or not SDS should be used by the gateway.  If
	// enabled, Image must also be specified.  Defaults to false.
	EnabledField `json:",inline"`
	// Image is the name of the image to use for the SDS container.  Defaults to
	// node-agent-k8s
	Image string `json:"image,omitempty"`
}

// ConfigMapVolume defines a ConfigMap Volume that will be added to a Deployment
type ConfigMapVolume struct {
	// Name of the Volume
	Name string `json:"name,omitempty"`
	// XXX: I suspect this will need to be added
	//MountPath  string `json:"mountPath,omitempty"`
	// ConfigMapName ...
	ConfigMapName string `json:"configMapName,omitempty"`
}

// SecretVolume defines a Secret Volume that will be added to a Deployment
type SecretVolume struct {
	// Name of the Volume
	Name string `json:"name,omitempty"`
	// MountPath for the Volume
	MountPath string `json:"mountPath,omitempty"`
	// SecretName ...
	SecretName string `json:"secretName,omitempty"`
}

// Mixer component

// MixerConfig is the configuration for the Mixer component
type MixerConfig struct {
	// Enabled specifies whether or not the mixer templates should be processed
	// Defaults to true.
	CommonComponentConfig `json:",inline"`
	// Defaults: Image: mixer, Env: { GODEBUG: gctrace=2 }
	// Autoscaler and ReplicaCount fields are unused
	DeploymentFields `json:",inline"`

	// Policy configuration.  Only Autoscaler and ReplicaCount fields
	// are used.
	Policy *MixerPolicyConfig `json:"policy,omitempty"`
	// Telemetry configuration.
	Telemetry *MixerTelemetryConfig `json:"telemetry,omitempty"`

	// Adapters is the configuration for Mixer adapters.
	Adapters *MixerAdaptersConfig `json:"adapters,omitempty"`
}

// MixerPolicyConfig is the configuration for Mixer's policy component
type MixerPolicyConfig struct {
	EnabledField     `json:",inline"`
	DeploymentFields `json:",inline"`
}

// MixerTelemetryConfig is the configuration for Mixer's telemetry component
type MixerTelemetryConfig struct {
	// Only Autoscaler and ReplicaCount fields are used
	DeploymentFields `json:",inline"`

	// SessionAffinityEnabled configures sessionAffinity: ClientIP on the
	// associated Service. Defaults to false
	SessionAffinityEnabled *bool `json:"sessionAffinityEnabled,omitempty"`
}

// MixerAdaptersConfig is the configuration for the Mixer adapters
type MixerAdaptersConfig struct {
	// Kubernetesenv is the configuration for a kubernetes handler
	KubernetesEnv *KubernetesEnvMixerAdapterConfig `json:"kubernetesenv,omitempty"`
	// Kubernetesenv is the configuration for a prometheus handler
	Prometheus *PrometheusMixerAdapterConfig `json:"prometheus,omitempty"`
	// Kubernetesenv is the configuration for a stdio handler
	Stdio *StdioMixerAdapterConfig `json:"stdio,omitempty"`
	// UseAdapterCRDs specifies whether or not CRDs are being used to configure
	// adapters.
	UseAdapterCRDs *bool `json:"useAdapterCRDs,omitempty"`
}

// KubernetesEnvMixerAdapterConfig is the configuration for the kubernetes env mixer adapter
type KubernetesEnvMixerAdapterConfig struct {
	// Defaults to true.
	EnabledField `json:",inline"`
}

// PrometheusMixerAdapterConfig is the configuration for the prometheus mixer adapter
type PrometheusMixerAdapterConfig struct {
	// Defaults to true.
	EnabledField `json:",inline"`
	// MetricExpiryDuration ... Defaults to 10m
	MetricExpiryDuration string
}

// StdioMixerAdapterConfig is the configuration for the stdio mixer adapter
type StdioMixerAdapterConfig struct {
	// Defaults to true.
	EnabledField `json:",inline"`
	// OutputAsJSON ...  Defaults to true.
	OutputAsJSON *bool
}

// Pilot component

// PilotConfig is the configuration for the Pilot component
type PilotConfig struct {
	// Enabled specifies whether or not the pilot templates should be processed
	// Defaults to true.
	CommonComponentConfig `json:",inline"`
	// Defaults: Image: pilot, AutoscaleEnabled: trueAutoscaleMin: 1, AutoscaleMax: 5,
	// CPU: { TargetAverageUtilization: 80 }, Resources: { Requests: { CPU: 500m, Memory: 2048Mi } },
	// Env: { PILOT_PUSH_THROTTLE_COUNT: 100, GODEBUG: gctrace=2 }
	DeploymentFields `json:",inline"`

	// Sidecar configures an Istio proxy sidecar on the Pilot Pods.
	// Defaults to true
	Sidecar *bool `json:"sidecar,omitempty"`
	// TraceSampling is the interval for random trace sampling.
	// Defaults to 100.0
	TraceSampling *float64 `json:"traceSampling,omitempty"`
}

// Security component

// SecurityConfig is the configuration for the Citadel component
type SecurityConfig struct {
	// Enabled specifies whether or not the security templates should be processed
	// Defaults to true.
	CommonComponentConfig `json:",inline"`
	// Defaults: Image: citadel, ReplicaCount: 1
	DeploymentFields `json:",inline"`

	// SelfSigned if using self-signed certificates.  Defaults to true.
	SelfSigned *bool `json:"selfSigned,omitempty"`
	// CreateMeshPolicy specifies whether or not a MeshPolicy should be created.
	// Defaults to true.
	CreateMeshPolicy *bool `json:"createMeshPolicy,omitempty"`
}

// Sidecar Injector component

// SidecarInjectorConfig is the configuration for the sidecar injector webhook
type SidecarInjectorConfig struct {
	// Enabled specifies whether or not the sidecar injector webhook templates
	// should be processed. Defaults to true.
	CommonComponentConfig `json:",inline"`
	// Defaults: Image: sidecar_injector, ReplicaCount: 1
	DeploymentFields `json:",inline"`

	// EnableNamespacesByDefault specifies whether injection for any given
	// namespace is opt-in vs. opt-out.  true means all namespaces will be
	// scanned for injection unless they are annotated with
	// istio-injection=disabled.  false means namespaces must be annotated with
	// istio-injection=enabled for automatic injection of sidecars.
	// Defaults to false.
	EnableNamespacesByDefault *bool `json:"enableNamespacesByDefault,omitempty"`
}

// Kiali component

// KialiConfig is the configuration for the Kiali component
type KialiConfig struct {
	CommonComponentConfig `json:",inline"`
	// Defaults: ReplicaCount: 1
	DeploymentFields `json:",inline"`
	// Hub is the name of the image registry/namespace from which the image should be pulled.
	// Defaults to docker.io/kiali
	Hub string `json:"hub,omitempty"`
	// Tag is the tag of the image
	// Defaults to v0.14
	Tag string `json:"tag,omitempty"`
	// ContextPath for the service
	ContextPath string                `json:"contextPath,omitempty"`
	Gateway     *EnabledField         `json:"gateway,omitempty"`
	Ingress     *IngressConfig        `json:"ingress,omitempty"`
	Dashboard   *KialiDashboardConfig `json:"dashboard,omitempty"`
	// PrometheusAddr for prometheus service
	PrometheusAddr string `json:"prometheusAddr,omitempty"`
	// CreateDemoSecret will cause a secret will be created with a default username
	// and password. Useful for demos.
	CreateDemoSecret *bool `json:"createDemoSecret,omitempty"`
}

// KialiDashboardConfig is the configuration for the Kiali dashboard
type KialiDashboardConfig struct {
	SecretName    string `json:"secretName,omitempty"`
	UsernameKey   string `json:"usernameKey,omitempty"`
	PassphraseKey string `json:"passphraseKey,omitempty"`
	User          string `json:"user,omitempty"`
	Passphrase    string `json:"passphrase,omitempty"`
}

// Shared structs used by multiple components

// IngressConfig is the Ingres configuration used by many components
type IngressConfig struct {
	// Defaults to false
	EnabledField `json:",inline"`
	// Annotations to configure on the Ingress
	Annotations AnnotationsType `json:"annotations,omitempty"`
	// Hosts to configure on the Ingress
	// Defaults to: [ prometheus.local ]
	Hosts []string `json:"hosts,omitempty"`
	// TLS configuration for the Ingress
	TLS []extensionsv1beta1.IngressTLS `json:"tls,omitempty"`
}

// CommonComponentConfig are settings common to most components
type CommonComponentConfig struct {
	EnabledField  `json:",inline"`
	NameOverrides `json:",inline"`

	// Global specifies component specific overrides for global values
	Global *GlobalConfig `json:"global,omitempty"`
}

// EnabledField is a helper for types which have an "enabled" field.
type EnabledField struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// NameOverrides are used by various components for specializing resource names.
type NameOverrides struct {
	// NameOverride typically overrides Chart.Name
	NameOverride string `json:"nameOverride,omitempty"`
	// FullnameOverride typically overrides Release.Name
	FullnameOverride string `json:"fullnameOverride,omitempty"`
}

// DeploymentFields are fields used in most deployment.yaml templates
type DeploymentFields struct {
	// Autoscaler specific configuration
	HorizontalPodAutoscalerFields `json:",inline"`

	// Image is the name of the image to use for the component's container.
	Image string `json:"image,omitempty"`
	// ReplicaCount for the Deployment.  Ignored if Autoscaler* is configured
	// and applicable to the component.  Defaults to 1.
	ReplicaCount *int32 `json:"replicaCount,omitempty"`
	// Resources are ResourceRequirements to be specified on the Deployment.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a set of key/value pairs to be used as the node selector
	// for the Pods.  If not specified, the DefaultNodeSelector is used.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Env values to be added to the Deployment
	Env map[string]string `json:"env,omitempty"`
	// PodAnnotations to be added to the Pods.
	PodAnnotations AnnotationsType `json:"podAnnotations,omitempty"`
}

// HorizontalPodAutoscalerFields used by most autoscale.yaml templates
type HorizontalPodAutoscalerFields struct {
	// AutoscaleEnabled specifies whether or not a HorizontalPodAutoscaler is
	// configured for the gateway Deployment.  If specified, AutoscaleMax and
	// AutoscaleMin must also be specified.  Defaults to true
	AutoscaleEnabled *bool `json:"autoscaleEnabled,omitempty"`
	// AutoscaleMax specifies the maximum number of replicas for the
	// HorizontalPodAutoscaler.  Defaults to 5
	AutoscaleMax *int32 `json:"autoscaleMax,omitempty"`
	// AutoscaleMax specifies the minimum number of replicas for the
	// HorizontalPodAutoscaler.  Defaults to 1
	AutoscaleMin *int32 `json:"autoscaleMin,omitempty"`
	// CPU metric setting used by the HorizontalPodAutoscaler.  TargetAverageUtilization
	// defaults to 80.
	CPU *ResourceMetricCPU `json:"cpu,omitempty"`
}

// ResourceMetricCPU wrapper for cpu field
type ResourceMetricCPU struct {
	TargetAverageUtilization *int32 `json:"targetAverageUtilization,omitempty"`
}

// AnnotationsType typedef for Annotations type fields
type AnnotationsType map[string]string
