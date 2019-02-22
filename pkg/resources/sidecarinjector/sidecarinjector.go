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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountName     = "istio-sidecar-injector-service-account"
	clusterRoleName        = "istio-sidecar-injector-cluster-role"
	clusterRoleBindingName = "istio-sidecar-injector-cluster-role-binding"
	configMapName          = "istio-sidecar-injector"
	webhookName            = "istio-sidecar-injector"
	deploymentName         = "istio-sidecar-injector"
	serviceName            = "istio-sidecar-injector"
)

var sidecarInjectorLabels = map[string]string{
	"app": "istio-sidecar-injector",
}

var labelSelector = map[string]string{
	"istio": "sidecar-injector",
}

type Reconciler struct {
	resources.Reconciler

	includeIPRanges string
	excludeIPRanges string
}

func New(configuration Configuration, client client.Client, config *istiov1beta1.Config) *Reconciler {
	if configuration.ExcludeIPRanges == "" && configuration.IncludeIPRanges == "" {
		configuration.IncludeIPRanges = "*"
	}

	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		includeIPRanges: configuration.IncludeIPRanges,
		excludeIPRanges: configuration.ExcludeIPRanges,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	for _, res := range []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
		r.deployment,
		r.service,
		r.webhook,
	} {
		o := res()
		err := k8sutil.Reconcile(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	err := r.labelNamespaces(log)
	if err != nil {
		return emperror.WrapWith(err, "failed to label namespaces")
	}
	return nil
}
