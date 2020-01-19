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

package mixerlesstelemetry

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "mixerless-telemetry"
)

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	exchangeFilterDesiredState := k8sutil.DesiredStateAbsent
	statsFilterDesiredState := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.Config.Spec.MixerlessTelemetry.Enabled) {
		exchangeFilterDesiredState = k8sutil.DesiredStatePresent
		statsFilterDesiredState = k8sutil.DesiredStatePresent
	}

	if util.PointerToBool(r.Config.Spec.Proxy.UseMetadataExchangeFilter) {
		exchangeFilterDesiredState = k8sutil.DesiredStatePresent
	}

	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.metaexchangeEnvoyFilter, DesiredState: exchangeFilterDesiredState},
		{DynamicResource: r.statsEnvoyFilter, DesiredState: statsFilterDesiredState},
	}

	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, dr.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}
