package istio

import (
	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/goph/emperror"
	"github.com/go-logr/logr"
)

func (r *ReconcileIstio) ReconcileCrds(log logr.Logger, istio *istiov1alpha1.Istio) error {
	crds := []*extensionsobj.CustomResourceDefinition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualservices.networking.istio.io",
				Labels: map[string]string{
					"app": "istio-pilot",
				},
			},
			Spec: extensionsobj.CustomResourceDefinitionSpec{
				Group:   "networking.istio.io",
				Version: "v1alpha3",
				Scope:   extensionsobj.NamespaceScoped,
				Names: extensionsobj.CustomResourceDefinitionNames{
					Plural:   "virtualservices",
					Kind:     "VirtualService",
					ListKind: "VirtualServiceList",
					Singular: "virtualservice",
					Categories: []string{
						"istio-io", "networking-istio-io",
					},
				},
			},
		},
	}

	for _, crd := range crds {
		controllerutil.SetControllerReference(istio, crd, r.scheme)
	}

	crdClient := r.crdClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range crds {
		oldCRD, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if _, err := crdClient.Create(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created", "crd", crd.Spec.Names.Kind)
		}
		if err == nil {
			crd.ResourceVersion = oldCRD.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated", "crd", crd.Spec.Names.Kind)
		}
	}
	return nil
}
