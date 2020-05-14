/*
Copyright 2020 Banzai Cloud.

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

package mixerlesstelemetry

import (
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	tcpMetadataExchangeFilterYAML15 = `
    - applyTo: NETWORK_FILTER
      match:
        context: SIDECAR_INBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        listener: {}
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.network.metadata_exchange
          config:
            protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: GATEWAY
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
`
)

func (r *Reconciler) tcpMetaExchangeEnvoyFilter15() *k8sutil.DynamicObject {
	return r.tcpMetaExchangeEnvoyFilter(proxyVersion15, tcpMetadataExchangeFilterYAML15)
}
