/*
Copyright 2021 Banzai Cloud.

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

package webhookcert

import (
	"context"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) configMap() (runtime.Object, error) {
	var err error
	var caBundle string

	if util.PointerToBool(r.Config.Spec.Pilot.SPIFFE.OperatorEndpoints.Enabled) {
		caBundle, err = GetWebhookCABundles(r.Client, r.operatorConfig.WebhookConfigurationName)
		if err != nil {
			return nil, err
		}
	}

	return &corev1.ConfigMap{
		ObjectMeta: templates.ObjectMetaWithRevision(r.Config.WithRevision(ConfigMapName), nil, r.Config),
		Data: map[string]string{
			"cacert.pem": caBundle,
		},
	}, nil
}

func GetWebhookCABundles(client client.Client, webhookConfigurationName string) (string, error) {
	bundles := make([]string, 0)

	vwh := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	err := client.Get(context.Background(), types.NamespacedName{
		Name: webhookConfigurationName,
	}, vwh)
	if err != nil {
		return "", err
	}

	for _, wh := range vwh.Webhooks {
		bundle := string(wh.ClientConfig.CABundle)
		if bundle != "" {
			bundles = append(bundles, bundle)
		}
	}

	return strings.Join(bundles, "\n"), nil
}
