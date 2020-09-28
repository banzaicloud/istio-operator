/*
Copyright 2020 Banzai Cloud.

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

package webhook

import (
	"context"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/istio"
	"github.com/goph/emperror"
)

type IstioResourceValidator struct {
	mgr     ctrl.Manager
	decoder *admission.Decoder
}

func NewIstioResourceValidator(mgr ctrl.Manager) *IstioResourceValidator {
	return &IstioResourceValidator{
		mgr: mgr,
	}
}

func (v *IstioResourceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &istiov1beta1.Istio{}
	err := v.decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, emperror.Wrap(err, "could not decode object"))
	}

	yes, err := istio.IsControlPlaneShouldBeRevisioned(v.mgr.GetClient(), obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, emperror.Wrap(err, "could not check control plane revisions"))
	}
	if yes && !obj.IsRevisionUsed() {
		return admission.Denied("'global' property must be set to false. A global Istio control plane already exists.")
	}

	return admission.Allowed("")
}

func (v *IstioResourceValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
