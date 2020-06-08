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
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
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
	ports := istiov1beta1.ServicePorts(r.gw.Spec.Ports).Convert()

	if name == defaultIngressgatewayName {
		if util.PointerToBool(r.Config.Spec.MeshExpansion) {
			ports = append(ports, apiv1.ServicePort{
				Port: 853, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(8853), Name: "tcp-dns-tls",
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

	ports = ensureHealthProbePort(ports)

	return ports
}

func ensureHealthProbePort(ports []apiv1.ServicePort) []apiv1.ServicePort {
	healthProbeTargetPort := intstr.FromInt(istiov1beta1.PortStatusPortNumber)

	protocols := sets.NewString()
	portName := istiov1beta1.PortStatusPortName
	portNameTaken := false
	portNumber := int32(istiov1beta1.PortStatusPortNumber)
	portNumberTaken := false
	for _, port := range ports {
		if port.Protocol == apiv1.ProtocolTCP && port.Port == portNumber && port.TargetPort == healthProbeTargetPort {
			// health probe port is included with proper protocol and target port. Nothing to do
			return ports
		}

		protocols.Insert(string(port.Protocol))
		if strings.ToLower(port.Name) == strings.ToLower(portName) {
			portNameTaken = true
		}
		if port.Protocol == apiv1.ProtocolTCP && port.Port == portNumber {
			portNumberTaken = true
		}
	}

	if protocols.Len() > 1 || (protocols.Len() == 1 && !protocols.Has(string(apiv1.ProtocolTCP))) {
		// mixed protocol types are not supported for LoadBalancer type services until
		// https://github.com/kubernetes/enhancements/pull/1438 is implemented. Health probe port cannot be included.
		// TODO check if type is LoadBalancer
		return ports
	}

	if portNameTaken || portNumberTaken {
		// random port name and number could be generated, but the reconciliation does not play nice in that case
		// (it's reconciling indefinitely), so let's just leave out the status port for now
		return ports
	}

	// status port should come first because of load balancer health checks
	return append([]apiv1.ServicePort{{
		Name:       portName,
		Protocol:   apiv1.ProtocolTCP,
		Port:       portNumber,
		TargetPort: healthProbeTargetPort,
	}}, ports...)
}
