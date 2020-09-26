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

package istiod

import (
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) webhooks() []admissionregistrationv1beta1.ValidatingWebhook {
	se := admissionregistrationv1beta1.SideEffectClassNone
	scope := admissionregistrationv1beta1.AllScopes
	return []admissionregistrationv1beta1.ValidatingWebhook{
		{
			Name: "validation.istio.io",
			ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
				Service: &admissionregistrationv1beta1.ServiceReference{
					Name:      r.Config.WithRevision(ServiceNameIstiod),
					Namespace: r.Config.Namespace,
					Path:      util.StrPointer("/validate"),
				},
				// patched at runtime when the webhook is ready
				CABundle: nil,
			},
			Rules: []admissionregistrationv1beta1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1beta1.OperationType{
						admissionregistrationv1beta1.Create,
						admissionregistrationv1beta1.Update,
					},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{"config.istio.io", "rbac.istio.io", "security.istio.io", "authentication.istio.io", "networking.istio.io"},
						APIVersions: []string{"*"},
						Resources:   []string{"*"},
						Scope:       &scope,
					},
				},
			},
			FailurePolicy: nil,
			SideEffects:   &se,
		},
	}
}

func (r *Reconciler) validatingWebhook() runtime.Object {
	return &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(validatingWebhookName, util.MergeMultipleStringMaps(istiodLabels, istiodLabelSelector, r.Config.RevisionLabels()), r.Config),
		Webhooks:   r.webhooks(),
	}
}
