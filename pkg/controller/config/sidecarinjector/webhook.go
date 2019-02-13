package sidecarinjector

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
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
