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

package istiod

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/config"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName                = "istiod"
	serviceAccountName           = "istiod-service-account"
	clusterRoleNameIstiod        = "istiod-cluster-role"
	roleNameIstiod               = "istiod-role"
	clusterRoleBindingNameIstiod = "istiod-cluster-role-binding"
	roleBindingNameIstiod        = "istiod-role-binding"
	deploymentName               = "istiod"
	ServiceNameIstiod            = "istiod"
	hpaName                      = "istiod-autoscaler"
	pdbName                      = "istiod"
	validatingWebhookName        = "istiod"
)

var istiodLabels = map[string]string{
	"app": "istiod",
}

var istiodLabelSelector = map[string]string{
	"istio": "istiod",
}

var pilotLabelSelector = map[string]string{
	"istio": "pilot",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface

	operatorConfig config.Configuration
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, operatorConfig config.Configuration, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
			Scheme: scheme,
		},
		dynamic: dc,

		operatorConfig: operatorConfig,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	var istiodDesiredState k8sutil.DesiredState
	var pdbDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		istiodDesiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
			pdbDesiredState = k8sutil.DesiredStatePresent
		} else {
			pdbDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		istiodDesiredState = k8sutil.DesiredStateAbsent
		pdbDesiredState = k8sutil.DesiredStateAbsent
	}

	deploymentDesiredState := istiodDesiredState
	// add specific desired state to support re-creation
	if deploymentDesiredState == k8sutil.DesiredStatePresent {
		deploymentDesiredState = k8sutil.NewRecreateAwareDeploymentDesiredState(r.Client, r.Scheme, log, util.MergeStringMaps(pilotLabelSelector, r.Config.RevisionLabels()))
	}

	overlays, err := k8sutil.GetObjectModifiersForOverlays(r.Scheme, r.Config.Spec.K8SOverlays)
	if err != nil {
		return emperror.WrapWith(err, "could not get k8s overlay object modifiers")
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: istiodDesiredState},
		{Resource: r.clusterRole, DesiredState: istiodDesiredState},
		{Resource: r.role, DesiredState: istiodDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: istiodDesiredState},
		{Resource: r.roleBinding, DesiredState: istiodDesiredState},
		{Resource: r.deployment, DesiredState: deploymentDesiredState},
		{Resource: r.service, DesiredState: istiodDesiredState},
		{Resource: r.horizontalPodAutoscaler, DesiredState: istiodDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
		{Resource: r.validatingWebhook, DesiredState: istiodDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.ReconcileWithObjectModifiers(log, r.Client, o, res.DesiredState, k8sutil.CombineObjectModifiers([]k8sutil.ObjectModifierFunc{k8sutil.GetGVKObjectModifier(r.Scheme)}, overlays))
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	var meshExpansionDesiredState k8sutil.DesiredState
	var meshExpansionDestinationRuleDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Gateways.Enabled) && (util.PointerToBool(r.Config.Spec.Gateways.Ingress.Enabled) || util.PointerToBool(r.Config.Spec.Gateways.MeshExpansion.Enabled)) && util.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutil.DesiredStatePresent
		meshExpansionDestinationRuleDesiredState = k8sutil.DesiredStatePresent
	} else {
		meshExpansionDesiredState = k8sutil.DesiredStateAbsent
		meshExpansionDestinationRuleDesiredState = k8sutil.DesiredStateAbsent
	}

	for _, dr := range []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.meshExpansionDestinationRule, DesiredState: meshExpansionDestinationRuleDesiredState},
		{DynamicResource: r.meshExpansionVirtualService, DesiredState: meshExpansionDesiredState},
	} {
		o := dr.DynamicResource()
		err := o.ReconcileWithObjectModifiers(log, r.dynamic, dr.DesiredState, overlays)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}

func ServiceName() string {
	return ServiceNameIstiod
}
