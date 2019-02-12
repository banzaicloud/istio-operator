package citadel

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var serviceLabels = map[string]string{
	"app": "istio-citadel",
}

func service(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, serviceLabels, owner),
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "grpc-citadel",
					Port:       8060,
					TargetPort: intstr.FromInt(8060),
					Protocol:   apiv1.ProtocolTCP,
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
