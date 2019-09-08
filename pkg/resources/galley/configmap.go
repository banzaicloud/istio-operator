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

package galley

import (
	"github.com/ghodss/yaml"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var cmLabels = map[string]string{
	"istio": "galley",
}

func (r *Reconciler) configMap() runtime.Object {
	configmap := &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, util.MergeLabels(galleyLabels, cmLabels), r.Config),
		Data:       make(map[string]string),
	}

	if util.PointerToBool(r.Config.Spec.Galley.ConfigValidation) {
		configmap.Data["validatingwebhookconfiguration.yaml"] = r.validatingWebhookConfig(r.Config.Namespace)
	}

	return configmap
}

func (r *Reconciler) validatingWebhookConfig(ns string) string {
	fail := admissionv1beta1.Fail
	se := admissionv1beta1.SideEffectClassNone
	webhook := admissionv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookName,
			Namespace: ns,
			Labels:    galleyLabels,
		},
		Webhooks: []admissionv1beta1.Webhook{
			{
				Name: "pilot.validation.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      serviceName,
						Namespace: ns,
						Path:      util.StrPointer("/admitpilot"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources:   []string{"httpapispecs", "httpapispecbindings", "quotaspecs", "quotaspecbindings"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"rbac.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"authentication.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"networking.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"destinationrules", "envoyfilters", "gateways", "serviceentries", "sidecars", "virtualservices"},
						},
					},
				},
				FailurePolicy: &fail,
				SideEffects:   &se,
			},
			{
				Name: "mixer.validation.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      serviceName,
						Namespace: ns,
						Path:      util.StrPointer("/admitmixer"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources: []string{
								"rules",
								"attributemanifests",
								"circonuses",
								"deniers",
								"fluentds",
								"kubernetesenvs",
								"listcheckers",
								"memquotas",
								"noops",
								"opas",
								"prometheuses",
								"rbacs",
								"solarwindses",
								"stackdrivers",
								"cloudwatches",
								"dogstatsds",
								"statsds",
								"stdios",
								"apikeys",
								"authorizations",
								"checknothings",
								"listentries",
								"logentries",
								"metrics",
								"quotas",
								"reportnothings",
								"tracespans",
								"adapters",
								"handlers",
								"instances",
								"templates",
								"zipkins",
							},
						},
					},
				},
				FailurePolicy: &fail,
				SideEffects:   &se,
			},
		},
	}
	// this is a static config, so we don't have to deal with errors
	marshaledConfig, _ := yaml.Marshal(webhook)
	return string(marshaledConfig)
}
