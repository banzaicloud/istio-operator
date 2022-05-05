## K8SObjectMeta

Generic k8s resource metadata

### labels (map[string]string, optional) {#k8sobjectmeta-labels}

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels +optional 

Default: -

### annotations (map[string]string, optional) {#k8sobjectmeta-annotations}

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations +optional 

Default: -

### - (struct{}, required) {#k8sobjectmeta--}

Default: -

### - ([]byte, required) {#k8sobjectmeta--}

Default: -

### - (int32, required) {#k8sobjectmeta--}

Default: -


## ContainerImageConfiguration

### hub (string, optional) {#containerimageconfiguration-hub}

Default hub for container images. 

Default: -

### tag (string, optional) {#containerimageconfiguration-tag}

Default tag for container images. 

Default: -

### imagePullPolicy (string, optional) {#containerimageconfiguration-imagepullpolicy}

Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. +optional 

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, required) {#containerimageconfiguration-imagepullsecrets}

ImagePullSecrets is an optional list of references to secrets to use for pulling any of the images. +optional 

Default: -

### - (struct{}, required) {#containerimageconfiguration--}

Default: -

### - ([]byte, required) {#containerimageconfiguration--}

Default: -

### - (int32, required) {#containerimageconfiguration--}

Default: -


## BaseKubernetesContainerConfiguration

### image (string, optional) {#basekubernetescontainerconfiguration-image}

Standard Kubernetes container image configuration 

Default: -

### env ([]v1.EnvVar, required) {#basekubernetescontainerconfiguration-env}

If present will be appended to the environment variables of the container 

Default: -

### resources (*ResourceRequirements, optional) {#basekubernetescontainerconfiguration-resources}

Standard Kubernetes resource configuration, memory and CPU resource requirements 

Default: -

### securityContext (*v1.SecurityContext, optional) {#basekubernetescontainerconfiguration-securitycontext}

Standard Kubernetes security context configuration 

Default: -

### volumeMounts ([]v1.VolumeMount, required) {#basekubernetescontainerconfiguration-volumemounts}

Pod volumes to mount into the container's filesystem. Cannot be updated. +optional +patchMergeKey=mountPath +patchStrategy=merge 

Default: -

### - (struct{}, required) {#basekubernetescontainerconfiguration--}

Default: -

### - ([]byte, required) {#basekubernetescontainerconfiguration--}

Default: -

### - (int32, required) {#basekubernetescontainerconfiguration--}

Default: -


## BaseKubernetesResourceConfig

### metadata (*K8SObjectMeta, optional) {#basekubernetesresourceconfig-metadata}

Generic k8s resource metadata 

Default: -

### image (string, optional) {#basekubernetesresourceconfig-image}

Standard Kubernetes container image configuration 

Default: -

### env ([]v1.EnvVar, required) {#basekubernetesresourceconfig-env}

If present will be appended to the environment variables of the container 

Default: -

### resources (*ResourceRequirements, optional) {#basekubernetesresourceconfig-resources}

Standard Kubernetes resource configuration, memory and CPU resource requirements 

Default: -

### nodeSelector (map[string]string, optional) {#basekubernetesresourceconfig-nodeselector}

Standard Kubernetes node selector configuration 

Default: -

### affinity (*v1.Affinity, optional) {#basekubernetesresourceconfig-affinity}

Standard Kubernetes affinity configuration 

Default: -

### securityContext (*v1.SecurityContext, optional) {#basekubernetesresourceconfig-securitycontext}

Standard Kubernetes security context configuration 

Default: -

### imagePullPolicy (string, optional) {#basekubernetesresourceconfig-imagepullpolicy}

Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if :latest tag is specified, or IfNotPresent otherwise. +optional 

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, required) {#basekubernetesresourceconfig-imagepullsecrets}

ImagePullSecrets is an optional list of references to secrets to use for pulling any of the images. +optional 

Default: -

### priorityClassName (string, optional) {#basekubernetesresourceconfig-priorityclassname}

If specified, indicates the pod's priority. "system-node-critical" and "system-cluster-critical" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default. +optional 

Default: -

### tolerations ([]v1.Toleration, required) {#basekubernetesresourceconfig-tolerations}

google.protobuf.Int32Value replicaCount = 1 [(gogoproto.wktpointer) = true]; If specified, the pod's tolerations. +optional 

Default: -

### volumes ([]v1.Volume, required) {#basekubernetesresourceconfig-volumes}

List of volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes +optional +patchMergeKey=name +patchStrategy=merge,retainKeys 

Default: -

### volumeMounts ([]v1.VolumeMount, required) {#basekubernetesresourceconfig-volumemounts}

Pod volumes to mount into the container's filesystem. Cannot be updated. +optional +patchMergeKey=mountPath +patchStrategy=merge 

Default: -

### replicas (*Replicas, optional) {#basekubernetesresourceconfig-replicas}

Replica configuration 

Default: -

### podMetadata (*K8SObjectMeta, optional) {#basekubernetesresourceconfig-podmetadata}

Standard Kubernetes pod annotation and label configuration 

Default: -

### podDisruptionBudget (*PodDisruptionBudget, optional) {#basekubernetesresourceconfig-poddisruptionbudget}

PodDisruptionBudget configuration 

Default: -

### deploymentStrategy (*DeploymentStrategy, optional) {#basekubernetesresourceconfig-deploymentstrategy}

DeploymentStrategy configuration 

Default: -

### podSecurityContext (*v1.PodSecurityContext, optional) {#basekubernetesresourceconfig-podsecuritycontext}

Standard Kubernetes pod security context configuration 

Default: -

### livenessProbe (*v1.Probe, optional) {#basekubernetesresourceconfig-livenessprobe}

Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes +optional 

Default: -

### readinessProbe (*v1.Probe, optional) {#basekubernetesresourceconfig-readinessprobe}

Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes +optional 

Default: -

### - (struct{}, required) {#basekubernetesresourceconfig--}

Default: -

### - ([]byte, required) {#basekubernetesresourceconfig--}

Default: -

### - (int32, required) {#basekubernetesresourceconfig--}

Default: -


## DeploymentStrategy

### type (string, optional) {#deploymentstrategy-type}

Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate. +optional 

Default: -

### rollingUpdate (*DeploymentStrategy_RollingUpdateDeployment, optional) {#deploymentstrategy-rollingupdate}

Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate. +optional 

Default: -

### - (struct{}, required) {#deploymentstrategy--}

Default: -

### - ([]byte, required) {#deploymentstrategy--}

Default: -

### - (int32, required) {#deploymentstrategy--}

Default: -


## DeploymentStrategy_RollingUpdateDeployment

### maxUnavailable (*IntOrString, optional) {#deploymentstrategy_rollingupdatedeployment-maxunavailable}

Default: -

### maxSurge (*IntOrString, optional) {#deploymentstrategy_rollingupdatedeployment-maxsurge}

Default: -

### - (struct{}, required) {#deploymentstrategy_rollingupdatedeployment--}

Default: -

### - ([]byte, required) {#deploymentstrategy_rollingupdatedeployment--}

Default: -

### - (int32, required) {#deploymentstrategy_rollingupdatedeployment--}

Default: -


## PodDisruptionBudget

PodDisruptionBudget is a description of a PodDisruptionBudget

### minAvailable (*IntOrString, optional) {#poddisruptionbudget-minavailable}

An eviction is allowed if at least "minAvailable" pods selected by "selector" will still be available after the eviction, i.e. even in the absence of the evicted pod.  So for example you can prevent all voluntary evictions by specifying "100%". +optional 

Default: -

### maxUnavailable (*IntOrString, optional) {#poddisruptionbudget-maxunavailable}

An eviction is allowed if at most "maxUnavailable" pods selected by "selector" are unavailable after the eviction, i.e. even in absence of the evicted pod. For example, one can prevent all voluntary evictions by specifying 0. This is a mutually exclusive setting with "minAvailable". +optional 

Default: -

### - (struct{}, required) {#poddisruptionbudget--}

Default: -

### - ([]byte, required) {#poddisruptionbudget--}

Default: -

### - (int32, required) {#poddisruptionbudget--}

Default: -


## Service

Service describes the attributes that a user creates on a service.

### metadata (*K8SObjectMeta, optional) {#service-metadata}

Default: -

### ports ([]ServicePort, required) {#service-ports}

The list of ports that are exposed by this service. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +patchMergeKey=port +patchStrategy=merge +listType=map +listMapKey=port +listMapKey=protocol 

Default: -

### selector (map[string]string, optional) {#service-selector}

Route service traffic to pods with label keys and values matching this selector. If empty or not present, the service is assumed to have an external process managing its endpoints, which Kubernetes will not modify. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/ +optional 

Default: -

### clusterIP (string, optional) {#service-clusterip}

clusterIP is the IP address of the service and is usually assigned randomly by the master. If an address is specified manually and is not in use by others, it will be allocated to the service; otherwise, creation of the service will fail. This field can not be changed through updates. Valid values are "None", empty string (""), or a valid IP address. "None" can be specified for headless services when proxying is not required. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +optional 

Default: -

### type (string, optional) {#service-type}

type determines how the Service is exposed. Defaults to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort, and LoadBalancer. "ExternalName" maps to the specified externalName. "ClusterIP" allocates a cluster-internal IP address for load-balancing to endpoints. Endpoints are determined by the selector or if that is not specified, by manual construction of an Endpoints object. If clusterIP is "None", no virtual IP is allocated and the endpoints are published as a set of endpoints rather than a stable IP. "NodePort" builds on ClusterIP and allocates a port on every node which routes to the clusterIP. "LoadBalancer" builds on NodePort and creates an external load-balancer (if supported in the current cloud) which routes to the clusterIP. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types +optional 

Default: -

### externalIPs ([]string, optional) {#service-externalips}

externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service.  These IPs are not managed by Kubernetes.  The user is responsible for ensuring that traffic arrives at a node with this IP.  A common example is external load-balancers that are not part of the Kubernetes system. +optional 

Default: -

### sessionAffinity (string, optional) {#service-sessionaffinity}

Supports "ClientIP" and "None". Used to maintain session affinity. Enable client IP based session affinity. Must be ClientIP or None. Defaults to None. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +optional 

Default: -

### loadBalancerIP (string, optional) {#service-loadbalancerip}

Only applies to Service Type: LoadBalancer LoadBalancer will get created with the IP specified in this field. This feature depends on whether the underlying cloud-provider supports specifying the loadBalancerIP when a load balancer is created. This field will be ignored if the cloud-provider does not support the feature. +optional 

Default: -

### loadBalancerSourceRanges ([]string, optional) {#service-loadbalancersourceranges}

If specified and supported by the platform, this will restrict traffic through the cloud-provider load-balancer will be restricted to the specified client IPs. This field will be ignored if the cloud-provider does not support the feature." More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/ +optional 

Default: -

### externalName (string, optional) {#service-externalname}

externalName is the external reference that kubedns or equivalent will return as a CNAME record for this service. No proxying will be involved. Must be a valid RFC-1123 hostname (https://tools.ietf.org/html/rfc1123) and requires Type to be ExternalName. +optional 

Default: -

### externalTrafficPolicy (string, optional) {#service-externaltrafficpolicy}

externalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints. "Local" preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. "Cluster" obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading. +optional 

Default: -

### healthCheckNodePort (int32, optional) {#service-healthchecknodeport}

healthCheckNodePort specifies the healthcheck nodePort for the service. If not specified, HealthCheckNodePort is created by the service api backend with the allocated nodePort. Will use user-specified nodePort value if specified by the client. Only effects when Type is set to LoadBalancer and ExternalTrafficPolicy is set to Local. +optional 

Default: -

### publishNotReadyAddresses (*bool, optional) {#service-publishnotreadyaddresses}

publishNotReadyAddresses, when set to true, indicates that DNS implementations must publish the notReadyAddresses of subsets for the Endpoints associated with the Service. The default value is false. The primary use case for setting this field is to use a StatefulSet's Headless Service to propagate SRV records for its Pods without respect to their readiness for purpose of peer discovery. +optional 

Default: -

### sessionAffinityConfig (*v1.SessionAffinityConfig, optional) {#service-sessionaffinityconfig}

sessionAffinityConfig contains the configurations of session affinity. +optional 

Default: -

### ipFamily (string, optional) {#service-ipfamily}

ipFamily specifies whether this Service has a preference for a particular IP family (e.g. IPv4 vs. IPv6).  If a specific IP family is requested, the clusterIP field will be allocated from that family, if it is available in the cluster.  If no IP family is requested, the cluster's primary IP family will be used. Other IP fields (loadBalancerIP, loadBalancerSourceRanges, externalIPs) and controllers which allocate external load-balancers should use the same IP family.  Endpoints for this Service will be of this family.  This field is immutable after creation. Assigning a ServiceIPFamily not available in the cluster (e.g. IPv6 in IPv4 only cluster) is an error condition and will fail during clusterIP assignment. +optional 

Default: -

### - (struct{}, required) {#service--}

Default: -

### - ([]byte, required) {#service--}

Default: -

### - (int32, required) {#service--}

Default: -


## UnprotectedService

Service describes the attributes that a user creates on a service.

### metadata (*K8SObjectMeta, optional) {#unprotectedservice-metadata}

Default: -

### ports ([]ServicePort, required) {#unprotectedservice-ports}

The list of ports that are exposed by this service. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +patchMergeKey=port +patchStrategy=merge +listType=map +listMapKey=port +listMapKey=protocol 

Default: -

### selector (map[string]string, optional) {#unprotectedservice-selector}

Route service traffic to pods with label keys and values matching this selector. If empty or not present, the service is assumed to have an external process managing its endpoints, which Kubernetes will not modify. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/ +optional 

Default: -

### clusterIP (string, optional) {#unprotectedservice-clusterip}

clusterIP is the IP address of the service and is usually assigned randomly by the master. If an address is specified manually and is not in use by others, it will be allocated to the service; otherwise, creation of the service will fail. This field can not be changed through updates. Valid values are "None", empty string (""), or a valid IP address. "None" can be specified for headless services when proxying is not required. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +optional 

Default: -

### type (string, optional) {#unprotectedservice-type}

type determines how the Service is exposed. Defaults to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort, and LoadBalancer. "ExternalName" maps to the specified externalName. "ClusterIP" allocates a cluster-internal IP address for load-balancing to endpoints. Endpoints are determined by the selector or if that is not specified, by manual construction of an Endpoints object. If clusterIP is "None", no virtual IP is allocated and the endpoints are published as a set of endpoints rather than a stable IP. "NodePort" builds on ClusterIP and allocates a port on every node which routes to the clusterIP. "LoadBalancer" builds on NodePort and creates an external load-balancer (if supported in the current cloud) which routes to the clusterIP. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types +optional 

Default: -

### externalIPs ([]string, optional) {#unprotectedservice-externalips}

externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service.  These IPs are not managed by Kubernetes.  The user is responsible for ensuring that traffic arrives at a node with this IP.  A common example is external load-balancers that are not part of the Kubernetes system. +optional 

Default: -

### sessionAffinity (string, optional) {#unprotectedservice-sessionaffinity}

Supports "ClientIP" and "None". Used to maintain session affinity. Enable client IP based session affinity. Must be ClientIP or None. Defaults to None. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies +optional 

Default: -

### loadBalancerIP (string, optional) {#unprotectedservice-loadbalancerip}

Only applies to Service Type: LoadBalancer LoadBalancer will get created with the IP specified in this field. This feature depends on whether the underlying cloud-provider supports specifying the loadBalancerIP when a load balancer is created. This field will be ignored if the cloud-provider does not support the feature. +optional 

Default: -

### loadBalancerSourceRanges ([]string, optional) {#unprotectedservice-loadbalancersourceranges}

If specified and supported by the platform, this will restrict traffic through the cloud-provider load-balancer will be restricted to the specified client IPs. This field will be ignored if the cloud-provider does not support the feature." More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/ +optional 

Default: -

### externalName (string, optional) {#unprotectedservice-externalname}

externalName is the external reference that kubedns or equivalent will return as a CNAME record for this service. No proxying will be involved. Must be a valid RFC-1123 hostname (https://tools.ietf.org/html/rfc1123) and requires Type to be ExternalName. +optional 

Default: -

### externalTrafficPolicy (string, optional) {#unprotectedservice-externaltrafficpolicy}

externalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints. "Local" preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. "Cluster" obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading. +optional 

Default: -

### healthCheckNodePort (int32, optional) {#unprotectedservice-healthchecknodeport}

healthCheckNodePort specifies the healthcheck nodePort for the service. If not specified, HealthCheckNodePort is created by the service api backend with the allocated nodePort. Will use user-specified nodePort value if specified by the client. Only effects when Type is set to LoadBalancer and ExternalTrafficPolicy is set to Local. +optional 

Default: -

### publishNotReadyAddresses (*bool, optional) {#unprotectedservice-publishnotreadyaddresses}

publishNotReadyAddresses, when set to true, indicates that DNS implementations must publish the notReadyAddresses of subsets for the Endpoints associated with the Service. The default value is false. The primary use case for setting this field is to use a StatefulSet's Headless Service to propagate SRV records for its Pods without respect to their readiness for purpose of peer discovery. +optional 

Default: -

### sessionAffinityConfig (*v1.SessionAffinityConfig, optional) {#unprotectedservice-sessionaffinityconfig}

sessionAffinityConfig contains the configurations of session affinity. +optional 

Default: -

### ipFamily (string, optional) {#unprotectedservice-ipfamily}

ipFamily specifies whether this Service has a preference for a particular IP family (e.g. IPv4 vs. IPv6).  If a specific IP family is requested, the clusterIP field will be allocated from that family, if it is available in the cluster.  If no IP family is requested, the cluster's primary IP family will be used. Other IP fields (loadBalancerIP, loadBalancerSourceRanges, externalIPs) and controllers which allocate external load-balancers should use the same IP family.  Endpoints for this Service will be of this family.  This field is immutable after creation. Assigning a ServiceIPFamily not available in the cluster (e.g. IPv6 in IPv4 only cluster) is an error condition and will fail during clusterIP assignment. +optional 

Default: -

### - (struct{}, required) {#unprotectedservice--}

Default: -

### - ([]byte, required) {#unprotectedservice--}

Default: -

### - (int32, required) {#unprotectedservice--}

Default: -


## ServicePort

ServicePort contains information on service's port.

### name (string, optional) {#serviceport-name}

The name of this port within the service. This must be a DNS_LABEL. All ports within a ServiceSpec must have unique names. When considering the endpoints for a Service, this must match the 'name' field in the EndpointPort. if only one ServicePort is defined on this service. +optional 

Default: -

### protocol (string, optional) {#serviceport-protocol}

The IP protocol for this port. Supports "TCP", "UDP", and "SCTP". Default is TCP. +optional 

Default: -

### port (int32, optional) {#serviceport-port}

The port that will be exposed by this service. 

Default: -

### targetPort (*IntOrString, optional) {#serviceport-targetport}

Number or name of the port to access on the pods targeted by the service. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME. If this is a string, it will be looked up as a named port in the target Pod's container ports. If this is not specified, the value of the 'port' field is used (an identity map). This field is ignored for services with clusterIP=None, and should be omitted or set equal to the 'port' field. More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service +optional 

Default: -

### nodePort (int32, optional) {#serviceport-nodeport}

The port on each node on which this service is exposed when type=NodePort or LoadBalancer. Usually assigned by the system. If specified, it will be allocated to the service if unused or else creation of the service will fail. Default is to auto-allocate a port if the ServiceType of this Service requires one. More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport +optional 

Default: -

### - (struct{}, required) {#serviceport--}

Default: -

### - ([]byte, required) {#serviceport--}

Default: -

### - (int32, required) {#serviceport--}

Default: -


## NamespacedName

### name (string, optional) {#namespacedname-name}

Name of the referenced Kubernetes resource 

Default: -

### namespace (string, optional) {#namespacedname-namespace}

Namespace of the referenced Kubernetes resource 

Default: -

### - (struct{}, required) {#namespacedname--}

Default: -

### - ([]byte, required) {#namespacedname--}

Default: -

### - (int32, required) {#namespacedname--}

Default: -


## ResourceRequirements

ResourceRequirements describes the compute resource requirements.

### limits (map[string]*Quantity, optional) {#resourcerequirements-limits}

Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/ +optional 

Default: -

### requests (map[string]*Quantity, optional) {#resourcerequirements-requests}

Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/ +optional 

Default: -

### - (struct{}, required) {#resourcerequirements--}

Default: -

### - ([]byte, required) {#resourcerequirements--}

Default: -

### - (int32, required) {#resourcerequirements--}

Default: -


## Replicas

Replicas contains pod replica configuration

### count (*int32, optional) {#replicas-count}

Standard Kubernetes replica count configuration 

Default: -

### min (*int32, optional) {#replicas-min}

min is the lower limit for the number of replicas to which the autoscaler can scale down. min and max both need to be set the turn on autoscaling. 

Default: -

### max (*int32, optional) {#replicas-max}

max is the upper limit for the number of replicas to which the autoscaler can scale up. min and max both need to be set the turn on autoscaling. It cannot be less than min. 

Default: -

### targetCPUUtilizationPercentage (*int32, optional) {#replicas-targetcpuutilizationpercentage}

target average CPU utilization (represented as a percentage of requested CPU) over all the pods; default 80% will be used if not specified. +optional 

Default: -

### - (struct{}, required) {#replicas--}

Default: -

### - ([]byte, required) {#replicas--}

Default: -

### - (int32, required) {#replicas--}

Default: -


## K8SResourceOverlayPatch

### groupVersionKind (K8SResourceOverlayPatch_GroupVersionKind, required) {#k8sresourceoverlaypatch-groupversionkind}

Default: -

### objectKey (*NamespacedName, optional) {#k8sresourceoverlaypatch-objectkey}

Default: -

### patches ([]K8SResourceOverlayPatch_Patch, required) {#k8sresourceoverlaypatch-patches}

Default: -

### - (struct{}, required) {#k8sresourceoverlaypatch--}

Default: -

### - ([]byte, required) {#k8sresourceoverlaypatch--}

Default: -

### - (int32, required) {#k8sresourceoverlaypatch--}

Default: -


## K8SResourceOverlayPatch_GroupVersionKind

### kind (string, optional) {#k8sresourceoverlaypatch_groupversionkind-kind}

Default: -

### version (string, optional) {#k8sresourceoverlaypatch_groupversionkind-version}

Default: -

### group (string, optional) {#k8sresourceoverlaypatch_groupversionkind-group}

Default: -

### - (struct{}, required) {#k8sresourceoverlaypatch_groupversionkind--}

Default: -

### - ([]byte, required) {#k8sresourceoverlaypatch_groupversionkind--}

Default: -

### - (int32, required) {#k8sresourceoverlaypatch_groupversionkind--}

Default: -


## K8SResourceOverlayPatch_Patch

### path (string, optional) {#k8sresourceoverlaypatch_patch-path}

Default: -

### value (string, optional) {#k8sresourceoverlaypatch_patch-value}

Default: -

### parseValue (bool, optional) {#k8sresourceoverlaypatch_patch-parsevalue}

Default: -

### type (K8SResourceOverlayPatch_Type, optional) {#k8sresourceoverlaypatch_patch-type}

Default: -

### - (struct{}, required) {#k8sresourceoverlaypatch_patch--}

Default: -

### - ([]byte, required) {#k8sresourceoverlaypatch_patch--}

Default: -

### - (int32, required) {#k8sresourceoverlaypatch_patch--}

Default: -


