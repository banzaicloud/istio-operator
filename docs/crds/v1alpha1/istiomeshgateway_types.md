## IstioMeshGateway

IstioMeshGateway is the Schema for the istiomeshgateways API

###  (metav1.TypeMeta, required) {#istiomeshgateway-}

Default: -

### metadata (metav1.ObjectMeta, optional) {#istiomeshgateway-metadata}

Default: -

### spec (*IstioMeshGatewaySpec, optional) {#istiomeshgateway-spec}

Default: -

### status (IstioMeshGatewayStatus, optional) {#istiomeshgateway-status}

Default: -


## IstioMeshGatewayWithProperties

### istiomeshgateway (*IstioMeshGateway, optional) {#istiomeshgatewaywithproperties-istiomeshgateway}

Default: -

### properties (IstioMeshGatewayProperties, optional) {#istiomeshgatewaywithproperties-properties}

Default: -


## IstioMeshGatewayProperties

Properties of the IstioMeshGateway

### revision (string, optional) {#istiomeshgatewayproperties-revision}

Default: -

### enablePrometheusMerge (*bool, optional) {#istiomeshgatewayproperties-enableprometheusmerge}

Default: -

### injectionTemplate (string, optional) {#istiomeshgatewayproperties-injectiontemplate}

Default: -

### injectionChecksum (string, optional) {#istiomeshgatewayproperties-injectionchecksum}

Default: -

### meshConfigChecksum (string, optional) {#istiomeshgatewayproperties-meshconfigchecksum}

Default: -

### istioControlPlane (*IstioControlPlane, optional) {#istiomeshgatewayproperties-istiocontrolplane}

Default: -

### generateExternalService (bool, optional) {#istiomeshgatewayproperties-generateexternalservice}

Default: -


## IstioMeshGatewayList

IstioMeshGatewayList contains a list of IstioMeshGateway

###  (metav1.TypeMeta, required) {#istiomeshgatewaylist-}

Default: -

### metadata (metav1.ListMeta, optional) {#istiomeshgatewaylist-metadata}

Default: -

### items ([]IstioMeshGateway, required) {#istiomeshgatewaylist-items}

Default: -


