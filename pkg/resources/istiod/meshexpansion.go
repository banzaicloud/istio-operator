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

package istiod

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) meshExpansionVirtualService() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "virtualservices",
		},
		Kind:      "VirtualService",
		Name:      r.Config.WithName("meshexpansion-vs-istiod"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(istiodLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"hosts": []string{
				r.Config.GetDiscoveryHost(),
			},
			"gateways": []string{
				r.Config.WithName("meshexpansion-gateway"),
			},
			"tcp": []map[string]interface{}{
				{
					"match": []map[string]interface{}{
						{
							"port": r.Config.GetDiscoveryPort(),
						},
					},
					"route": []map[string]interface{}{
						{
							"destination": map[string]interface{}{
								"host": r.Config.GetDiscoveryHost(),
								"port": map[string]interface{}{
									"number": r.Config.GetDiscoveryPort(),
								},
							},
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) meshExpansionDestinationRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      r.Config.WithName("meshexpansion-dr-istiod"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(istiodLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"host": r.Config.GetDiscoveryHost(),
			"trafficPolicy": map[string]interface{}{
				"portLevelSettings": []map[string]interface{}{
					{
						"port": map[string]interface{}{
							"number": r.Config.GetDiscoveryPort(),
						},
						"tls": map[string]interface{}{
							"mode": "DISABLE",
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}
