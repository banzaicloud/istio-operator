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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName          = "citadel"
	serviceAccountName     = "istio-citadel-service-account"
	clusterRoleName        = "istio-citadel-cluster-role"
	clusterRoleBindingName = "istio-citadel-cluster-role-binding"
	deploymentName         = "istio-citadel"
	serviceName            = "istio-citadel"
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

func New(configuration Configuration, client client.Client, dc dynamic.Interface, config *istiov1beta1.Config) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic:       dc,
		configuration: configuration,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	for _, res := range []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.deployment,
		r.service,
	} {
		o := res()
		err := k8sutil.Reconcile(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	var drs []resources.DynamicResource

	if !r.configuration.DeployMeshPolicy {
		return nil
	}

	if r.Config.Spec.MTLS {
		drs = []resources.DynamicResource{
			r.meshPolicyMTLS,
			r.destinationRuleDefaultMtls,
			r.destinationRuleApiServerMtls,
		}
	} else {
		drs = []resources.DynamicResource{
			r.meshPolicy,
		}
	}

	for _, dr := range drs {
		o := dr()
		err := o.Reconcile(log, r.dynamic)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}
