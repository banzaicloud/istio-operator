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

var gatewaySelector = map[string]interface{}{
	"istio": "ingressgateway",
}

func (r *Reconciler) gateway() *k8sutil.DynamicObject {
	spec := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"port": map[string]interface{}{
					"name":     "http",
					"protocol": "HTTP2",
					"number":   80,
				},
				"hosts": util.EmptyTypedStrSlice("*"),
			},
		},
		"selector": gatewaySelector,
	}

	if util.PointerToBool(r.Config.Spec.Gateways.K8sIngress.EnableHttps) {
		spec["servers"] = append([]interface{}{spec["servers"]}, map[string]interface{}{
			"port": map[string]interface{}{
				"name":     "https-default",
				"protocol": "HTTPS",
				"number":   443,
			},
			"hosts": util.EmptyTypedStrSlice("*"),
			"tls": map[string]interface{}{
				"mode":              "SIMPLE",
				"serverCertificate": "/etc/istio/ingressgateway-certs/tls.crt",
				"privateKey":        "/etc/istio/ingressgateway-certs/tls.key",
			},
		})
	}

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      defaultGatewayName,
		Namespace: r.Config.Namespace,
		Spec:      spec,
		Owner:     r.Config,
	}
}

func (r *Reconciler) clusterAwareGateway() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      "cluster-aware-gateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"port": map[string]interface{}{
						"name":     "tls",
						"protocol": "TLS",
						"number":   443,
					},
					"tls": map[string]interface{}{
						"mode": "AUTO_PASSTHROUGH",
					},
					"hosts": util.EmptyTypedStrSlice("*.local"),
				},
			},
			"selector": gatewaySelector,
		},
		Owner: r.Config,
	}
}
