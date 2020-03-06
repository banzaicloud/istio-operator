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

package pilot

import (
	"github.com/banzaicloud/istio-operator/pkg/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) servicePorts() []apiv1.ServicePort {
	ports := []apiv1.ServicePort{
		{
			Name:       "grpc-xds",
			Port:       15010,
			TargetPort: intstr.FromInt(15010),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "https-xds",
			Port:       15011,
			TargetPort: intstr.FromInt(15011),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "https-dns",
			Port:       15012,
			TargetPort: intstr.FromInt(15012),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "http-legacy-discovery",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "http-monitoring",
			Port:       15014,
			TargetPort: intstr.FromInt(15014),
			Protocol:   apiv1.ProtocolTCP,
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		ports = append(ports, apiv1.ServicePort{
			Name:       "https-webhook",
			Port:       443,
			TargetPort: intstr.FromInt(15017),
			Protocol:   apiv1.ProtocolTCP,
		})
	}

	return ports
}

func (r *Reconciler) service() runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, util.MergeStringMaps(pilotLabels, pilotLabelSelector), r.Config),
		Spec: apiv1.ServiceSpec{
			Ports:    r.servicePorts(),
			Selector: pilotLabelSelector,
		},
	}
}

func (r *Reconciler) serviceIstiod() runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceNameIstiod, istiodLabels, r.Config),
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "https-dns",
					Port:       15012,
					TargetPort: intstr.FromInt(15012),
					Protocol:   apiv1.ProtocolTCP,
				},
				{
					Name:       "https-webhook",
					Port:       443,
					TargetPort: intstr.FromInt(15017),
					Protocol:   apiv1.ProtocolTCP,
				},
			},
			Selector: util.MergeStringMaps(istiodLabels, pilotLabelSelector),
		},
	}
}
