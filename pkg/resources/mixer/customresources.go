package mixer

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *Reconciler) dynamicCustomResources(istio *istiov1beta1.Config) []k8sutil.DynamicResource {
	return []k8sutil.DynamicResource{
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "attributemanifests",
			},
			Kind:      "attributemanifest",
			Name:      "istioproxy",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"attributes": map[string]interface{}{
					"origin.ip":                        map[string]interface{}{"valueType": "IP_ADDRESS"},
					"origin.uid":                       map[string]interface{}{"valueType": "STRING"},
					"origin.user":                      map[string]interface{}{"valueType": "STRING"},
					"request.headers":                  map[string]interface{}{"valueType": "STRING_MAP"},
					"request.id":                       map[string]interface{}{"valueType": "STRING"},
					"request.host":                     map[string]interface{}{"valueType": "STRING"},
					"request.method":                   map[string]interface{}{"valueType": "STRING"},
					"request.path":                     map[string]interface{}{"valueType": "STRING"},
					"request.reason":                   map[string]interface{}{"valueType": "STRING"},
					"request.referer":                  map[string]interface{}{"valueType": "STRING"},
					"request.scheme":                   map[string]interface{}{"valueType": "STRING"},
					"request.total_size":               map[string]interface{}{"valueType": "INT64"},
					"request.size":                     map[string]interface{}{"valueType": "INT64"},
					"request.time":                     map[string]interface{}{"valueType": "TIMESTAMP"},
					"request.useragent":                map[string]interface{}{"valueType": "STRING"},
					"response.code":                    map[string]interface{}{"valueType": "INT64"},
					"response.duration":                map[string]interface{}{"valueType": "DURATION"},
					"response.headers":                 map[string]interface{}{"valueType": "STRING_MAP"},
					"response.total_size":              map[string]interface{}{"valueType": "INT64"},
					"response.size":                    map[string]interface{}{"valueType": "INT64"},
					"response.time":                    map[string]interface{}{"valueType": "TIMESTAMP"},
					"source.uid":                       map[string]interface{}{"valueType": "STRING"},
					"source.user":                      map[string]interface{}{"valueType": "STRING"},
					"source.principal":                 map[string]interface{}{"valueType": "STRING"},
					"destination.uid":                  map[string]interface{}{"valueType": "STRING"},
					"destination.port":                 map[string]interface{}{"valueType": "INT64"},
					"destination.principal":            map[string]interface{}{"valueType": "STRING"},
					"connection.event":                 map[string]interface{}{"valueType": "STRING"},
					"connection.id":                    map[string]interface{}{"valueType": "STRING"},
					"connection.received.bytes":        map[string]interface{}{"valueType": "INT64"},
					"connection.received.bytes_total":  map[string]interface{}{"valueType": "INT64"},
					"connection.sent.bytes":            map[string]interface{}{"valueType": "INT64"},
					"connection.sent.bytes_total":      map[string]interface{}{"valueType": "INT64"},
					"connection.duration":              map[string]interface{}{"valueType": "DURATION"},
					"connection.mtls":                  map[string]interface{}{"valueType": "BOOL"},
					"connection.requested_server_name": map[string]interface{}{"valueType": "STRING"},
					"context.protocol":                 map[string]interface{}{"valueType": "STRING"},
					"context.timestamp":                map[string]interface{}{"valueType": "TIMESTAMP"},
					"context.time":                     map[string]interface{}{"valueType": "TIMESTAMP"},
					"context.reporter.local":           map[string]interface{}{"valueType": "BOOL"},
					"context.reporter.kind":            map[string]interface{}{"valueType": "STRING"},
					"context.reporter.uid":             map[string]interface{}{"valueType": "STRING"},
					"api.service":                      map[string]interface{}{"valueType": "STRING"},
					"api.version":                      map[string]interface{}{"valueType": "STRING"},
					"api.operation":                    map[string]interface{}{"valueType": "STRING"},
					"api.protocol":                     map[string]interface{}{"valueType": "STRING"},
					"request.auth.principal":           map[string]interface{}{"valueType": "STRING"},
					"request.auth.audiences":           map[string]interface{}{"valueType": "STRING"},
					"request.auth.presenter":           map[string]interface{}{"valueType": "STRING"},
					"request.auth.claims":              map[string]interface{}{"valueType": "STRING_MAP"},
					"request.auth.raw_claims":          map[string]interface{}{"valueType": "STRING"},
					"request.api_key":                  map[string]interface{}{"valueType": "STRING"},
				},
			},
			Owner: istio,
		},
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "attributemanifests",
			},
			Kind:      "attributemanifest",
			Name:      "kubernetes",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"attributes": map[string]interface{}{
					"source.ip":                      map[string]interface{}{"valueType": "IP_ADDRESS"},
					"source.labels":                  map[string]interface{}{"valueType": "STRING_MAP"},
					"source.metadata":                map[string]interface{}{"valueType": "STRING_MAP"},
					"source.name":                    map[string]interface{}{"valueType": "STRING"},
					"source.namespace":               map[string]interface{}{"valueType": "STRING"},
					"source.owner":                   map[string]interface{}{"valueType": "STRING"},
					"source.service":                 map[string]interface{}{"valueType": "STRING"},
					"source.serviceAccount":          map[string]interface{}{"valueType": "STRING"},
					"source.services":                map[string]interface{}{"valueType": "STRING"},
					"source.workload.uid":            map[string]interface{}{"valueType": "STRING"},
					"source.workload.name":           map[string]interface{}{"valueType": "STRING"},
					"source.workload.namespace":      map[string]interface{}{"valueType": "STRING"},
					"destination.ip":                 map[string]interface{}{"valueType": "IP_ADDRESS"},
					"destination.labels":             map[string]interface{}{"valueType": "STRING_MAP"},
					"destination.metadata":           map[string]interface{}{"valueType": "STRING_MAP"},
					"destination.owner":              map[string]interface{}{"valueType": "STRING"},
					"destination.name":               map[string]interface{}{"valueType": "STRING"},
					"destination.container.name":     map[string]interface{}{"valueType": "STRING"},
					"destination.namespace":          map[string]interface{}{"valueType": "STRING"},
					"destination.service":            map[string]interface{}{"valueType": "STRING"},
					"destination.service.uid":        map[string]interface{}{"valueType": "STRING"},
					"destination.service.name":       map[string]interface{}{"valueType": "STRING"},
					"destination.service.namespace":  map[string]interface{}{"valueType": "STRING"},
					"destination.service.host":       map[string]interface{}{"valueType": "STRING"},
					"destination.serviceAccount":     map[string]interface{}{"valueType": "STRING"},
					"destination.workload.uid":       map[string]interface{}{"valueType": "STRING"},
					"destination.workload.name":      map[string]interface{}{"valueType": "STRING"},
					"destination.workload.namespace": map[string]interface{}{"valueType": "STRING"},
				},
			},
			Owner: istio,
		},
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "stdios",
			},
			Kind:      "stdio",
			Name:      "handler",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"outputAsJson": true,
			},
			Owner: istio,
		},
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "logentries",
			},
			Kind:      "logentry",
			Name:      "accesslog",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"severity":  `"Info"`,
				"timestamp": "request.time",
				"variables": map[string]interface{}{
					"sourceIp":                   `source.ip | ip("0.0.0.0")`,
					"sourceApp":                  `source.labels["app"] | ""`,
					"sourcePrincipal":            `source.principal | ""`,
					"sourceName":                 `source.name | ""`,
					"sourceWorkload":             `source.workload.name | ""`,
					"sourceNamespace":            `source.namespace | ""`,
					"sourceOwner":                `source.owner | ""`,
					"destinationApp":             `destination.labels["app"] | ""`,
					"destinationIp":              `destination.ip | ip("0.0.0.0")`,
					"destinationServiceHost":     `destination.service.host | ""`,
					"destinationWorkload":        `destination.workload.name | ""`,
					"destinationName":            `destination.name | ""`,
					"destinationNamespace":       `destination.namespace | ""`,
					"destinationOwner":           `destination.owner | ""`,
					"destinationPrincipal":       `destination.principal | ""`,
					"apiClaims":                  `request.auth.raw_claims | ""`,
					"apiKey":                     `request.api_key | request.headers["x-api-key"] | ""`,
					"protocol":                   `request.scheme | context.protocol | "http"`,
					"method":                     `request.method | ""`,
					"url":                        `request.path | ""`,
					"responseCode":               `response.code | 0`,
					"responseSize":               `response.size | 0`,
					"requestSize":                `request.size | 0`,
					"requestId":                  `request.headers["x-request-id"] | ""`,
					"clientTraceId":              `request.headers["x-client-trace-id"] | ""`,
					"latency":                    `response.duration | "0ms"`,
					"connection_security_policy": `conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))`,
					"requestedServerName":        `connection.requested_server_name | ""`,
					"userAgent":                  `request.useragent | ""`,
					"responseTimestamp":          `response.time`,
					"receivedBytes":              `request.total_size | 0`,
					"sentBytes":                  `response.total_size | 0`,
					"referer":                    `request.referer | ""`,
					"httpAuthority":              `request.headers[":authority"] | request.host | ""`,
					"xForwardedFor":              `request.headers["x-forwarded-for"] | "0.0.0.0"`,
					"reporter":                   `conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")`,
				},
				"monitored_resource_type": `"global"`,
			},
			Owner: istio,
		},
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "logentries",
			},
			Kind:      "logentry",
			Name:      "tcpaccesslog",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"severity":  `"Info"`,
				"timestamp": `context.time | timestamp("2017-01-01T00:00:00Z")`,
				"variables": map[string]interface{}{
					"connectionEvent":            `connection.event | ""`,
					"sourceIp":                   `source.ip | ip("0.0.0.0")`,
					"sourceApp":                  `source.labels["app"] | ""`,
					"sourcePrincipal":            `source.principal | ""`,
					"sourceName":                 `source.name | ""`,
					"sourceWorkload":             `source.workload.name | ""`,
					"sourceNamespace":            `source.namespace | ""`,
					"sourceOwner":                `source.owner | ""`,
					"destinationApp":             `destination.labels["app"] | ""`,
					"destinationIp":              `destination.ip | ip("0.0.0.0")`,
					"destinationServiceHost":     `destination.service.host | ""`,
					"destinationWorkload":        `destination.workload.name | ""`,
					"destinationName":            `destination.name | ""`,
					"destinationNamespace":       `destination.namespace | ""`,
					"destinationOwner":           `destination.owner | ""`,
					"destinationPrincipal":       `destination.principal | ""`,
					"protocol":                   `context.protocol | "tcp"`,
					"connectionDuration":         `connection.duration | "0ms"`,
					"connection_security_policy": `conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))`,
					"requestedServerName":        `connection.requested_server_name | ""`,
					"receivedBytes":              `connection.received.bytes | 0`,
					"sentBytes":                  `connection.sent.bytes | 0`,
					"totalReceivedBytes":         `connection.received.bytes_total | 0`,
					"totalSentBytes":             `connection.sent.bytes_total | 0`,
					"reporter":                   `conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")`,
				},
				"monitored_resource_type": `"global"`,
			},
			Owner: istio,
		},
		rule("stdio", `context.protocol == "http" || context.protocol == "grpc"`, "handler.stdio", []string{"accesslog.logentry"}, istio),
		rule("stdiotcp", `context.protocol == "tcp"`, "handler.stdio", []string{"tcpaccesslog.logentry"}, istio),
		metric("requestcount", `"1"`, metricDimensions(), istio),
		metric("requestduration", `response.duration | "0ms"`, metricDimensions(), istio),
		metric("requestsize", `request.size | 0`, metricDimensions(), istio),
		metric("responsesize", `response.size | 0`, metricDimensions(), istio),
		metric("tcpbytesent", `connection.sent.bytes | 0`, tcpMetricDimensions(), istio),
		metric("tcpbytereceived", `connection.received.bytes | 0`, tcpMetricDimensions(), istio),
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "prometheuses",
			},
			Kind:      "prometheus",
			Name:      "handler",
			Namespace: istio.Namespace,
			Spec: map[string]interface{}{
				"metrics": []map[string]interface{}{
					{
						"name":          "requests_total",
						"instance_name": "requestcount.metric." + istio.Namespace,
						"kind":          "COUNTER",
						"label_names":   metricLabels(),
					},
					{
						"name":          "request_duration_seconds",
						"instance_name": "requestduration.metric." + istio.Namespace,
						"kind":          "DISTRIBUTION",
						"label_names":   metricLabels(),
						"buckets": map[string]interface{}{
							"explicit_buckets": map[string]interface{}{
								"bounds": fSlice(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
							},
						},
					},
					{
						"name":          "request_bytes",
						"instance_name": "requestsize.metric." + istio.Namespace,
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
						"instance_name": "responsesize.metric." + istio.Namespace,
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
						"instance_name": "tcpbytesent.metric." + istio.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
					{
						"name":          "tcp_received_bytes_total",
						"instance_name": "tcpbytereceived.metric." + istio.Namespace,
						"kind":          "COUNTER",
						"label_names":   tcpMetricLabels(),
					},
				},
			},
			Owner: istio,
		},
		rule("promhttp", `context.protocol == "http" || context.protocol == "grpc"`, "handler.prometheus",
			[]string{"requestcount.metric", "requestduration.metric", "requestsize.metric", "responsesize.metric"}, istio),
		rule("promtcp", `context.protocol == "tcp"`, "handler.prometheus",
			[]string{"tcpbytesent.metric", "tcpbytereceived.metric"}, istio),
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "kubernetesenvs",
			},
			Kind:      "kubernetesenv",
			Name:      "handler",
			Namespace: istio.Namespace,
			Spec:      nil,
			Owner:     istio,
		},
		rule("kubeattrgenrulerule", "", "handler.kubernetesenv", []string{"attributes.kubernetes"}, istio),
		rule("tcpkubeattrgenrulerule", `context.protocol == "tcp"`, "handler.kubernetesenv", []string{"attributes.kubernetes"}, istio),
		{
			Gvr: schema.GroupVersionResource{
				Group:    "config.istio.io",
				Version:  "v1alpha2",
				Resource: "kuberneteses",
			},
			Kind:      "kubernetes",
			Name:      "attributes",
			Namespace: istio.Namespace,
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
			Owner: istio,
		},
		{
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
			Owner: istio,
		},
	}
}

func iSlice(s ...string) []interface{} {
	ret := make([]interface{}, len(s))
	for i := 0; i < len(s); i++ {
		ret[i] = s[i]
	}
	return ret
}

func fSlice(f ...float64) []interface{} {
	ret := make([]interface{}, len(f))
	for i := 0; i < len(f); i++ {
		ret[i] = f[i]
	}
	return ret
}

func metricLabels() []interface{} {
	ml := tcpMetricLabels()
	ml = append(ml, "request_protocol")
	ml = append(ml, "response_code")
	return ml
}

func tcpMetricLabels() []interface{} {
	return iSlice([]string{
		"reporter",
		"source_app",
		"source_principal",
		"source_workload",
		"source_workload_namespace",
		"source_version",
		"destination_app",
		"destination_principal",
		"destination_workload",
		"destination_workload_namespace",
		"destination_version",
		"destination_service",
		"destination_service_name",
		"destination_service_namespace",
		"connection_security_policy",
	}...)
}

func rule(name, match, handler string, instances []string, owner *istiov1beta1.Config) k8sutil.DynamicResource {
	d := k8sutil.DynamicResource{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      name,
		Namespace: owner.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   handler,
					"instances": iSlice(instances...),
				},
			},
		},
		Owner: owner,
	}
	if len(match) > 0 {
		d.Spec["match"] = match
	}
	return d
}

func metric(name, value string, dimensions map[string]interface{}, owner *istiov1beta1.Config) k8sutil.DynamicResource {
	return k8sutil.DynamicResource{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "metrics",
		},
		Kind:      "metric",
		Name:      name,
		Namespace: owner.Namespace,
		Spec: map[string]interface{}{
			"value":                   value,
			"dimensions":              dimensions,
			"monitored_resource_type": `"UNSPECIFIED"`,
		},
		Owner: owner,
	}
}

func metricDimensions() map[string]interface{} {
	md := tcpMetricDimensions()
	md["request_protocol"] = `api.protocol | context.protocol | "unknown"`
	md["response_code"] = `response.code | 200`
	md["destination_service"] = `destination.service.host | "unknown"`
	return md
}

func tcpMetricDimensions() map[string]interface{} {
	return map[string]interface{}{
		"reporter":                       `conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")`,
		"source_workload":                `source.workload.name | "unknown"`,
		"source_workload_namespace":      `source.workload.namespace | "unknown"`,
		"source_principal":               `source.principal | "unknown"`,
		"source_app":                     `source.labels["app"] | "unknown"`,
		"source_version":                 `source.labels["version"] | "unknown"`,
		"destination_workload":           `destination.workload.name | "unknown"`,
		"destination_workload_namespace": `destination.workload.namespace | "unknown"`,
		"destination_principal":          `destination.principal | "unknown"`,
		"destination_app":                `destination.labels["app"] | "unknown"`,
		"destination_version":            `destination.labels["version"] | "unknown"`,
		"destination_service":            `destination.service.name | "unknown"`,
		"destination_service_name":       `destination.service.name | "unknown"`,
		"destination_service_namespace":  `destination.service.namespace | "unknown"`,
		"connection_security_policy":     `conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))`,
	}
}
