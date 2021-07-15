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

package cni

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName                = "cni"
	serviceAccountName           = "istio-cni"
	clusterRoleName              = "istio-cni"
	clusterRoleRepairName        = "istio-cni-repair-role"
	clusterRoleTaintName         = "istio-cni-taint-role"
	clusterRoleBindingName       = "istio-cni"
	clusterRoleBindingRepairName = "istio-cni-repair-rolebinding"
	clusterRoleBindingTaintName  = "istio-cni-taint-rolebinding"
	daemonSetName                = "istio-cni-node"
	configMapName                = "istio-cni-config"
	taintConfigMapName           = "istio-cni-taint-configmap"
)

var cniLabels = map[string]string{
	"k8s-app": "istio-cni-node",
}

var defaultLabels = map[string]string{
	"app": "istio-cni",
}

var labelSelector = map[string]string{
	"k8s-app": "istio-cni-node",
}

type Reconciler struct {
	resources.Reconciler
}

func New(client client.Client, config *istiov1beta1.Istio, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
			Scheme: scheme,
		},
	}
}

func (r *Reconciler) Cleanup(log logr.Logger) error {
	return nil
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	desiredState := k8sutil.DesiredStatePresent
	desiredStateRepair := k8sutil.DesiredStatePresent
	desiredStateTaint := k8sutil.DesiredStatePresent
	if !util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled) {
		desiredState = k8sutil.DesiredStateAbsent
		desiredStateRepair = k8sutil.DesiredStateAbsent
	}
	if !util.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Taint.Enabled) {
		desiredStateTaint = k8sutil.DesiredStateAbsent
	}

	log.Info("Reconciling")

	overlays, err := k8sutil.GetObjectModifiersForOverlays(r.Scheme, r.Config.Spec.K8SOverlays)
	if err != nil {
		return emperror.WrapWith(err, "could not get k8s overlay object modifiers")
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: desiredState},
		{Resource: r.clusterRole, DesiredState: desiredState},
		{Resource: r.clusterRoleRepair, DesiredState: desiredStateRepair},
		{Resource: r.clusterRoleTaint, DesiredState: desiredStateTaint},
		{Resource: r.clusterRoleBinding, DesiredState: desiredState},
		{Resource: r.clusterRoleBindingRepair, DesiredState: desiredStateRepair},
		{Resource: r.clusterRoleBindingTaint, DesiredState: desiredStateTaint},
		{Resource: r.configMap, DesiredState: desiredState},
		{Resource: r.configMapTaint, DesiredState: desiredStateTaint},
		{Resource: r.daemonSet, DesiredState: desiredState},
	} {
		o := res.Resource()
		err := k8sutil.ReconcileWithObjectModifiers(log, r.Client, o, res.DesiredState, k8sutil.CombineObjectModifiers([]k8sutil.ObjectModifierFunc{k8sutil.GetGVKObjectModifier(r.Scheme)}, overlays))
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
