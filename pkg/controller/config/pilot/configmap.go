package pilot

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	"fmt"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func configMap(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, cmLabels, owner),
		Data: map[string]string{
			"mesh": meshConfig(owner.Namespace),
		},
	}
}

func meshConfig(ns string) string {
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
