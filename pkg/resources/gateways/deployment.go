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
	"fmt"

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

func (r *Reconciler) deployment(gw string) runtime.Object {
	gwConfig := r.getGatewayConfig(gw)

	var initContainers []apiv1.Container
	if r.Config.Spec.Proxy.EnableCoreDump {
		initContainers = []apiv1.Container{GetCoreDumpContainer(r.Config)}
	}

	var containers = make([]apiv1.Container, 0)
	if gwConfig.SDS.Enabled {
		containers = append(containers, apiv1.Container{
			Name:            "ingress-sds",
			Image:           gwConfig.SDS.Image,
			ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
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
	containers = append(containers, apiv1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args: []string{
			"proxy",
			"router",
			"--domain", "$(POD_NAMESPACE).svc.cluster.local",
			"--log_output_level", "info",
			"--drainDuration", "45s",
			"--parentShutdownDuration", "1m0s",
			"--connectTimeout", "10s",
			"--serviceCluster", fmt.Sprintf("istio-%s", gw),
			"--zipkinAddress", r.Config.Spec.Tracing.Zipkin.Address,
			"--proxyAdminPort", "15000",
			"--statusPort", "15020",
			"--controlPlaneAuthPolicy", templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
			"--discoveryAddress", fmt.Sprintf("istio-pilot.%s:%s", r.Config.Namespace, r.discoveryPort()),
		},
		Ports: r.ports(gw),
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
		Env:                      append(templates.IstioProxyEnv(), r.envVars(gwConfig)...),
		Resources:                templates.DefaultResources(),
		VolumeMounts:             r.volumeMounts(gw, gwConfig),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	})

	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(gatewayName(gw), util.MergeLabels(labelSelector(gw), gwLabels(gw)), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName(gw),
				Namespace: r.Config.Namespace,
			}, gwConfig.ReplicaCount)),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeLabels(labelSelector(gw), gwLabels(gw)),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeLabels(labelSelector(gw), gwLabels(gw)),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName(gw),
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            r.volumes(gw, gwConfig),
					Affinity:           &apiv1.Affinity{},
				},
			},
		},
	}
}

func (r *Reconciler) ports(gw string) []apiv1.ContainerPort {
	switch gw {
	case ingress:
		return []apiv1.ContainerPort{
			{ContainerPort: 80, Protocol: apiv1.ProtocolTCP, Name: "http2"},
			{ContainerPort: 443, Protocol: apiv1.ProtocolTCP, Name: "https"},
			{ContainerPort: 31400, Protocol: apiv1.ProtocolTCP, Name: "tcp"},
			{ContainerPort: 15029, Protocol: apiv1.ProtocolTCP, Name: "https-kiali"},
			{ContainerPort: 15030, Protocol: apiv1.ProtocolTCP, Name: "https-prom"},
			{ContainerPort: 15031, Protocol: apiv1.ProtocolTCP, Name: "https-grafana"},
			{ContainerPort: 15032, Protocol: apiv1.ProtocolTCP, Name: "https-tracing"},
			{ContainerPort: 15443, Protocol: apiv1.ProtocolTCP, Name: "tls"},
			{ContainerPort: 15020, Protocol: apiv1.ProtocolTCP, Name: "status-port"},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	case egress:
		return []apiv1.ContainerPort{
			{ContainerPort: 80, Protocol: apiv1.ProtocolTCP, Name: "http2"},
			{ContainerPort: 443, Protocol: apiv1.ProtocolTCP, Name: "https"},
			{ContainerPort: 15443, Protocol: apiv1.ProtocolTCP, Name: "tls"},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	}
	return nil
}

func (r *Reconciler) discoveryPort() string {
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return "15011"
	}
	return "15010"
}

func (r *Reconciler) envVars(gwConfig *istiov1beta1.GatewayConfiguration) []apiv1.EnvVar {
	envVars := []apiv1.EnvVar{
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
	}
	if gwConfig.SDS.Enabled {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "ISTIO_META_USER_SDS",
			Value: "true",
		})
	}
	return envVars
}

func (r *Reconciler) volumeMounts(gw string, gwConfig *istiov1beta1.GatewayConfiguration) []apiv1.VolumeMount {
	vms := []apiv1.VolumeMount{
		{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		},
		{
			Name:      fmt.Sprintf("%s-certs", gw),
			MountPath: fmt.Sprintf("/etc/istio/%s-certs", gw),
			ReadOnly:  true,
		},
		{
			Name:      fmt.Sprintf("%s-ca-certs", gw),
			MountPath: fmt.Sprintf("/etc/istio/%s-ca-certs", gw),
			ReadOnly:  true,
		},
	}
	if r.Config.Spec.SDS.Enabled {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "sdsudspath",
			MountPath: "/var/run/sds/uds_path",
			ReadOnly:  true,
		})
		if r.Config.Spec.SDS.UseTrustworthyJwt {
			vms = append(vms, apiv1.VolumeMount{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
			})
		}
	}
	if gwConfig.SDS.Enabled {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "ingressgatewaysdsudspath",
			MountPath: "/var/run/ingress_gateway",
		})
	}
	return vms
}

func (r *Reconciler) volumes(gw string, gwConfig *istiov1beta1.GatewayConfiguration) []apiv1.Volume {
	volumes := []apiv1.Volume{
		{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio.%s", serviceAccountName(gw)),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-certs", gw),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio-%s-certs", gw),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-ca-certs", gw),
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio-%s-ca-certs", gw),
					Optional:    util.BoolPointer(true),
					DefaultMode: util.IntPointer(420),
				},
			},
		},
	}
	if r.Config.Spec.SDS.Enabled {
		hostPathType := apiv1.HostPathSocket
		volumes = append(volumes, apiv1.Volume{
			Name: "sdsudspath",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/run/sds/uds_path",
					Type: &hostPathType,
				},
			},
		})
		if r.Config.Spec.SDS.UseTrustworthyJwt {
			volumes = append(volumes, apiv1.Volume{
				Name: "istio-token",
				VolumeSource: apiv1.VolumeSource{
					Projected: &apiv1.ProjectedVolumeSource{
						Sources: []apiv1.VolumeProjection{
							{
								ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
									Path:              "istio-token",
									ExpirationSeconds: util.Int64Pointer(43200),
									Audience:          "",
								},
							},
						},
						DefaultMode: util.IntPointer(420),
					},
				},
			})
		}
	}
	if gwConfig.SDS.Enabled {
		volumes = append(volumes, apiv1.Volume{
			Name: "ingressgatewaysdsudspath",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		})
	}
	return volumes
}

// GetCoreDumpContainer get core dump init container for Envoy proxies
func GetCoreDumpContainer(config *istiov1beta1.Istio) apiv1.Container {
	return apiv1.Container{
		Name:            "enable-core-dump",
		Image:           config.Spec.ProxyInit.Image,
		ImagePullPolicy: config.Spec.ImagePullPolicy,
		Command: []string{
			"/bin/sh",
		},
		Args: []string{
			"-c",
			"sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited",
		},
		SecurityContext: &apiv1.SecurityContext{
			Privileged: util.BoolPointer(true),
		},
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}
