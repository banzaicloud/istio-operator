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
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "mixerless-telemetry"
)

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	desiredState := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.Config.Spec.MixerlessTelemetry.Enabled) {
		desiredState = k8sutil.DesiredStatePresent
	}

	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.metaexchangeEnvoyFilter},
		{DynamicResource: r.statsEnvoyFilter},
	}

	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, desiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}

func (r *Reconciler) metaexchangeEnvoyFilter() *k8sutil.DynamicObject {
	filterInbound := map[string]interface{}{
		"filterConfig": map[string]interface{}{
			"configuration": "envoy.wasm.metadata_exchange",
			"vm_config": map[string]interface{}{
				"code": map[string]interface{}{
					"inline_string": "envoy.wasm.metadata_exchange",
				},
				"vm": "envoy.wasm.vm.null",
			},
		},
		"filterName": "envoy.wasm",
		"filterType": "HTTP",
		"insertPosition": map[string]interface{}{
			"index": "FIRST",
		},
		"listenerMatch": map[string]interface{}{
			"listenerProtocol": "HTTP",
			"listenerType":     "SIDECAR_INBOUND",
		},
	}

	var filterOutbound map[string]interface{}
	filterOutbound, _ = util.Map(filterInbound)
	filterOutbound["listenerMatch"].(map[string]interface{})["listenerType"] = "SIDECAR_OUTBOUND"

	var filterGateway map[string]interface{}
	filterGateway, _ = util.Map(filterInbound)
	filterGateway["listenerMatch"].(map[string]interface{})["listenerType"] = "GATEWAY"

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
			"filters": []map[string]interface{}{
				filterInbound,
				filterOutbound,
				filterGateway,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) statsEnvoyFilter() *k8sutil.DynamicObject {
	filterInbound := map[string]interface{}{
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
			"listenerType":     "SIDECAR_INBOUND",
		},
	}

	var filterOutbound map[string]interface{}
	filterOutbound, _ = util.Map(filterInbound)
	filterOutbound["listenerMatch"].(map[string]interface{})["listenerType"] = "SIDECAR_OUTBOUND"

	var filterGateway map[string]interface{}
	filterGateway, _ = util.Map(filterInbound)
	filterGateway["listenerMatch"].(map[string]interface{})["listenerType"] = "GATEWAY"

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
				filterInbound,
				filterOutbound,
				filterGateway,
			},
		},
		Owner: r.Config,
	}
}
