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

package mixer

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/base"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) deployment(t string) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMetaWithRevision(deploymentName(t), labelSelector, r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName(t),
				Namespace: r.Config.Namespace,
			}, util.PointerToInt32(r.k8sResourceConfig.ReplicaCount))),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeMultipleStringMaps(labelSelector, mixerTypeLabel(t), r.Config.RevisionLabels()),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeMultipleStringMaps(labelSelector, appLabel(t), mixerTypeLabel(t), mixerTLSModeLabel, r.Config.RevisionLabels()),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: r.Config.WithName(serviceAccountName),
					Volumes:            r.volumes(t),
					Affinity:           r.k8sResourceConfig.Affinity,
					NodeSelector:       r.k8sResourceConfig.NodeSelector,
					Tolerations:        r.k8sResourceConfig.Tolerations,
					Containers: []apiv1.Container{
						r.mixerContainer(t, r.Config.Namespace),
						r.istioProxyContainer(t),
					},
					PriorityClassName: r.Config.Spec.PriorityClassName,
					SecurityContext:   util.GetPodSecurityContextFromSecurityContext(r.Config.Spec.Mixer.SecurityContext),
				},
			},
		},
	}
}

func (r *Reconciler) volumes(t string) []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio.%s", r.Config.WithName(serviceAccountName)),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
		{
			Name: "uds-socket",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: fmt.Sprintf("%s-adapter-secret", t),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-adapter-secret", t),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
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
	}

	if t == telemetryComponentName {
		volumes = append(volumes, apiv1.Volume{
			Name: r.Config.WithName(configMapNameEnvoy),
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: r.Config.WithName(configMapNameEnvoy),
					},
					DefaultMode: util.IntPointer(420),
				},
			},
		})
	}

	volumes = append(volumes, apiv1.Volume{
		Name: "config-volume",
		VolumeSource: apiv1.VolumeSource{
			ConfigMap: &apiv1.ConfigMapVolumeSource{
				LocalObjectReference: apiv1.LocalObjectReference{
					Name: r.Config.WithName(base.IstioConfigMapName),
				},
				DefaultMode: util.IntPointer(420),
				Optional:    util.BoolPointer(true),
			},
		},
	})

	if r.Config.Spec.Pilot.CertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
		volumes = append(volumes, apiv1.Volume{
			Name: "istiod-ca-cert",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "istio-ca-root-cert",
					},
				},
			},
		})
	}

	if r.Config.Spec.JWTPolicy == istiov1beta1.JWTPolicyThirdPartyJWT {
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

func (r *Reconciler) containerEnvs(t string) []apiv1.EnvVar {
	envs := []apiv1.EnvVar{
		{
			Name:  "GOMAXPROCS",
			Value: "6",
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}

	envs = append(envs, []apiv1.EnvVar{
		{
			Name:  "JWT_POLICY",
			Value: string(r.Config.Spec.JWTPolicy),
		},
		{
			Name:  "PILOT_CERT_PROVIDER",
			Value: string(r.Config.Spec.Pilot.CertProvider),
		},
		{
			Name:  "ISTIO_META_USER_SDS",
			Value: "true",
		},
	}...)

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envs = append(envs, apiv1.EnvVar{
			Name:  "CA_ADDR",
			Value: r.Config.GetCAAddress(),
		})
	}

	envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Mixer.AdditionalEnvVars)

	switch t {
	case telemetryComponentName:
		envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Telemetry.AdditionalEnvVars)
	case policyComponentName:
		envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Policy.AdditionalEnvVars)
	}

	return envs
}

func (r *Reconciler) containerArgs(t string, ns string) []string {
	containerArgs := []string{
		"--address",
		"unix:///sock/mixer.socket",
		"--configDefaultNamespace",
		ns,
		"--monitoringPort",
		"15014",
	}

	if r.Config.Spec.Logging.Level != nil {
		containerArgs = append(containerArgs, fmt.Sprintf("--log_output_level=%s", util.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		containerArgs = append(containerArgs, "--trace_zipkin_url",
			"http://"+r.Config.Spec.Tracing.Zipkin.Address+"/api/v1/spans")
	}

	if util.PointerToBool(r.Config.Spec.UseMCP) {
		if r.Config.Spec.ControlPlaneSecurityEnabled {
			containerArgs = append(containerArgs, "--configStoreURL", "mcps://istio-galley."+r.Config.Namespace+".svc:9901")
			if t == telemetryComponentName {
				containerArgs = append(containerArgs, "--certFile", "/etc/certs/cert-chain.pem")
				containerArgs = append(containerArgs, "--keyFile", "/etc/certs/key.pem")
				containerArgs = append(containerArgs, "--caCertFile", "/etc/certs/root-cert.pem")
			}
		} else {
			containerArgs = append(containerArgs, "--configStoreURL", "mcp://istio-galley."+r.Config.Namespace+".svc:9901")
		}
	} else {
		containerArgs = append(containerArgs, "--configStoreURL", "k8s://")
	}

	if r.Config.Spec.WatchAdapterCRDs {
		containerArgs = append(containerArgs, "--useAdapterCRDs=true")
	} else {
		containerArgs = append(containerArgs, "--useAdapterCRDs=false")
	}

	containerArgs = append(containerArgs, "--useTemplateCRDs=false")

	if t == telemetryComponentName {
		containerArgs = append(containerArgs, "--averageLatencyThreshold", "100ms")
		containerArgs = append(containerArgs, "--loadsheddingMode", "enforce")
	}

	if len(r.Config.Spec.Mixer.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.Mixer.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) mixerContainer(t string, ns string) apiv1.Container {
	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      "uds-socket",
			MountPath: "/sock",
		},
		{
			Name:      fmt.Sprintf("%s-adapter-secret", t),
			MountPath: fmt.Sprintf("/var/run/secrets/istio.io/%s/adapter", t),
			ReadOnly:  true,
		},
	}
	if util.PointerToBool(r.Config.Spec.UseMCP) {
		volumeMounts = append(volumeMounts, apiv1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		volumeMounts = append(volumeMounts, apiv1.VolumeMount{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
			ReadOnly:  true,
		})

		if r.Config.Spec.Pilot.CertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
			volumeMounts = append(volumeMounts, apiv1.VolumeMount{
				Name:      "istiod-ca-cert",
				MountPath: "/var/run/secrets/istio",
			})
		}
		if r.Config.Spec.JWTPolicy == istiov1beta1.JWTPolicyThirdPartyJWT {
			volumeMounts = append(volumeMounts, apiv1.VolumeMount{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
				ReadOnly:  true,
			})
		}
	}

	return apiv1.Container{
		Name:            t,
		Image:           util.PointerToString(r.k8sResourceConfig.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Ports: []apiv1.ContainerPort{
			{
				ContainerPort: 15014,
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				ContainerPort: 42422,
				Protocol:      apiv1.ProtocolTCP,
			},
		},
		Args: r.containerArgs(t, ns),
		Env:  r.containerEnvs(t),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.k8sResourceConfig.Resources,
			r.Config.Spec.DefaultResources,
		),
		VolumeMounts: volumeMounts,
		LivenessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path:   "/version",
					Port:   intstr.FromInt(15014),
					Scheme: apiv1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		},
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
		SecurityContext:          r.Config.Spec.Mixer.SecurityContext,
	}
}

func (r *Reconciler) istioProxyContainer(t string) apiv1.Container {
	templateFile := fmt.Sprintf("/var/lib/istio/envoy/envoy_%s.yaml.tmpl", t)
	if t == telemetryComponentName && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		templateFile = "/var/lib/envoy/envoy.yaml.tmpl"
	}

	args := []string{
		"proxy",
		"--serviceCluster",
		fmt.Sprintf("istio-%s", t),
		"--templateFile",
		templateFile,
		"--controlPlaneAuthPolicy",
		templates.ControlPlaneAuthPolicy(util.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"--domain",
		fmt.Sprintf("$(POD_NAMESPACE).svc.%s", r.Config.Spec.Proxy.ClusterDomain),
		"--trust-domain",
		r.Config.Spec.TrustDomain,
	}
	if r.Config.Spec.Proxy.LogLevel != "" {
		args = append(args, fmt.Sprintf("--proxyLogLevel=%s", r.Config.Spec.Proxy.LogLevel))
	}
	if r.Config.Spec.Proxy.ComponentLogLevel != "" {
		args = append(args, fmt.Sprintf("--proxyComponentLogLevel=%s", r.Config.Spec.Proxy.ComponentLogLevel))
	}
	if r.Config.Spec.Logging.Level != nil {
		args = append(args, fmt.Sprintf("--log_output_level=%s", util.PointerToString(r.Config.Spec.Logging.Level)))
	}

	vms := []apiv1.VolumeMount{
		{
			Name:      "uds-socket",
			MountPath: "/sock",
		},
		{
			Name:      "istio-envoy",
			MountPath: "/etc/istio/proxy",
		},
	}

	if t == telemetryComponentName {
		vms = append(vms, apiv1.VolumeMount{
			Name:      r.Config.WithName(configMapNameEnvoy),
			MountPath: "/var/lib/envoy",
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
		})

		if r.Config.Spec.Pilot.CertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
			vms = append(vms, apiv1.VolumeMount{
				Name:      "istiod-ca-cert",
				MountPath: "/var/run/secrets/istio",
			})
		}
		if r.Config.Spec.JWTPolicy == istiov1beta1.JWTPolicyThirdPartyJWT {
			vms = append(vms, apiv1.VolumeMount{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
				ReadOnly:  true,
			})
		}
	}

	return apiv1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Ports: []apiv1.ContainerPort{
			{ContainerPort: 9091, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15004, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		},
		Args: args,
		Env:  templates.IstioProxyEnv(r.Config),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Proxy.Resources,
			r.Config.Spec.DefaultResources,
		),
		VolumeMounts:             vms,
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
		SecurityContext:          r.Config.Spec.Proxy.SecurityContext,
	}
}
