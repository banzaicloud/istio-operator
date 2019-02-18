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
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// meshPolicyMTLS returns an authentication policy to enable mutual TLS for all services (that have sidecar) in the mesh
// https://istio.io/docs/tasks/security/authn-policy/
func (r *Reconciler) meshPolicyMTLS() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "authentication.istio.io",
			Version:  "v1alpha1",
			Resource: "meshpolicies",
		},
		Kind: "MeshPolicy",
		Name: "default",
		Labels: map[string]string{
			"app": "istio-security",
		},
		Spec: map[string]interface{}{
			"peers": []map[string]interface{}{
				{
					"mtls": map[string]interface{}{},
				},
			},
		},
		Owner: r.Config,
	}
}

// defaultMTLS returns a destination rule to configure client side to use mutual TLS when talking to
// any service (host) in the mesh
func (r *Reconciler) defaultMTLS() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "default",
		Namespace: "default",
		Labels: map[string]string{
			"app": "istio-security",
		},
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

// apiServerMTLS returns a destination rule to disable (m)TLS when talking to API server, as API server doesn't have sidecar
// User should add similar destination rules for other services that don't have sidecar
func (r *Reconciler) apiServerMTLS() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "api-server",
		Labels: map[string]string{
			"app": "istio-security",
		},
		Spec: map[string]interface{}{
			"host": "kubernetes.default.svc.cluster.local",
			"trafficPolicy": map[string]interface{}{
				"tls": map[string]interface{}{
					"mode": "DISABLE",
				},
			},
		},
		Owner: r.Config,
	}
}
