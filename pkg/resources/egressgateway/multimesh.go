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

package egressgateway

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	multimeshResourceNamePrefix = "istio-multicluster"
)

func (r *Reconciler) multimeshEgressGateway() *k8sutil.DynamicObject {
	hosts := make([]string, 0)
	for _, domain := range r.Config.Spec.GetMultiMeshExpansion().GetDomains() {
		hosts = append(hosts, fmt.Sprintf("*.%s", domain))
	}

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      r.Config.WithRevision(multimeshResourceNamePrefix + "-egressgateway"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"hosts": hosts,
					"port": map[string]interface{}{
						"name":     "tls",
						"protocol": "TLS",
						"number":   15443,
					},
					"tls": map[string]interface{}{
						"mode": "AUTO_PASSTHROUGH",
					},
				},
			},
			"selector": r.labels(),
		},
		Owner: r.Config,
	}
}
