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

package galley

import (
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) serviceAccount() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMetaWithRevision(serviceAccountName, galleyLabels, r.Config),
	}
}

func (r *Reconciler) rules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		{
			// For reading Istio resources
			APIGroups: []string{"authentication.istio.io", "config.istio.io", "networking.istio.io", "rbac.istio.io", "security.istio.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			// For updating Istio resource statuses
			APIGroups: []string{"authentication.istio.io", "config.istio.io", "networking.istio.io", "rbac.istio.io", "security.istio.io"},
			Resources: []string{"*/status"},
			Verbs:     []string{"update"},
		},
		{
			APIGroups: []string{"extensions", "apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "nodes", "services", "endpoints", "namespaces"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"extensions"},
			Resources: []string{"ingresses"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{"extensions"},
			Resources:     []string{"deployments/finalizers"},
			ResourceNames: []string{"istio-galley"},
			Verbs:         []string{"update"},
		},
		{
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterroles"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}

	if !util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		rules = append(rules, []rbacv1.PolicyRule{
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"validatingwebhookconfigurations"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"networking.istio.io"},
				Resources: []string{"gateways"},
				Verbs:     []string{"create"},
			},
		}...)
	}

	return rules
}

func (r *Reconciler) clusterRole() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleName, galleyLabels, r.Config),
		Rules:      r.rules(),
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleBindingName, galleyLabels, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     util.CombinedName(clusterRoleName, r.Config.Name, r.Config.Namespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      util.CombinedName(serviceAccountName, r.Config.Name),
				Namespace: r.Config.Namespace,
			},
		},
	}
}
