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

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/sidecarinjector"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeStringMaps(galleyLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: r.Config.Spec.Galley.ReplicaCount,
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       util.IntstrPointer(1),
					MaxUnavailable: util.IntstrPointer(0),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeStringMaps(galleyLabels, labelSelector),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeStringMaps(galleyLabels, labelSelector),
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.Galley.PodAnnotations),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []apiv1.Container{
						{
							Name:            "galley",
							Image:           util.PointerToString(r.Config.Spec.Galley.Image),
							ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 9443,
									Protocol:      apiv1.ProtocolTCP,
								},
								{
									ContainerPort: 15014,
									Protocol:      apiv1.ProtocolTCP,
								},
								{
									ContainerPort: 9901,
									Protocol:      apiv1.ProtocolTCP,
								},
							},
							Command: []string{"/usr/local/bin/galley"},
							Args:    r.containerArgs(),
							SecurityContext: &apiv1.SecurityContext{
								RunAsUser:    util.Int64Pointer(1337),
								RunAsGroup:   util.Int64Pointer(1337),
								RunAsNonRoot: util.BoolPointer(true),
								Capabilities: &apiv1.Capabilities{
									Drop: []apiv1.Capability{
										"ALL",
									},
								},
							},
							VolumeMounts:   r.volumeMounts(),
							LivenessProbe:  r.galleyProbe("/tmp/healthliveness"),
							ReadinessProbe: r.galleyProbe("/tmp/healthready"),
							Resources: templates.GetResourcesRequirementsOrDefault(
								r.Config.Spec.Galley.Resources,
								r.Config.Spec.DefaultResources,
							),
							Env:                      r.containerEnvs(),
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					SecurityContext: &apiv1.PodSecurityContext{
						FSGroup: util.Int64Pointer(1337),
					},
					Volumes:      r.volumes(),
					Affinity:     r.Config.Spec.Galley.Affinity,
					NodeSelector: r.Config.Spec.Galley.NodeSelector,
					Tolerations:  r.Config.Spec.Galley.Tolerations,
				},
			},
		},
	}
}

func (r *Reconciler) containerEnvs() []apiv1.EnvVar {
	envs := make([]apiv1.EnvVar, 0)

	envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Galley.AdditionalEnvVars)

	return envs
}

func (r *Reconciler) containerArgs() []string {
	containerArgs := []string{
		"server",
		"--meshConfigFile=/etc/mesh-config/mesh",
		"--livenessProbeInterval=1s",
		"--livenessProbePath=/tmp/healthliveness",
		"--readinessProbePath=/tmp/healthready",
		"--readinessProbeInterval=1s",
		fmt.Sprintf("--deployment-namespace=%s", r.Config.Namespace),
		"--monitoringPort=15014",
		"--enable-reconcileWebhookConfiguration=true",
	}

	if r.Config.Spec.Logging.Level != nil {
		containerArgs = append(containerArgs, fmt.Sprintf("--log_output_level=%s", util.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if util.PointerToBool(r.Config.Spec.Galley.ConfigValidation) && !util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		containerArgs = append(containerArgs, "--enable-validation=true")
	} else {
		containerArgs = append(containerArgs, "--enable-validation=false")
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled {
		containerArgs = append(containerArgs, "--insecure=false")
	} else {
		containerArgs = append(containerArgs, "--insecure=true")
	}

	if util.PointerToBool(r.Config.Spec.Galley.EnableServiceDiscovery) {
		containerArgs = append(containerArgs, "--enableServiceDiscovery=true")
	}

	if !util.PointerToBool(r.Config.Spec.UseMCP) {
		containerArgs = append(containerArgs, "--enable-server=false")
	}

	if util.PointerToBool(r.Config.Spec.Galley.EnableAnalysis) {
		containerArgs = append(containerArgs, "--enableAnalysis=true")
	}

	if len(r.Config.Spec.Certificates) != 0 {
		containerArgs = append(containerArgs, "--validation.tls.clientCertificate=/etc/dnscerts/cert-chain.pem")
		containerArgs = append(containerArgs, "--validation.tls.privateKey=/etc/dnscerts/key.pem")
		containerArgs = append(containerArgs, "--validation.tls.caCertificates=/etc/dnscerts/root-cert.pem")
	}

	if len(r.Config.Spec.Galley.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.Galley.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) volumeMounts() []apiv1.VolumeMount {
	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/config",
			ReadOnly:  true,
		},
		{
			Name:      "mesh-config",
			MountPath: "/etc/mesh-config",
			ReadOnly:  true,
		},
	}

	if util.PointerToBool(r.Config.Spec.Galley.ConfigValidation) && !util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		volumeMounts = append(volumeMounts, apiv1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	if len(r.Config.Spec.Certificates) != 0 {
		volumeMounts = append(volumeMounts, apiv1.VolumeMount{
			Name:      "dnscerts",
			MountPath: "/etc/dnscerts",
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func (r *Reconciler) galleyProbe(path string) *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/galley",
					"probe",
					fmt.Sprintf("--probe-path=%s", path),
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

func (r *Reconciler) volumes() []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			// galley expects /etc/config to exist even though it doesn't include any files.
			Name: "config",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{
					Medium: apiv1.StorageMediumMemory,
				},
			},
		},
		{
			Name: "mesh-config",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: sidecarinjector.IstioConfigMapName,
					},
					DefaultMode: util.IntPointer(420),
				},
			},
		},
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled || (util.PointerToBool(r.Config.Spec.Galley.ConfigValidation) && !util.PointerToBool(r.Config.Spec.Istiod.Enabled)) {
		volumes = append(volumes, apiv1.Volume{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio.%s", serviceAccountName),
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	if len(r.Config.Spec.Certificates) != 0 {
		volumes = append(volumes, apiv1.Volume{
			Name: "dnscerts",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("dns.%s", serviceAccountName),
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled {
		volumes = append(volumes, apiv1.Volume{
			Name: "envoy-config",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "galley-envoy-config",
					},
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	return volumes
}
