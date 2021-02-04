/*
Copyright 2021 Banzai Cloud.

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
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	httpStatsFilterYAML19 = `
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_OUTBOUND
    proxy:
      proxyVersion: '^1\.9.*'
      %[4]s
    listener:
      filterChain:
        filter:
          name: envoy.filters.network.http_connection_manager
          subFilter:
            name: envoy.filters.http.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: istio.stats
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
        value:
          config:
            configuration:
              "@type": "type.googleapis.com/google.protobuf.StringValue"
              value: |
                {
                  "debug": "false",
                  "stat_prefix": "istio",
                  "metrics": [
                    {
                      "dimensions": {
                        "source_cluster": "node.metadata['CLUSTER_ID']",
                        "destination_cluster": "upstream_peer.cluster_id"
                      }
                    }
                  ]
                }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              allow_precompiled: %[3]s
              vm_id: stats_outbound
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_INBOUND
    proxy:
      proxyVersion: '^1\.9.*'
      %[4]s
    listener:
      filterChain:
        filter:
          name: envoy.filters.network.http_connection_manager
          subFilter:
            name: envoy.filters.http.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: istio.stats
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
        value:
          config:
            configuration:
              "@type": "type.googleapis.com/google.protobuf.StringValue"
              value: |
                {
                  "debug": "false",
                  "stat_prefix": "istio",
                  "metrics": [
                    {
                      "dimensions": {
                        "destination_cluster": "node.metadata['CLUSTER_ID']",
                        "source_cluster": "downstream_peer.cluster_id"
                     }
                    }
                  ]
                }
            root_id: stats_inbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              allow_precompiled: %[3]s
              vm_id: stats_inbound
- applyTo: HTTP_FILTER
  match:
    context: GATEWAY
    proxy:
      proxyVersion: '^1\.9.*'
      %[4]s
    listener:
      filterChain:
        filter:
          name: envoy.filters.network.http_connection_manager
          subFilter:
            name: envoy.filters.http.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: istio.stats
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
        value:
          config:
            configuration:
              "@type": "type.googleapis.com/google.protobuf.StringValue"
              value: |
                {
                  "debug": "false",
                  "stat_prefix": "istio",
                  "metrics": [
                    {
                      "dimensions": {
                        "source_cluster": "node.metadata['CLUSTER_ID']",
                        "destination_cluster": "upstream_peer.cluster_id"
                     }
                    }
                  ]
                }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              allow_precompiled: %[3]s
              vm_id: stats_outbound
`
)

func (r *Reconciler) httpStatsFilter19() *k8sutil.DynamicObject {
	return r.httpStatsFilter(proxyVersion19, httpStatsFilterYAML19)
}
