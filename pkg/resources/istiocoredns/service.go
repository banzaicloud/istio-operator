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

package istiocoredns

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) service() runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, util.MergeLabels(labels, labelSelector), r.Config),
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "dns",
					Port:       53,
					Protocol:   apiv1.ProtocolUDP,
					TargetPort: intstr.FromInt(53),
				},
				{
					Name:       "dns-tcp",
					Port:       53,
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromInt(53),
				},
			},
			Selector: labelSelector,
		},
	}
}
