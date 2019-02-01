package apis

import (
	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"istio.io/api/pkg/kube/apis/networking/v1alpha3"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v1alpha3.AddToScheme)
}
