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

package galley

import (
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(galleyLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: &r.Config.Spec.Galley.ReplicaCount,
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
							Image:           r.Config.Spec.Galley.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 443,
									Protocol:      apiv1.ProtocolTCP,
								},
								{
									ContainerPort: 9093,
									Protocol:      apiv1.ProtocolTCP,
								},
							},
							Command: []string{
								"/usr/local/bin/galley",
								"validator",
								fmt.Sprintf("--deployment-namespace=%s", r.Config.Namespace),
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
							LivenessProbe:            r.galleyProbe(),
							ReadinessProbe:           r.galleyProbe(),
							Resources:                templates.DefaultResources(),
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName:  fmt.Sprintf("istio.%s", serviceAccountName),
									DefaultMode: util.IntPointer(420),
								},
							},
						},
						{
							Name: "config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: configMapName,
									},
									DefaultMode: util.IntPointer(420),
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
		FailureThreshold:    3,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}
}
