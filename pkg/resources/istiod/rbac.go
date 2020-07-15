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

package istiod

import (
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) serviceAccount() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMetaWithRevision(serviceAccountName, istiodLabels, r.Config),
	}
}

func (r *Reconciler) clusterRole() runtime.Object {
	rules := []rbacv1.PolicyRule{
		// sidecar injection controller
		{
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"mutatingwebhookconfigurations"},
			Verbs:     []string{"get", "list", "watch", "patch"},
		},
		// configuration validation webhook controller
		{
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"validatingwebhookconfigurations"},
			Verbs:     []string{"get", "list", "watch", "update"},
		},
		// permissions to verify the webhook is ready and rejecting
		// invalid config. We use --server-dry-run so no config is persisted.
		{
			APIGroups: []string{"networking.istio.io"},
			Resources: []string{"gateways"},
			Verbs:     []string{"create"},
		},
		// istio configuration
		{
			APIGroups: []string{"config.istio.io", "rbac.istio.io", "security.istio.io", "networking.istio.io", "authentication.istio.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "watch", "list", "update"},
		},
		// auto-detect installed CRD definitions
		{
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
			Verbs:     []string{"get", "watch", "list"},
		},
		// discovery and routing
		{
			APIGroups: []string{"extensions", "apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"endpoints", "pods", "services", "namespaces", "nodes"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"discovery.k8s.io"},
			Resources: []string{"endpointslices"},
			Verbs:     []string{"get", "list", "watch"},
		},
		// ingress controller
		{
			APIGroups: []string{"extensions", "networking.k8s.io"},
			Resources: []string{"ingresses"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"extensions", "networking.k8s.io"},
			Resources: []string{"ingresses/status"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses", "ingressclasses"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses/status"},
			Verbs:     []string{"*"},
		},
		// required for CA's namespace controller
		{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"create", "get", "list", "watch", "update"},
		},
		// Istiod and bootstrap
		{
			APIGroups: []string{"certificates.k8s.io"},
			Resources: []string{"certificatesigningrequests", "certificatesigningrequests/approval", "certificatesigningrequests/status"},
			Verbs:     []string{"update", "create", "get", "delete", "watch"},
		},
		{
			APIGroups:     []string{"certificates.k8s.io"},
			Resources:     []string{"signers"},
			ResourceNames: []string{"kubernetes.io/legacy-unknown"},
			Verbs:         []string{"approve"},
		},
		// Used by Istiod to verify the JWT tokens
		{
			APIGroups: []string{"authentication.k8s.io"},
			Resources: []string{"tokenreviews"},
			Verbs:     []string{"create"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"create", "get", "watch", "list", "update", "delete"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"serviceaccounts"},
			Verbs:     []string{"get", "watch", "list"},
		},
		// Use for Kubernetes Service APIs
		{
			APIGroups: []string{"networking.x.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "watch", "list"},
		},
	}

	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleNameIstiod, istiodLabels, r.Config),
		Rules:      rules,
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(clusterRoleBindingNameIstiod, istiodLabels, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(clusterRoleNameIstiod),
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
