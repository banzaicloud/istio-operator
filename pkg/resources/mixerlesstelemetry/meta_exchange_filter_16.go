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
	metadataExchangeFilterYAML16 = `
- applyTo: HTTP_FILTER
  match:
    context: ANY # inbound, outbound, and gateway
    proxy:
      proxyVersion: '^1\.6.*'
      metadata:
        REVISION: %[3]s
    listener:
      filterChain:
        filter:
          name: "envoy.http_connection_manager"
  patch:
    operation: INSERT_BEFORE
    value:
      name: istio.metadata_exchange
      typed_config:
        "@type": type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
        value:
          config:
            configuration: |
              {}
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
`
)

func (r *Reconciler) metaExchangeEnvoyFilter16() *k8sutil.DynamicObject {
	return r.metaExchangeEnvoyFilter(proxyVersion16, metadataExchangeFilterYAML16)
}
