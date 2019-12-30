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
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

var statsFilterYAML = `applyTo: HTTP_FILTER
match:
  context: %CONTEXT%
  listener:
    filterChain:
      filter:
        name: "envoy.http_connection_manager"
        subFilter:
          name: "envoy.router"
patch:
  operation: INSERT_BEFORE
  value:
    name: envoy.filters.http.wasm
    config:
      config:
        root_id: stats_outbound
        configuration: |
          {
            "debug": "false",
            "stat_prefix": "istio",
          }
        vm_config:
          vm_id: stats_outbound
          runtime: envoy.wasm.runtime.null
          code:
            inline_string: envoy.wasm.stats
`

func (r *Reconciler) statsEnvoyFilter() *k8sutil.DynamicObject {
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
			"configPatches": []map[string]interface{}{
				r.getStatFilter("SIDECAR_INBOUND"),
				r.getStatFilter("SIDECAR_OUTBOUND"),
				r.getStatFilter("GATEWAY"),
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) getStatFilter(listenerType string) map[string]interface{} {
	var y map[string]interface{}
	yaml.Unmarshal([]byte(strings.Replace(statsFilterYAML, "%CONTEXT%", listenerType, 1)), &y)

	return y
}
