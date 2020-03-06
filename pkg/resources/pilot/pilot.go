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

package pilot

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
	componentName                = "pilot"
	serviceAccountName           = "istiod-service-account"
	clusterRoleName              = "istio-pilot-cluster-role"
	clusterRoleNameIstiod        = "istiod-cluster-role"
	clusterRoleNameGalley        = "istio-galley-cluster-role"
	clusterRoleBindingName       = "istio-pilot-cluster-role-binding"
	clusterRoleBindingNameIstiod = "istiod-cluster-role-binding"
	configMapName                = "istio"
	configMapNameInjector        = "istio-sidecar-injector"
	configMapNameEnvoy           = "pilot-envoy-config"
	deploymentName               = "istiod"
	serviceName                  = "istio-pilot"
	serviceNameIstiod            = "istiod"
	hpaName                      = "istiod-autoscaler"
	pdbName                      = "istiod"
	mutatingWebhookName          = "istio-sidecar-injector"
)

var pilotLabels = map[string]string{
	"app": "pilot",
}

var istiodLabels = map[string]string{
	"app": "istiod",
}

var labelSelector = map[string]string{
	"istio": "pilot",
}

var galleyLabels = map[string]string{
	"app": "istio-galley",
}

var sidecarInjectorLabels = map[string]string{
	"app": "sidecar-injector",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
	remote  bool
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, isRemote bool) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
		remote:  isRemote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	var pilotDesiredState k8sutil.DesiredState
	var pdbDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Pilot.Enabled) {
		pilotDesiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
			pdbDesiredState = k8sutil.DesiredStatePresent
		} else {
			pdbDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		pilotDesiredState = k8sutil.DesiredStateAbsent
		pdbDesiredState = k8sutil.DesiredStateAbsent
	}

	var istiodDesiredState k8sutil.DesiredState
	var galleyDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		istiodDesiredState = k8sutil.DesiredStatePresent
		if !util.PointerToBool(r.Config.Spec.Galley.Enabled) {
			galleyDesiredState = k8sutil.DesiredStatePresent
		} else {
			galleyDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		istiodDesiredState = k8sutil.DesiredStateAbsent
		galleyDesiredState = k8sutil.DesiredStateAbsent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: pilotDesiredState},
		{Resource: r.clusterRole, DesiredState: pilotDesiredState},
		{Resource: r.clusterRoleIstiod, DesiredState: pilotDesiredState},
		{Resource: r.clusterRoleGalley, DesiredState: galleyDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: pilotDesiredState},
		{Resource: r.clusterRoleBindingIstiod, DesiredState: istiodDesiredState},
		{Resource: r.configMap, DesiredState: istiodDesiredState},
		{Resource: r.configMapInjector, DesiredState: istiodDesiredState},
		{Resource: r.configMapEnvoy, DesiredState: pilotDesiredState},
		{Resource: r.deployment, DesiredState: pilotDesiredState},
		{Resource: r.service, DesiredState: pilotDesiredState},
		{Resource: r.serviceIstiod, DesiredState: istiodDesiredState},
		{Resource: r.horizontalPodAutoscaler, DesiredState: pilotDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
		{Resource: r.mutatingWebhook, DesiredState: istiodDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	var meshExpansionDesiredState k8sutil.DesiredState
	var meshExpansionDestinationRuleDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutil.DesiredStatePresent
		if r.Config.Spec.ControlPlaneSecurityEnabled {
			meshExpansionDestinationRuleDesiredState = k8sutil.DesiredStatePresent
		}
	} else {
		meshExpansionDesiredState = k8sutil.DesiredStateAbsent
		meshExpansionDestinationRuleDesiredState = k8sutil.DesiredStateAbsent
	}

	for _, dr := range []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.meshExpansionDestinationRule, DesiredState: meshExpansionDestinationRuleDesiredState},
		{DynamicResource: r.meshExpansionVirtualService, DesiredState: meshExpansionDesiredState},
	} {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, dr.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}
