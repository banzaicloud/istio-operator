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
		"--caCertFile=/etc/istio/certs/root-cert.pem",
		"--tlsCertFile=/etc/istio/certs/cert-chain.pem",
		"--tlsKeyFile=/etc/istio/certs/key.pem",
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

	envs = append(envs, apiv1.EnvVar{
		Name:  "REVISION",
		Value: r.Config.NamespacedRevision(),
	})

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
					Labels:      util.MergeMultipleStringMaps(sidecarInjectorLabels, labelSelector, r.Config.RevisionLabels()),
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
									ReadOnly:  true,
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
				},
			},
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && util.PointerToBool(r.Config.Spec.Istiod.MultiClusterSupport) {
		deployment.Spec.Template.Spec.InitContainers = []apiv1.Container{
			r.certFetcherContainer(),
		}
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

func (r *Reconciler) certFetcherContainer() apiv1.Container {
	args := []string{
		"proxy",
		"sidecar",
		"--serviceCluster",
		"istio-si-cert-fetcher",
		"--controlPlaneAuthPolicy",
		string(r.Config.GetControlPlaneAuthPolicy()),
		"--domain",
		r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
		"--discoveryAddress", r.Config.GetDiscoveryAddress(),
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		args = append(args, []string{
			"--envoyAccessLogService",
			r.Config.Spec.Proxy.EnvoyAccessLogService.GetDataJSON(),
		}...)
	}

	return apiv1.Container{
		Name:                     "cert-fetcher",
		Image:                    r.Config.Spec.Proxy.Image,
		ImagePullPolicy:          r.Config.Spec.ImagePullPolicy,
		Args:                     args,
		Env:                      append(templates.IstioProxyEnv(r.Config), r.cfEnvVars()...),
		VolumeMounts:             r.cfVolumeMounts(),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}

func (r *Reconciler) cfEnvVars() []apiv1.EnvVar {
	serviceHostnames := []string{
		fmt.Sprintf("%s.%s", r.Config.WithRevision(serviceName), r.Config.Namespace),
		fmt.Sprintf("%s.%s.svc", r.Config.WithRevision(serviceName), r.Config.Namespace),
		fmt.Sprintf("%s.%s.svc.%s", r.Config.WithRevision(serviceName), r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain),
	}
	envVars := []apiv1.EnvVar{
		{
			Name:  "CERT_CUSTOM_DNS_NAMES",
			Value: strings.Join(serviceHostnames, ","),
		},
		{
			Name:  "SECRET_TTL",
			Value: "8640h",
		},
		{
			Name:  "OUTPUT_CERTS",
			Value: "/etc/certs",
		},
		{
			Name:  "CERT_GENERATION_ONLY",
			Value: "true",
		},
	}

	envVars = append(envVars, apiv1.EnvVar{
		Name:  "CA_ADDR",
		Value: r.Config.GetCAAddress(),
	})

	if r.Config.Spec.ClusterName != "" {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_CLUSTER_ID",
			Value: r.Config.Spec.ClusterName,
		})
	} else {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_CLUSTER_ID",
			Value: "Kubernetes",
		})
	}

	return envVars
}

func (r *Reconciler) cfVolumeMounts() []apiv1.VolumeMount {
	vms := []apiv1.VolumeMount{}

	vms = append(vms, []apiv1.VolumeMount{
		{
			Name:      "certs",
			MountPath: "/etc/certs",
		},
		{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
		},
		{
			Name:      "istiod-ca-cert",
			MountPath: "/var/run/secrets/istio",
		},
		{
			Name:      "istio-envoy",
			MountPath: "/etc/istio/proxy",
		},
	}...)

	if r.Config.Spec.JWTPolicy == v1beta1.JWTPolicyThirdPartyJWT {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "istio-token",
			MountPath: "/var/run/secrets/tokens",
			ReadOnly:  true,
		})
	}

	return vms
}
