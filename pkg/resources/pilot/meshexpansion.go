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

package pilot

import (
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/gvr"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) meshExpansionVirtualService() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr:       gvr.VirtualService,
		Kind:      "VirtualService",
		Name:      r.Config.WithRevision("meshexpansion-vs-pilot"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(pilotLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"hosts": []string{
				r.Config.GetDiscoveryHost(true),
			},
			"gateways": []string{
				"meshexpansion-gateway",
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
								"host": r.Config.GetDiscoveryHost(true),
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
		Gvr:       gvr.DestinationRule,
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision("meshexpansion-dr-pilot"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(pilotLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"host": r.Config.GetDiscoveryHost(true),
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
