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
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName          = "pilot"
	serviceAccountName     = "istio-pilot-service-account"
	clusterRoleName        = "istio-pilot-cluster-role"
	clusterRoleBindingName = "istio-pilot-cluster-role-binding"
	deploymentName         = "istio-pilot"
	serviceName            = "istio-pilot"
	hpaName                = "istio-pilot-autoscaler"
	pdbName                = "istio-pilot"
)

var pilotLabels = map[string]string{
	"app": "istio-pilot",
}

var labelSelector = map[string]string{
	"istio": "pilot",
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

	resources := []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: pilotDesiredState},
		{Resource: r.clusterRole, DesiredState: pilotDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: pilotDesiredState},
		{Resource: r.deployment, DesiredState: pilotDesiredState},
		{Resource: r.service, DesiredState: pilotDesiredState},
		{Resource: r.horizontalPodAutoscaler, DesiredState: pilotDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
	}

	for _, res := range resources {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
