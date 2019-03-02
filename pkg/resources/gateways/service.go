/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gateways

import (
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) service(gw string) runtime.Object {
	gwConfig := r.getGatewayConfig(gw)
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(gatewayName(gw), util.MergeLabels(labelSelector(gw), gwConfig.ServiceLabels), gwConfig.ServiceAnnotations, r.Config),
		Spec: apiv1.ServiceSpec{
			Type:     serviceType(gw),
			Ports:    servicePorts(gw),
			Selector: labelSelector(gw),
		},
	}
}

func servicePorts(gw string) []apiv1.ServicePort {
	switch gw {
	case ingress:
		return []apiv1.ServicePort{
			{Port: 80, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80), Name: "http2", NodePort: 31380},
			{Port: 443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443), Name: "https", NodePort: 31390},
			{Port: 31400, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(31400), Name: "tcp", NodePort: 31400},
			{Port: 15011, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls", NodePort: 31410},
			{Port: 8060, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls", NodePort: 31420},
			{Port: 853, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls", NodePort: 31430},
			{Port: 15030, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15030), Name: "http2-prometheus", NodePort: 31440},
			{Port: 15031, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15031), Name: "http2-grafana", NodePort: 31450},
		}
	case egress:
		return []apiv1.ServicePort{
			{Port: 80, Name: "http2", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80)},
			{Port: 443, Name: "https", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443)},
		}
	}
	return []apiv1.ServicePort{}
}

func serviceType(gw string) apiv1.ServiceType {
	switch gw {
	case ingress:
		return apiv1.ServiceTypeLoadBalancer
	case egress:
		return apiv1.ServiceTypeClusterIP
	}
	return ""
}
