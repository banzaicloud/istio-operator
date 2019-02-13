package galley

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/ghodss/yaml"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cmLabels = map[string]string{
	"istio": "mixer",
}

func (r *Reconciler) configMap(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, util.MergeLabels(galleyLabels, cmLabels), owner),
		Data: map[string]string{
			"validatingwebhookconfiguration.yaml": r.validatingWebhookConfig(owner.Namespace),
		},
	}
}

func (r *Reconciler) validatingWebhookConfig(ns string) string {
	fail := admissionv1beta1.Fail
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
							Resources:   []string{"destinationrules", "envoyfilters", "gateways", "serviceentries", "virtualservices"},
						},
					},
				},
				FailurePolicy: &fail,
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
								"servicecontrols",
								"solarwindses",
								"stackdrivers",
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
								"servicecontrolreports",
								"tracespans"},
						},
					},
				},
				FailurePolicy: &fail,
			},
		},
	}
	// this is a static config, so we don't have to deal with errors
	marshaledConfig, _ := yaml.Marshal(webhook)
	return string(marshaledConfig)
}
