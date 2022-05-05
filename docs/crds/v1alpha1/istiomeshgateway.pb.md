## IstioMeshGatewaySpec

IstioMeshGateway defines an Istio ingress or egress gateway

<!-- crd generation tags
+cue-gen:IstioMeshGateway:groupName:servicemesh.cisco.com
+cue-gen:IstioMeshGateway:version:v1alpha1
+cue-gen:IstioMeshGateway:storageVersion
+cue-gen:IstioMeshGateway:annotations:helm.sh/resource-policy=keep
+cue-gen:IstioMeshGateway:subresource:status
+cue-gen:IstioMeshGateway:scope:Namespaced
+cue-gen:IstioMeshGateway:resource:shortNames=imgw,istiomgw
+cue-gen:IstioMeshGateway:printerColumn:name="Type",type="string",JSONPath=".spec.type",description="Type of the gateway"
+cue-gen:IstioMeshGateway:printerColumn:name="Service Type",type="string",JSONPath=".spec.service.type",description="Type of the service"
+cue-gen:IstioMeshGateway:printerColumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
+cue-gen:IstioMeshGateway:printerColumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
+cue-gen:IstioMeshGateway:printerColumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
+cue-gen:IstioMeshGateway:printerColumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
+cue-gen:IstioMeshGateway:printerColumn:name="Control Plane",type="string",JSONPath=".spec.istioControlPlane"
+cue-gen:IstioMeshGateway:preserveUnknownFields:false
+cue-gen:IstioMeshGateway:specIsRequired
-->

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### deployment (*BaseKubernetesResourceConfig, optional) {#istiomeshgatewayspec-deployment}

Deployment spec 

Default: -

### service (*Service, optional) {#istiomeshgatewayspec-service}

Service spec 

Default: -

### runAsRoot (*bool, optional) {#istiomeshgatewayspec-runasroot}

Whether to run the gateway in a privileged container 

Default: -

### type (GatewayType, optional) {#istiomeshgatewayspec-type}

Type of gateway, either ingress or egress 

Default: -

### istioControlPlane (*NamespacedName, optional) {#istiomeshgatewayspec-istiocontrolplane}

Istio CR to which this gateway belongs to 

Default: -

### k8sResourceOverlays ([]*K8SResourceOverlayPatch, optional) {#istiomeshgatewayspec-k8sresourceoverlays}

K8s resource overlay patches 

Default: -

### - (struct{}, required) {#istiomeshgatewayspec--}

Default: -

### - ([]byte, required) {#istiomeshgatewayspec--}

Default: -

### - (int32, required) {#istiomeshgatewayspec--}

Default: -


## Properties

### name (string, optional) {#properties-name}

Default: -

### - (struct{}, required) {#properties--}

Default: -

### - ([]byte, required) {#properties--}

Default: -

### - (int32, required) {#properties--}

Default: -


## IstioMeshGatewayStatus

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### Status (ConfigState, optional) {#istiomeshgatewaystatus-status}

Reconciliation status of the istio mesh gateway 

Default: -

### GatewayAddress ([]string, optional) {#istiomeshgatewaystatus-gatewayaddress}

Current address for the gateway 

Default: -

### ErrorMessage (string, optional) {#istiomeshgatewaystatus-errormessage}

Reconciliation error message if any 

Default: -

### - (struct{}, required) {#istiomeshgatewaystatus--}

Default: -

### - ([]byte, required) {#istiomeshgatewaystatus--}

Default: -

### - (int32, required) {#istiomeshgatewaystatus--}

Default: -


