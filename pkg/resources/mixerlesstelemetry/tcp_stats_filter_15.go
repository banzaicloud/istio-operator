/*
Copyright 2020 Banzai Cloud.

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
	tcpStatsFilterYAML15 = `
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_INBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: ^1\.5.*
      %[3]s
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_inbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: tcp_stats_inbound
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_OUTBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: ^1\.5.*
      %[3]s
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: tcp_stats_outbound
- applyTo: NETWORK_FILTER
  match:
    context: GATEWAY
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: ^1\.5.*
      %[3]s
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: tcp_stats_outbound
`
)

func (r *Reconciler) tcpStatsFilter15() *k8sutil.DynamicObject {
	return r.tcpStatsFilter(proxyVersion15, tcpStatsFilterYAML15)
}
