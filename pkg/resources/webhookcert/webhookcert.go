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
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/config"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "webhookcert"
	ConfigMapName = "istio-operator-webhook-cabundle"
)

type Reconciler struct {
	resources.Reconciler

	operatorConfig config.Configuration
}

func New(client client.Client, config *istiov1beta1.Istio, operatorConfig config.Configuration, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
			Scheme: scheme,
		},
		operatorConfig: operatorConfig,
	}
}

func (r *Reconciler) Cleanup(log logr.Logger) error {
	return nil
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	desired := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.Config.Spec.Pilot.SPIFFE.OperatorEndpoints.Enabled) {
		desired = k8sutil.DesiredStatePresent
	}

	cm, err := r.configMap()
	if err != nil {
		return emperror.Wrap(err, "could not generate operator webhook cabundle configmap")
	}

	overlays, err := k8sutil.GetObjectModifiersForOverlays(r.Scheme, r.Config.Spec.K8SOverlays)
	if err != nil {
		return emperror.WrapWith(err, "could not get k8s overlay object modifiers")
	}

	for _, res := range []resources.Resource{
		func() runtime.Object { return cm },
	} {
		o := res()
		err := k8sutil.ReconcileWithObjectModifiers(log, r.Client, o, desired, k8sutil.CombineObjectModifiers([]k8sutil.ObjectModifierFunc{k8sutil.GetGVKObjectModifier(r.Scheme)}, overlays))
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
