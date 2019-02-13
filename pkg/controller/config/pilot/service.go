package pilot

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) service(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, pilotLabels, owner),
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "grpc-xds",
					Port: 15010,
				},
				{
					Name: "https-xds",
					Port: 15011},
				{
					Name: "http-legacy-discovery",
					Port: 8080,
				},
				{
					Name: "http-monitoring",
					Port: 9093,
				},
			},
			Selector: labelSelector,
		},
	}
}
