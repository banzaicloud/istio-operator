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

func (r *Reconciler) service() runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(r.gatewayName(), util.MergeStringMaps(r.gw.Spec.ServiceLabels, r.labels()), r.gw.Spec.ServiceAnnotations, r.gw),
		Spec: apiv1.ServiceSpec{
			LoadBalancerIP: r.gw.Spec.LoadBalancerIP,
			Type:           r.gw.Spec.ServiceType,
			Ports:          r.servicePorts(r.gw.Name),
			Selector:       r.labelSelector(),
		},
	}
}

func (r *Reconciler) servicePorts(name string) []apiv1.ServicePort {
	if name == defaultIngressgatewayName {
		ports := r.gw.Spec.Ports
		if util.PointerToBool(r.Config.Spec.MeshExpansion) {
			ports = append(ports, apiv1.ServicePort{
				Port: 853, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls",
			})

			if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
				ports = append(ports, apiv1.ServicePort{
					Port: 15012, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15012), Name: "tcp-istiod-grpc-tls",
				})
			}
			if util.PointerToBool(r.Config.Spec.Pilot.Enabled) {
				ports = append(ports, apiv1.ServicePort{
					Port: 15011, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls",
				})
			}
			if util.PointerToBool(r.Config.Spec.Telemetry.Enabled) {
				ports = append(ports, apiv1.ServicePort{
					Port: 15004, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15004), Name: "tcp-mixer-grpc-tls",
				})
			}
			if util.PointerToBool(r.Config.Spec.Citadel.Enabled) {
				ports = append(ports, apiv1.ServicePort{
					Port: 8060, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls",
				})
			}
		}
		return ports
	}
	return r.gw.Spec.Ports
}
