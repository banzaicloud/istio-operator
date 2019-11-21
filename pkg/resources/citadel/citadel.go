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

package citadel

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/pkg/util"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName          = "citadel"
	serviceAccountName     = "istio-citadel-service-account"
	clusterRoleName        = "istio-citadel-cluster-role"
	clusterRoleBindingName = "istio-citadel-cluster-role-binding"
	deploymentName         = "istio-citadel"
	serviceName            = "istio-citadel"
	pdbName                = "istio-citadel"

	SelfSignedCASecretName = "istio-ca-secret"
)

var citadelLabels = map[string]string{
	"app": "security",
}

var labelSelector = map[string]string{
	"istio": "citadel",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface

	configuration Configuration
}

func New(configuration Configuration, client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic:       dc,
		configuration: configuration,
	}
}

func GetDeploymentName() string {
	return deploymentName
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	var citadelDesiredState k8sutil.DesiredState
	var pdbDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Citadel.Enabled) {
		citadelDesiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
			pdbDesiredState = k8sutil.DesiredStatePresent
		} else {
			pdbDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		citadelDesiredState = k8sutil.DesiredStateAbsent
		pdbDesiredState = k8sutil.DesiredStateAbsent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: citadelDesiredState},
		{Resource: r.clusterRole, DesiredState: citadelDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: citadelDesiredState},
		{Resource: r.deployment, DesiredState: citadelDesiredState},
		{Resource: r.service, DesiredState: citadelDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	if !r.configuration.DeployMeshPolicy {
		return nil
	}

	var meshExpansionDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutil.DesiredStatePresent
	} else {
		meshExpansionDesiredState = k8sutil.DesiredStateAbsent
	}

	var meshPolicyDesiredState k8sutil.DesiredState
	var mTLSDesiredState k8sutil.DesiredState
	if util.PointerToBool(r.Config.Spec.Citadel.Enabled) {
		if r.Config.Spec.MTLS {
			meshPolicyDesiredState = k8sutil.DesiredStatePresent

			if !r.Config.Spec.AutoMTLS {
				mTLSDesiredState = k8sutil.DesiredStatePresent
			} else {
				mTLSDesiredState = k8sutil.DesiredStateAbsent
			}
		} else {
			meshPolicyDesiredState = k8sutil.DesiredStatePresent
			mTLSDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		meshPolicyDesiredState = k8sutil.DesiredStateAbsent
		mTLSDesiredState = k8sutil.DesiredStateAbsent
	}
	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.meshPolicy, DesiredState: meshPolicyDesiredState},
		{DynamicResource: r.destinationRuleDefaultMtls, DesiredState: mTLSDesiredState},
		{DynamicResource: r.destinationRuleApiServerMtls, DesiredState: mTLSDesiredState},
		{DynamicResource: r.meshExpansion, DesiredState: meshExpansionDesiredState},
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
