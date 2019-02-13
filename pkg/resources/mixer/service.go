package mixer

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) service(t string, owner *istiov1beta1.Config) runtime.Object {
	svc := &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName(t), labelSelector, owner),
		Spec: apiv1.ServiceSpec{
			Ports:    r.servicePorts(t),
			Selector: util.MergeLabels(labelSelector, mixerTypeLabel(t)),
		},
	}
	return svc
}

func (r *Reconciler) servicePorts(t string) []apiv1.ServicePort {
	switch t {
	case "policy":
		return r.commonPorts()
	case "telemetry":
		ports := r.commonPorts()
		ports = append(ports, apiv1.ServicePort{
			Name: "prometheus",
			Port: 42422,
		})
		return ports
	}
	return nil
}

func (r *Reconciler) commonPorts() []apiv1.ServicePort {
	return []apiv1.ServicePort{
		{
			Name: "grpc-mixer",
			Port: 9091,
		},
		{
			Name: "grpc-mixer-mtls",
			Port: 15004,
		},
		{
			Name: "http-monitoring",
			Port: 9093,
		},
	}
}
