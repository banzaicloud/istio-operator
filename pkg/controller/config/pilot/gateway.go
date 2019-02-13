package pilot

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
)

var gatewaySelector = map[string]string{
	"istio": "ingress",
}

func (r *Reconciler) gateway(owner *istiov1beta1.Config) runtime.Object {
	return &networkingv1alpha3.Gateway{
		ObjectMeta: templates.ObjectMeta(gatewayName, nil, owner),
		Spec: v1alpha3.Gateway{
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Name:     "http",
						Protocol: "HTTP2",
						Number:   80,
					},
					Hosts: []string{"*"},
				},
			},
			Selector: gatewaySelector,
		},
	}
}
