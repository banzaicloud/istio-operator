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
	noneSideEffects := admissionv1beta1.SideEffectClassNone
	service := serviceName
	if !util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		service = istiod.ServiceNameIstiod
	}

	globalMatchExpression := []metav1.LabelSelectorRequirement{
		{
			Key:      "istio-injection",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"enabled"},
		},
	}

	defaultAllowAnyMatchExpression := []metav1.LabelSelectorRequirement{
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
	}

	noRevisionLabelsMatchExpression := []metav1.LabelSelectorRequirement{
		{
			Key:      "istio-env",
			Operator: metav1.LabelSelectorOpDoesNotExist,
		},
		{
			Key:      "istio.io/rev",
			Operator: metav1.LabelSelectorOpDoesNotExist,
		},
	}

	revisionLabelMatchExpression := []metav1.LabelSelectorRequirement{
		{
			Key:      "istio-injection",
			Operator: metav1.LabelSelectorOpDoesNotExist,
		},
		{
			Key:      "istio.io/rev",
			Operator: metav1.LabelSelectorOpIn,
			Values: []string{
				r.Config.NamespacedRevision(),
			},
		},
	}

	webhookConfiguration := &admissionv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(webhookName, sidecarInjectorLabels, r.Config),
		Webhooks:   []admissionv1beta1.Webhook{},
	}

	webhook := &admissionv1beta1.Webhook{
		Name: "sidecar-injector.istio.io",
		ClientConfig: admissionv1beta1.WebhookClientConfig{
			Service: &admissionv1beta1.ServiceReference{
				Name:      r.Config.WithRevision(service),
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
					APIVersions: []string{"v1"},
				},
			},
		},
		FailurePolicy:     &fail,
		NamespaceSelector: &metav1.LabelSelector{},
		SideEffects:       &noneSideEffects,
	}

	matchExpression := make([]metav1.LabelSelectorRequirement, 0)

	if util.PointerToBool(r.Config.Spec.SidecarInjector.EnableNamespacesByDefault) {
		matchExpression = append(matchExpression, defaultAllowAnyMatchExpression...)
		if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
			matchExpression = append(matchExpression, noRevisionLabelsMatchExpression...)
		}
	} else if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		if r.Config.IsRevisionUsed() {
			matchExpression = append(matchExpression, revisionLabelMatchExpression...)
		} else {
			matchExpression = append(matchExpression, globalMatchExpression...)
			wh := webhook.DeepCopy()
			wh.NamespaceSelector.MatchExpressions = revisionLabelMatchExpression
			webhookConfiguration.Webhooks = append(webhookConfiguration.Webhooks, *wh)
		}
	}

	webhook.NamespaceSelector.MatchExpressions = matchExpression
	webhookConfiguration.Webhooks = append(webhookConfiguration.Webhooks, *webhook)

	return webhookConfiguration
}
