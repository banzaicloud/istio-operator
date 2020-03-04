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

package mixerlesstelemetry

import (
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const metadataExchangeFilterYAML = `
- applyTo: HTTP_FILTER
  match:
    context: ANY # inbound, outbound, and gateway
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: "envoy.http_connection_manager"
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        "@type": type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: envoy.wasm.metadata_exchange
            vm_config:
              runtime: envoy.wasm.runtime.null
              code:
                local:
                  inline_string: envoy.wasm.metadata_exchange
`

const TCPMetadataExchangeFilterYAML = `
    - applyTo: NETWORK_FILTER
      match:
        context: SIDECAR_INBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        listener: {}
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.network.metadata_exchange
          config:
            protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: GATEWAY
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
`

func (r *Reconciler) metaexchangeEnvoyFilter() *k8sutil.DynamicObject {
	var y []map[string]interface{}
	yaml.Unmarshal([]byte(metadataExchangeFilterYAML), &y)

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-metadata-exchange",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) TCPMetaexchangeEnvoyFilter() *k8sutil.DynamicObject {
	var y []map[string]interface{}
	yaml.Unmarshal([]byte(TCPMetadataExchangeFilterYAML), &y)

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-tcp-metadata-exchange",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}
