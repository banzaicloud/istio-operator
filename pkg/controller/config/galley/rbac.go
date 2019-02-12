package galley

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func serviceAccount(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMeta(serviceAccountName, galleyLabels, owner),
	}
}

func clusterRole(owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMeta(clusterRoleName, galleyLabels, owner),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"validatingwebhookconfigurations"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"config.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups:     []string{"*"},
				Resources:     []string{"deployments"},
				ResourceNames: []string{"istio-galley"},
				Verbs:         []string{"get"},
			},
			{
				APIGroups:     []string{"*"},
				Resources:     []string{"endpoints"},
				ResourceNames: []string{"istio-galley"},
				Verbs:         []string{"get"},
			},
		},
	}
}

func clusterRoleBinding(owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMeta(clusterRoleBindingName, galleyLabels, owner),
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