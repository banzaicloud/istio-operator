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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) prometheusHandler() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "handlers",
		},
		Kind:      "handler",
		Name:      "prometheus",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledAdapter": "prometheus",
			"params": map[string]interface{}{
				"metrics": []map[string]interface{}{
					{
						"name":          "requests_total",
						"instance_name": "requestcount.instance." + r.Config.Namespace,
						"kind":          "COUNTER",
						"label_names":   metricLabels(),
					},
					{
						"name":          "request_duration_seconds",
						"instance_name": "requestduration.instance." + r.Config.Namespace,
						"kind":          "DISTRIBUTION",
						"label_names":   metricLabels(),
						"buckets": map[string]interface{}{
							"explicit_buckets": map[string]interface{}{
								"bounds": util.EmptyTypedFloatSlice(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
							},
						},
					},
					{
						"name":          "request_bytes",
						"instance_name": "requestsize.instance." + r.Config.Namespace,
						"kind":          "DISTRIBUTION",
						"label_names":   metricLabels(),
						"buckets": map[string]interface{}{
							"exponentialBuckets": map[string]interface{}{
								"numFiniteBuckets": 8,
								"scale":            1,
								"growthFactor":     10,
							},
						},
					},
					{
						"name":          "response_bytes",
						"instance_name": "responsesize.instance." + r.Config.Namespace,
						"kind":          "DISTRIBUTION",
						"label_names":   metricLabels(),
						"buckets": map[string]interface{}{
							"exponentialBuckets": map[string]interface{}{
								"numFiniteBuckets": 8,
								"scale":            1,
								"growthFactor":     10,
							},
						},
					},
					{
						"name":          "tcp_sent_bytes_total",
						"instance_name": "tcpbytesent.instance." + r.Config.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
					{
						"name":          "tcp_received_bytes_total",
						"instance_name": "tcpbytereceived.instance." + r.Config.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
					{
						"name":          "tcp_connections_opened_total",
						"instance_name": "tcpconnectionsopened.instance." + r.Config.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
					{
						"name":          "tcp_connections_closed_total",
						"instance_name": "tcpconnectionsclosed.instance." + r.Config.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) requestCountMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "requestcount",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   "1",
				"dimensions":              metricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) requestDurationMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "requestduration",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   `response.duration | "0ms"`,
				"dimensions":              metricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) requestSizeMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "requestsize",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   `request.size | 0`,
				"dimensions":              metricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) responseSizeMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "responsesize",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   `response.size | 0`,
				"dimensions":              metricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpByteSentMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "tcpbytesent",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   `connection.sent.bytes | 0`,
				"dimensions":              tcpMetricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpByteReceivedMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "tcpbytereceived",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   `connection.received.bytes | 0`,
				"dimensions":              tcpMetricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpConnectionsOpenedMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "tcpconnectionsopened",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   "1",
				"dimensions":              tcpMetricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpConnectionsClosedMetric() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "instances",
		},
		Kind:      "instance",
		Name:      "tcpconnectionsclosed",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"compiledTemplate": "metric",
			"params": map[string]interface{}{
				"value":                   "1",
				"dimensions":              tcpMetricDimensions(),
				"monitored_resource_type": `"UNSPECIFIED"`,
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) promHttpRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "promhttp",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "prometheus",
					"instances": util.EmptyTypedStrSlice("requestcount", "requestduration", "requestsize", "responsesize"),
				},
			},
			"match": `(context.protocol == "http" || context.protocol == "grpc") && (match((request.useragent | "-"), "kube-probe*") == false)  && (match((request.useragent | "-"), "Prometheus*") == false)`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) promTcpRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "promtcp",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "prometheus",
					"instances": util.EmptyTypedStrSlice("tcpbytesent", "tcpbytereceived"),
				},
			},
			"match": `context.protocol == "tcp"`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) promTcpConnectionOpenRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "promtcpconnectionopen",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "prometheus",
					"instances": util.EmptyTypedStrSlice("tcpconnectionsopened"),
				},
			},
			"match": `context.protocol == "tcp" && ((connection.event | "na") == "open")`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) promTcpConnectionClosedRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "promtcpconnectionclosed",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "prometheus",
					"instances": util.EmptyTypedStrSlice("tcpconnectionsclosed"),
				},
			},
			"match": `context.protocol == "tcp" && ((connection.event | "na") == "close")`,
		},
		Owner: r.Config,
	}
}

func metricDimensions() map[string]interface{} {
	md := tcpMetricDimensions()
	md["request_protocol"] = `api.protocol | context.protocol | "unknown"`
	md["response_code"] = `response.code | 200`
	md["destination_service"] = `destination.service.host | request.host | "unknown"`
	md["response_flags"] = `context.proxy_error_code | "-"`
	md["permissive_response_code"] = `rbac.permissive.response_code | "none"`
	md["permissive_response_policyid"] = `rbac.permissive.effective_policy_id | "none"`
	return md
}

func tcpMetricDimensions() map[string]interface{} {
	return map[string]interface{}{
		"reporter":                  `conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")`,
		"source_workload":           `source.workload.name | "unknown"`,
		"source_workload_namespace": `source.workload.namespace | "unknown"`,
		"source_principal":          `source.principal | "unknown"`,
		"source_app":                `source.labels["app"] | "unknown"`,
		"source_version":            `source.labels["version"] | "unknown"`,
		// "source_cluster_id":              `source.cluster.id | "unknown"`,
		"destination_workload":           `destination.workload.name | "unknown"`,
		"destination_workload_namespace": `destination.workload.namespace | "unknown"`,
		"destination_principal":          `destination.principal | "unknown"`,
		"destination_app":                `destination.labels["app"] | "unknown"`,
		"destination_version":            `destination.labels["version"] | "unknown"`,
		"destination_service":            `destination.service.host | "unknown"`,
		"destination_service_name":       `destination.service.name | "unknown"`,
		"destination_service_namespace":  `destination.service.namespace | "unknown"`,
		// "destination_cluster_id":         `destination.cluster.id | "unknown"`,
		"connection_security_policy": `conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))`,
		"response_flags":             `context.proxy_error_code | "-"`,
	}
}

func metricLabels() []interface{} {
	ml := tcpMetricLabels()
	ml = append(ml, "request_protocol")
	ml = append(ml, "response_code")
	ml = append(ml, "permissive_response_code")
	ml = append(ml, "permissive_response_policyid")
	return ml
}

func tcpMetricLabels() []interface{} {
	return util.EmptyTypedStrSlice(
		"reporter",
		"source_app",
		"source_principal",
		"source_workload",
		"source_workload_namespace",
		"source_version",
		// "source_cluster_id",
		"destination_app",
		"destination_principal",
		"destination_workload",
		"destination_workload_namespace",
		"destination_version",
		"destination_service",
		"destination_service_name",
		"destination_service_namespace",
		// "destination_cluster_id",
		"connection_security_policy",
		"response_flags",
	)
}
