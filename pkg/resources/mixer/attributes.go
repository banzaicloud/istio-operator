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
)

func (r *Reconciler) istioProxyAttributeManifest() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "attributemanifests",
		},
		Kind:      "attributemanifest",
		Name:      "istioproxy",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"attributes": map[string]interface{}{
				"origin.ip":                           map[string]interface{}{"valueType": "IP_ADDRESS"},
				"origin.uid":                          map[string]interface{}{"valueType": "STRING"},
				"origin.user":                         map[string]interface{}{"valueType": "STRING"},
				"request.headers":                     map[string]interface{}{"valueType": "STRING_MAP"},
				"request.id":                          map[string]interface{}{"valueType": "STRING"},
				"request.host":                        map[string]interface{}{"valueType": "STRING"},
				"request.method":                      map[string]interface{}{"valueType": "STRING"},
				"request.path":                        map[string]interface{}{"valueType": "STRING"},
				"request.url_path":                    map[string]interface{}{"valueType": "STRING"},
				"request.query_params":                map[string]interface{}{"valueType": "STRING_MAP"},
				"request.reason":                      map[string]interface{}{"valueType": "STRING"},
				"request.referer":                     map[string]interface{}{"valueType": "STRING"},
				"request.scheme":                      map[string]interface{}{"valueType": "STRING"},
				"request.total_size":                  map[string]interface{}{"valueType": "INT64"},
				"request.size":                        map[string]interface{}{"valueType": "INT64"},
				"request.time":                        map[string]interface{}{"valueType": "TIMESTAMP"},
				"request.useragent":                   map[string]interface{}{"valueType": "STRING"},
				"response.code":                       map[string]interface{}{"valueType": "INT64"},
				"response.duration":                   map[string]interface{}{"valueType": "DURATION"},
				"response.headers":                    map[string]interface{}{"valueType": "STRING_MAP"},
				"response.total_size":                 map[string]interface{}{"valueType": "INT64"},
				"response.size":                       map[string]interface{}{"valueType": "INT64"},
				"response.time":                       map[string]interface{}{"valueType": "TIMESTAMP"},
				"response.grpc_status":                map[string]interface{}{"valueType": "STRING"},
				"response.grpc_message":               map[string]interface{}{"valueType": "STRING"},
				"source.uid":                          map[string]interface{}{"valueType": "STRING"},
				"source.user":                         map[string]interface{}{"valueType": "STRING"},
				"source.principal":                    map[string]interface{}{"valueType": "STRING"},
				"destination.uid":                     map[string]interface{}{"valueType": "STRING"},
				"destination.port":                    map[string]interface{}{"valueType": "INT64"},
				"destination.principal":               map[string]interface{}{"valueType": "STRING"},
				"connection.event":                    map[string]interface{}{"valueType": "STRING"},
				"connection.id":                       map[string]interface{}{"valueType": "STRING"},
				"connection.received.bytes":           map[string]interface{}{"valueType": "INT64"},
				"connection.received.bytes_total":     map[string]interface{}{"valueType": "INT64"},
				"connection.sent.bytes":               map[string]interface{}{"valueType": "INT64"},
				"connection.sent.bytes_total":         map[string]interface{}{"valueType": "INT64"},
				"connection.duration":                 map[string]interface{}{"valueType": "DURATION"},
				"connection.mtls":                     map[string]interface{}{"valueType": "BOOL"},
				"connection.requested_server_name":    map[string]interface{}{"valueType": "STRING"},
				"context.protocol":                    map[string]interface{}{"valueType": "STRING"},
				"context.proxy_error_code":            map[string]interface{}{"valueType": "STRING"},
				"context.timestamp":                   map[string]interface{}{"valueType": "TIMESTAMP"},
				"context.time":                        map[string]interface{}{"valueType": "TIMESTAMP"},
				"context.reporter.local":              map[string]interface{}{"valueType": "BOOL"},
				"context.reporter.kind":               map[string]interface{}{"valueType": "STRING"},
				"context.reporter.uid":                map[string]interface{}{"valueType": "STRING"},
				"api.service":                         map[string]interface{}{"valueType": "STRING"},
				"api.version":                         map[string]interface{}{"valueType": "STRING"},
				"api.operation":                       map[string]interface{}{"valueType": "STRING"},
				"api.protocol":                        map[string]interface{}{"valueType": "STRING"},
				"request.auth.principal":              map[string]interface{}{"valueType": "STRING"},
				"request.auth.audiences":              map[string]interface{}{"valueType": "STRING"},
				"request.auth.presenter":              map[string]interface{}{"valueType": "STRING"},
				"request.auth.claims":                 map[string]interface{}{"valueType": "STRING_MAP"},
				"request.auth.raw_claims":             map[string]interface{}{"valueType": "STRING"},
				"request.api_key":                     map[string]interface{}{"valueType": "STRING"},
				"rbac.permissive.response_code":       map[string]interface{}{"valueType": "STRING"},
				"rbac.permissive.effective_policy_id": map[string]interface{}{"valueType": "STRING"},
				"check.error_code":                    map[string]interface{}{"valueType": "INT64"},
				"check.error_message":                 map[string]interface{}{"valueType": "STRING"},
				"check.cache_hit":                     map[string]interface{}{"valueType": "BOOL"},
				"quota.cache_hit":                     map[string]interface{}{"valueType": "BOOL"},
				"context.proxy_version":               map[string]interface{}{"valueType": "STRING"},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) kubernetesAttributeManifest() *k8sutil.DynamicObject {
	return &k8sutil.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "config.istio.io",
			Version:  "v1alpha2",
			Resource: "attributemanifests",
		},
		Kind:      "attributemanifest",
		Name:      "kubernetes",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"attributes": map[string]interface{}{
				"source.ip":                      map[string]interface{}{"valueType": "IP_ADDRESS"},
				"source.labels":                  map[string]interface{}{"valueType": "STRING_MAP"},
				"source.metadata":                map[string]interface{}{"valueType": "STRING_MAP"},
				"source.name":                    map[string]interface{}{"valueType": "STRING"},
				"source.namespace":               map[string]interface{}{"valueType": "STRING"},
				"source.owner":                   map[string]interface{}{"valueType": "STRING"},
				"source.serviceAccount":          map[string]interface{}{"valueType": "STRING"},
				"source.services":                map[string]interface{}{"valueType": "STRING"},
				"source.workload.uid":            map[string]interface{}{"valueType": "STRING"},
				"source.workload.name":           map[string]interface{}{"valueType": "STRING"},
				"source.workload.namespace":      map[string]interface{}{"valueType": "STRING"},
				"source.cluster.id":              map[string]interface{}{"valueType": "STRING"},
				"destination.ip":                 map[string]interface{}{"valueType": "IP_ADDRESS"},
				"destination.labels":             map[string]interface{}{"valueType": "STRING_MAP"},
				"destination.metadata":           map[string]interface{}{"valueType": "STRING_MAP"},
				"destination.owner":              map[string]interface{}{"valueType": "STRING"},
				"destination.name":               map[string]interface{}{"valueType": "STRING"},
				"destination.container.name":     map[string]interface{}{"valueType": "STRING"},
				"destination.namespace":          map[string]interface{}{"valueType": "STRING"},
				"destination.service.uid":        map[string]interface{}{"valueType": "STRING"},
				"destination.service.name":       map[string]interface{}{"valueType": "STRING"},
				"destination.service.namespace":  map[string]interface{}{"valueType": "STRING"},
				"destination.service.host":       map[string]interface{}{"valueType": "STRING"},
				"destination.serviceAccount":     map[string]interface{}{"valueType": "STRING"},
				"destination.workload.uid":       map[string]interface{}{"valueType": "STRING"},
				"destination.workload.name":      map[string]interface{}{"valueType": "STRING"},
				"destination.workload.namespace": map[string]interface{}{"valueType": "STRING"},
				"destination.cluster.id":         map[string]interface{}{"valueType": "STRING"},
			},
		},
		Owner: r.Config,
	}
}
