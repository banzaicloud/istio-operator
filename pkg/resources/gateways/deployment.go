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

func (r *Reconciler) deployment() runtime.Object {
	var initContainers []apiv1.Container
	if util.PointerToBool(r.Config.Spec.Proxy.EnableCoreDump) && r.Config.Spec.Proxy.CoreDumpImage != "" {
		initContainers = []apiv1.Container{GetCoreDumpContainer(r.Config)}
	}

	args := []string{
		"proxy",
		"router",
		"--domain", fmt.Sprintf("$(POD_NAMESPACE).svc.%s", r.Config.Spec.Proxy.ClusterDomain),
		"--log_output_level", "info",
		"--serviceCluster", r.gw.Name,
		"--trust-domain", r.Config.Spec.TrustDomain,
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

	var containers = make([]apiv1.Container, 0)
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
					Port:   intstr.FromInt(15021),
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
		SecurityContext:          r.securityContext(),
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
					SecurityContext:    r.podSecurityContext(),
					ServiceAccountName: r.serviceAccountName(),
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            r.volumes(),
					Affinity:           r.gw.Spec.Affinity,
					NodeSelector:       r.gw.Spec.NodeSelector,
					Tolerations:        r.gw.Spec.Tolerations,
					PriorityClassName:  r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}

func (r *Reconciler) podSecurityContext() *apiv1.PodSecurityContext {
	if util.PointerToBool(r.gw.Spec.RunAsRoot) {
		return &apiv1.PodSecurityContext{}
	}

	return &apiv1.PodSecurityContext{
		RunAsUser:    util.Int64Pointer(1337),
		RunAsGroup:   util.Int64Pointer(1337),
		RunAsNonRoot: util.BoolPointer(true),
		FSGroup:      util.Int64Pointer(1337),
	}
}

func (r *Reconciler) securityContext() *apiv1.SecurityContext {
	if util.PointerToBool(r.gw.Spec.RunAsRoot) {
		return &apiv1.SecurityContext{}
	}

	return &apiv1.SecurityContext{
		RunAsUser:    util.Int64Pointer(1337),
		RunAsGroup:   util.Int64Pointer(1337),
		RunAsNonRoot: util.BoolPointer(true),
		Capabilities: &apiv1.Capabilities{
			Drop: []apiv1.Capability{
				"ALL",
			},
		},
		Privileged:             util.BoolPointer(false),
		ReadOnlyRootFilesystem: util.BoolPointer(true),
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
		var portNumber int32
		var portName string
		switch port.TargetPort.Type {
		case intstr.String:
			portNumber = port.Port
			portName = port.TargetPort.StrVal
		case intstr.Int:
			portNumber = port.TargetPort.IntVal
			portName = port.Name
		}
		ports = append(ports, apiv1.ContainerPort{
			ContainerPort: portNumber, Protocol: port.Protocol, Name: portName,
		})
	}
	ports = append(ports, apiv1.ContainerPort{
		ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom",
	})
	return ports
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
			Name: "CANONICAL_SERVICE",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath: "metadata.labels['service.istio.io/canonical-name']",
				},
			},
		},
		{
			Name: "CANONICAL_REVISION",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					FieldPath: "metadata.labels['service.istio.io/canonical-revision']",
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

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "CA_ADDR",
			Value: r.Config.GetCAAddress(),
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

	meshConfig := base.MeshConfig(r.Config, false)
	proxyConfig := meshConfig["defaultConfig"]
	proxyConfigJSON, err := json.Marshal(proxyConfig)
	if err == nil {
		envVars = append(envVars, apiv1.EnvVar{
			Name:  "PROXY_CONFIG",
			Value: string(proxyConfigJSON),
		})
	}

	if r.Config.Spec.NetworkName != "" {
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

	envVars = k8sutil.MergeEnvVars(envVars, r.gw.Spec.AdditionalEnvVars)

	return envVars
}

func (r *Reconciler) volumeMounts() []apiv1.VolumeMount {
	vms := []apiv1.VolumeMount{
		{
			Name:      "istio-envoy",
			MountPath: "/etc/istio/proxy",
		},
		{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
		},
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

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.Pilot.CertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
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

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && util.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		vms = append(vms, apiv1.VolumeMount{
			Name:      "ingressgatewaysdsudspath",
			MountPath: "/var/run/ingress_gateway",
		})
	}

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && util.PointerToBool(r.Config.Spec.MountMtlsCerts) {
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

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.Pilot.CertProvider == istiov1beta1.PilotCertProviderTypeIstiod {
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

	volumes = append(volumes, apiv1.Volume{
		Name: "istio-envoy",
		VolumeSource: apiv1.VolumeSource{
			EmptyDir: &apiv1.EmptyDirVolumeSource{},
		},
	})

	if r.gw.Spec.Type == istiov1beta1.GatewayTypeIngress && (util.PointerToBool(r.Config.Spec.Istiod.Enabled)) {
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

	if util.PointerToBool(r.Config.Spec.Istiod.Enabled) && util.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		volumes = append(volumes, apiv1.Volume{
			Name: "istio-certs",
			VolumeSource: apiv1.VolumeSource{
				Secret: &apiv1.SecretVolumeSource{
					SecretName: fmt.Sprintf("istio.%s", r.serviceAccountName()),
					Optional:   util.BoolPointer(true),
				},
			},
		})
	}

	volumes = append(volumes, apiv1.Volume{
		Name: "config-volume",
		VolumeSource: apiv1.VolumeSource{
			ConfigMap: &apiv1.ConfigMapVolumeSource{
				LocalObjectReference: apiv1.LocalObjectReference{
					Name: base.IstioConfigMapName,
				},
				DefaultMode: util.IntPointer(420),
				Optional:    util.BoolPointer(true),
			},
		},
	})

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
			"sysctl -w kernel.core_pattern=/var/lib/istio/data/core.proxy && ulimit -c unlimited",
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
