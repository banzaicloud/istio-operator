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
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName          = "cni"
	serviceAccountName     = "istio-cni"
	clusterRoleName        = "istio-cni"
	clusterRoleBindingName = "istio-cni"
	daemonSetName          = "istio-cni-node"
	configMapName          = "istio-cni-config"
)

var cniLabels = map[string]string{
	"k8s-app": "istio-cni-node",
}

var labelSelector = map[string]string{
	"k8s-app": "istio-cni-node",
}

type Reconciler struct {
	resources.Reconciler
}

func New(client client.Client, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	desiredState := k8sutil.DesiredStatePresent
	if !*r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled {
		desiredState = k8sutil.DesiredStateAbsent
	}

	log.Info("Reconciling")

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: desiredState},
		{Resource: r.clusterRole, DesiredState: desiredState},
		{Resource: r.clusterRoleBinding, DesiredState: desiredState},
		{Resource: r.configMap, DesiredState: desiredState},
		{Resource: r.daemonSet, DesiredState: desiredState},
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
