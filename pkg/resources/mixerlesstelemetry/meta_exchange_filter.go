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
	"strings"

	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	metaExchangeWasmLocal   = "filename: /etc/istio/extensions/metadata-exchange-filter.wasm"
	metaExchangeNoWasmLocal = "inline_string: envoy.wasm.metadata_exchange"
)

func (r *Reconciler) metadataMatch(padCount int) string {
	metadataMatch := ""
	if r.Config.IsRevisionUsed() {
		metadataMatch = fmt.Sprintf("metadata:\n%sREVISION: %s", strings.Repeat(" ", padCount), r.Config.NamespacedRevision())
	}

	return metadataMatch
}

func (r *Reconciler) metaExchangeEnvoyFilter(version string, metadataExchangeFilterYAML string) *k8sutil.DynamicObject {
	wasmEnabled := util.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := metaExchangeNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = metaExchangeWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(metadataExchangeFilterYAML, vmConfigLocal, vmConfigRuntime, r.metadataMatch(8))), &y)

	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      r.Config.WithRevision(fmt.Sprintf("%s-metadata-exchange-%s", componentName, version)),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}
