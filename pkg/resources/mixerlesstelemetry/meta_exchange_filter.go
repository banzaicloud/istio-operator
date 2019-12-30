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

const metadataExchangeFilterYAML = `applyTo: HTTP_FILTER
match:
  context: ANY # inbound, outbound, and gateway
  listener:
    filterChain:
      filter:
        name: "envoy.http_connection_manager"
patch:
  operation: INSERT_BEFORE
  value:
    name: envoy.filters.http.wasm
    config:
      config:
        configuration: envoy.wasm.metadata_exchange
        vm_config:
          runtime: envoy.wasm.runtime.null
          code:
            inline_string: envoy.wasm.metadata_exchange
`

func (r *Reconciler) metaexchangeEnvoyFilter() *k8sutil.DynamicObject {
	var y map[string]interface{}
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
			"configPatches": []map[string]interface{}{
				y,
			},
		},
		Owner: r.Config,
	}
}
