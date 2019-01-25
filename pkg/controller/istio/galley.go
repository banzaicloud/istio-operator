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
)

func (r *ReconcileIstio) ReconcileGalley(log logr.Logger, istio *istiov1alpha1.Istio) error {
	ns := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: Namespace,
			Labels: map[string]string{
				"name": "istio-system",
			},
		},
	}
	controllerutil.SetControllerReference(istio, ns, r.scheme)
	err := k8sutil.ReconcileResource(log, r.client, "", ns.Name, ns)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile istio namespace", "namespace", ns.Name)
	}

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
	err = k8sutil.ReconcileResource(log, r.client, "", galleyCr.Name, galleyCr)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile cluster role", "clusterRole", galleyCr.Name)
	}

	galleySa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley-service-account",
			Namespace: Namespace,
			Labels: map[string]string{
				"app": "istio-galley",
			},
		},
	}
	controllerutil.SetControllerReference(istio, galleySa, r.scheme)
	err = k8sutil.ReconcileResource(log, r.client, Namespace, galleySa.Name, galleySa)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile service account", "serviceAccount", galleySa.Name)
	}
	return nil
}
