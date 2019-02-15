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

package mixer

import (
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) destinationRule(t string) runtime.Object {
	return &networkingv1alpha3.DestinationRule{
		ObjectMeta: templates.ObjectMeta(destinationRuleName(t), nil, r.Config),
		Spec: v1alpha3.DestinationRule{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", t, r.Config.Namespace),
			TrafficPolicy: &v1alpha3.TrafficPolicy{
				ConnectionPool: &v1alpha3.ConnectionPoolSettings{
					Http: &v1alpha3.ConnectionPoolSettings_HTTPSettings{
						Http2MaxRequests:         10000,
						MaxRequestsPerConnection: 10000,
					},
				},
			},
		},
	}
}
