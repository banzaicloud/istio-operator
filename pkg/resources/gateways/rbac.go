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

package gateways

import (
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) serviceAccount() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMeta(r.serviceAccountName(), r.labels(), r.gw),
	}
}

func (r *Reconciler) clusterRole() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScope(r.clusterRoleName(), r.labelSelector(), r.gw),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"networking.istio.io"},
				Resources: []string{"virtualservices", "destinationrules", "gateways"},
				Verbs:     []string{"get", "watch", "list", "update"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScope(r.clusterRoleBindingName(), r.labelSelector(), r.gw),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.clusterRoleName(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.serviceAccountName(),
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) role() runtime.Object {
	return &rbacv1.Role{
		ObjectMeta: templates.ObjectMeta(r.roleName(), r.labelSelector(), r.gw),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
}

func (r *Reconciler) roleBinding() runtime.Object {
	return &rbacv1.RoleBinding{
		ObjectMeta: templates.ObjectMeta(r.roleBindingName(), r.labelSelector(), r.gw),
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.roleName(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.serviceAccountName(),
				Namespace: r.Config.Namespace,
			},
		},
	}
}
