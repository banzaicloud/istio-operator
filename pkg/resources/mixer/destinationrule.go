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

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

func (r *Reconciler) policyDestinationRule() *k8sutil.DynamicObject {
	dr := &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "istio-policy",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"traffic_policy": r.trafficPolicy(),
			"host":           fmt.Sprintf("istio-policy.%s.svc.%s", r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain),
		},
		Owner: r.Config,
	}

	exportTo := r.Config.Spec.GetDefaultConfigVisibility()
	if exportTo != "" {
		dr.Spec["exportTo"] = exportTo
	}

	return dr
}

func (r *Reconciler) telemetryDestinationRule() *k8sutil.DynamicObject {
	dr := &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "istio-telemetry",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"traffic_policy": r.trafficPolicy(),
			"host":           fmt.Sprintf("istio-telemetry.%s.svc.%s", r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain),
		},
		Owner: r.Config,
	}

	exportTo := r.Config.Spec.GetDefaultConfigVisibility()
	if exportTo != "" {
		dr.Spec["exportTo"] = exportTo
	}

	return dr
}

func (r *Reconciler) connectionPool() map[string]interface{} {
	return map[string]interface{}{
		"http": map[string]interface{}{
			"http2_max_requests":          10000,
			"max_requests_per_connection": 10000,
		},
	}
}

func (r *Reconciler) trafficPolicy() map[string]interface{} {
	tp := map[string]interface{}{
		"connection_pool": r.connectionPool(),
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		tp["portLevelSettings"] = []map[string]interface{}{
			{
				"port": map[string]interface{}{
					"number": 15004,
				},
				"tls": map[string]interface{}{
					"mode": "ISTIO_MUTUAL",
				},
			},
		}
	}
	return tp
}
