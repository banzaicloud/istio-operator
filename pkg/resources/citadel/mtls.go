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

package citadel

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

// spec returns a map to configure the mesh-wide PeerAuthentication
func (r *Reconciler) spec() map[string]interface{} {
	if r.Config.Spec.MeshPolicy.MTLSMode == v1beta1.DISABLED {
		return map[string]interface{}{
			"mtls": map[string]interface{}{
				"mode": "DISABLE",
			},
		}
	}

	return map[string]interface{}{
		"mtls": map[string]interface{}{
			"mode": r.Config.Spec.MeshPolicy.MTLSMode,
		},
	}
}

// peerAuthentication returns a peerAuthentication resource to either require, allow or disable mutual TLS
// for all services (that have sidecar) in the mesh
// https://istio.io/docs/tasks/security/authn-policy/
func (r *Reconciler) peerAuthentication() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "security.istio.io",
			Version:  "v1beta1",
			Resource: "peerauthentications",
		},
		Kind:      "PeerAuthentication",
		Name:      r.Config.WithRevision("default"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(citadelLabels, r.Config.RevisionLabels()),
		Spec:      r.spec(),
		Owner:     r.Config,
	}
}

// destinationRuleDefaultMtls returns a destination rule to configure client side to use mutual TLS when talking to
// any service (host) in the mesh
func (r *Reconciler) destinationRuleDefaultMtls() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision("default"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(citadelLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"host": "*.local",
			"trafficPolicy": map[string]interface{}{
				"tls": map[string]interface{}{
					"mode": "ISTIO_MUTUAL",
				},
			},
		},
		Owner: r.Config,
	}
}

// destinationRuleApiServerMtls returns a destination rule to disable (m)TLS when talking to API server, as API server doesn't have sidecar
// User should add similar destination rules for other services that don't have sidecar
func (r *Reconciler) destinationRuleApiServerMtls() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      r.Config.WithRevision("api-server"),
		Namespace: r.Config.Namespace,
		Labels:    util.MergeStringMaps(citadelLabels, r.Config.RevisionLabels()),
		Spec: map[string]interface{}{
			"host": "kubernetes.default.svc." + r.Config.Spec.Proxy.ClusterDomain,
			"trafficPolicy": map[string]interface{}{
				"tls": map[string]interface{}{
					"mode": "DISABLE",
				},
			},
		},
		Owner: r.Config,
	}
}
