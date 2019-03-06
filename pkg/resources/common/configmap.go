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

package common

import (
	"fmt"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(IstioConfigMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh":         r.meshConfig(),
			"meshNetworks": r.meshNetworks(),
		},
	}
}

func (r *Reconciler) meshConfig() string {
	meshConfig := map[string]interface{}{
		"disablePolicyChecks": false,
		"enableTracing":       true,
		"accessLogFile":       "/dev/stdout",
		"accessLogFormat":     "",
		"accessLogEncoding":   "TEXT",
		"mixerCheckServer":    r.mixerServer("policy"),
		"mixerReportServer":   r.mixerServer("telemetry"),
		"policyCheckFailOpen": false,
		"ingressService":      "istio-ingressgateway",
		"sdsUdsPath":          "",
		"enableSdsTokenMount": false,
		"sdsUseK8sSaJwt":      false,
		"trustDomain":         "",
		"outboundTrafficPolicy": map[string]interface{}{
			"mode": "ALLOW_ANY",
		},
		"defaultConfig": map[string]interface{}{
			"connectTimeout":         "10s",
			"configPath":             "/etc/istio/proxy",
			"binaryPath":             "/usr/local/bin/envoy",
			"serviceCluster":         "istio-proxy",
			"drainDuration":          "45s",
			"parentShutdownDuration": "1m0s",
			"proxyAdminPort":         15000,
			"concurrency":            0,
			"tracing": map[string]interface{}{
				"zipkin": map[string]interface{}{
					"address": fmt.Sprintf("zipkin.%s:9411", r.Config.Namespace),
				},
			},
			"controlPlaneAuthPolicy": templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
			"discoveryAddress":       fmt.Sprintf("istio-pilot.%s:%s", r.Config.Namespace, r.discoveryPort()),
		},
	}

	if r.Config.Spec.UseMCP {
		meshConfig["configSources"] = []map[string]interface{}{
			r.defaultConfigSource(),
		}
	}

	marshaledConfig, _ := yaml.Marshal(meshConfig)
	return string(marshaledConfig)
}

func (r *Reconciler) meshNetworks() string {
	meshNetworks := map[string]interface{}{
		"networks": map[string]interface{}{},
	}
	marshaledConfig, _ := yaml.Marshal(meshNetworks)
	return string(marshaledConfig)
}

func (r *Reconciler) mixerServer(mixerType string) string {
	if r.remote {
		return fmt.Sprintf("istio-%s.%s:%s", mixerType, r.Config.Namespace, "15004")
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return fmt.Sprintf("istio-%s.%s.svc.cluster.local:%s", mixerType, r.Config.Namespace, "15004")
	}
	return fmt.Sprintf("istio-%s.%s.svc.cluster.local:%s", mixerType, r.Config.Namespace, "9091")
}

func (r *Reconciler) defaultConfigSource() map[string]interface{} {
	cs := map[string]interface{}{
		"address": fmt.Sprintf("istio-galley.%s.svc:9901", r.Config.Namespace),
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		cs["tlsSettings"] = map[string]interface{}{
			"mode": "ISTIO_MUTUAL",
		}
	}
	return cs
}

func (r *Reconciler) discoveryPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15011"
	}
	return "15010"
}
