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

package sidecarinjector

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName          = "sidecarinjector"
	serviceAccountName     = "istio-sidecar-injector-service-account"
	clusterRoleName        = "istio-sidecar-injector-cluster-role"
	clusterRoleBindingName = "istio-sidecar-injector-cluster-role-binding"
	configMapNameInjector  = "istio-sidecar-injector"
	webhookName            = "istio-sidecar-injector"
	deploymentName         = "istio-sidecar-injector"
	serviceName            = "istio-sidecar-injector"
	pdbName                = "istio-sidecar-injector"
)

var sidecarInjectorLabels = map[string]string{
	"app": "istio-sidecar-injector",
}

var labelSelector = map[string]string{
	"istio": "sidecar-injector",
}

type Reconciler struct {
	resources.Reconciler
}

func New(client client.Client, config *istiov1beta1.Istio) *Reconciler {
	if config.Spec.ExcludeIPRanges == "" && config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = "*"
	}

	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	var sidecarInjectorDesiredState k8sutil.DesiredState
	var istiodDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) {
		sidecarInjectorDesiredState = k8sutil.DesiredStatePresent
		istiodDesiredState = k8sutil.DesiredStatePresent
	} else if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		sidecarInjectorDesiredState = k8sutil.DesiredStateAbsent
		istiodDesiredState = k8sutil.DesiredStatePresent
	} else {
		sidecarInjectorDesiredState = k8sutil.DesiredStateAbsent
		istiodDesiredState = k8sutil.DesiredStateAbsent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: sidecarInjectorDesiredState},
		{Resource: r.clusterRole, DesiredState: sidecarInjectorDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: sidecarInjectorDesiredState},
		{Resource: r.configMapInjector, DesiredState: istiodDesiredState},
		{Resource: r.deployment, DesiredState: sidecarInjectorDesiredState},
		{Resource: r.service, DesiredState: sidecarInjectorDesiredState},
		{Resource: r.webhook, DesiredState: istiodDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	if util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) || util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		err := r.reconcileLegacyAutoInjectionLabels(log)
		if err != nil {
			return emperror.WrapWith(err, "failed to label namespaces")
		}
	}

	err := r.reconcileAutoInjectionLabels(log)
	if err != nil {
		return emperror.WrapWith(err, "failed to label namespaces")
	}

	log.Info("Reconciled")

	return nil
}
