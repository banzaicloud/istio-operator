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

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/ingressgateway"
	"github.com/banzaicloud/istio-operator/pkg/resources/meshexpansion"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) service() runtime.Object {
	service := &apiv1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(r.gatewayName(), util.MergeStringMaps(r.gw.Spec.ServiceLabels, r.labels()), r.gw.Spec.ServiceAnnotations, r.gw),
		Spec: apiv1.ServiceSpec{
			LoadBalancerIP: r.gw.Spec.LoadBalancerIP,
			Type:           r.gw.Spec.ServiceType,
			Ports:          r.servicePorts(r.gw.Name),
			Selector:       r.labelSelector(),
		},
	}

	externalTrafficPolicy := r.gw.Spec.ServiceExternalTrafficPolicy
	if externalTrafficPolicy != "" {
		service.Spec.ExternalTrafficPolicy = externalTrafficPolicy
	}

	return service
}

func (r *Reconciler) servicePorts(name string) []apiv1.ServicePort {
	ports := istiov1beta1.ServicePorts(r.gw.Spec.Ports).Convert()

	ports = r.ensureMeshExpansionPorts(name, ports)
	ports = ensureStatusPort(ports)

	return ports
}

func (r *Reconciler) ensureMeshExpansionPorts(name string, ports []apiv1.ServicePort) []apiv1.ServicePort {
	newPorts := make([]apiv1.ServicePort, 0)
	newPorts = append(newPorts, ports...)

	if !util.PointerToBool(r.Config.Spec.MeshExpansion) {
		return newPorts
	}

	if (name == r.Config.WithRevision(meshexpansion.ResourceName) &&
		util.PointerToBool(r.Config.Spec.Gateways.MeshExpansion.Enabled)) ||
		(name == r.Config.WithRevision(ingressgateway.ResourceName) &&
			!util.PointerToBool(r.Config.Spec.Gateways.MeshExpansion.Enabled)) {

		newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
			Port: 15443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15443), Name: "tcp-mtls",
		}, false)

		if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
			newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
				Port: 15012, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15012), Name: "tcp-istiod-grpc-tls",
			}, false)
			if util.PointerToBool(r.Config.Spec.Istiod.ExposeWebhookPort) {
				newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
					Port: 15017, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15017), Name: "tcp-istiodwebhook",
				}, false)
			}
		}
		if util.PointerToBool(r.Config.Spec.Pilot.Enabled) {
			newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
				Port: 15011, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls",
			}, false)
		}
		if util.PointerToBool(r.Config.Spec.Telemetry.Enabled) {
			newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
				Port: 15004, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15004), Name: "tcp-mixer-grpc-tls",
			}, false)
		}
		if util.PointerToBool(r.Config.Spec.Citadel.Enabled) {
			newPorts = ensureServicePort(newPorts, apiv1.ServicePort{
				Port: 8060, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(8060), Name: "tcp-citadel-grpc-tls",
			}, false)
		}
	}

	return newPorts
}

func ensureStatusPort(ports []apiv1.ServicePort) []apiv1.ServicePort {
	return ensureServicePort(ports, apiv1.ServicePort{
		Name:       istiov1beta1.PortStatusPortName,
		Port:       int32(istiov1beta1.PortStatusPortNumber),
		Protocol:   apiv1.ProtocolTCP,
		TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
	}, true)
}

func ensureServicePort(ports []apiv1.ServicePort, servicePort apiv1.ServicePort, prepend bool) []apiv1.ServicePort {
	newPorts := make([]apiv1.ServicePort, 0)
	newPorts = append(newPorts, ports...)

	portNameTaken := false
	portNumberTaken := false
	for _, port := range ports {
		if port.Protocol == servicePort.Protocol && port.Port == servicePort.Port && port.TargetPort == servicePort.TargetPort {
			// Service probe port is included with proper protocol and target port. Nothing to do
			return newPorts
		}

		if strings.ToLower(port.Name) == strings.ToLower(servicePort.Name) {
			portNameTaken = true
			break
		}

		if port.Port == servicePort.Port {
			portNumberTaken = true
			break
		}
	}

	// mixed protocol types are not supported for LoadBalancer type services until
	// https://github.com/kubernetes/enhancements/pull/1438 is implemented. Health probe port cannot be included.
	// TODO check if type is LoadBalancer
	if portNumberTaken {
		return ports
	}

	// random port name and number could be generated, but the reconciliation does not play nice in that case
	// (it's reconciling indefinitely), so let's just leave out the status port for now
	if portNameTaken {
		return ports
	}

	if prepend {
		return append([]apiv1.ServicePort{servicePort}, ports...)
	}

	return append(ports, servicePort)
}
