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
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) serviceAccount() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMetaWithRevision(serviceAccountName, defaultLabels, r.Config),
	}
}

func (r *Reconciler) clusterRole() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleName, defaultLabels, r.Config),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "nodes", "namespaces"},
				Verbs:     []string{"get"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleRepair() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleRepairName, defaultLabels, r.Config),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch", "delete", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list", "watch", "delete", "patch", "update", "create"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleTaint() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleTaintName, defaultLabels, r.Config),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "list", "create", "update"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleBindingName, defaultLabels, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(clusterRoleName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(serviceAccountName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) clusterRoleBindingRepair() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleBindingRepairName, defaultLabels, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(clusterRoleRepairName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(serviceAccountName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) clusterRoleBindingTaint() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleBindingTaintName, defaultLabels, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(clusterRoleTaintName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(serviceAccountName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}
