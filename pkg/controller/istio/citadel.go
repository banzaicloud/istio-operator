package istio

import (
	"fmt"

	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *ReconcileIstio) ReconcileCitadel(log logr.Logger, istio *istiov1alpha1.Istio) error {

	citadelResources := make(map[string]runtime.Object)

	citadelSa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-citadel-service-account",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "security",
			},
		},
	}
	controllerutil.SetControllerReference(istio, citadelSa, r.scheme)
	citadelResources[citadelSa.Name] = citadelSa

	citadelCr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-citadel-cluster-role",
			Labels: map[string]string{
				"app": "security",
			},
		},
		Rules: []rbacv1.PolicyRule{
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
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
	controllerutil.SetControllerReference(istio, citadelCr, r.scheme)
	citadelResources[citadelCr.Name] = citadelCr

	citadelCrb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "istio-citadel-cluster-role-binding",
			Labels: map[string]string{
				"app": "security",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "istio-citadel-cluster-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "istio-citadel-service-account",
				Namespace: istio.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(istio, citadelCrb, r.scheme)
	citadelResources[citadelCrb.Name] = citadelCrb

	citadelDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-citadel-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "security",
				"istio": "citadel",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio": "citadel",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio": "citadel",
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject":                    "false",
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: "istio-citadel-service-account",
					Containers: []apiv1.Container{
						{
							Name:            "citadel",
							Image:           "docker.io/istio/citadel:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"--append-dns-names=true",
								"--grpc-port=8060",
								"--grpc-hostname=citadel",
								fmt.Sprintf("--citadel-storage-namespace=%s", istio.Namespace),
								fmt.Sprintf("--custom-dns-names=istio-pilot-service-account.%[1]s:istio-pilot.%[1]s,istio-ingressgateway-service-account.%[1]s:istio-ingressgateway.%[1]s", istio.Namespace),
								"--self-signed-ca=true",
							},
							Resources: defaultResources(),
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, citadelDeploy, r.scheme)
	citadelResources[citadelDeploy.Name] = citadelDeploy

	citadelSvc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-citadel",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app": "istio-citadel",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "grpc-citadel",
					Port:       8060,
					TargetPort: intstr.FromInt(8060),
					Protocol:   apiv1.ProtocolTCP,
				},
				{
					Name: "http-monitoring",
					Port: 9093,
				},
			},
			Selector: map[string]string{
				"istio": "citadel",
			},
		},
	}
	controllerutil.SetControllerReference(istio, citadelSvc, r.scheme)
	citadelResources[citadelSvc.Name] = citadelSvc

	for name, res := range citadelResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}
