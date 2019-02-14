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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) webhook(owner *istiov1beta1.Config) runtime.Object {
	fail := admissionv1beta1.Fail
	return &admissionv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMeta(webhookName, sidecarInjectorLabels, owner),
		Webhooks: []admissionv1beta1.Webhook{
			{
				Name: "sidecar-injector.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      serviceName,
						Namespace: owner.Namespace,
						Path:      util.StrPointer("/inject"),
					},
					CABundle: []byte{},
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
			},
		},
	}
}
