package sidecarinjector

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) serviceAccount(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMeta(serviceAccountName, sidecarInjectorLabels, owner),
	}
}

func (r *Reconciler) clusterRole(owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMeta(clusterRoleName, sidecarInjectorLabels, owner),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"configmaps"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding(owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMeta(clusterRoleBindingName, sidecarInjectorLabels, owner),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     clusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: owner.Namespace,
			},
		},
	}
}
