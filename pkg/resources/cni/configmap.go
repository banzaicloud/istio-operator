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

package cni

import (
	"encoding/json"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, nil, r.Config),
		Data: map[string]string{
			"cni_network_config": r.networkConfig(),
		},
	}
}

func (r *Reconciler) networkConfig() string {
	config := map[string]interface{}{
		"type":      "istio-cni",
		"log_level": r.Config.Spec.SidecarInjector.InitCNIConfiguration.LogLevel,
		"kubernetes": map[string]interface{}{
			"kubeconfig":         "__KUBECONFIG_FILEPATH__",
			"cni_bin_dir":        r.Config.Spec.SidecarInjector.InitCNIConfiguration.BinDir,
			"exclude_namespaces": r.Config.Spec.SidecarInjector.InitCNIConfiguration.ExcludeNamespaces,
		},
	}

	marshaledConfig, _ := json.Marshal(config)
	return string(marshaledConfig)
}
