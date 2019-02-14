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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

var gatewaySelector = map[string]string{
	"istio": "ingress",
}

func (r *Reconciler) gateway(owner *istiov1beta1.Config) runtime.Object {
	return &networkingv1alpha3.Gateway{
		ObjectMeta: templates.ObjectMeta(gatewayName, nil, owner),
		Spec: v1alpha3.Gateway{
			Servers: []*v1alpha3.Server{
				{
					Port: &v1alpha3.Port{
						Name:     "http",
						Protocol: "HTTP2",
						Number:   80,
					},
					Hosts: []string{"*"},
				},
			},
			Selector: gatewaySelector,
		},
	}
}
