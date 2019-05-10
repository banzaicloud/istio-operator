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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) service(gw string) runtime.Object {
	gwConfig := r.getGatewayConfig(gw)
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(gatewayName(gw), util.MergeLabels(gwConfig.ServiceLabels, labelSelector(gw)), gwConfig.ServiceAnnotations, r.Config),
		Spec: apiv1.ServiceSpec{
			Type:     r.serviceType(gw),
			Ports:    r.servicePorts(gw),
			Selector: labelSelector(gw),
		},
	}
}

func (r *Reconciler) servicePorts(gw string) []apiv1.ServicePort {
	switch gw {
	case ingress:
		ports := r.Config.Spec.Gateways.IngressConfig.Ports
		if util.PointerToBool(r.Config.Spec.MeshExpansion) {
			ports = append(ports, []apiv1.ServicePort{
				{Port: 15011, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls", NodePort: 31470},
				{Port: 15004, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15004), Name: "tcp-mixer-grpc-tls", NodePort: 31480},
				{Port: 8060, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls", NodePort: 31490},
				{Port: 853, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls", NodePort: 31500},
			}...)
		}
		return ports
	case egress:
		return r.Config.Spec.Gateways.EgressConfig.Ports
	}
	return []apiv1.ServicePort{}
}

func (r *Reconciler) serviceType(gw string) apiv1.ServiceType {
	switch gw {
	case ingress:
		return r.Config.Spec.Gateways.IngressConfig.ServiceType
	case egress:
		return r.Config.Spec.Gateways.EgressConfig.ServiceType
	}
	return ""
}
