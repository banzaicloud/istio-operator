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

package istiod

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/base"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) containerArgs() []string {
	containerArgs := []string{
		"discovery",
		"--monitoringAddr=:15014",
		"--domain",
		r.Config.Spec.Proxy.ClusterDomain,
		"--keepaliveMaxServerConnectionAge",
		"30m",
		"--trust-domain",
		r.Config.Spec.TrustDomain,
	}

	if r.Config.Spec.Logging.Level != nil {
		containerArgs = append(containerArgs, fmt.Sprintf("--log_output_level=%s", util.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if r.Config.Spec.WatchOneNamespace {
		containerArgs = append(containerArgs, "-a", r.Config.Namespace)
	}

	if len(r.Config.Spec.Pilot.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.Pilot.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) containerEnvs() []apiv1.EnvVar {
	envs := []apiv1.EnvVar{
		{
			Name:  "PILOT_PUSH_THROTTLE",
			Value: "100",
		},
		{
			Name:  "PILOT_TRACE_SAMPLING",
			Value: fmt.Sprintf("%.2f", r.Config.Spec.Pilot.TraceSampling),
		},
		{
			Name:  "MESHNETWORKS_HASH",
			Value: r.Config.Spec.GetMeshNetworksHash(),
		},
		{
			Name:  "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_OUTBOUND",
			Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Pilot.EnableProtocolSniffingOutbound)),
		},
		{
			Name:  "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_INBOUND",
			Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Pilot.EnableProtocolSniffingInbound)),
		},
		{
			Name:  "INJECTION_WEBHOOK_CONFIG_NAME",
			Value: r.Config.WithNamespacedName("istio-sidecar-injector"),
		},
		{
			Name:  "CENTRAL_ISTIOD",
			Value: "true",
		},
		{
			Name:  "REVISION",
			Value: r.Config.Name,
		},
		{
			Name:  "ISTIOD_CUSTOM_HOST",
			Value: fmt.Sprintf("%s.%s.svc", r.Config.WithName(ServiceNameIstiod), r.Config.Namespace),
		},
	}

	envs = append(envs, templates.IstioProxyEnv(r.Config)...)

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envs = append(envs, apiv1.EnvVar{
			Name:  "ISTIOD_ADDR",
			Value: r.Config.GetDiscoveryAddress(),
		})
	}

	if r.Config.Spec.LocalityLB != nil && util.PointerToBool(r.Config.Spec.LocalityLB.Enabled) {
		envs = append(envs, apiv1.EnvVar{
			Name:  "PILOT_ENABLE_LOCALITY_LOAD_BALANCING",
			Value: "1",
		})
	}

	envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Pilot.AdditionalEnvVars)

	return envs
}

func (r *Reconciler) containerPorts() []apiv1.ContainerPort {
	return []apiv1.ContainerPort{
		{ContainerPort: 8080, Protocol: apiv1.ProtocolTCP},
		{ContainerPort: 15010, Protocol: apiv1.ProtocolTCP},
		{ContainerPort: 15017, Protocol: apiv1.ProtocolTCP},
		{ContainerPort: 15053, Protocol: apiv1.ProtocolTCP},
	}
}

func (r *Reconciler) containers() []apiv1.Container {
	discoveryContainer := apiv1.Container{
		Name:            "discovery",
		Image:           util.PointerToString(r.Config.Spec.Pilot.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args:            r.containerArgs(),
		Ports:           r.containerPorts(),
		ReadinessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path:   "/ready",
					Port:   intstr.FromInt(8080),
					Scheme: apiv1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			TimeoutSeconds:      5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
		Env: r.containerEnvs(),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Pilot.Resources,
			r.Config.Spec.DefaultResources,
		),
		SecurityContext:          r.Config.Spec.Pilot.SecurityContext,
		VolumeMounts:             r.volumeMounts(),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}

	containers := []apiv1.Container{
		discoveryContainer,
	}

	return containers
}

func (r *Reconciler) volumeMounts() []apiv1.VolumeMount {
	vms := []apiv1.VolumeMount{
		{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		if r.Config.Spec.JWTPolicy == v1beta1.JWTPolicyThirdPartyJWT {
			vms = append(vms, apiv1.VolumeMount{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
				ReadOnly:  true,
			})
		}

		if r.Config.Spec.Citadel.CASecretName != "" {
			vms = append(vms, apiv1.VolumeMount{
				Name:      "cacerts",
				MountPath: "/etc/cacerts",
				ReadOnly:  true,
			})
		}

		vms = append(vms, []apiv1.VolumeMount{
			{
				Name:      "local-certs",
				MountPath: "/var/run/secrets/istio-dns",
			},
			{
				Name:      "inject",
				MountPath: "/var/lib/istio/inject",
				ReadOnly:  true,
			},
		}...)
	}

	return vms
}

func (r *Reconciler) volumes() []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			Name: "config-volume",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: r.Config.WithName(base.IstioConfigMapName),
					},
					DefaultMode: util.IntPointer(420),
				},
			},
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		volumes = append(volumes, apiv1.Volume{
			Name: "local-certs",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{
					Medium: apiv1.StorageMediumMemory,
				},
			},
		})

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

		if r.Config.Spec.Citadel.CASecretName != "" {
			volumes = append(volumes, apiv1.Volume{
				Name: "cacerts",
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName:  r.Config.Spec.Citadel.CASecretName,
						Optional:    util.BoolPointer(false),
						DefaultMode: util.IntPointer(420),
					},
				},
			})
		}

		volumes = append(volumes, apiv1.Volume{
			Name: "inject",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: r.Config.WithName("istio-sidecar-injector"),
					},
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	return volumes
}

func (r *Reconciler) deployment() runtime.Object {
	deployment := &appsv1.Deployment{
		ObjectMeta: templates.ObjectMetaWithRevision(deploymentName, util.MergeStringMaps(istiodLabels, pilotLabelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName,
				Namespace: r.Config.Namespace,
			}, util.PointerToInt32(r.Config.Spec.Pilot.ReplicaCount))),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeStringMaps(pilotLabelSelector, r.Config.RevisionLabels()),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeMultipleStringMaps(istiodLabels, pilotLabelSelector, r.Config.RevisionLabels()),
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.Pilot.PodAnnotations),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: r.Config.WithName(serviceAccountName),
					SecurityContext:    util.GetPodSecurityContextFromSecurityContext(r.Config.Spec.Pilot.SecurityContext),
					Containers:         r.containers(),
					Volumes:            r.volumes(),
					Affinity:           r.Config.Spec.Pilot.Affinity,
					NodeSelector:       r.Config.Spec.Pilot.NodeSelector,
					Tolerations:        r.Config.Spec.Pilot.Tolerations,
				},
			},
		},
	}

	return deployment
}
