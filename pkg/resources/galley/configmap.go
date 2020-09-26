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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
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
		ObjectMeta: templates.ObjectMetaWithRevision(configMapName, util.MergeStringMaps(galleyLabels, cmLabels), r.Config),
		Data:       make(map[string]string),
	}

	if util.PointerToBool(r.Config.Spec.Galley.ConfigValidation) {
		configmap.Data["validatingwebhookconfiguration.yaml"] = r.validatingWebhookConfig()
	}

	return configmap
}

func (r *Reconciler) validatingWebhookConfig() string {
	ignore := admissionregistrationv1.Ignore
	se := admissionregistrationv1.SideEffectClassNone
	webhook := admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   webhookName,
			Labels: galleyLabels,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name:                    "pilot.validation.istio.io",
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      serviceName,
						Namespace: r.Config.Namespace,
						Path:      util.StrPointer("/admitpilot"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources:   []string{"httpapispecs", "httpapispecbindings", "quotaspecs", "quotaspecbindings"},
						},
					},
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"security.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"authentication.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"networking.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"destinationrules", "envoyfilters", "gateways", "serviceentries", "sidecars", "virtualservices"},
						},
					},
				},
				FailurePolicy: &ignore,
				SideEffects:   &se,
			},
			{
				Name:                    "mixer.validation.istio.io",
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      serviceName,
						Namespace: r.Config.Namespace,
						Path:      util.StrPointer("/admitmixer"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources: []string{
								"rules",
								"attributemanifests",
								"adapters",
								"handlers",
								"instances",
								"templates",
							},
						},
					},
				},
				FailurePolicy: &ignore,
				SideEffects:   &se,
			},
		},
	}
	// this is a static config, so we don't have to deal with errors
	marshaledConfig, _ := yaml.Marshal(webhook)
	return string(marshaledConfig)
}
