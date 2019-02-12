package galley

import (
	"k8s.io/apimachinery/pkg/runtime"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
)

var serviceLabels = map[string]string{
	"istio": "galley",
}

func service(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, serviceLabels, owner),
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "https-validation",
					Port: 443,
				},
				{
					Name: "https-monitoring",
					Port: 9093,
				},
			},
			Selector: labelSelector,
		},
	}
}
