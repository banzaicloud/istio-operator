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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/base"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) containerArgs() []string {
	containerArgs := []string{
		"--caCertFile=/etc/istio/certs/ca.crt",
		"--tlsCertFile=/etc/istio/certs/tls.crt",
		"--tlsKeyFile=/etc/istio/certs/tls.key",
		"--injectConfig=/etc/istio/inject/config",
		"--meshConfig=/etc/istio/config/mesh",
		"--healthCheckInterval=2s",
		"--healthCheckFile=/tmp/health",
		"--reconcileWebhookConfig=true",
		fmt.Sprintf("--webhookConfigName=%s", r.Config.WithNamespacedRevision(configMapNameInjector)),
	}

	if len(r.Config.Spec.SidecarInjector.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.SidecarInjector.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) containerEnvs() []apiv1.EnvVar {
	envs := make([]apiv1.EnvVar, 0)

	serviceHostnames := []string{
		fmt.Sprintf("%s.%s", r.Config.WithRevision(serviceName), r.Config.Namespace),
		fmt.Sprintf("%s.%s.svc", r.Config.WithRevision(serviceName), r.Config.Namespace),
		fmt.Sprintf("%s.%s.svc.%s", r.Config.WithRevision(serviceName), r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain),
	}

	envs = append(envs, []apiv1.EnvVar{
		{
			Name:  "REVISION",
			Value: r.Config.NamespacedRevision(),
		},
		{
			Name:  "CERT_DNS_NAMES",
			Value: strings.Join(serviceHostnames, ","),
		},
	}...)

	envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.SidecarInjector.AdditionalEnvVars)

	return envs
}

func (r *Reconciler) deployment() runtime.Object {
	deployment := &appsv1.Deployment{
		ObjectMeta: templates.ObjectMetaWithRevision(deploymentName, util.MergeStringMaps(sidecarInjectorLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: r.Config.Spec.SidecarInjector.ReplicaCount,
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeMultipleStringMaps(sidecarInjectorLabels, labelSelector, r.Config.RevisionLabels()),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeMultipleStringMaps(sidecarInjectorLabels, labelSelector, r.Config.RevisionLabels(), v1beta1.DisableInjectionLabel),
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.SidecarInjector.PodAnnotations),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: r.Config.WithRevision(serviceAccountName),
					Containers: []apiv1.Container{
						{
							Name:            "sidecar-injector-webhook",
							Image:           util.PointerToString(r.Config.Spec.SidecarInjector.Image),
							ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
							Args:            r.containerArgs(),
							SecurityContext: r.Config.Spec.SidecarInjector.SecurityContext,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/istio/config",
									ReadOnly:  true,
								},
								{
									Name:      "certs",
									MountPath: "/etc/istio/certs",
									ReadOnly:  false,
								},
								{
									Name:      "inject-config",
									MountPath: "/etc/istio/inject",
									ReadOnly:  true,
								},
							},
							ReadinessProbe: siProbe(),
							LivenessProbe:  siProbe(),
							Resources: templates.GetResourcesRequirementsOrDefault(
								r.Config.Spec.SidecarInjector.Resources,
								r.Config.Spec.DefaultResources,
							),
							Env:                      r.containerEnvs(),
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					Volumes:           r.volumes(),
					Affinity:          r.Config.Spec.SidecarInjector.Affinity,
					NodeSelector:      r.Config.Spec.SidecarInjector.NodeSelector,
					Tolerations:       r.Config.Spec.SidecarInjector.Tolerations,
					PriorityClassName: r.Config.Spec.PriorityClassName,
					SecurityContext:   util.GetPodSecurityContextFromSecurityContext(r.Config.Spec.SidecarInjector.SecurityContext),
					ImagePullSecrets:  r.Config.Spec.ImagePullSecrets,
				},
			},
		},
	}

	return deployment
}

func siProbe() *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/sidecar-injector",
					"probe",
					"--probe-path=/tmp/health",
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

func (r *Reconciler) volumes() []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			Name: "config-volume",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: r.Config.WithRevision(base.IstioConfigMapName),
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
						Name: r.Config.WithRevision(configMapNameInjector),
					},
					Items: []apiv1.KeyToPath{
						{
							Key:  "config",
							Path: "config",
						},
						{
							Key:  "values",
							Path: "values",
						},
					},
					DefaultMode: util.IntPointer(420),
				},
			},
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && util.PointerToBool(r.Config.Spec.Istiod.MultiClusterSupport) {
		volumes = append(volumes, []apiv1.Volume{
			{
				Name: "certs",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{
						Medium: apiv1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "istio-envoy",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{
						Medium: apiv1.StorageMediumMemory,
					},
				},
			},
		}...)
	} else {
		var secretPrefix string
		if len(r.Config.Spec.Certificates) != 0 {
			secretPrefix = "dns"
		} else {
			secretPrefix = "istio"
		}
		secretName := fmt.Sprintf("%s.%s", secretPrefix, serviceAccountName)
		volumes = append(volumes, apiv1.Volume{
			Name: "certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  r.Config.WithRevision(secretName),
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.Pilot.CertProvider == v1beta1.PilotCertProviderTypeIstiod {
		volumes = append(volumes, apiv1.Volume{
			Name: "istiod-ca-cert",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: r.Config.WithRevisionIf("istio-ca-root-cert", util.PointerToBool(r.Config.Spec.Istiod.MultiControlPlaneSupport)),
					},
				},
			},
		})
	}

	if r.Config.Spec.JWTPolicy == v1beta1.JWTPolicyThirdPartyJWT {
		volumes = append(volumes, apiv1.Volume{
			Name: "istio-token",
			VolumeSource: apiv1.VolumeSource{
				Projected: &apiv1.ProjectedVolumeSource{
					Sources: []apiv1.VolumeProjection{
						{
							ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
								Audience:          r.Config.Spec.SDS.TokenAudience,
								ExpirationSeconds: util.Int64Pointer(43200),
								Path:              "istio-token",
							},
						},
					},
				},
			},
		})
	}

	return volumes
}
