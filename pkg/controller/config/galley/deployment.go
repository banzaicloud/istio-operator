package galley

import (
	"fmt"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) deployment(owner *istiov1beta1.Config) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(galleyLabels, labelSelector), owner),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       util.IntstrPointer(1),
					MaxUnavailable: util.IntstrPointer(0),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labelSelector,
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
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
								fmt.Sprintf("--deployment-namespace=%s", owner.Namespace),
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
							LivenessProbe:  r.galleyProbe(),
							ReadinessProbe: r.galleyProbe(),
							Resources:      templates.DefaultResources(),
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio.%s", serviceAccountName),
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
}

func (r *Reconciler) galleyProbe() *apiv1.Probe {
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
