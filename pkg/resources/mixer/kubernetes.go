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
	"github.com/banzaicloud/istio-operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *Reconciler) kubernetesEnvHandler() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "kubernetesenvs",
		},
		Kind:      "kubernetesenv",
		Name:      "handler",
		Namespace: r.Config.Namespace,
		Spec:      nil,
		Owner:     r.Config,
	}
}

func (r *Reconciler) attributesKubernetes() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "kuberneteses",
		},
		Kind:      "kubernetes",
		Name:      "attributes",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"source_uid":       `source.uid | ""`,
			"source_ip":        `source.ip | ip("0.0.0.0")`,
			"destination_uid":  `destination.uid | ""`,
			"destination_port": `destination.port | 0`,
			"attribute_bindings": map[string]interface{}{
				"source.ip":                      `$out.source_pod_ip | ip("0.0.0.0")`,
				"source.uid":                     `$out.source_pod_uid | "unknown"`,
				"source.labels":                  `$out.source_labels | emptyStringMap()`,
				"source.name":                    `$out.source_pod_name | "unknown"`,
				"source.namespace":               `$out.source_namespace | "default"`,
				"source.owner":                   `$out.source_owner | "unknown"`,
				"source.serviceAccount":          `$out.source_service_account_name | "unknown"`,
				"source.workload.uid":            `$out.source_workload_uid | "unknown"`,
				"source.workload.name":           `$out.source_workload_name | "unknown"`,
				"source.workload.namespace":      `$out.source_workload_namespace | "unknown"`,
				"destination.ip":                 `$out.destination_pod_ip | ip("0.0.0.0")`,
				"destination.uid":                `$out.destination_pod_uid | "unknown"`,
				"destination.labels":             `$out.destination_labels | emptyStringMap()`,
				"destination.name":               `$out.destination_pod_name | "unknown"`,
				"destination.container.name":     `$out.destination_container_name | "unknown"`,
				"destination.namespace":          `$out.destination_namespace | "default"`,
				"destination.owner":              `$out.destination_owner | "unknown"`,
				"destination.serviceAccount":     `$out.destination_service_account_name | "unknown"`,
				"destination.workload.uid":       `$out.destination_workload_uid | "unknown"`,
				"destination.workload.name":      `$out.destination_workload_name | "unknown"`,
				"destination.workload.namespace": `$out.destination_workload_namespace | "unknown"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) kubeAttrRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "kubeattrgenrulerule",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "handler.kubernetesenv",
					"instances": util.EmptyTypedStrSlice("attributes.kubernetes"),
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpKubeAttrRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "tcpkubeattrgenrulerule",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "handler.kubernetesenv",
					"instances": util.EmptyTypedStrSlice("attributes.kubernetes"),
				},
			},
			"match": `context.protocol == "tcp"`,
		},
		Owner: r.Config,
	}
}
