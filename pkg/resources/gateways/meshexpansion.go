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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) meshExpansionGateway() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      "meshexpansion-gateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"port": map[string]interface{}{
						"name":     "tcp-pilot",
						"protocol": "TCP",
						"number":   15011,
					},
					"hosts": util.EmptyTypedStrSlice("*"),
				},
				{
					"port": map[string]interface{}{
						"name":     "tcp-citadel",
						"protocol": "TCP",
						"number":   8060,
					},
					"hosts": util.EmptyTypedStrSlice("*"),
				},
				{
					"port": map[string]interface{}{
						"name":     "tls-mixer",
						"protocol": "TLS",
						"number":   15004,
					},
					"tls": map[string]interface{}{
						"mode": "AUTO_PASSTHROUGH",
					},
					"hosts": util.EmptyTypedStrSlice("*"),
				},
			},
			"selector": map[string]interface{}{
				"istio": "ingressgateway",
			},
		},
		Owner: r.Config,
	}
}
