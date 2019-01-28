package istio

import (
	"github.com/go-logr/logr"
	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *ReconcileIstio) ReconcileGalley(log logr.Logger, istio *istiov1alpha1.Istio) error {

	galleyResources := make(map[string]runtime.Object)

	galleySa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley-service-account",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-galley",
			},
		},
	}
	controllerutil.SetControllerReference(istio, galleySa, r.scheme)
	galleyResources[galleySa.Name] = galleySa

	galleyCr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-galley-cluster-role",
			Labels: map[string]string{
				"app": "istio-galley",
			},
		},
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
	controllerutil.SetControllerReference(istio, galleyCr, r.scheme)
	galleyResources[galleyCr.Name] = galleyCr

	galleyCrb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-galley-admin-role-binding",
			Labels: map[string]string{
				"app": "istio-galley",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "istio-galley-cluster-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "istio-galley-service-account",
				Namespace: istio.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(istio, galleyCrb, r.scheme)
	galleyResources[galleyCrb.Name] = galleyCrb

	for name, res := range galleyResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}
