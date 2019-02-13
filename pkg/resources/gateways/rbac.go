package gateways

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) serviceAccount(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ServiceAccount{
		ObjectMeta: templates.ObjectMeta(serviceAccountName(gw), gwLabels(gw), owner),
	}
}

func (r *Reconciler) clusterRole(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: templates.ObjectMeta(clusterRoleName(gw), gwLabels(gw), owner),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"thirdpartyresources", "virtualservices", "destinationrules", "gateways"},
				Verbs:     []string{"get", "watch", "list", "update"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: templates.ObjectMeta(clusterRoleBindingName(gw), gwLabels(gw), owner),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     clusterRoleName(gw),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName(gw),
				Namespace: owner.Namespace,
			},
		},
	}
}
