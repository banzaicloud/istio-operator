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

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *Reconciler) policyDestinationRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "istio-policy",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"traffic_policy": map[string]interface{}{
				"connection_pool": map[string]interface{}{
					"http": map[string]interface{}{
						"http2_max_requests":          10000,
						"max_requests_per_connection": 10000,
					},
				},
			},
			"host": fmt.Sprintf("istio-policy.%s.svc.cluster.local", r.Config.Namespace),
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) telemetryDestinationRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "istio-telemetry",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"traffic_policy": map[string]interface{}{
				"connection_pool": map[string]interface{}{
					"http": map[string]interface{}{
						"http2_max_requests":          10000,
						"max_requests_per_connection": 10000,
					},
				},
			},
			"host": fmt.Sprintf("istio-telemetry.%s.svc.cluster.local", r.Config.Namespace),
		},
		Owner: r.Config,
	}
}
