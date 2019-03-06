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

func (r *Reconciler) stdioHandler() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "handlers",
		},
		Kind:      "handler",
		Name:      "stdio",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"params": map[string]interface{}{
				"outputAsJson": true,
			},
			"compiledAdapter": "stdio",
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) accessLogLogentry() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "logentries",
		},
		Kind:      "logentry",
		Name:      "accesslog",
		Namespace: r.Config.Namespace,
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
				"responseFlags":              `context.proxy_error_code | ""`,
				"responseSize":               `response.size | 0`,
				"permissiveResponseCode":     `rbac.permissive.response_code | "none"`,
				"permissiveResponsePolicyID": `rbac.permissive.effective_policy_id | "none"`,
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
				"grpcStatus":                 `response.grpc_status | ""`,
				"grpcMessage":                `response.grpc_message | ""`,
			},
			"monitored_resource_type": `"global"`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpAccessLogLogentry() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "logentries",
		},
		Kind:      "logentry",
		Name:      "tcpaccesslog",
		Namespace: r.Config.Namespace,
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
				"responseFlags":              `context.proxy_error_code | ""`,
			},
			"monitored_resource_type": `"global"`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) stdioRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "stdio",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "stdio",
					"instances": util.EmptyTypedStrSlice("accesslog.logentry"),
				},
			},
			"match": `context.protocol == "http" || context.protocol == "grpc"`,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) stdioTcpRule() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "rules",
		},
		Kind:      "rule",
		Name:      "stdiotcp",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"handler":   "stdio",
					"instances": util.EmptyTypedStrSlice("tcpaccesslog.logentry"),
				},
			},
			"match": `context.protocol == "tcp"`,
		},
		Owner: r.Config,
	}
}
