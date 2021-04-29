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
	"github.com/banzaicloud/istio-operator/pkg/resources/gvr"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) kubernetesEnvHandler() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr:       gvr.IstioConfigHandler,
		Kind:      "handler",
		Name:      r.Config.WithRevision("kubernetesenv"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"compiledAdapter": "kubernetesenv",
			"params":          map[string]interface{}{},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) attributesKubernetes() *k8sutil.DynamicObject {
	attributes := &k8sutil.DynamicObject{
		Gvr:       gvr.IstioConfigInstance,
		Kind:      "instance",
		Name:      r.Config.WithRevision("attributes"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"compiledTemplate": "kubernetes",
			"params": map[string]interface{}{
				"source_uid":       `source.uid | ""`,
				"source_ip":        `source.ip | ip("0.0.0.0")`,
				"destination_uid":  `destination.uid | ""`,
				"destination_ip":   `destination.ip | ip("0.0.0.0")`,
				"destination_port": `destination.port | 0`,
			},
			"attributeBindings": map[string]string{
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
				"source.cluster.id":              `"unknown"`,
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
				"destination.cluster.id":         `"unknown"`,
			},
		},
		Owner: r.Config,
	}

	if util.PointerToBool(r.Config.Spec.Mixer.MultiClusterSupport) {
		if bindings, ok := attributes.Spec["attributeBindings"].(map[string]string); ok {
			bindings["source.cluster.id"] = `$out.source_cluster_id | "unknown"`
			bindings["destination.cluster.id"] = `$out.destination_cluster_id | "unknown"`
			attributes.Spec["attributeBindings"] = bindings
		}
	}

	return attributes
}

func (r *Reconciler) kubeAttrRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr:       gvr.IstioConfigRule,
		Kind:      "rule",
		Name:      r.Config.WithRevision("kubeattrgenrulerule"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   r.Config.WithRevision("kubernetesenv"),
					"instances": util.EmptyTypedStrSlice(r.Config.WithRevision("attributes")),
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpKubeAttrRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr:       gvr.IstioConfigRule,
		Kind:      "rule",
		Name:      r.Config.WithRevision("tcpkubeattrgenrulerule"),
		Namespace: r.Config.Namespace,
		Labels:    r.Config.RevisionLabels(),
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   r.Config.WithRevision("kubernetesenv"),
					"instances": util.EmptyTypedStrSlice(r.Config.WithRevision("attributes")),
				},
			},
			"match": `context.protocol == "tcp"`,
		},
		Owner: r.Config,
	}
}
