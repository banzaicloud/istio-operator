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

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(IstioConfigMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh": r.meshConfig(r.Config.Namespace),
		},
	}
}

func (r *Reconciler) meshConfig(ns string) string {
	meshConfig := map[string]interface{}{
		"disablePolicyChecks": false,
		"enableTracing":       true,
		"accessLogFile":       "/dev/stdout",
		"mixerCheckServer":    fmt.Sprintf("istio-policy.%s.svc.cluster.local:%s", ns, r.mixerPort()),
		"mixerReportServer":   fmt.Sprintf("istio-telemetry.%s.svc.cluster.local:%s", ns, r.mixerPort()),
		"policyCheckFailOpen": false,
		"sdsUdsPath":          "",
		"sdsRefreshDelay":     "15s",
		"defaultConfig": map[string]interface{}{
			"connectTimeout":         "10s",
			"configPath":             "/etc/istio/proxy",
			"binaryPath":             "/usr/local/bin/envoy",
			"serviceCluster":         "istio-proxy",
			"drainDuration":          "45s",
			"parentShutdownDuration": "1m0s",
			"proxyAdminPort":         15000,
			"concurrency":            0,
			"zipkinAddress":          fmt.Sprintf("zipkin.%s:9411", ns),
			"controlPlaneAuthPolicy": templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
			"discoveryAddress":       fmt.Sprintf("istio-pilot.%s:%s", ns, r.discoveryPort()),
		},
	}
	marshaledConfig, _ := yaml.Marshal(meshConfig)
	return string(marshaledConfig)
}

func (r *Reconciler) mixerPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15004"
	}
	return "9091"
}

func (r *Reconciler) discoveryPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15005"
	}
	return "15007"
}
