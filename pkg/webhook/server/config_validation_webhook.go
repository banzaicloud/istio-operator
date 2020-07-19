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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/istio"
)

func init() {
	webhooks = append(webhooks, NewConfigValidationWebhook)
}

const (
	configValidationWebhookName = "istio.validation.banzaicloud.io"
	configValidationWebhookPath = "/validate-istio-config"
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
	sc := runtime.NewScheme()
	istiov1beta1.AddToScheme(sc)
	s := json.NewSerializer(json.DefaultMetaFactory, sc, sc, false)
	obj, _, err := s.Decode([]byte(req.AdmissionRequest.Object.Raw), nil, nil)
	if err != nil {
		return admission.ValidationResponse(false, emperror.Wrap(err, "could not decode object").Error())
	}

	if config, ok := obj.(*istiov1beta1.Istio); ok {
		yes, err := istio.IsControlPlaneShouldBeRevisioned(wh.client, config)
		if err != nil {
			return admission.ValidationResponse(false, emperror.Wrap(err, "could not check control plane revisions").Error())
		}
		if yes && !config.IsRevisionUsed() {
			return admission.ValidationResponse(false, fmt.Sprintf("'useRevision' must be set to true. A main Istio control plane is already exists in the '%s' namespace", config.Namespace))
		}

		return admission.ValidationResponse(true, "")
	}

	return admission.ValidationResponse(false, "could not check control plane revision")
}

// istioConfigValidator implements inject.Client.
var _ inject.Client = &istioConfigValidator{}

// InjectClient injects the client into the istioConfigValidator
func (wh *istioConfigValidator) InjectClient(c client.Client) error {
	wh.client = c
	return nil
}
