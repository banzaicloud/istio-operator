package istio

import (
	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	yamlv2 "gopkg.in/yaml.v2"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/util/intstr"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
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

	webhookConfig, err := validatingWebhookConfig(istio.Namespace)
	if err != nil {
		return emperror.With(err)
	}
	galleyCm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley-configuration",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio-galley",
				"istio": "mixer",
			},
		},
		Data: map[string]string{
			"validatingwebhookconfiguration.yaml": webhookConfig,
		},
	}
	controllerutil.SetControllerReference(istio, galleyCm, r.scheme)
	galleyResources[galleyCm.Name] = galleyCm

	galleyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley-deployment",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"app":   "istio-galley",
				"istio": "galley",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: intPointer(1),
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       intstrPointer(1),
					MaxUnavailable: intstrPointer(0),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"istio": "galley",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio": "galley",
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject":                    "false",
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: "istio-galley-service-account",
					PriorityClassName:  "",
					Containers: []apiv1.Container{
						{
							Name:            "validator",
							Image:           "docker.io/istio/galley:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 443,
								},
								{
									ContainerPort: 9093,
								},
							},
							Command: []string{
								"/usr/local/bin/galley",
								"validator",
								fmt.Sprintf("--deployment-namespace=%s", istio.Namespace),
								"--caCertFile=/etc/istio/certs/root-cert.pem",
								"--tlsCertFile=/etc/istio/certs/cert-chain.pem",
								"--tlsKeyFile=/etc/istio/certs/key.pem",
								"--healthCheckInterval=1s",
								"--healthCheckFile=/health",
								"--webhook-config-file",
								"/etc/istio/config/validatingwebhookconfiguration.yaml",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "certs",
									MountPath: "/etc/istio/certs",
									ReadOnly:  true,
								},
								{
									Name:      "config",
									MountPath: "/etc/istio/config",
									ReadOnly:  true,
								},
							},
							LivenessProbe:  galleyProbe(),
							ReadinessProbe: galleyProbe(),
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("10m"),
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: "istio.istio-galley-service-account",
								},
							},
						},
						{
							Name: "config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "istio-galley-configuration",
									},
								},
							},
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
	controllerutil.SetControllerReference(istio, galleyDeploy, r.scheme)
	galleyResources[galleyDeploy.Name] = galleyDeploy

	galleySvc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley",
			Namespace: istio.Namespace,
			Labels: map[string]string{
				"istio": "galley",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "https-validation",
					Port: 443,
				},
				{
					Name: "https-monitoring",
					Port: 9093,
				},
			},
			Selector: map[string]string{
				"istio": "galley",
			},
		},
	}
	controllerutil.SetControllerReference(istio, galleySvc, r.scheme)
	galleyResources[galleySvc.Name] = galleySvc

	for name, res := range galleyResources {
		err := k8sutil.ReconcileResource(log, r.client, istio.Namespace, name, res)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", res.GetObjectKind().GroupVersionKind().Kind, "name", name)
		}
	}

	return nil
}

func validatingWebhookConfig(ns string) (string, error) {
	fail := admissionv1beta1.Fail
	webhook := admissionv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-galley",
			Namespace: ns,
			Labels: map[string]string{
				"app": "istio-galley",
			},
		},
		Webhooks: []admissionv1beta1.Webhook{
			{
				Name: "pilot.validation.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      "istio-galley",
						Namespace: ns,
						Path:      strPointer("/admitpilot"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources:   []string{"httpapispecs", "httpapispecbindings", "quotaspecs", "quotaspecbindings"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"rbac.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"authentication.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"networking.istio.io"},
							APIVersions: []string{"*"},
							Resources:   []string{"destinationrules", "envoyfilters", "gateways", "serviceentries", "virtualservices"},
						},
					},
				},
				FailurePolicy: &fail,
			},
			{
				Name: "mixer.validation.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      "istio-galley",
						Namespace: ns,
						Path:      strPointer("/admitmixer"),
					},
					CABundle: []byte{},
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
							admissionv1beta1.Update,
						},
						Rule: admissionv1beta1.Rule{
							APIGroups:   []string{"config.istio.io"},
							APIVersions: []string{"v1alpha2"},
							Resources: []string{
								"rules",
								"attributemanifests",
								"circonuses",
								"deniers",
								"fluentds",
								"kubernetesenvs",
								"listcheckers",
								"memquotas",
								"noops",
								"opas",
								"prometheuses",
								"rbacs",
								"servicecontrols",
								"solarwindses",
								"stackdrivers",
								"statsds",
								"stdios",
								"apikeys",
								"authorizations",
								"checknothings",
								"listentries",
								"logentries",
								"metrics",
								"quotas",
								"reportnothings",
								"servicecontrolreports",
								"tracespans"},
						},
					},
				},
				FailurePolicy: &fail,
			},
		},
	}
	marshaledConfig, err := yamlv2.Marshal(webhook)
	if err != nil {
		return "", emperror.Wrap(err, "failed to marshal webhook config")
	}
	return string(marshaledConfig), nil
}

func galleyProbe() *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/galley",
					"probe",
					"--probe-path=/health",
					"--interval=10s",
				},
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
	}
}

func strPointer(s string) *string {
	return &s
}

func intPointer(i int32) *int32 {
	return &i
}

func intstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}
