package gateways

import (
	"k8s.io/apimachinery/pkg/runtime"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) service(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(gatewayName(gw), labelSelector(gw), owner),
		Spec: apiv1.ServiceSpec{
			Type:     serviceType(gw),
			Ports:    servicePorts(gw),
			Selector: labelSelector(gw),
		},
	}
}

func servicePorts(gw string) []apiv1.ServicePort {
	switch gw {
	case "ingressgateway":
		return []apiv1.ServicePort{
			{Port: 80, TargetPort: intstr.FromInt(80), Name: "http2", NodePort: 31380},
			{Port: 443, Name: "https", NodePort: 31390},
			{Port: 31400, Name: "tcp", NodePort: 31400},
			{Port: 15011, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls"},
			{Port: 8060, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls"},
			{Port: 853, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls"},
			{Port: 15030, TargetPort: intstr.FromInt(15030), Name: "http2-prometheus"},
			{Port: 15031, TargetPort: intstr.FromInt(15031), Name: "http2-grafana"},
		}
	case "egressgateway":
		return []apiv1.ServicePort{
			{Port: 80, Name: "http2"},
			{Port: 443, Name: "https"},
		}
	}
	return []apiv1.ServicePort{}
}

func serviceType(gw string) apiv1.ServiceType {
	switch gw {
	case "ingressgateway":
		return apiv1.ServiceTypeLoadBalancer
	case "egressgateway":
		return apiv1.ServiceTypeClusterIP
	}
	return ""
}
