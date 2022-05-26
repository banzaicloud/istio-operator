## IstioControlPlane

IstioControlPlane is the Schema for the istiocontrolplanes API

###  (metav1.TypeMeta, required) {#istiocontrolplane-}

Default: -

### metadata (metav1.ObjectMeta, optional) {#istiocontrolplane-metadata}

Default: -

### spec (*IstioControlPlaneSpec, optional) {#istiocontrolplane-spec}

Default: -

### status (IstioControlPlaneStatus, optional) {#istiocontrolplane-status}

Default: -


## IstioControlPlaneWithProperties

### istioControlPlane (*IstioControlPlane, optional) {#istiocontrolplanewithproperties-istiocontrolplane}

Default: -

### properties (IstioControlPlaneProperties, optional) {#istiocontrolplanewithproperties-properties}

Default: -


## IstioControlPlaneProperties

Properties of the IstioControlPlane

### mesh (*IstioMesh, optional) {#istiocontrolplaneproperties-mesh}

Default: -

### meshNetworks (*v1alpha1.MeshNetworks, optional) {#istiocontrolplaneproperties-meshnetworks}

Default: -

### trustedRootCACertificatePEMs ([]string, optional) {#istiocontrolplaneproperties-trustedrootcacertificatepems}

Default: -


## IstioControlPlaneList

IstioControlPlaneList contains a list of IstioControlPlane

###  (metav1.TypeMeta, required) {#istiocontrolplanelist-}

Default: -

### metadata (metav1.ListMeta, optional) {#istiocontrolplanelist-metadata}

Default: -

### items ([]IstioControlPlane, required) {#istiocontrolplanelist-items}

Default: -


## PeerIstioControlPlane

PeerIstioControlPlane is the Schema for the clone of the istiocontrolplanes API

###  (metav1.TypeMeta, required) {#peeristiocontrolplane-}

Default: -

### metadata (metav1.ObjectMeta, optional) {#peeristiocontrolplane-metadata}

Default: -

### spec (*IstioControlPlaneSpec, optional) {#peeristiocontrolplane-spec}

Default: -

### status (IstioControlPlaneStatus, optional) {#peeristiocontrolplane-status}

Default: -


## PeerIstioControlPlaneList

PeerIstioControlPlaneList contains a list of PeerIstioControlPlane

###  (metav1.TypeMeta, required) {#peeristiocontrolplanelist-}

Default: -

### metadata (metav1.ListMeta, optional) {#peeristiocontrolplanelist-metadata}

Default: -

### items ([]PeerIstioControlPlane, required) {#peeristiocontrolplanelist-items}

Default: -


