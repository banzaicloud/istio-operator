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
	"strings"

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
	"github.com/banzaicloud/istio-operator/pkg/trustbundle"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	vaultTLSVolumeName = "vault-tls"
	vaultTLSVolumePath = "/vault/tls"
	caCertsVolumeName  = "cacerts"
	caCertsVolumePath  = "/etc/cacerts"
)

func (r *Reconciler) containerArgs() []string {
	containerArgs := []string{
		"discovery",
		"--monitoringAddr=:15014",
		"--domain",
		r.Config.Spec.Proxy.ClusterDomain,
		"--keepaliveMaxServerConnectionAge",
		"30m",
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
			Value: r.Config.WithNamespacedRevision("istio-sidecar-injector"),
		},
		{
			Name:  "CENTRAL_ISTIOD",
			Value: "true",
		},
		{
			Name:  "REVISION",
			Value: r.Config.NamespacedRevision(),
		},
		{
			Name:  "ISTIOD_CUSTOM_HOST",
			Value: fmt.Sprintf("%s.%s.svc", r.Config.WithRevision(ServiceNameIstiod), r.Config.Namespace),
		},
		{
			Name:  "VALIDATION_WEBHOOK_CONFIG_NAME",
			Value: r.Config.WithNamespacedRevision("istiod"),
		},
		{
			Name:  "PILOT_ENABLE_ANALYSIS",
			Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Istiod.EnableAnalysis)),
		},
		{
			Name:  "PILOT_ENABLE_STATUS",
			Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Istiod.EnableStatus)),
		},
	}

	if util.PointerToBool(r.Config.Spec.Istiod.MultiControlPlaneSupport) {
		envs = append(envs, []apiv1.EnvVar{
			{
				Name:  "NAMESPACE_LE_NAME",
				Value: r.Config.WithRevision("istio-namespace-controller-election"),
			},
			{
				Name:  "VALIDATION_LE_NAME",
				Value: r.Config.WithRevision("istio-validation-controller-election"),
			},
			{
				Name:  "INGRESS_LE_NAME",
				Value: r.Config.WithRevision("istio-leader"),
			},
			{
				Name:  "CACERT_CONFIG_NAME",
				Value: r.Config.WithRevision("istio-ca-root-cert"),
			},
			{
				Name:  "PILOT_ENDPOINT_TELEMETRY_LABEL",
				Value: "true",
			},
		}...)
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

	envs = r.mergeSPIFFEOperatorEndpoints(envs)

	return envs
}

func (r *Reconciler) mergeSPIFFEOperatorEndpoints(envs []apiv1.EnvVar) []apiv1.EnvVar {
	if !util.PointerToBool(r.Config.Spec.Pilot.SPIFFE.OperatorEndpoints.Enabled) {
		return envs
	}

	env := apiv1.EnvVar{
		Name: "SPIFFE_BUNDLE_ENDPOINTS",
	}

	for i, e := range envs {
		if e.Name == env.Name {
			env = e
			envs = append(envs[:i], envs[i+1:]...)
			break
		}
	}

	p := strings.Split(env.Value, "||")
	if len(p) == 1 && p[0] == "" {
		p = make([]string, 0)
	}

	for _, domain := range append(r.Config.Spec.TrustDomainAliases, r.Config.Spec.TrustDomain) {
		p = append(p, fmt.Sprintf("%s|https://%s/%s?trustDomain=%s&revision=%s#insecure", domain, r.operatorConfig.WebhookServiceAddress, trustbundle.WebhookEndpointPath, domain, r.Config.NamespacedRevision()))
	}

	env.Value = strings.Join(p, "||")

	envs = append(envs, env)

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

func (r *Reconciler) initContainers() []apiv1.Container {
	containers := make([]apiv1.Container, 0)

	if util.PointerToBool(r.Config.Spec.Istiod.CA.Vault.Enabled) {
		containers = append(containers, r.vaultENVInitContainer())
	}

	return containers
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
			InitialDelaySeconds: 1,
			PeriodSeconds:       3,
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

func (r *Reconciler) vaultENVInitContainer() apiv1.Container {
	certChainPath := util.PointerToString(r.Config.Spec.Istiod.CA.Vault.CertPath)
	if r.Config.Spec.Istiod.CA.Vault.CertChainPath != nil {
		certChainPath = util.PointerToString(r.Config.Spec.Istiod.CA.Vault.CertChainPath)
	}

	vaultEnvContainer := apiv1.Container{
		Name:  "vault-env",
		Image: util.PointerToString(r.Config.Spec.Istiod.CA.Vault.VaultEnvImage),
		Args: []string{
			"vault-env",
			"sh",
			"-c",
			`echo "$ISTIO_CA_CERT" > ca-cert.pem; echo "$ISTIO_CA_CERT_CHAIN" > root-cert.pem; echo -e "$ISTIO_CA_CERT\n$ISTIO_CA_CERT_CHAIN" > cert-chain.pem; echo "$ISTIO_CA_KEY" > ca-key.pem`,
		},
		Env: []apiv1.EnvVar{
			{
				Name:  "VAULT_ADDR",
				Value: util.PointerToString(r.Config.Spec.Istiod.CA.Vault.Address),
			},
			{
				Name:  "VAULT_ROLE",
				Value: util.PointerToString(r.Config.Spec.Istiod.CA.Vault.Role),
			},
			{
				Name:  "VAULT_CACERT",
				Value: fmt.Sprintf("%s/ca.crt", vaultTLSVolumePath),
			},
			{
				Name:  "ISTIO_CA_CERT",
				Value: util.PointerToString(r.Config.Spec.Istiod.CA.Vault.CertPath),
			},
			{
				Name:  "ISTIO_CA_KEY",
				Value: util.PointerToString(r.Config.Spec.Istiod.CA.Vault.KeyPath),
			},
			{
				Name:  "ISTIO_CA_CERT_CHAIN",
				Value: certChainPath,
			},
		},
		WorkingDir: caCertsVolumePath,
		VolumeMounts: []apiv1.VolumeMount{
			{
				MountPath: caCertsVolumePath,
				Name:      caCertsVolumeName,
			},
			{
				MountPath: vaultTLSVolumePath,
				Name:      vaultTLSVolumeName,
			},
		},
	}

	return vaultEnvContainer
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

		if util.PointerToBool(r.Config.Spec.Istiod.CA.Vault.Enabled) || r.Config.Spec.Citadel.CASecretName != "" {
			vms = append(vms, apiv1.VolumeMount{
				Name:      caCertsVolumeName,
				MountPath: caCertsVolumePath,
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
						Name: r.Config.WithRevision(base.IstioConfigMapName),
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

		if util.PointerToBool(r.Config.Spec.Istiod.CA.Vault.Enabled) {
			volumes = append(volumes, []apiv1.Volume{
				{
					Name: caCertsVolumeName,
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: vaultTLSVolumeName,
					VolumeSource: apiv1.VolumeSource{
						Projected: &apiv1.ProjectedVolumeSource{
							DefaultMode: util.IntPointer(420),
							Sources: []apiv1.VolumeProjection{
								{
									Secret: &apiv1.SecretProjection{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: vaultTLSVolumeName,
										},
										Items: []apiv1.KeyToPath{
											{
												Key:  "ca.crt",
												Path: "ca.crt",
											},
										},
									},
								},
							},
						},
					},
				},
			}...)
		} else if r.Config.Spec.Citadel.CASecretName != "" {
			volumes = append(volumes, apiv1.Volume{
				Name: caCertsVolumeName,
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
						Name: r.Config.WithRevision("istio-sidecar-injector"),
					},
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
				Name:      r.Config.WithRevision(hpaName),
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
					ServiceAccountName: r.Config.WithRevision(serviceAccountName),
					SecurityContext:    util.GetPodSecurityContextFromSecurityContext(r.Config.Spec.Pilot.SecurityContext),
					InitContainers:     r.initContainers(),
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
