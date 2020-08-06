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

package base

import (
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) serviceAccountReader() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMetaWithRevision(istioReaderServiceAccountName, istioReaderLabel, r.Config),
	}
}

func (r *Reconciler) serviceAccount() runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMetaWithRevision(istiodServiceAccountName, istiodLabel, r.Config),
	}
}

func (r *Reconciler) role() runtime.Object {
	return &rbacv1.Role{
		ObjectMeta: templates.ObjectMetaWithRevision(istiodName, istiodLabel, r.Config),
		Rules: []rbacv1.PolicyRule{
			{
				// permissions to verify the webhook is ready and rejecting
				// invalid config. We use --server-dry-run so no config is persisted.
				APIGroups: []string{"networking.istio.io"},
				Resources: []string{"gateways"},
				Verbs:     []string{"create"},
			},
			{
				// For storing CA secret
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create", "get", "watch", "list", "update", "delete"},
			},
		},
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
		// auto-detect installed CRD definitions
		{
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
			Verbs:     []string{"get", "list", "watch"},
		},
		// discovery and routing
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "nodes", "services", "namespaces", "endpoints"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"discovery.k8s.io"},
			Resources: []string{"endpointslices"},
			Verbs:     []string{"get", "list", "watch"},
		},
		// ingress controller
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
		// Use for Kubernetes Service APIs
		{
			APIGroups: []string{"networking.x-k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"get", "watch", "list"},
		},
		// Needed for multicluster secret reading, possibly ingress certs in the future
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "watch", "list"},
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.EnableAnalysis) {
		rules = append(rules, []rbacv1.PolicyRule{
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
		}...)
	}

	verbs := []string{"get", "watch", "list"}
	if util.PointerToBool(r.Config.Spec.Istiod.EnableAnalysis) {
		verbs = append(verbs, "update")
	}
	rules = append(rules, rbacv1.PolicyRule{
		// istio configuration
		APIGroups: []string{"config.istio.io", "security.istio.io", "networking.istio.io", "authentication.istio.io"},
		Resources: []string{"*"},
		Verbs:     verbs,
	})

	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(istiodName, istiodLabel, r.Config),
		Rules:      rules,
	}
}

func (r *Reconciler) clusterRoleReader() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(istioReaderName, istioReaderLabel, r.Config),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"config.istio.io", "security.istio.io", "networking.istio.io", "authentication.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints", "pods", "services", "nodes", "replicationcontrollers", "namespaces"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"replicasets"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBindingReader() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(istioReaderName, istioReaderLabel, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(istioReaderName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(istioReaderServiceAccountName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMetaClusterScopeWithRevision(istiodPilotName, pilotLabel, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(istiodName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(istiodName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) roleBinding() runtime.Object {
	return &rbacv1.RoleBinding{
		ObjectMeta: templates.ObjectMetaWithRevision(istiodName, pilotLabel, r.Config),
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.Config.WithNamespacedRevision(istiodName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.WithRevision(istiodName),
				Namespace: r.Config.Namespace,
			},
		},
	}
}
