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

	"github.com/banzaicloud/istio-operator/pkg/util"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName          = "sidecarinjector"
	serviceAccountName     = "istio-sidecar-injector-service-account"
	clusterRoleName        = "istio-sidecar-injector-cluster-role"
	clusterRoleBindingName = "istio-sidecar-injector-cluster-role-binding"
	configMapNameInjector  = "istio-sidecar-injector-old"
	istioConfigMapName     = "injector-mesh"
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
	remote bool
}

func New(client client.Client, config *istiov1beta1.Istio, isRemote bool) *Reconciler {
	if config.Spec.ExcludeIPRanges == "" && config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = "*"
	}

	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		remote: isRemote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	var sidecarInjectorDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) {
		sidecarInjectorDesiredState = k8sutil.DesiredStatePresent
	} else {
		sidecarInjectorDesiredState = k8sutil.DesiredStateAbsent
	}

	ress := []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
		r.configMapInjector,
		r.deployment,
		r.service,
	}

	if sidecarInjectorDesiredState == k8sutil.DesiredStatePresent {
		ress = append(ress, r.webhook)
	}

	for _, res := range ress {
		o := res()
		err := k8sutil.Reconcile(log, r.Client, o, sidecarInjectorDesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	if util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) {
		err := r.reconcileAutoInjectionLabels(log)
		if err != nil {
			return emperror.WrapWith(err, "failed to label namespaces")
		}
	}

	log.Info("Reconciled")

	return nil
}
