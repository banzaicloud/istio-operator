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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	supportedIstioMinorVersionRegex = "^1.7"
)

var (
	SupportedIstioVersion = "1.7.1"
	Version               = "0.7.3"
)

// IstioVersion stores the intended Istio version
type IstioVersion string

// BaseK8sResourceConfiguration defines basic K8s resource spec configurations
type BaseK8sResourceConfiguration struct {
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector    map[string]string            `json:"nodeSelector,omitempty"`
	Affinity        *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations     []corev1.Toleration          `json:"tolerations,omitempty"`
	PodAnnotations  map[string]string            `json:"podAnnotations,omitempty"`
	SecurityContext *corev1.SecurityContext      `json:"securityContext,omitempty"`
}

type BaseK8sResourceConfigurationWithImage struct {
	Image                        *string `json:"image,omitempty"`
	BaseK8sResourceConfiguration `json:",inline"`
}

type BaseK8sResourceConfigurationWithReplicas struct {
	// +kubebuilder:validation:Minimum=0
	ReplicaCount                          *int32 `json:"replicaCount,omitempty"`
	BaseK8sResourceConfigurationWithImage `json:",inline"`
}

type BaseK8sResourceConfigurationWithHPA struct {
	// +kubebuilder:validation:Minimum=0
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// +kubebuilder:validation:Minimum=0
	MaxReplicas                              *int32 `json:"maxReplicas,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
}

type BaseK8sResourceConfigurationWithHPAWithoutImage struct {
	// +kubebuilder:validation:Minimum=0
	ReplicaCount *int32 `json:"replicaCount,omitempty"`
	// +kubebuilder:validation:Minimum=0
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// +kubebuilder:validation:Minimum=0
	MaxReplicas                  *int32 `json:"maxReplicas,omitempty"`
	BaseK8sResourceConfiguration `json:",inline"`
}

// SDSConfiguration defines Secret Discovery Service config options
type SDSConfiguration struct {
	// If set to true, mTLS certificates for the sidecars will be
	// distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates.
	Enabled *bool `json:"enabled,omitempty"`
	// Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
	// key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.
	UdsPath string `json:"udsPath,omitempty"`
	// The JWT token for SDS and the aud field of such JWT. See RFC 7519, section 4.1.3.
	// When a CSR is sent from Citadel Agent to the CA (e.g. Citadel), this aud is to make sure the
	// 	JWT is intended for the CA.
	TokenAudience string `json:"tokenAudience,omitempty"`

	CustomTokenDirectory string `json:"customTokenDirectory,omitempty"`
}

// IstiodConfiguration defines config options for Istiod
type IstiodConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// If enabled, pilot will run Istio analyzers and write analysis errors to the Status field of any Istio Resources
	EnableAnalysis *bool `json:"enableAnalysis,omitempty"`
	// If enabled, pilot will update the CRD Status field of all Istio resources with reconciliation status
	EnableStatus             *bool `json:"enableStatus,omitempty"`
	MultiClusterSupport      *bool `json:"multiClusterSupport,omitempty"`
	MultiControlPlaneSupport *bool `json:"multiControlPlaneSupport,omitempty"`
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	Sidecar                             *bool   `json:"sidecar,omitempty"`
	TraceSampling                       float32 `json:"traceSampling,omitempty"`
	// If enabled, protocol sniffing will be used for outbound listeners whose port protocol is not specified or unsupported
	EnableProtocolSniffingOutbound *bool `json:"enableProtocolSniffingOutbound,omitempty"`
	// If enabled, protocol sniffing will be used for inbound listeners whose port protocol is not specified or unsupported
	EnableProtocolSniffingInbound *bool `json:"enableProtocolSniffingInbound,omitempty"`
	// Configure the certificate provider for control plane communication.
	// Currently, two providers are supported: "kubernetes" and "istiod".
	// As some platforms may not have kubernetes signing APIs,
	// Istiod is the default
	// +kubebuilder:validation:Enum=kubernetes,istiod
	CertProvider PilotCertProviderType `json:"certProvider,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

// CitadelConfiguration defines config options for Citadel
type CitadelConfiguration struct {
	Enabled                               *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithImage `json:",inline"`
	CASecretName                          string `json:"caSecretName,omitempty"`
	// Enable health checking on the Citadel CSR signing API. https://istio.io/docs/tasks/security/health-check/
	HealthCheck *bool `json:"healthCheck,omitempty"`
	// For the workloads running in Kubernetes, the lifetime of their Istio certificates is controlled by the workload-cert-ttl flag on Citadel. The default value is 90 days. This value should be no greater than max-workload-cert-ttl of Citadel.
	WorkloadCertTTL string `json:"workloadCertTTL,omitempty"`
	// Citadel uses a flag max-workload-cert-ttl to control the maximum lifetime for Istio certificates issued to workloads. The default value is 90 days. If workload-cert-ttl on Citadel or node agent is greater than max-workload-cert-ttl, Citadel will fail issuing the certificate.
	MaxWorkloadCertTTL string `json:"maxWorkloadCertTTL,omitempty"`

	// Determines Citadel default behavior if the ca.istio.io/env or ca.istio.io/override
	// labels are not found on a given namespace.
	//
	// For example: consider a namespace called "target", which has neither the "ca.istio.io/env"
	// nor the "ca.istio.io/override" namespace labels. To decide whether or not to generate secrets
	// for service accounts created in this "target" namespace, Citadel will defer to this option. If the value
	// of this option is "true" in this case, secrets will be generated for the "target" namespace.
	// If the value of this option is "false" Citadel will not generate secrets upon service account creation.
	EnableNamespacesByDefault *bool `json:"enableNamespacesByDefault,omitempty"`

	// Whether SDS is enabled.
	SDSEnabled *bool `json:"sdsEnabled,omitempty"`

	// Select the namespaces for the Citadel to listen to, separated by comma. If set to empty,
	// Citadel listens to all namespaces.
	ListenedNamespaces *string `json:"listenedNamespaces,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

// GalleyConfiguration defines config options for Galley
type GalleyConfiguration struct {
	Enabled                                  *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
	ConfigValidation                         *bool `json:"configValidation,omitempty"`
	EnableServiceDiscovery                   *bool `json:"enableServiceDiscovery,omitempty"`
	EnableAnalysis                           *bool `json:"enableAnalysis,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
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

type ServicePort struct {
	corev1.ServicePort `json:",inline"`
	TargetPort         *int32 `json:"targetPort,omitempty"`
}

type ServicePorts []ServicePort

func (ps ServicePorts) Convert() []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0)
	for _, po := range ps {
		port := corev1.ServicePort{
			Name:     po.Name,
			Protocol: po.Protocol,
			Port:     po.Port,
			NodePort: po.NodePort,
		}
		if po.TargetPort != nil {
			port.TargetPort = intstr.FromInt(int(util.PointerToInt32(po.TargetPort)))
		}
		ports = append(ports, port)
	}

	return ports
}

type GatewayConfiguration struct {
	MeshGatewayConfiguration `json:",inline"`
	Ports                    []ServicePort `json:"ports,omitempty"`
	Enabled                  *bool         `json:"enabled,omitempty"`
	// Whether to fully reconcile the MGW resource or just take care that it exists
	CreateOnly *bool `json:"createOnly,omitempty"`
}

type K8sIngressConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// enableHttps will add port 443 on the ingress.
	// It REQUIRES that the certificates are installed  in the
	// expected secrets - enabling this option without certificates
	// will result in LDS rejection and the ingress will not work.
	EnableHttps *bool `json:"enableHttps,omitempty"`
}

// MixerConfiguration defines config options for Mixer
type MixerConfiguration struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	PolicyConfigurationSpec             `json:",inline"`
	TelemetryConfigurationSpec          `json:",inline"`
	// Turn it on if you use mixer that supports multi cluster telemetry
	MultiClusterSupport *bool `json:"multiClusterSupport,omitempty"`
	// stdio is a debug adapter in Istio telemetry, it is not recommended for production use
	StdioAdapterEnabled *bool `json:"stdioAdapterEnabled,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

type PolicyConfiguration struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	PolicyConfigurationSpec             `json:",inline"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

type PolicyConfigurationSpec struct {
	ChecksEnabled *bool `json:"checksEnabled,omitempty"`
}

type TelemetryConfiguration struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	TelemetryConfigurationSpec          `json:",inline"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

type TelemetryConfigurationSpec struct {
	// Set reportBatchMaxEntries to 0 to use the default batching behavior (i.e., every 100 requests).
	// A positive value indicates the number of requests that are batched before telemetry data
	// is sent to the mixer server
	ReportBatchMaxEntries *int32 `json:"reportBatchMaxEntries,omitempty"`
	// Set reportBatchMaxTime to 0 to use the default batching behavior (i.e., every 1 second).
	// A positive time value indicates the maximum wait time since the last request will telemetry data
	// be batched before being sent to the mixer server
	ReportBatchMaxTime *string `json:"reportBatchMaxTime,omitempty"`
	// Set whether to create a STRICT_DNS type cluster for istio-telemetry.
	SessionAffinityEnabled *bool `json:"sessionAffinityEnabled,omitempty"`
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
	// List of namespaces to include for Istio pod check
	IncludeNamespaces []string `json:"includeNamespaces,omitempty"`
	// Logging level for CNI binary
	LogLevel string                 `json:"logLevel,omitempty"`
	Affinity *corev1.Affinity       `json:"affinity,omitempty"`
	Chained  *bool                  `json:"chained,omitempty"`
	Repair   CNIRepairConfiguration `json:"repair,omitempty"`
}

// CNIRepairConfiguration defines config for the repair CNI container
type CNIRepairConfiguration struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	Hub                 *string `json:"hub,omitempty"`
	Tag                 *string `json:"tag,omitempty"`
	LabelPods           *bool   `json:"labelPods,omitempty"`
	DeletePods          *bool   `json:"deletePods,omitempty"`
	InitContainerName   *string `json:"initContainerName,omitempty"`
	BrokenPodLabelKey   *string `json:"brokenPodLabelKey,omitempty"`
	BrokenPodLabelValue *string `json:"brokenPodLabelValue,omitempty"`
}

// SidecarInjectorInitConfiguration defines options for init containers in the sidecar
type SidecarInjectorInitConfiguration struct {
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// SidecarInjectorConfiguration defines config options for SidecarInjector
type SidecarInjectorConfiguration struct {
	Enabled                                  *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
	Init                                     SidecarInjectorInitConfiguration `json:"init,omitempty"`
	InitCNIConfiguration                     InitCNIConfiguration             `json:"initCNIConfiguration,omitempty"`
	// If true, sidecar injector will rewrite PodSpec for liveness
	// health check to redirect request to sidecar. This makes liveness check work
	// even when mTLS is enabled.
	RewriteAppHTTPProbe *bool `json:"rewriteAppHTTPProbe,omitempty"`
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
	// injectedAnnotations are additional annotations that will be added to the pod spec after injection
	// This is primarily to support PSP annotations. For example, if you defined a PSP with the annotations:
	//
	// annotations:
	//   apparmor.security.beta.kubernetes.io/allowedProfileNames: runtime/default
	//   apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
	//
	// The PSP controller would add corresponding annotations to the pod spec for each container. However, this happens before
	// the inject adds additional containers, so we must specify them explicitly here. With the above example, we could specify:
	// injectedAnnotations:
	//   container.apparmor.security.beta.kubernetes.io/istio-init: runtime/default
	//   container.apparmor.security.beta.kubernetes.io/istio-proxy: runtime/default
	InjectedAnnotations map[string]string `json:"injectedAnnotations,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	InjectedContainerAdditionalArgs []string `json:"injectedContainerAdditionalArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	InjectedContainerAdditionalEnvVars []corev1.EnvVar `json:"injectedContainerAdditionalEnvVars,omitempty"`
}

// ProxyWasmConfiguration defines config options for Envoy wasm
type ProxyWasmConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// NodeAgentConfiguration defines config options for NodeAgent
type NodeAgentConfiguration struct {
	Enabled                               *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithImage `json:",inline"`
}

type EnvoyStatsD struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Host    string `json:"host,omitempty"`
	Port    int32  `json:"port,omitempty"`
}

type EnvoyServiceCommonConfiguration struct {
	Enabled      *bool         `json:"enabled,omitempty"`
	Host         string        `json:"host,omitempty"`
	Port         int32         `json:"port,omitempty"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty"`
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`
}

func (c EnvoyServiceCommonConfiguration) GetData() map[string]interface{} {
	data := map[string]interface{}{
		"address": fmt.Sprintf("%s:%d", c.Host, c.Port),
	}
	if c.TLSSettings != nil {
		data["tlsSettings"] = c.TLSSettings
	}
	if c.TCPKeepalive != nil {
		data["tcpKeepalive"] = c.TCPKeepalive
	}

	return data
}

func (c EnvoyServiceCommonConfiguration) GetDataJSON() string {
	j, err := json.Marshal(c.GetData())
	if err != nil {
		return ""
	}

	return string(j)
}

type TLSSettings struct {
	// +kubebuilder:validation:Enum=DISABLE,SIMPLE,MUTUAL,ISTIO_MUTUAL
	Mode              string   `json:"mode,omitempty"`
	ClientCertificate string   `json:"clientCertificate,omitempty"`
	PrivateKey        string   `json:"privateKey,omitempty"`
	CACertificates    string   `json:"caCertificates,omitempty"`
	SNI               string   `json:"sni,omitempty"`
	SubjectAltNames   []string `json:"subjectAltNames,omitempty"`
}

type TCPKeepalive struct {
	Probes   int32  `json:"probes,omitempty"`
	Time     string `json:"time,omitempty"`
	Interval string `json:"interval,omitempty"`
}

// ProxyConfiguration defines config options for Proxy
type ProxyConfiguration struct {
	Image string `json:"image,omitempty"`
	// Configures the access log for each sidecar.
	// Options:
	//   "" - disables access log
	//   "/dev/stdout" - enables access log
	// +kubebuilder:validation:Enum=,/dev/stdout
	AccessLogFile *string `json:"accessLogFile,omitempty"`
	// Configure how and what fields are displayed in sidecar access log. Setting to
	// empty string will result in default log format.
	// If accessLogEncoding is TEXT, value will be used directly as the log format
	// example: "[%START_TIME%] %REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\n"
	// If AccessLogEncoding is JSON, value will be parsed as map[string]string
	// example: '{"start_time": "%START_TIME%", "req_method": "%REQ(:METHOD)%"}'
	AccessLogFormat *string `json:"accessLogFormat,omitempty"`
	// Configure the access log for sidecar to JSON or TEXT.
	// +kubebuilder:validation:Enum=JSON,TEXT
	AccessLogEncoding *string `json:"accessLogEncoding,omitempty"`
	// If set to true, istio-proxy container will have privileged securityContext
	Privileged bool `json:"privileged,omitempty"`
	// If set, newly injected sidecars will have core dumps enabled.
	EnableCoreDump *bool `json:"enableCoreDump,omitempty"`
	// Image used to enable core dumps. This is only used, when "EnableCoreDump" is set to true.
	CoreDumpImage string `json:"coreDumpImage,omitempty"`
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
	// cluster domain. Default value is "cluster.local"
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// Controls if sidecar is injected at the front of the container list and blocks the start of the other containers until the proxy is ready
	HoldApplicationUntilProxyStarts *bool `json:"holdApplicationUntilProxyStarts,omitempty"`

	EnvoyStatsD               EnvoyStatsD                     `json:"envoyStatsD,omitempty"`
	EnvoyMetricsService       EnvoyServiceCommonConfiguration `json:"envoyMetricsService,omitempty"`
	EnvoyAccessLogService     EnvoyServiceCommonConfiguration `json:"envoyAccessLogService,omitempty"`
	ProtocolDetectionTimeout  *string                         `json:"protocolDetectionTimeout,omitempty"`
	UseMetadataExchangeFilter *bool                           `json:"useMetadataExchangeFilter,omitempty"`

	Lifecycle corev1.Lifecycle `json:"lifecycle,omitempty"`

	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	SecurityContext *corev1.SecurityContext      `json:"securityContext,omitempty"`
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
	// TLS setting for Zipkin endpoint.
	TLSSettings *TLSSettings `json:"tlsSettings,omitempty"`
}

func (c ZipkinConfiguration) GetData() map[string]interface{} {
	data := map[string]interface{}{
		"address": c.Address,
	}

	return data
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

type StrackdriverConfiguration struct {
	// enables trace output to stdout.
	Debug *bool `json:"debug,omitempty"`
	// The global default max number of attributes per span.
	MaxNumberOfAttributes *int32 `json:"maxNumberOfAttributes,omitempty"`
	// The global default max number of annotation events per span.
	MaxNumberOfAnnotations *int32 `json:"maxNumberOfAnnotations,omitempty"`
	// The global default max number of message events per span.
	MaxNumberOfMessageEvents *int32 `json:"maxNumberOfMessageEvents,omitempty"`
}

type TracerType string

const (
	TracerTypeZipkin      TracerType = "zipkin"
	TracerTypeLightstep   TracerType = "lightstep"
	TracerTypeDatadog     TracerType = "datadog"
	TracerTypeStackdriver TracerType = "stackdriver"
)

type TracingConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Enum=zipkin,lightstep,datadog,stackdriver
	Tracer       TracerType                `json:"tracer,omitempty"`
	Zipkin       ZipkinConfiguration       `json:"zipkin,omitempty"`
	Lightstep    LightstepConfiguration    `json:"lightstep,omitempty"`
	Datadog      DatadogConfiugration      `json:"datadog,omitempty"`
	Strackdriver StrackdriverConfiguration `json:"stackdriver,omitempty"`
}

type IstioCoreDNS struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	PluginImage                         string `json:"pluginImage,omitempty"`
}

// Describes how traffic originating in the 'from' zone is
// distributed over a set of 'to' zones. Syntax for specifying a zone is
// {region}/{zone} and terminal wildcards are allowed on any
// segment of the specification. Examples:
// * - matches all localities
// us-west/* - all zones and sub-zones within the us-west region
type LocalityLBDistributeConfiguration struct {
	// Originating locality, '/' separated, e.g. 'region/zone'.
	From string `json:"from,omitempty"`
	// Map of upstream localities to traffic distribution weights. The sum of
	// all weights should be == 100. Any locality not assigned a weight will
	// receive no traffic.
	To map[string]uint32 `json:"to,omitempty"`
}

// Specify the traffic failover policy across regions. Since zone
// failover is supported by default this only needs to be specified for
// regions when the operator needs to constrain traffic failover so that
// the default behavior of failing over to any endpoint globally does not
// apply. This is useful when failing over traffic across regions would not
// improve service health or may need to be restricted for other reasons
// like regulatory controls.
type LocalityLBFailoverConfiguration struct {
	// Originating region.
	From string `json:"from,omitempty"`
	// Destination region the traffic will fail over to when endpoints in
	// the 'from' region becomes unhealthy.
	To string `json:"to,omitempty"`
}

// Locality-weighted load balancing allows administrators to control the
// distribution of traffic to endpoints based on the localities of where the
// traffic originates and where it will terminate.
type LocalityLBConfiguration struct {
	// If set to true, locality based load balancing will be enabled
	Enabled *bool `json:"enabled,omitempty"`
	// Optional: only one of distribute or failover can be set.
	// Explicitly specify loadbalancing weight across different zones and geographical locations.
	// Refer to [Locality weighted load balancing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/load_balancing/locality_weight)
	// If empty, the locality weight is set according to the endpoints number within it.
	Distribute []*LocalityLBDistributeConfiguration `json:"distribute,omitempty"`
	// Optional: only failover or distribute can be set.
	// Explicitly specify the region traffic will land on when endpoints in local region becomes unhealthy.
	// Should be used together with OutlierDetection to detect unhealthy endpoints.
	// Note: if no OutlierDetection specified, this will not take effect.
	Failover []*LocalityLBFailoverConfiguration `json:"failover,omitempty"`
}

// CertificateConfig configures DNS certificates provisioned through Chiron linked into Pilot
type CertificateConfig struct {
	SecretName *string  `json:"secretName,omitempty"`
	DNSNames   []string `json:"dnsNames,omitempty"`
}

// Comma-separated minimum per-scope logging level of messages to output, in the form of <scope>:<level>,<scope>:<level>
// The control plane has different scopes depending on component, but can configure default log level across all components
// If empty, default scope and level will be used as configured in code
type LoggingConfiguration struct {
	// +kubebuilder:validation:Pattern=^([a-zA-Z]+:[a-zA-Z]+,?)+$
	Level *string `json:"level,omitempty"`
}

// MeshPolicyConfiguration configures the mesh-wide PeerAuthentication resource
type MeshPolicyConfiguration struct {
	// MTLSMode sets the mesh-wide mTLS policy
	// +kubebuilder:validation:Enum=STRICT,PERMISSIVE,DISABLED
	MTLSMode MTLSMode `json:"mtlsMode,omitempty"`
}

type MTLSMode string

const (
	STRICT     MTLSMode = "STRICT"
	PERMISSIVE MTLSMode = "PERMISSIVE"
	DISABLED   MTLSMode = "DISABLED"
)

type PilotCertProviderType string

const (
	PilotCertProviderTypeKubernetes PilotCertProviderType = "kubernetes"
	PilotCertProviderTypeIstiod     PilotCertProviderType = "istiod"
)

type JWTPolicyType string

const (
	JWTPolicyThirdPartyJWT JWTPolicyType = "third-party-jwt"
	JWTPolicyFirstPartyJWT JWTPolicyType = "first-party-jwt"
)

type ControlPlaneAuthPolicyType string

const (
	ControlPlaneAuthPolicyMTLS ControlPlaneAuthPolicyType = "MUTUAL_TLS"
	ControlPlaneAuthPolicyNone ControlPlaneAuthPolicyType = "NONE"
)

type HTTPProxyEnvs struct {
	HTTPProxy  string `json:"httpProxy,omitemtpy"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// Contains the intended Istio version
	// +kubebuilder:validation:Pattern=^1.
	Version IstioVersion `json:"version"`

	// Logging configurations
	Logging LoggingConfiguration `json:"logging,omitempty"`

	// MeshPolicy configures the mesh-wide PeerAuthentication resource
	MeshPolicy MeshPolicyConfiguration `json:"meshPolicy,omitempty"`

	// DEPRECATED: Use meshPolicy instead.
	// MTLS enables or disables global mTLS
	MTLS *bool `json:"mtls,omitempty"`

	// If set to true, and a given service does not have a corresponding DestinationRule configured,
	// or its DestinationRule does not have TLSSettings specified, Istio configures client side
	// TLS configuration automatically, based on the server side mTLS authentication policy and the
	// availability of sidecars.
	AutoMTLS *bool `json:"autoMtls,omitempty"`

	// IncludeIPRanges the range where to capture egress traffic
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`

	// ExcludeIPRanges the range where not to capture egress traffic
	ExcludeIPRanges string `json:"excludeIPRanges,omitempty"`

	// List of namespaces to label with sidecar auto injection enabled
	AutoInjectionNamespaces []string `json:"autoInjectionNamespaces,omitempty"`

	// ControlPlaneAuthPolicy defines how the proxy is authenticated when it connects to the control plane
	// +kubebuilder:validation:Enum=MUTUAL_TLS,NONE
	ControlPlaneAuthPolicy ControlPlaneAuthPolicyType `json:"controlPlaneAuthPolicy,omitempty"`

	// Use the user-specified, secret volume mounted key and certs for Pilot and workloads.
	MountMtlsCerts *bool `json:"mountMtlsCerts,omitempty"`

	// DefaultResources are applied for all Istio components by default, can be overridden for each component
	DefaultResources *corev1.ResourceRequirements `json:"defaultResources,omitempty"`

	// If SDS is configured, mTLS certificates for the sidecars will be distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates
	SDS SDSConfiguration `json:"sds,omitempty"`

	// Istiod configuration
	Istiod IstiodConfiguration `json:"istiod,omitempty"`

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

	// Policy configuration options
	Policy PolicyConfiguration `json:"policy,omitempty"`

	// Telemetry configuration options
	Telemetry TelemetryConfiguration `json:"telemetry,omitempty"`

	// SidecarInjector configuration options
	SidecarInjector SidecarInjectorConfiguration `json:"sidecarInjector,omitempty"`

	// ProxyWasm configuration options
	ProxyWasm ProxyWasmConfiguration `json:"proxyWasm,omitempty"`

	// NodeAgent configuration options
	NodeAgent NodeAgentConfiguration `json:"nodeAgent,omitempty"`

	// Proxy configuration options
	Proxy ProxyConfiguration `json:"proxy,omitempty"`

	// Proxy Init configuration options
	ProxyInit ProxyInitConfiguration `json:"proxyInit,omitempty"`

	// Whether to restrict the applications namespace the controller manages
	WatchOneNamespace bool `json:"watchOneNamespace,omitempty"`

	// Prior to Kubernetes v1.17.0 it was not allowed to use the system-cluster-critical and system-node-critical
	// PriorityClass outside of the kube-system namespace, so it is advised to create your own PriorityClass
	// and use its name here
	// On Kubernetes >=v1.17.0 it is possible to configure system-cluster-critical and
	// system-node-critical PriorityClass in order to make sure your Istio pods
	// will not be killed because of low priority class.
	// Refer to https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
	// for more detail.
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// Use the Mesh Control Protocol (MCP) for configuring Mixer and Pilot. Requires an MCP source.
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

	// Locality based load balancing distribution or failover settings.
	LocalityLB *LocalityLBConfiguration `json:"localityLB,omitempty"`

	// Should be set to the name of the cluster this installation will run in.
	// This is required for sidecar injection to properly label proxies
	ClusterName string `json:"clusterName,omitempty"`

	// Network defines the network this cluster belong to. This name
	// corresponds to the networks in the map of mesh networks.
	NetworkName string `json:"networkName,omitempty"`

	// Mesh ID means Mesh Identifier. It should be unique within the scope where
	// meshes will interact with each other, but it is not required to be
	// globally/universally unique.
	MeshID string `json:"meshID,omitempty"`

	// Mixerless telemetry configuration
	MixerlessTelemetry *MixerlessTelemetryConfiguration `json:"mixerlessTelemetry,omitempty"`

	meshNetworks *MeshNetworks

	// The domain serves to identify the system with SPIFFE. (default "cluster.local")
	TrustDomain string `json:"trustDomain,omitempty"`

	//  The trust domain aliases represent the aliases of trust_domain.
	//  For example, if we have
	//  trustDomain: td1
	//  trustDomainAliases: ["td2", "td3"]
	//  Any service with the identity "td1/ns/foo/sa/a-service-account", "td2/ns/foo/sa/a-service-account",
	//  or "td3/ns/foo/sa/a-service-account" will be treated the same in the Istio mesh.
	TrustDomainAliases []string `json:"trustDomainAliases,omitempty"`

	// Configures DNS certificates provisioned through Chiron linked into Pilot.
	// The DNS names in this file are all hard-coded; please ensure the namespaces
	// in dnsNames are consistent with those of your services.
	// Example:
	// certificates:
	// certificates:
	//   - secretName: dns.istiod-service-account
	//     dnsNames: [istiod.istio-system.svc, istiod.istio-system]
	// +k8s:deepcopy-gen:interfaces=Certificates
	Certificates []CertificateConfig `json:"certificates,omitempty"`

	// Configure the policy for validating JWT.
	// Currently, two options are supported: "third-party-jwt" and "first-party-jwt".
	// +kubebuilder:validation:Enum=third-party-jwt,first-party-jwt
	JWTPolicy JWTPolicyType `json:"jwtPolicy,omitempty"`

	// The customized CA address to retrieve certificates for the pods in the cluster.
	//CSR clients such as the Istio Agent and ingress gateways can use this to specify the CA endpoint.
	CAAddress string `json:"caAddress,omitempty"`

	// Upstream HTTP proxy properties to be injected as environment variables to the pod containers
	HTTPProxyEnvs HTTPProxyEnvs `json:"httpProxyEnvs,omitempty"`

	// Specifies whether the control plane is a global one or revisioned. There must be only one global control plane.
	Global *bool `json:"global,omitempty"`
}

type MixerlessTelemetryConfiguration struct {
	// If set to true, experimental Mixerless http telemetry will be enabled
	Enabled *bool `json:"enabled,omitempty"`
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

func (c *Istio) GetControlPlaneAuthPolicy() ControlPlaneAuthPolicyType {
	if c.Spec.ControlPlaneAuthPolicy != "" {
		return c.Spec.ControlPlaneAuthPolicy
	}

	return ControlPlaneAuthPolicyMTLS
}

func (c *Istio) GetCAAddress() string {
	if c.Spec.CAAddress != "" {
		return c.Spec.CAAddress
	}

	return c.GetDiscoveryAddress()
}

func (c *Istio) GetDiscoveryHost(withClusterDomain bool) string {
	svcName := "istio-pilot"
	if util.PointerToBool(c.Spec.Istiod.Enabled) {
		svcName = "istiod"
	}
	if withClusterDomain {
		return fmt.Sprintf("%s.%s.svc.%s", c.WithRevision(svcName), c.Namespace, c.Spec.Proxy.ClusterDomain)
	}
	return fmt.Sprintf("%s.%s.svc", c.WithRevision(svcName), c.Namespace)
}

func (c *Istio) GetDiscoveryAddress() string {
	return fmt.Sprintf("%s:%d", c.GetDiscoveryHost(false), c.GetDiscoveryPort())
}

func (c *Istio) GetDiscoveryPort() int {
	if util.PointerToBool(c.Spec.Istiod.Enabled) {
		return 15012
	}
	return 15011
}

func (c *Istio) GetWebhookPort() int {
	return 15017
}

func (c *Istio) IsRevisionUsed() bool {
	if c.Spec.Global == nil {
		return false
	}

	return !*c.Spec.Global
}

func NamespacedRevision(revision, namespace string) string {
	return fmt.Sprintf("%s.%s", revision, namespace)
}

func (c *Istio) Revision() string {
	return strings.Replace(c.Name, ".", "-", -1)
}

func (c *Istio) NamespacedRevision() string {
	return NamespacedRevision(c.Revision(), c.Namespace)
}

func (c *Istio) RevisionLabels() map[string]string {
	return map[string]string{
		"istio.io/rev": c.NamespacedRevision(),
	}
}

func (c *Istio) WithRevision(s string) string {
	if !c.IsRevisionUsed() {
		return s
	}

	return strings.Join([]string{s, c.Revision()}, "-")
}

func (c *Istio) WithRevisionIf(s string, condition bool) string {
	if !condition {
		return s
	}

	return c.WithRevision(s)
}

func (c *Istio) WithNamespacedRevision(s string) string {
	if !c.IsRevisionUsed() {
		return s
	}
	return strings.Join([]string{c.WithRevision(s), c.Namespace}, "-")
}

type SortableIstioItems []Istio

func (list SortableIstioItems) Len() int {
	return len(list)
}

func (list SortableIstioItems) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableIstioItems) Less(i, j int) bool {
	return list[i].CreationTimestamp.Time.Before(list[j].CreationTimestamp.Time)
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	Status         ConfigState `json:"Status,omitempty"`
	GatewayAddress []string    `json:"GatewayAddress,omitempty"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Istio is the Schema for the istios API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
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
