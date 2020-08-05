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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) configMapEnvoy() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMetaWithRevision(configMapNameEnvoy, nil, r.Config),
		Data: map[string]string{
			"envoy.yaml.tmpl": r.envoyConfig(),
		},
	}
}

func (r *Reconciler) envoyConfig() string {
	yaml := `
    admin:
      access_log_path: /dev/null
      address:
        socket_address:
          address: 127.0.0.1
          port_value: 15000
    stats_config:
      use_all_default_tags: false
      stats_tags:
      - tag_name: cluster_name
        regex: '^cluster\.((.+?(\..+?\.svc\.cluster\.local)?)\.)'
      - tag_name: tcp_prefix
        regex: '^tcp\.((.*?)\.)\w+?$'
      - tag_name: response_code
        regex: '_rq(_(\d{3}))$'
      - tag_name: response_code_class
        regex: '_rq(_(\dxx))$'
      - tag_name: http_conn_manager_listener_prefix
        regex: '^listener(?=\.).*?\.http\.(((?:[_.[:digit:]]*|[_\[\]aAbBcCdDeEfF[:digit:]]*))\.)'
      - tag_name: http_conn_manager_prefix
        regex: '^http\.(((?:[_.[:digit:]]*|[_\[\]aAbBcCdDeEfF[:digit:]]*))\.)'
      - tag_name: listener_address
        regex: '^listener\.(((?:[_.[:digit:]]*|[_\[\]aAbBcCdDeEfF[:digit:]]*))\.)'
    static_resources:
      clusters:
      - name: prometheus_stats
        type: STATIC
        connect_timeout: 0.250s
        lb_policy: ROUND_ROBIN
        hosts:
        - socket_address:
            protocol: TCP
            address: 127.0.0.1
            port_value: 15000
      - name: sds-grpc
        type: STATIC
        http2_protocol_options: {}
        connect_timeout: 0.250s
        lb_policy: ROUND_ROBIN
        hosts:
        - pipe:
            path: "/etc/istio/proxy/SDS"
      - name: inbound_9092
        circuit_breakers:
          thresholds:
          - max_connections: 100000
            max_pending_requests: 100000
            max_requests: 100000
            max_retries: 3
        connect_timeout: 1.000s
        hosts:
        - pipe:
            path: /sock/mixer.socket
        http2_protocol_options: {}
      - name: out.galley.15019
        http2_protocol_options: {}
        connect_timeout: 1.000s
        type: STRICT_DNS
        circuit_breakers:
          thresholds:
            - max_connections: 100000
              max_pending_requests: 100000
              max_requests: 100000
              max_retries: 3
        tls_context:
          common_tls_context:
            tls_certificate_sds_secret_configs:
            - name: default
              sds_config:
                api_config_source:
                  api_type: GRPC
                  grpc_services:
                  - envoy_grpc:
                      cluster_name: sds-grpc
            combined_validation_context:
              default_validation_context:
                verify_subject_alt_name:
                - spiffe://` + r.Config.Spec.TrustDomain + `/ns/` + r.Config.Namespace + `/sa/istio-galley-service-account
              validation_context_sds_secret_config:
                name: ROOTCA
                sds_config:
                  api_config_source:
                    api_type: GRPC
                    grpc_services:
                    - envoy_grpc:
                        cluster_name: sds-grpc
        hosts:
          - socket_address:
              address: istio-galley.` + r.Config.Namespace + `
              port_value: 15019
      listeners:
      - name: "15090"
        address:
          socket_address:
            protocol: TCP
            address: 0.0.0.0
            port_value: 15090
        filter_chains:
        - filters:
          - name: envoy.http_connection_manager
            config:
              codec_type: AUTO
              stat_prefix: stats
              route_config:
                virtual_hosts:
                - name: backend
                  domains:
                  - '*'
                  routes:
                  - match:
                      prefix: /stats/prometheus
                    route:
                      cluster: prometheus_stats
              http_filters:
              - name: envoy.router
      - name: "15004"
        address:
          socket_address:
            address: 0.0.0.0
            port_value: 15004
        filter_chains:
        - filters:
          - config:
              codec_type: HTTP2
              http2_protocol_options:
                max_concurrent_streams: 1073741824
              generate_request_id: true
              http_filters:
              - config:
                  default_destination_service: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                  service_configs:
                    ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `:
                      disable_check_calls: true
    {{- if .DisableReportCalls }}
                      disable_report_calls: true
    {{- end }}
                      mixer_attributes:
                        attributes:
                          destination.service.host:
                            string_value: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.uid:
                            string_value: istio://` + r.Config.Namespace + `/services/` + serviceNameWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.name:
                            string_value: ` + serviceNameWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.namespace:
                            string_value: ` + r.Config.Namespace + `
                          destination.uid:
                            string_value: kubernetes://{{ .PodName }}.` + r.Config.Namespace + `
                          destination.namespace:
                            string_value: {{.Release.Namespace }}
                          destination.ip:
                            bytes_value: {{ .PodIP }}
                          destination.port:
                            int64_value: 15004
                          context.reporter.kind:
                            string_value: inbound
                          context.reporter.uid:
                            string_value: kubernetes://{{ .PodName }}.` + r.Config.Namespace + `
                  transport:
                    check_cluster: mixer_check_server
                    report_cluster: inbound_9092
                name: mixer
              - name: envoy.router
              route_config:
                name: "15004"
                virtual_hosts:
                - domains:
                  - '*'
                  name: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                  routes:
                  - decorator:
                      operation: Report
                    match:
                      prefix: /
                    route:
                      cluster: inbound_9092
                      timeout: 0.000s
              stat_prefix: "15004"
            name: envoy.http_connection_manager
`

	if r.Config.Spec.ControlPlaneSecurityEnabled {
		yaml += `
          tls_context:
            require_client_certificate: true
            common_tls_context:
              alpn_protocols:
              - h2
              tls_certificate_sds_secret_configs:
              - name: default
                sds_config:
                  api_config_source:
                    api_type: GRPC
                    grpc_services:
                    - envoy_grpc:
                        cluster_name: sds-grpc
              validation_context_sds_secret_config:
                name: ROOTCA
                sds_config:
                  api_config_source:
                    api_type: GRPC
                    grpc_services:
                    - envoy_grpc:
                        cluster_name: sds-grpc
`
	}

	yaml += `
      - name: "9091"
        address:
          socket_address:
            address: 0.0.0.0
            port_value: 9091
        filter_chains:
        - filters:
          - config:
              codec_type: HTTP2
              http2_protocol_options:
                max_concurrent_streams: 1073741824
              generate_request_id: true
              http_filters:
              - config:
                  default_destination_service: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                  service_configs:
                    ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `:
                      disable_check_calls: true
    {{- if .DisableReportCalls }}
                      disable_report_calls: true
    {{- end }}
                      mixer_attributes:
                        attributes:
                          destination.service.host:
                            string_value: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.uid:
                            string_value: istio://` + r.Config.Namespace + `/services/` + serviceNameWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.name:
                            string_value: ` + serviceNameWithRevision(r.Config, telemetryComponentName) + `
                          destination.service.namespace:
                            string_value: ` + r.Config.Namespace + `
                          destination.uid:
                            string_value: kubernetes://{{ .PodName }}.` + r.Config.Namespace + `
                          destination.namespace:
                            string_value: {{.Release.Namespace }}
                          destination.ip:
                            bytes_value: {{ .PodIP }}
                          destination.port:
                            int64_value: 9091
                          context.reporter.kind:
                            string_value: inbound
                          context.reporter.uid:
                            string_value: kubernetes://{{ .PodName }}.` + r.Config.Namespace + `
                  transport:
                    check_cluster: mixer_check_server
                    report_cluster: inbound_9092
                name: mixer
              - name: envoy.router
              route_config:
                name: "9091"
                virtual_hosts:
                - domains:
                  - '*'
                  name: ` + serviceHostWithRevision(r.Config, telemetryComponentName) + `
                  routes:
                  - decorator:
                      operation: Report
                    match:
                      prefix: /
                    route:
                      cluster: inbound_9092
                      timeout: 0.000s
              stat_prefix: "9091"
            name: envoy.http_connection_manager
      - name: "local.15019"
        address:
          socket_address:
            address: 127.0.0.1
            port_value: 15019
        filter_chains:
          - filters:
              - name: envoy.http_connection_manager
                config:
                  codec_type: HTTP2
                  stat_prefix: "15019"
                  stream_idle_timeout: 0s
                  http2_protocol_options:
                    max_concurrent_streams: 1073741824
                  access_log:
                    - name: envoy.file_access_log
                      config:
                        path: /dev/stdout
                  http_filters:
                    - name: envoy.router
                  route_config:
                    name: "15019"
                    virtual_hosts:
                      - name: istio-galley
                        domains:
                          - '*'
                        routes:
                          - match:
                              prefix: /
                            route:
                              cluster: out.galley.15019
                              timeout: 0.000s
`

	return yaml
}
