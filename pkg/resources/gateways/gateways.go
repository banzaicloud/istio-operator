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

package gateways

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
	componentName             = "meshgateway"
	defaultIngressgatewayName = "istio-ingressgateway"
)

type Reconciler struct {
	resources.Reconciler
	gw      *istiov1beta1.MeshGateway
	dynamic dynamic.Interface
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, gw *istiov1beta1.MeshGateway) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		gw:      gw,
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	pdbDesiredState := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
		pdbDesiredState = k8sutil.DesiredStatePresent
	}

	sdsDesiredState := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.gw.Spec.SDS.Enabled) || util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		sdsDesiredState = k8sutil.DesiredStatePresent
	}

	hpaDesiredState := k8sutil.DesiredStateAbsent
	if r.gw.Spec.MinReplicas != nil && r.gw.Spec.MaxReplicas != nil && *r.gw.Spec.MinReplicas > 1 && *r.gw.Spec.MinReplicas != *r.gw.Spec.MaxReplicas {
		hpaDesiredState = k8sutil.DesiredStatePresent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.clusterRole, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.clusterRoleBinding, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.deployment, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.service, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.horizontalPodAutoscaler, DesiredState: hpaDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
		{Resource: r.role, DesiredState: sdsDesiredState},
		{Resource: r.roleBinding, DesiredState: sdsDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
