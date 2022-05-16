## IstioControlPlaneSpec

IstioControlPlane defines an Istio control plane

<!-- crd generation tags
+cue-gen:IstioControlPlane:groupName:servicemesh.cisco.com
+cue-gen:IstioControlPlane:version:v1alpha1
+cue-gen:IstioControlPlane:storageVersion
+cue-gen:IstioControlPlane:annotations:helm.sh/resource-policy=keep
+cue-gen:IstioControlPlane:subresource:status
+cue-gen:IstioControlPlane:scope:Namespaced
+cue-gen:IstioControlPlane:resource:shortNames=icp,istiocp
+cue-gen:IstioControlPlane:printerColumn:name="Mode",type="string",JSONPath=".spec.mode",description="Mode for the Istio control plane"
+cue-gen:IstioControlPlane:printerColumn:name="Network",type="string",JSONPath=".spec.networkName",description="The network this cluster belongs to"
+cue-gen:IstioControlPlane:printerColumn:name="Status",type="string",JSONPath=".status.status",description="Status of the resource"
+cue-gen:IstioControlPlane:printerColumn:name="Mesh expansion",type="string",JSONPath=".spec.meshExpansion.enabled",description="Whether mesh expansion is enabled"
+cue-gen:IstioControlPlane:printerColumn:name="Expansion GW IPs",type="string",JSONPath=".status.gatewayAddress",description="IP addresses of the mesh expansion gateway"
+cue-gen:IstioControlPlane:printerColumn:name="Error",type="string",JSONPath=".status.errorMessage",description="Error message"
+cue-gen:IstioControlPlane:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp
representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.
Clients may not set this value. It is represented in RFC3339 form and is in UTC.
Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
+cue-gen:IstioControlPlane:preserveUnknownFields:false
+cue-gen:IstioControlPlane:aliases:PeerIstioControlPlane
-->

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### version (string, optional) {#istiocontrolplanespec-version}

Contains the intended version for the Istio control plane. 

Default: -

### mode (ModeType, optional) {#istiocontrolplanespec-mode}

Configure the mode for this control plane. Currently, two options are supported: "ACTIVE" and "PASSIVE". ACTIVE mode means that a full-fledged Istio control plane will be deployed and operated (usually called primary cluster in upstream Istio terminology). PASSIVE mode means that only a few resources will be installed for sidecar injection and cross-cluster communication, it is used for multi cluster setups (this is the remote cluster in upstream Istio terminology). 

Default: -

### logging (*LoggingConfiguration, optional) {#istiocontrolplanespec-logging}

Logging configurations. 

Default: -

### mountMtlsCerts (*bool, optional) {#istiocontrolplanespec-mountmtlscerts}

Use the user-specified, secret volume mounted key and certs for Pilot and workloads. 

Default: -

### istiod (*IstiodConfiguration, optional) {#istiocontrolplanespec-istiod}

Istiod configuration. 

Default: -

### proxy (*ProxyConfiguration, optional) {#istiocontrolplanespec-proxy}

Proxy configuration options. 

Default: -

### proxyInit (*ProxyInitConfiguration, optional) {#istiocontrolplanespec-proxyinit}

Proxy Init configuration options. 

Default: -

### telemetryV2 (*TelemetryV2Configuration, optional) {#istiocontrolplanespec-telemetryv2}

Telemetry V2 configuration. 

Default: -

### sds (*SDSConfiguration, optional) {#istiocontrolplanespec-sds}

If SDS is configured, mTLS certificates for the sidecars will be distributed through the SecretDiscoveryService instead of using K8S secrets to mount the certificates. 

Default: -

### proxyWasm (*ProxyWasmConfiguration, optional) {#istiocontrolplanespec-proxywasm}

ProxyWasm configuration options. 

Default: -

### watchOneNamespace (*bool, optional) {#istiocontrolplanespec-watchonenamespace}

Whether to restrict the applications namespace the controller manages. If not set, controller watches all namespaces 

Default: -

### jwtPolicy (JWTPolicyType, optional) {#istiocontrolplanespec-jwtpolicy}

Configure the policy for validating JWT. Currently, two options are supported: "third-party-jwt" and "first-party-jwt". 

Default: -

### caAddress (string, optional) {#istiocontrolplanespec-caaddress}

The customized CA address to retrieve certificates for the pods in the cluster. CSR clients such as the Istio Agent and ingress gateways can use this to specify the CA endpoint. 

Default: -

### caProvider (string, optional) {#istiocontrolplanespec-caprovider}

The name of the CA for workload certificates. 

Default: -

### distribution (string, optional) {#istiocontrolplanespec-distribution}

Contains the intended distribution for the Istio control plane. The official distribution is used by default unless special preserved distribution value is set. The only preserved distribution is "cisco" as of now. 

Default: -

### httpProxyEnvs (*HTTPProxyEnvsConfiguration, optional) {#istiocontrolplanespec-httpproxyenvs}

Upstream HTTP proxy properties to be injected as environment variables to the pod containers. 

Default: -

### meshConfig (*v1alpha1.MeshConfig, optional) {#istiocontrolplanespec-meshconfig}

Defines mesh-wide settings for the Istio control plane. 

Default: -

### k8sResourceOverlays ([]*K8SResourceOverlayPatch, optional) {#istiocontrolplanespec-k8sresourceoverlays}

K8s resource overlay patches 

Default: -

### meshID (string, optional) {#istiocontrolplanespec-meshid}

Name of the Mesh to which this control plane belongs. 

Default: -

### containerImageConfiguration (*ContainerImageConfiguration, optional) {#istiocontrolplanespec-containerimageconfiguration}

Global configuration for container images. 

Default: -

### meshExpansion (*MeshExpansionConfiguration, optional) {#istiocontrolplanespec-meshexpansion}

Mesh expansion configuration 

Default: -

### clusterID (string, optional) {#istiocontrolplanespec-clusterid}

Cluster ID 

Default: -

### networkName (string, optional) {#istiocontrolplanespec-networkname}

Network defines the network this cluster belongs to. This name corresponds to the networks in the map of mesh networks. +default=network1 

Default: -

### sidecarInjector (*SidecarInjectorConfiguration, optional) {#istiocontrolplanespec-sidecarinjector}

Standalone sidecar injector configuration. 

Default: -

### - (struct{}, required) {#istiocontrolplanespec--}

Default: -

### - ([]byte, required) {#istiocontrolplanespec--}

Default: -

### - (int32, required) {#istiocontrolplanespec--}

Default: -


## SidecarInjectorConfiguration

### deployment (*BaseKubernetesResourceConfig, optional) {#sidecarinjectorconfiguration-deployment}

Deployment spec 

Default: -

### service (*Service, optional) {#sidecarinjectorconfiguration-service}

Service spec 

Default: -

### - (struct{}, required) {#sidecarinjectorconfiguration--}

Default: -

### - ([]byte, required) {#sidecarinjectorconfiguration--}

Default: -

### - (int32, required) {#sidecarinjectorconfiguration--}

Default: -


## MeshExpansionConfiguration

### enabled (*bool, optional) {#meshexpansionconfiguration-enabled}

Default: -

### gateway (*MeshExpansionConfiguration_IstioMeshGatewayConfiguration, optional) {#meshexpansionconfiguration-gateway}

Default: -

### istiod (*MeshExpansionConfiguration_Istiod, optional) {#meshexpansionconfiguration-istiod}

istiod component configuration 

Default: -

### webhook (*MeshExpansionConfiguration_Webhook, optional) {#meshexpansionconfiguration-webhook}

webhook component configuration 

Default: -

### clusterServices (*MeshExpansionConfiguration_ClusterServices, optional) {#meshexpansionconfiguration-clusterservices}

cluster services configuration 

Default: -

### - (struct{}, required) {#meshexpansionconfiguration--}

Default: -

### - ([]byte, required) {#meshexpansionconfiguration--}

Default: -

### - (int32, required) {#meshexpansionconfiguration--}

Default: -


## MeshExpansionConfiguration_Istiod

### expose (*bool, optional) {#meshexpansionconfiguration_istiod-expose}

Default: -

### - (struct{}, required) {#meshexpansionconfiguration_istiod--}

Default: -

### - ([]byte, required) {#meshexpansionconfiguration_istiod--}

Default: -

### - (int32, required) {#meshexpansionconfiguration_istiod--}

Default: -


## MeshExpansionConfiguration_Webhook

### expose (*bool, optional) {#meshexpansionconfiguration_webhook-expose}

Default: -

### - (struct{}, required) {#meshexpansionconfiguration_webhook--}

Default: -

### - ([]byte, required) {#meshexpansionconfiguration_webhook--}

Default: -

### - (int32, required) {#meshexpansionconfiguration_webhook--}

Default: -


## MeshExpansionConfiguration_ClusterServices

### expose (*bool, optional) {#meshexpansionconfiguration_clusterservices-expose}

Default: -

### - (struct{}, required) {#meshexpansionconfiguration_clusterservices--}

Default: -

### - ([]byte, required) {#meshexpansionconfiguration_clusterservices--}

Default: -

### - (int32, required) {#meshexpansionconfiguration_clusterservices--}

Default: -


## MeshExpansionConfiguration_IstioMeshGatewayConfiguration

### metadata (*K8SObjectMeta, optional) {#meshexpansionconfiguration_istiomeshgatewayconfiguration-metadata}

Istio Mesh gateway metadata 

Default: -

### deployment (*BaseKubernetesResourceConfig, optional) {#meshexpansionconfiguration_istiomeshgatewayconfiguration-deployment}

Deployment spec 

Default: -

### service (*UnprotectedService, optional) {#meshexpansionconfiguration_istiomeshgatewayconfiguration-service}

Service spec 

Default: -

### runAsRoot (*bool, optional) {#meshexpansionconfiguration_istiomeshgatewayconfiguration-runasroot}

Whether to run the gateway in a privileged container 

Default: -

### k8sResourceOverlays ([]*K8SResourceOverlayPatch, optional) {#meshexpansionconfiguration_istiomeshgatewayconfiguration-k8sresourceoverlays}

K8s resource overlay patches 

Default: -

### - (struct{}, required) {#meshexpansionconfiguration_istiomeshgatewayconfiguration--}

Default: -

### - ([]byte, required) {#meshexpansionconfiguration_istiomeshgatewayconfiguration--}

Default: -

### - (int32, required) {#meshexpansionconfiguration_istiomeshgatewayconfiguration--}

Default: -


## LoggingConfiguration

Comma-separated minimum per-scope logging level of messages to output, in the form of <scope>:<level>,<scope>:<level>
The control plane has different scopes depending on component, but can configure default log level across all components
If empty, default scope and level will be used as configured in code

### level (string, optional) {#loggingconfiguration-level}

Default: -

### - (struct{}, required) {#loggingconfiguration--}

Default: -

### - ([]byte, required) {#loggingconfiguration--}

Default: -

### - (int32, required) {#loggingconfiguration--}

Default: -


## SDSConfiguration

SDSConfiguration defines Secret Discovery Service config options

### tokenAudience (string, optional) {#sdsconfiguration-tokenaudience}

The JWT token for SDS and the aud field of such JWT. See RFC 7519, section 4.1.3. When a CSR is sent from Citadel Agent to the CA (e.g. Citadel), this aud is to make sure the JWT is intended for the CA. 

Default: -

### - (struct{}, required) {#sdsconfiguration--}

Default: -

### - ([]byte, required) {#sdsconfiguration--}

Default: -

### - (int32, required) {#sdsconfiguration--}

Default: -


## ProxyConfiguration

ProxyConfiguration defines config options for Proxy

### image (string, optional) {#proxyconfiguration-image}

Default: -

### privileged (*bool, optional) {#proxyconfiguration-privileged}

If set to true, istio-proxy container will have privileged securityContext 

Default: -

### enableCoreDump (*bool, optional) {#proxyconfiguration-enablecoredump}

If set, newly injected sidecars will have core dumps enabled. 

Default: -

### logLevel (ProxyLogLevel, optional) {#proxyconfiguration-loglevel}

Log level for proxy, applies to gateways and sidecars. If left empty, "warning" is used. Expected values are: trace|debug|info|warning|error|critical|off 

Default: -

### componentLogLevel (string, optional) {#proxyconfiguration-componentloglevel}

Per Component log level for proxy, applies to gateways and sidecars. If a component level is not set, then the "LogLevel" will be used. If left empty, "misc:error" is used. 

Default: -

### clusterDomain (string, optional) {#proxyconfiguration-clusterdomain}

cluster domain. Default value is "cluster.local" 

Default: -

### holdApplicationUntilProxyStarts (*bool, optional) {#proxyconfiguration-holdapplicationuntilproxystarts}

Controls if sidecar is injected at the front of the container list and blocks the start of the other containers until the proxy is ready Default value is 'false'. 

Default: -

### lifecycle (*v1.Lifecycle, optional) {#proxyconfiguration-lifecycle}

Default: -

### resources (*ResourceRequirements, optional) {#proxyconfiguration-resources}

Default: -

### includeIPRanges (string, optional) {#proxyconfiguration-includeipranges}

IncludeIPRanges the range where to capture egress traffic 

Default: -

### excludeIPRanges (string, optional) {#proxyconfiguration-excludeipranges}

ExcludeIPRanges the range where not to capture egress traffic 

Default: -

### excludeInboundPorts (string, optional) {#proxyconfiguration-excludeinboundports}

ExcludeInboundPorts the comma separated list of inbound ports to be excluded from redirection to Envoy 

Default: -

### excludeOutboundPorts (string, optional) {#proxyconfiguration-excludeoutboundports}

ExcludeOutboundPorts the comma separated list of outbound ports to be excluded from redirection to Envoy 

Default: -

### - (struct{}, required) {#proxyconfiguration--}

Default: -

### - ([]byte, required) {#proxyconfiguration--}

Default: -

### - (int32, required) {#proxyconfiguration--}

Default: -


## ProxyInitConfiguration

ProxyInitConfiguration defines config options for Proxy Init containers

### image (string, optional) {#proxyinitconfiguration-image}

Default: -

### resources (*ResourceRequirements, optional) {#proxyinitconfiguration-resources}

Default: -

### cni (*CNIConfiguration, optional) {#proxyinitconfiguration-cni}

Default: -

### - (struct{}, required) {#proxyinitconfiguration--}

Default: -

### - ([]byte, required) {#proxyinitconfiguration--}

Default: -

### - (int32, required) {#proxyinitconfiguration--}

Default: -


## CNIConfiguration

### enabled (*bool, optional) {#cniconfiguration-enabled}

Default: -

### chained (*bool, optional) {#cniconfiguration-chained}

Default: -

### binDir (string, optional) {#cniconfiguration-bindir}

Default: -

### confDir (string, optional) {#cniconfiguration-confdir}

Default: -

### excludeNamespaces ([]string, optional) {#cniconfiguration-excludenamespaces}

Default: -

### includeNamespaces ([]string, optional) {#cniconfiguration-includenamespaces}

Default: -

### logLevel (string, optional) {#cniconfiguration-loglevel}

Default: -

### confFileName (string, optional) {#cniconfiguration-conffilename}

Default: -

### pspClusterRoleName (string, optional) {#cniconfiguration-pspclusterrolename}

Default: -

### repair (*CNIConfiguration_RepairConfiguration, optional) {#cniconfiguration-repair}

Default: -

### taint (*CNIConfiguration_TaintConfiguration, optional) {#cniconfiguration-taint}

Default: -

### resourceQuotas (*CNIConfiguration_ResourceQuotas, optional) {#cniconfiguration-resourcequotas}

Default: -

### daemonset (*BaseKubernetesResourceConfig, optional) {#cniconfiguration-daemonset}

Default: -

### - (struct{}, required) {#cniconfiguration--}

Default: -

### - ([]byte, required) {#cniconfiguration--}

Default: -

### - (int32, required) {#cniconfiguration--}

Default: -


## CNIConfiguration_RepairConfiguration

### enabled (*bool, optional) {#cniconfiguration_repairconfiguration-enabled}

Default: -

### labelPods (*bool, optional) {#cniconfiguration_repairconfiguration-labelpods}

Default: -

### deletePods (*bool, optional) {#cniconfiguration_repairconfiguration-deletepods}

Default: -

### initContainerName (string, optional) {#cniconfiguration_repairconfiguration-initcontainername}

Default: -

### brokenPodLabelKey (string, optional) {#cniconfiguration_repairconfiguration-brokenpodlabelkey}

Default: -

### brokenPodLabelValue (string, optional) {#cniconfiguration_repairconfiguration-brokenpodlabelvalue}

Default: -

### - (struct{}, required) {#cniconfiguration_repairconfiguration--}

Default: -

### - ([]byte, required) {#cniconfiguration_repairconfiguration--}

Default: -

### - (int32, required) {#cniconfiguration_repairconfiguration--}

Default: -


## CNIConfiguration_TaintConfiguration

### enabled (*bool, optional) {#cniconfiguration_taintconfiguration-enabled}

Default: -

### container (*BaseKubernetesContainerConfiguration, optional) {#cniconfiguration_taintconfiguration-container}

Default: -

### - (struct{}, required) {#cniconfiguration_taintconfiguration--}

Default: -

### - ([]byte, required) {#cniconfiguration_taintconfiguration--}

Default: -

### - (int32, required) {#cniconfiguration_taintconfiguration--}

Default: -


## CNIConfiguration_ResourceQuotas

### enabled (*bool, optional) {#cniconfiguration_resourcequotas-enabled}

Default: -

### pods (string, optional) {#cniconfiguration_resourcequotas-pods}

Default: -

### priorityClasses ([]string, optional) {#cniconfiguration_resourcequotas-priorityclasses}

Default: -

### - (struct{}, required) {#cniconfiguration_resourcequotas--}

Default: -

### - ([]byte, required) {#cniconfiguration_resourcequotas--}

Default: -

### - (int32, required) {#cniconfiguration_resourcequotas--}

Default: -


## IstiodConfiguration

IstiodConfiguration defines config options for Istiod

### deployment (*BaseKubernetesResourceConfig, optional) {#istiodconfiguration-deployment}

Deployment spec 

Default: -

### enableAnalysis (*bool, optional) {#istiodconfiguration-enableanalysis}

If enabled, pilot will run Istio analyzers and write analysis errors to the Status field of any Istio Resources 

Default: -

### enableStatus (*bool, optional) {#istiodconfiguration-enablestatus}

If enabled, pilot will update the CRD Status field of all Istio resources with reconciliation status 

Default: -

### externalIstiod (*ExternalIstiodConfiguration, optional) {#istiodconfiguration-externalistiod}

Settings for local istiod to control remote clusters as well 

Default: -

### traceSampling (*float32, optional) {#istiodconfiguration-tracesampling}

Default: -

### enableProtocolSniffingOutbound (*bool, optional) {#istiodconfiguration-enableprotocolsniffingoutbound}

If enabled, protocol sniffing will be used for outbound listeners whose port protocol is not specified or unsupported 

Default: -

### enableProtocolSniffingInbound (*bool, optional) {#istiodconfiguration-enableprotocolsniffinginbound}

If enabled, protocol sniffing will be used for inbound listeners whose port protocol is not specified or unsupported 

Default: -

### certProvider (PilotCertProviderType, optional) {#istiodconfiguration-certprovider}

Configure the certificate provider for control plane communication. Currently, two providers are supported: "kubernetes" and "istiod". As some platforms may not have kubernetes signing APIs, Istiod is the default 

Default: -

### spiffe (*SPIFFEConfiguration, optional) {#istiodconfiguration-spiffe}

SPIFFE configuration of Pilot 

Default: -

### - (struct{}, required) {#istiodconfiguration--}

Default: -

### - ([]byte, required) {#istiodconfiguration--}

Default: -

### - (int32, required) {#istiodconfiguration--}

Default: -


## ExternalIstiodConfiguration

ExternalIstiodConfiguration defines settings for local istiod to control remote clusters as well

### enabled (*bool, optional) {#externalistiodconfiguration-enabled}

Default: -

### - (struct{}, required) {#externalistiodconfiguration--}

Default: -

### - ([]byte, required) {#externalistiodconfiguration--}

Default: -

### - (int32, required) {#externalistiodconfiguration--}

Default: -


## SPIFFEConfiguration

SPIFFEConfiguration is for SPIFFE configuration of Pilot

### operatorEndpoints (*OperatorEndpointsConfiguration, optional) {#spiffeconfiguration-operatorendpoints}

Default: -

### - (struct{}, required) {#spiffeconfiguration--}

Default: -

### - ([]byte, required) {#spiffeconfiguration--}

Default: -

### - (int32, required) {#spiffeconfiguration--}

Default: -


## OperatorEndpointsConfiguration

OperatorEndpointsConfiguration defines config options for automatic SPIFFE endpoints

### enabled (*bool, optional) {#operatorendpointsconfiguration-enabled}

Default: -

### - (struct{}, required) {#operatorendpointsconfiguration--}

Default: -

### - ([]byte, required) {#operatorendpointsconfiguration--}

Default: -

### - (int32, required) {#operatorendpointsconfiguration--}

Default: -


## TelemetryV2Configuration

### enabled (*bool, optional) {#telemetryv2configuration-enabled}

Default: -

### - (struct{}, required) {#telemetryv2configuration--}

Default: -

### - ([]byte, required) {#telemetryv2configuration--}

Default: -

### - (int32, required) {#telemetryv2configuration--}

Default: -


## ProxyWasmConfiguration

ProxyWasmConfiguration defines config options for Envoy wasm

### enabled (*bool, optional) {#proxywasmconfiguration-enabled}

Default: -

### - (struct{}, required) {#proxywasmconfiguration--}

Default: -

### - ([]byte, required) {#proxywasmconfiguration--}

Default: -

### - (int32, required) {#proxywasmconfiguration--}

Default: -


## PDBConfiguration

PDBConfiguration holds Pod Disruption Budget related config options

### enabled (*bool, optional) {#pdbconfiguration-enabled}

Default: -

### - (struct{}, required) {#pdbconfiguration--}

Default: -

### - ([]byte, required) {#pdbconfiguration--}

Default: -

### - (int32, required) {#pdbconfiguration--}

Default: -


## HTTPProxyEnvsConfiguration

### httpProxy (string, optional) {#httpproxyenvsconfiguration-httpproxy}

Default: -

### httpsProxy (string, optional) {#httpproxyenvsconfiguration-httpsproxy}

Default: -

### noProxy (string, optional) {#httpproxyenvsconfiguration-noproxy}

Default: -

### - (struct{}, required) {#httpproxyenvsconfiguration--}

Default: -

### - ([]byte, required) {#httpproxyenvsconfiguration--}

Default: -

### - (int32, required) {#httpproxyenvsconfiguration--}

Default: -


## IstioControlPlaneStatus

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### status (ConfigState, optional) {#istiocontrolplanestatus-status}

Reconciliation status of the Istio control plane 

Default: -

### clusterID (string, optional) {#istiocontrolplanestatus-clusterid}

Cluster ID 

Default: -

### istioControlPlaneName (string, optional) {#istiocontrolplanestatus-istiocontrolplanename}

Name of the IstioControlPlane resource It is used on remote clusters in the PeerIstioControlPlane resource status to identify the original Istio control plane 

Default: -

### gatewayAddress ([]string, optional) {#istiocontrolplanestatus-gatewayaddress}

Current addresses for the corresponding gateways 

Default: -

### istiodAddresses ([]string, optional) {#istiocontrolplanestatus-istiodaddresses}

Current addresses for the corresponding istiod pods 

Default: -

### injectionNamespaces ([]string, optional) {#istiocontrolplanestatus-injectionnamespaces}

Namespaces which are set for injection for this control plane 

Default: -

### caRootCertificate (string, optional) {#istiocontrolplanestatus-carootcertificate}

Istio CA root certificate 

Default: -

### errorMessage (string, optional) {#istiocontrolplanestatus-errormessage}

Reconciliation error message if any 

Default: -

### meshConfig (*v1alpha1.MeshConfig, optional) {#istiocontrolplanestatus-meshconfig}

Default: -

### checksums (*StatusChecksums, optional) {#istiocontrolplanestatus-checksums}

Default: -

### - (struct{}, required) {#istiocontrolplanestatus--}

Default: -

### - ([]byte, required) {#istiocontrolplanestatus--}

Default: -

### - (int32, required) {#istiocontrolplanestatus--}

Default: -


## StatusChecksums

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### meshConfig (string, optional) {#statuschecksums-meshconfig}

Default: -

### sidecarInjector (string, optional) {#statuschecksums-sidecarinjector}

Default: -

### - (struct{}, required) {#statuschecksums--}

Default: -

### - ([]byte, required) {#statuschecksums--}

Default: -

### - (int32, required) {#statuschecksums--}

Default: -


