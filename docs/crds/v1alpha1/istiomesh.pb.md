## IstioMeshSpec

Mesh defines an Istio service mesh

<!-- crd generation tags
+cue-gen:IstioMesh:groupName:servicemesh.cisco.com
+cue-gen:IstioMesh:version:v1alpha1
+cue-gen:IstioMesh:storageVersion
+cue-gen:IstioMesh:annotations:helm.sh/resource-policy=keep
+cue-gen:IstioMesh:subresource:status
+cue-gen:IstioMesh:scope:Namespaced
+cue-gen:IstioMesh:resource:shortNames="im,imesh",plural="istiomeshes"
+cue-gen:IstioMesh:printerColumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
+cue-gen:IstioMesh:preserveUnknownFields:false
-->

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### config (*v1alpha1.MeshConfig, optional) {#istiomeshspec-config}

Default: -

### - (struct{}, required) {#istiomeshspec--}

Default: -

### - ([]byte, required) {#istiomeshspec--}

Default: -

### - (int32, required) {#istiomeshspec--}

Default: -


## IstioMeshStatus

<!-- go code generation tags
+genclient
+k8s:deepcopy-gen=true
-->

### status (ConfigState, optional) {#istiomeshstatus-status}

Reconciliation status of the Istio mesh 

Default: -

### errorMessage (string, optional) {#istiomeshstatus-errormessage}

Reconciliation error message if any 

Default: -

### - (struct{}, required) {#istiomeshstatus--}

Default: -

### - ([]byte, required) {#istiomeshstatus--}

Default: -

### - (int32, required) {#istiomeshstatus--}

Default: -


