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

package sidecarinjector

import (
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/istiod"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) webhook() runtime.Object {
	fail := admissionv1beta1.Fail
	unknownSideEffects := admissionv1beta1.SideEffectClassUnknown
	service := serviceName
	if !util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		service = istiod.ServiceNameIstiod
	}
	webhook := &admissionv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScope(webhookName, sidecarInjectorLabels, r.Config),
		Webhooks: []admissionv1beta1.Webhook{
			{
				Name: "sidecar-injector.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      service,
						Namespace: r.Config.Namespace,
						Path:      util.StrPointer("/inject"),
					},
					CABundle: nil,
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
						},
						Rule: admissionv1beta1.Rule{
							Resources:   []string{"pods"},
							APIGroups:   []string{""},
							APIVersions: []string{"*"},
						},
					},
				},
				FailurePolicy: &fail,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"istio-injection": "enabled",
					},
				},
				SideEffects: &unknownSideEffects,
			},
		},
	}

	if util.PointerToBool(r.Config.Spec.SidecarInjector.EnableNamespacesByDefault) {
		webhook.Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{r.Config.Namespace},
				},
				{
					Key:      "istio-injection",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"disabled"},
				},
			},
		}
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		webhook.Webhooks[0].NamespaceSelector.MatchExpressions = append(webhook.Webhooks[0].NamespaceSelector.MatchExpressions, []metav1.LabelSelectorRequirement{
			{
				Key:      "istio-env",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
			{
				Key:      "istio.io/rev",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		}...)
	}

	return webhook
}
