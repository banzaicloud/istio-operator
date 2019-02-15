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

package mixer

import (
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *Reconciler) meshPolicy() *k8sutil.DynamicObject {
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
					"mtls": map[string]interface{}{
						"mode": "PERMISSIVE",
					},
				},
			},
		},
		Owner: r.Config,
	}
}
