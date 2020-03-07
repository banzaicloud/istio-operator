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

package gateways

import (
	"encoding/json"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) deployment() runtime.Object {
	var initContainers []apiv1.Container
	if util.PointerToBool(r.Config.Spec.Proxy.EnableCoreDump) && r.Config.Spec.Proxy.CoreDumpImage != "" {
		initContainers = []apiv1.Container{GetCoreDumpContainer(r.Config)}
	}

	var containers = make([]apiv1.Container, 0)
	if !util.PointerToBool(r.Config.Spec.Istiod.Enabled) && util.PointerToBool(r.gw.Spec.SDS.Enabled) {
		containers = append(containers, apiv1.Container{
			Name:            "ingress-sds",
			Image:           r.gw.Spec.SDS.Image,
			ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
			Resources: templates.GetResourcesRequirementsOrDefault(
				r.gw.Spec.SDS.Resources,
				r.Config.Spec.DefaultResources,
			),
			Env: []apiv1.EnvVar{
				{
					Name:  "ENABLE_WORKLOAD_SDS",
					Value: "false",
				},
				{
					Name:  "ENABLE_INGRESS_GATEWAY_SDS",
					Value: "true",
				},
				{
					Name: "INGRESS_GATEWAY_NAMESPACE",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.namespace",
						},
					},
				},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ingressgatewaysdsudspath",
					MountPath: "/var/run/ingress_gateway",
				},
			},
			TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
			TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
		})
	}
	args := []string{
		"proxy",
		"router",
		"--domain", fmt.Sprintf("$(POD_NAMESPACE).svc.%s", r.Config.Spec.Proxy.ClusterDomain),
		"--log_output_level", "info",
		"--drainDuration", "45s",
		"--parentShutdownDuration", "1m0s",
		"--connectTimeout", "10s",
		"--serviceCluster", r.gw.Name,
		"--proxyAdminPort", "15000",
		"--statusPort", "15020",
		"--controlPlaneAuthPolicy", templates.ControlPlaneAuthPolicy(util.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"--discoveryAddress", r.discoveryAddress(),
		"--trust-domain", r.Config.Spec.TrustDomain,
	}

	if util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		if r.Config.Spec.Tracing.Tracer == istiov1beta1.TracerTypeLightstep {
			args = append(args, "--lightstepAddress", r.Config.Spec.Tracing.Lightstep.Address)
			args = append(args, "--lightstepAccessToken", r.Config.Spec.Tracing.Lightstep.AccessToken)
			args = append(args, fmt.Sprintf("--lightstepSecure=%t", r.Config.Spec.Tracing.Lightstep.Secure))
			args = append(args, "--lightstepCacertPath", r.Config.Spec.Tracing.Lightstep.CacertPath)
		} else if r.Config.Spec.Tracing.Tracer == istiov1beta1.TracerTypeZipkin {
			args = append(args, "--zipkinAddress", r.Config.Spec.Tracing.Zipkin.Address)
		} else if r.Config.Spec.Tracing.Tracer == istiov1beta1.TracerTypeDatadog {
			args = append(args, "--datadogAgentAddress", r.Config.Spec.Tracing.Datadog.Address)
		}
	}

	if r.Config.Spec.Proxy.LogLevel != "" {
		args = append(args, "--proxyLogLevel", r.Config.Spec.Proxy.LogLevel)
	}

	if r.Config.Spec.Proxy.ComponentLogLevel != "" {
		args = append(args, "--proxyComponentLogLevel", r.Config.Spec.Proxy.ComponentLogLevel)
	}

	if r.Config.Spec.Logging.Level != nil {
		args = append(args, fmt.Sprintf("--log_output_level=%s", util.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyMetricsService.Enabled) {
		envoyMetricsServiceJSON, err := r.getEnvoyServiceConfigurationJSON(r.Config.Spec.Proxy.EnvoyMetricsService)
		if err == nil {
			args = append(args, "--envoyMetricsService", fmt.Sprintf("%s", string(envoyMetricsServiceJSON)))
		}
	}

	if util.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		envoyAccessLogServiceJSON, err := r.getEnvoyServiceConfigurationJSON(r.Config.Spec.Proxy.EnvoyAccessLogService)
		if err == nil {
			args = append(args, "--envoyAccessLogService", fmt.Sprintf("%s", string(envoyAccessLogServiceJSON)))
		}
	}

	containers = append(containers, apiv1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args:            args,
		Ports:           r.ports(),
		ReadinessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path:   "/healthz/ready",
					Port:   intstr.FromInt(15020),
					Scheme: apiv1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 1,
			PeriodSeconds:       2,
			FailureThreshold:    30,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		},
		Env: append(templates.IstioProxyEnv(r.Config), r.envVars()...),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.gw.Spec.Resources,
			r.Config.Spec.Proxy.Resources,
		),
		VolumeMounts:             r.volumeMounts(),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	})

	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(r.gatewayName(), r.labels(), r.gw),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      r.hpaName(),
				Namespace: r.Config.Namespace,
			}, *r.gw.Spec.ReplicaCount)),
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labels(),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.labels(),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: r.serviceAccountName(),
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            r.volumes(),
					Affinity:           r.gw.Spec.Affinity,
					NodeSelector:       r.gw.Spec.NodeSelector,
					Tolerations:        r.gw.Spec.Tolerations,
				},
			},
		},
	}
}

func (r *Reconciler) getEnvoyServiceConfigurationJSON(config istiov1beta1.EnvoyServiceCommonConfiguration) (string, error) {
	type Properties struct {
		Address      string                     `json:"address,omitempty"`
		TLSSettings  *istiov1beta1.TLSSettings  `json:"tlsSettings,omitempty"`
		TCPKeepalive *istiov1beta1.TCPKeepalive `json:"tcpKeepalive,omitempty"`
	}

	properties := Properties{
		Address:      fmt.Sprintf("%s:%d", config.Host, config.Port),
		TLSSettings:  config.TLSSettings,
		TCPKeepalive: config.TCPKeepalive,
	}

	data, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	return string(data), err
}

func (r *Reconciler) ports() []apiv1.ContainerPort {
	var ports []apiv1.ContainerPort
	for _, port := range r.gw.Spec.Ports {
		ports = append(ports, apiv1.ContainerPort{
			ContainerPort: port.Port, Protocol: port.Protocol, Name: port.Name,
		})
	}
	ports = append(ports, apiv1.ContainerPort{
		ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom",
	})
	return ports
}

func (r *Reconciler) discoveryAddress() string {
	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		return fmt.Sprintf("istio-pilot.%s.svc:15012", r.Config.Namespace)
	}
	return fmt.Sprintf("istio-pilot.%s:%s", r.Config.Namespace, r.discoveryPort())
}

func (r *Reconciler) discoveryPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15011"
	}
	return "15010"
}

func (r *Reconciler) envVars() []apiv1.EnvVar {
	envVars := []apiv1.EnvVar{
		{
			Name: "HOST_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
		{
			Name: "SERVICE_ACCOUNT",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath:  "spec.serviceAccountName",
					APIVersion: "v1",
				},
			},
		},
		{
			Name: "ISTIO_META_POD_NAME",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath:  "metadata.name",
					APIVersion: "v1",
				},
			},
		},
		{
			Name: "ISTIO_META_CONFIG_NAMESPACE",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath:  "metadata.namespace",
					APIVersion: "v1",
				},
			},
		},
		{
			Name:  "ISTIO_META_ROUTER_MODE",
			Value: "sni-dnat",
		},
		{
			Name: "NODE_NAME",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath:  "spec.nodeName",
					APIVersion: "v1",
				},
			},
		},
		{
			Name:  "ISTIO_META_WORKLOAD_NAME",
			Value: r.gatewayName(),
		},
		{
			Name:  "ISTIO_META_OWNER",
			Value: fmt.Sprintf("kubernetes://apis/apps/v1/namespaces/%s/deployments/%s", r.Config.Namespace, r.gatewayName()),
		},
	}

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && (util.PointerToBool(r.Config.Spec.Istiod.Enabled) || util.PointerToBool(r.gw.Spec.SDS.Enabled)) {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_USER_SDS",
			Value: "true",
		})
	}

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "CA_ADDR",
			Value: fmt.Sprintf("istio-pilot.%s.svc:15012", r.Config.Namespace),
		})
	}

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

	if util.PointerToBool(r.Config.Spec.MeshExpansion) && r.Config.Spec.NetworkName != "" {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_NETWORK",
			Value: r.Config.Spec.NetworkName,
		})
	}
	if r.gw.Spec.RequestedNetworkView != "" {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_REQUESTED_NETWORK_VIEW",
			Value: r.gw.Spec.RequestedNetworkView,
		})
	}
	if util.PointerToBool(r.Config.Spec.AutoMTLS) {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_AUTO_MTLS_ENABLED",
			Value: "true",
		})
	}

	if r.Config.Spec.MeshID != "" {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_MESH_ID",
			Value: r.Config.Spec.MeshID,
		})
	} else if r.Config.Spec.TrustDomain != "" {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_MESH_ID",
			Value: r.Config.Spec.TrustDomain,
		})
	}

	if util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		if r.Config.Spec.Tracing.Tracer == istiov1beta1.TracerTypeDatadog {
			envVars = append(envVars, apiv1.EnvVar{
				Name: "HOST_IP",
				ValueFrom: &apiv1.EnvVarSource{
					FieldRef: &apiv1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			})
		} else if r.Config.Spec.Tracing.Tracer == istiov1beta1.TracerTypeStackdriver {
			envVars = append(envVars, apiv1.EnvVar{
				Name:  "STACKDRIVER_TRACING_ENABLED",
				Value: "true",
			})
			envVars = append(envVars, apiv1.EnvVar{
				Name:  "STACKDRIVER_TRACING_DEBUG",
				Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Tracing.Strackdriver.Debug)),
			})
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations != nil {
				envVars = append(envVars, apiv1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_ANNOTATIONS",
					Value: string(util.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations)),
				})
			}
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes != nil {
				envVars = append(envVars, apiv1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_ATTRIBUTES",
					Value: string(util.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes)),
				})
			}
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents != nil {
				envVars = append(envVars, apiv1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_MESSAGE_EVENTS",
					Value: string(util.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents)),
				})
			}
		}
	}

	return envVars
}

func (r *Reconciler) volumeMounts() []apiv1.VolumeMount {
	vms := []apiv1.VolumeMount{
		{
			Name:      fmt.Sprintf("%s-certs", r.gw.Name),
			MountPath: fmt.Sprintf("/etc/istio/%s-certs", r.gw.Spec.Type+"gateway"),
			ReadOnly:  true,
		},
		{
			Name:      fmt.Sprintf("%s-ca-certs", r.gw.Name),
			MountPath: fmt.Sprintf("/etc/istio/%s-ca-certs", r.gw.Spec.Type+"gateway"),
			ReadOnly:  true,
		},
	}

	if r.Config.Spec.PilotCertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "istiod-ca-cert",
			MountPath: "/var/run/secrets/istio",
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.JWTPolicy == istiov1beta1.JWTPolicyThirdPartyJWT {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "istio-token",
			MountPath: "/var/run/secrets/tokens",
			ReadOnly:  true,
		})
	}

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && (util.PointerToBool(r.Config.Spec.Istiod.Enabled) || util.PointerToBool(r.gw.Spec.SDS.Enabled)) {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "ingressgatewaysdsudspath",
			MountPath: "/var/run/ingress_gateway",
		})
	}

	if util.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	vms = append(vms, apiv1.VolumeMount{
		Name:      "podinfo",
		MountPath: "/etc/istio/pod",
	})

	return vms
}

func (r *Reconciler) volumes() []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			Name: fmt.Sprintf("%s-certs", r.gw.Name),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-certs", r.gw.Name),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-ca-certs", r.gw.Name),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-ca-certs", r.gw.Name),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
	}

	if r.Config.Spec.PilotCertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
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

	volumes = append(volumes, apiv1.Volume{
		Name: "podinfo",
		VolumeSource: apiv1.VolumeSource{
			DownwardAPI: &apiv1.DownwardAPIVolumeSource{
				DefaultMode: util.IntPointer(420),
				Items: []apiv1.DownwardAPIVolumeFile{
					{
						Path: "labels",
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.labels",
						},
					},
					{
						Path: "annotations",
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.annotations",
						},
					},
				},
			},
		},
	})

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && (util.PointerToBool(r.Config.Spec.Istiod.Enabled) || util.PointerToBool(r.gw.Spec.SDS.Enabled)) {
		volumes = append(volumes, apiv1.Volume{
			Name: "ingressgatewaysdsudspath",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.JWTPolicy == istiov1beta1.JWTPolicyThirdPartyJWT {
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

	if util.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		volumes = append(volumes, apiv1.Volume{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: "istio.default",
					Optional:   util.BoolPointer(true),
				},
			},
		})
	}

	return volumes
}

// GetCoreDumpContainer get core dump init container for Envoy proxies
func GetCoreDumpContainer(config *istiov1beta1.Istio) apiv1.Container {
	return apiv1.Container{
		Name:            "enable-core-dump",
		Image:           config.Spec.Proxy.CoreDumpImage,
		ImagePullPolicy: config.Spec.ImagePullPolicy,
		Command: []string{
			"/bin/sh",
		},
		Args: []string{
			"-c",
			"sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited",
		},
		Resources: templates.GetResourcesRequirementsOrDefault(config.Spec.SidecarInjector.Init.Resources, config.Spec.DefaultResources),
		SecurityContext: &apiv1.SecurityContext{
			AllowPrivilegeEscalation: util.BoolPointer(true),
			Capabilities: &apiv1.Capabilities{
				Add: []apiv1.Capability{
					"SYS_ADMIN",
				},
				Drop: []apiv1.Capability{
					"ALL",
				},
			},
			Privileged:             util.BoolPointer(true),
			ReadOnlyRootFilesystem: util.BoolPointer(false),
			RunAsGroup:             util.Int64Pointer(0),
			RunAsNonRoot:           util.BoolPointer(false),
			RunAsUser:              util.Int64Pointer(0),
		},
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}
