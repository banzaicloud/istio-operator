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

package meshexpansion

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	multimeshResourceNamePrefix = "istio-multicluster"
)

func (r *Reconciler) multimeshIngressGateway(selector map[string]string) *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      r.Config.WithRevision(multimeshResourceNamePrefix + "-ingressgateway"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"hosts": util.EmptyTypedStrSlice("*.global"),
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
			"selector": selector,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) multimeshEnvoyFilter(selector map[string]string) *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      r.Config.WithRevision(multimeshResourceNamePrefix + "-ingressgateway"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"workloadSelector": map[string]interface{}{
				"labels": selector,
			},
			"configPatches": []map[string]interface{}{
				{
					"applyTo": "NETWORK_FILTER",
					"match": map[string]interface{}{
						"context": "GATEWAY",
						"listener": map[string]interface{}{
							"portNumber": 15443,
							"filterChain": map[string]interface{}{
								"filter": map[string]interface{}{
									"name": "envoy.filters.network.sni_cluster",
								},
							},
						},
					},
					"patch": map[string]interface{}{
						"operation": "INSERT_AFTER",
						"value": map[string]interface{}{
							"name": "envoy.filters.network.tcp_cluster_rewrite",
							"typed_config": map[string]interface{}{
								"@type":               "type.googleapis.com/istio.envoy.config.filter.network.tcp_cluster_rewrite.v2alpha1.TcpClusterRewrite",
								"cluster_pattern":     "\\.global$",
								"cluster_replacement": ".svc." + r.Config.Spec.Proxy.ClusterDomain,
							},
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) multimeshDestinationRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision(multimeshResourceNamePrefix + "-destinationrule"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"exportTo": util.EmptyTypedStrSlice("*"),
			"host":     fmt.Sprintf("*.global"),
			"trafficPolicy": map[string]interface{}{
				"tls": map[string]interface{}{
					"mode": "ISTIO_MUTUAL",
				},
			},
		},
		Owner: r.Config,
	}
}
