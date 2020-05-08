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
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	httpStatsFilterYAML = `
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_OUTBOUND
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio"
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_INBOUND
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio"
              }
            root_id: stats_inbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_inbound
- applyTo: HTTP_FILTER
  match:
    context: GATEWAY
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
                "disable_host_header_fallback": true,
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
`
	tcpStatsFilterYAML = `
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_INBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
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
              vm_id: stats_inbound
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_OUTBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
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
              vm_id: stats_outbound
- applyTo: NETWORK_FILTER
  match:
    context: GATEWAY
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
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
              vm_id: stats_outbound
`
	statsWasmLocal   = "filename: /etc/istio/extensions/stats-filter.wasm"
	statsNoWasmLocal = "inline_string: envoy.wasm.stats"
)

func (r *Reconciler) httpStatsFilter() *k8sutil.DynamicObject {

	wasmEnabled := util.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := statsNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = statsWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(httpStatsFilterYAML, vmConfigLocal, vmConfigRuntime)), &y)

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-stats",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpStatsFilter() *k8sutil.DynamicObject {

	wasmEnabled := util.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := statsNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = statsWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(tcpStatsFilterYAML, vmConfigLocal, vmConfigRuntime)), &y)

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-tcp-stats",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}
