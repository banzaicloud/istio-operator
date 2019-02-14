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

package pilot

import (
	"fmt"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func (r *Reconciler) configMap(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, cmLabels, owner),
		Data: map[string]string{
			"mesh": r.meshConfig(owner.Namespace),
		},
	}
}

func (r *Reconciler) meshConfig(ns string) string {
	meshConfig := map[string]interface{}{
		"disablePolicyChecks": false,
		"enableTracing":       true,
		"accessLogFile":       "/dev/stdout",
		"mixerCheckServer":    fmt.Sprintf("istio-policy.%s.svc.cluster.local:9091", ns),
		"mixerReportServer":   fmt.Sprintf("istio-telemetry.%s.svc.cluster.local:9091", ns),
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
			"controlPlaneAuthPolicy": "NONE",
			"discoveryAddress":       fmt.Sprintf("istio-pilot.%s:15007", ns),
		},
	}
	marshaledConfig, _ := yaml.Marshal(meshConfig)
	return string(marshaledConfig)
}
