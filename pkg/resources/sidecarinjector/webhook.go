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
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/istiod"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) webhook() runtime.Object {
	fail := admissionv1.Fail
	noneSideEffects := admissionv1.SideEffectClassNone
	scope := admissionv1.AllScopes
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

	webhookConfiguration := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(webhookName, sidecarInjectorLabels, r.Config),
		Webhooks:   []admissionv1.MutatingWebhook{},
	}

	webhook := &admissionv1.MutatingWebhook{
		Name: "sidecar-injector.istio.io",
		ClientConfig: admissionv1.WebhookClientConfig{
			Service: &admissionv1.ServiceReference{
				Name:      r.Config.WithRevision(service),
				Namespace: r.Config.Namespace,
				Path:      util.StrPointer("/inject"),
			},
			CABundle: nil,
		},
		Rules: []admissionv1.RuleWithOperations{
			{
				Operations: []admissionv1.OperationType{
					admissionv1.Create,
				},
				Rule: admissionv1.Rule{
					Resources:   []string{"pods"},
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Scope:       &scope,
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
		}
	}

	webhook.NamespaceSelector.MatchExpressions = matchExpression
	webhookConfiguration.Webhooks = append(webhookConfiguration.Webhooks, *webhook)

	if !util.PointerToBool(r.Config.Spec.SidecarInjector.EnableNamespacesByDefault) && util.PointerToBool(r.Config.Spec.Istiod.Enabled) && !r.Config.IsRevisionUsed() {
		wh := webhook.DeepCopy()
		wh.Name = "secondary." + wh.Name
		wh.NamespaceSelector.MatchExpressions = revisionLabelMatchExpression
		webhookConfiguration.Webhooks = append(webhookConfiguration.Webhooks, *wh)
	}

	return webhookConfiguration
}
