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

package sidecarinjector

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(sidecarInjectorLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
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
							Name:            "sidecar-injector-webhook",
							Image:           "docker.io/istio/sidecar_injector:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"--caCertFile=/etc/istio/certs/root-cert.pem",
								"--tlsCertFile=/etc/istio/certs/cert-chain.pem",
								"--tlsKeyFile=/etc/istio/certs/key.pem",
								"--injectConfig=/etc/istio/inject/config",
								"--meshConfig=/etc/istio/config/mesh",
								"--healthCheckInterval=2s",
								"--healthCheckFile=/health",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/istio/config",
									ReadOnly:  true,
								},
								{
									Name:      "certs",
									MountPath: "/etc/istio/certs",
									ReadOnly:  true,
								},
								{
									Name:      "inject-config",
									MountPath: "/etc/istio/inject",
									ReadOnly:  true,
								},
							},
							ReadinessProbe:           siProbe(),
							LivenessProbe:            siProbe(),
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
							Name: "config-volume",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: common.IstioConfigMapName,
									},
									DefaultMode: util.IntPointer(420),
								},
							},
						},
						{
							Name: "inject-config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  "config",
											Path: "config",
										},
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

func siProbe() *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/sidecar-injector",
					"probe",
					"--probe-path=/health",
					"--interval=4s",
				},
			},
		},
		InitialDelaySeconds: 4,
		PeriodSeconds:       4,
		FailureThreshold:    3,
		TimeoutSeconds:      1,
		SuccessThreshold:    1,
	}
}
