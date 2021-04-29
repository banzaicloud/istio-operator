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
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/gvr"
)

func (r *Reconciler) policyDestinationRule() *k8sutil.DynamicObject {
	dr := &k8sutil.DynamicObject{
		Gvr:       gvr.DestinationRule,
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision("istio-policy"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"traffic_policy": r.trafficPolicy(),
			"host":           serviceHostWithRevision(r.Config, policyComponentName),
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
		Gvr:       gvr.DestinationRule,
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision("istio-telemetry"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"traffic_policy": r.trafficPolicy(),
			"host":           serviceHostWithRevision(r.Config, telemetryComponentName),
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
		"portLevelSettings": []map[string]interface{}{
			{
				"port": map[string]interface{}{
					"number": 15004, // grpc-mixer-mtls
				},
				"tls": map[string]interface{}{
					"mode": "ISTIO_MUTUAL",
				},
			},
			{
				"port": map[string]interface{}{
					"number": 9091, // grpc-mixer
				},
				"tls": map[string]interface{}{
					"mode": "DISABLE",
				},
			},
		},
		"connection_pool": r.connectionPool(),
	}

	return tp
}
