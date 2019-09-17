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
	"github.com/MakeNowJust/heredoc"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

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
			"filters": []map[string]interface{}{
				r.getStatFilter("SIDECAR_INBOUND"),
				r.getStatFilter("SIDECAR_OUTBOUND"),
				r.getStatFilter("GATEWAY"),
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) getStatFilter(listenerType string) map[string]interface{} {
	return map[string]interface{}{
		"filterConfig": map[string]interface{}{
			"configuration": heredoc.Doc(`
			  {
			    "debug": "false",
			    "stat_prefix": "istio",
			  }
			`),
			"vm_config": map[string]interface{}{
				"code": map[string]interface{}{
					"inline_string": "envoy.wasm.stats",
				},
				"vm": "envoy.wasm.vm.null",
			},
		},
		"filterName": "envoy.wasm",
		"filterType": "HTTP",
		"insertPosition": map[string]interface{}{
			"index":      "BEFORE",
			"relativeTo": "envoy.router",
		},
		"listenerMatch": map[string]interface{}{
			"listenerProtocol": "HTTP",
			"listenerType":     listenerType,
		},
	}
}
