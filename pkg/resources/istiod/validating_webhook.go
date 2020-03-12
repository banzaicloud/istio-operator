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
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) webhooks() []admissionv1beta1.Webhook {
	// ignore := admissionv1beta1.Fail
	se := admissionv1beta1.SideEffectClassNone
	return []admissionv1beta1.Webhook{
		{
			Name: "validation.istio.io",
			ClientConfig: admissionv1beta1.WebhookClientConfig{
				Service: &admissionv1beta1.ServiceReference{
					Name:      ServiceNameIstiod,
					Namespace: r.Config.Namespace,
					Path:      util.StrPointer("/validate"),
				},
				CABundle: nil,
			},
			Rules: []admissionv1beta1.RuleWithOperations{
				{
					Operations: []admissionv1beta1.OperationType{
						admissionv1beta1.Create,
						admissionv1beta1.Update,
					},
					Rule: admissionv1beta1.Rule{
						APIGroups:   []string{"config.istio.io", "rbac.istio.io", "security.istio.io", "authentication.istio.io", "networking.istio.io"},
						APIVersions: []string{"*"},
						Resources:   []string{"*"},
					},
				},
			},
			FailurePolicy: nil,
			SideEffects:   &se,
		},
	}
}

func (r *Reconciler) validatingWebhook() runtime.Object {
	return &admissionv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScope(validatingWebhookName, util.MergeStringMaps(istiodLabels, istiodLabelSelector), r.Config),
		Webhooks:   r.webhooks(),
	}
}
