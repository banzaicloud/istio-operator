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

package server

import (
	"context"

	"github.com/go-logr/logr"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

func init() {
	webhooks = append(webhooks, NewConfigValidationWebhook)
}

const (
	configValidationWebhookName               = "istio.validation.banzaicloud.io"
	configValidationWebhookPath               = "/validate-istio-config"
	configValidationWebhookAlreadyExistsError = "istio config resource already exists"
)

// NewConfigValidationWebhook initializes an Istio config resource validator webhook configuration
func NewConfigValidationWebhook(mgr manager.Manager, logger logr.Logger) (*admission.Webhook, error) {
	return builder.NewWebhookBuilder().
		Name(configValidationWebhookName).
		Path(configValidationWebhookPath).
		FailurePolicy(admissionregistrationv1beta1.Ignore).
		Validating().
		NamespaceSelector(&metav1.LabelSelector{}).
		Operations(admissionregistrationv1beta1.Create).
		ForType(&istiov1beta1.Istio{}).
		Handlers(&istioConfigValidator{
			logger: logger,
		}).
		WithManager(mgr).
		Build()
}

type istioConfigValidator struct {
	client client.Client
	logger logr.Logger
}

// istioConfigValidator implements admission.Handler.
var _ admission.Handler = &istioConfigValidator{}

// Automatically generate RBAC rules to allow the Controller to validate IstioConfigs
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios,verbs=get;list;watch
func (wh *istioConfigValidator) Handle(ctx context.Context, req types.Request) types.Response {
	var configs istiov1beta1.IstioList

	err := wh.client.List(context.TODO(), &client.ListOptions{}, &configs)
	if err != nil {
		wh.logger.Error(err, "could not list istio config objects")
	}
	if len(configs.Items) == 0 {
		return admission.ValidationResponse(true, "")
	}

	return admission.ValidationResponse(false, configValidationWebhookAlreadyExistsError)
}

// istioConfigValidator implements inject.Client.
var _ inject.Client = &istioConfigValidator{}

// InjectClient injects the client into the istioConfigValidator
func (wh *istioConfigValidator) InjectClient(c client.Client) error {
	wh.client = c
	return nil
}
