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

package helm

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func Convert(c *istiov1beta1.IstioSpec) *IstioHelmValues {
	globalConfig := &GlobalConfig{
		ImageHub: c.ImageHub,
		ImageTag: c.ImageTag,
		MTLS: &MTLSConfig{
			EnabledField{
				Enabled: &c.MTLS,
			},
		},
		ControlPlaneSecurityEnabled: &c.ControlPlaneSecurityEnabled,
		Proxy: &ProxyConfig{
			Image:           c.Proxy.Image,
			Privileged:      &c.Proxy.Privileged,
			EnableCoreDump:  &c.Proxy.EnableCoreDump,
			IncludeIPRanges: c.IncludeIPRanges,
			ExcludeIPRanges: c.ExcludeIPRanges,
		},
		ProxyInit: &ProxyInitConfig{
			Image: c.ProxyInit.Image,
		},
		SDS: &SDSConfig{
			EnabledField: EnabledField{
				Enabled: &c.SDS.Enabled,
			},
			UDSPath:           c.SDS.UdsPath,
			UseNormalJWT:      &c.SDS.UseNormalJwt,
			UseTrustworthyJWT: &c.SDS.UseTrustworthyJwt,
		},
		KubernetesIngress: &KubernetesIngressConfig{
			EnabledField: EnabledField{
				Enabled: &c.Gateways.K8sIngress.Enabled,
			},
		},
		DefaultPodDisruptionBudget: &PodDisruptionBudget{
			EnabledField: EnabledField{
				Enabled: &c.DefaultPodDisruptionBudget.Enabled,
			},
		},
		OutboundTrafficPolicy: &OutboundTrafficPolicyConfig{
			Mode: OutboundTrafficPolicyMode(c.OutboundTrafficPolicy.Mode),
		},
		UseMCP:                          &c.UseMCP,
		DefaultConfigVisibilitySettings: defaultConfigVisibilitySettings(c.DefaultConfigVisibility),
		Tracer: &ProxyTracerConfig{
			Zipkin: &ProxyTracerZipkinConfig{
				Address: c.Tracing.Zipkin.Address,
			},
		},
		OneNamespace: &c.WatchOneNamespace,
	}

	return &IstioHelmValues{
		Global: globalConfig,
		Pilot: &PilotConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image:        c.Pilot.Image,
				ReplicaCount: &c.Pilot.ReplicaCount,
				HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
					AutoscaleMin: &c.Pilot.MinReplicas,
					AutoscaleMax: &c.Pilot.MaxReplicas,
				},
			},
			TraceSampling: &c.Pilot.TraceSampling,
		},
		Security: &SecurityConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image:        c.Citadel.Image,
				ReplicaCount: &c.Citadel.ReplicaCount,
			},
		},
		Galley: &GalleyConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image:        c.Galley.Image,
				ReplicaCount: &c.Galley.ReplicaCount,
			},
		},
		Mixer: &MixerConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image:        c.Mixer.Image,
				ReplicaCount: &c.Mixer.ReplicaCount,
				HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
					AutoscaleMin: &c.Mixer.MinReplicas,
					AutoscaleMax: &c.Mixer.MaxReplicas,
				},
			},
			Policy: &MixerPolicyConfig{
				DeploymentFields: DeploymentFields{
					Image:        c.Mixer.Image,
					ReplicaCount: &c.Mixer.ReplicaCount,
					HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
						AutoscaleMin: &c.Mixer.MinReplicas,
						AutoscaleMax: &c.Mixer.MaxReplicas,
					},
				},
			},
			Telemetry: &MixerTelemetryConfig{
				DeploymentFields: DeploymentFields{
					Image:        c.Mixer.Image,
					ReplicaCount: &c.Mixer.ReplicaCount,
					HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
						AutoscaleMin: &c.Mixer.MinReplicas,
						AutoscaleMax: &c.Mixer.MaxReplicas,
					},
				},
			},
			Adapters: &MixerAdaptersConfig{
				UseAdapterCRDs: &c.WatchAdapterCRDs,
			},
		},
		SidecarInjector: &SidecarInjectorConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image:        c.SidecarInjector.Image,
				ReplicaCount: &c.SidecarInjector.ReplicaCount,
			},
			RewriteAppHTTPProbe: &c.SidecarInjector.RewriteAppHTTPProbe,
		},
		Gateways: &GatewaysConfig{
			CommonComponentConfig: CommonComponentConfig{
				Global: globalConfig,
			},
			Gateways: map[string]GatewayConfig{
				"istio-ingressgateway": {
					DeploymentFields: DeploymentFields{
						ReplicaCount: &c.Gateways.IngressConfig.ReplicaCount,
						HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
							AutoscaleMin: &c.Gateways.IngressConfig.MinReplicas,
							AutoscaleMax: &c.Gateways.IngressConfig.MaxReplicas,
						},
					},
					SDS: &SDSContainerConfig{
						EnabledField: EnabledField{
							Enabled: &c.Gateways.IngressConfig.SDS.Enabled,
						},
						Image: c.Gateways.IngressConfig.SDS.Image,
					},
					ServiceAnnotations: c.Gateways.IngressConfig.ServiceAnnotations,
				},
				"istio-egressgateway": {
					DeploymentFields: DeploymentFields{
						ReplicaCount: &c.Gateways.EgressConfig.ReplicaCount,
						HorizontalPodAutoscalerFields: HorizontalPodAutoscalerFields{
							AutoscaleMin: &c.Gateways.EgressConfig.MinReplicas,
							AutoscaleMax: &c.Gateways.EgressConfig.MaxReplicas,
						},
					},
					SDS: &SDSContainerConfig{
						EnabledField: EnabledField{
							Enabled: &c.Gateways.EgressConfig.SDS.Enabled,
						},
						Image: c.Gateways.EgressConfig.SDS.Image,
					},
					ServiceAnnotations: c.Gateways.EgressConfig.ServiceAnnotations,
				},
			},
		},
		NodeAgent: &NodeAgentConfig{
			CommonComponentConfig: CommonComponentConfig{
				EnabledField: EnabledField{
					Enabled: &c.NodeAgent.Enabled,
				},
				Global: globalConfig,
			},
			DeploymentFields: DeploymentFields{
				Image: c.NodeAgent.Image,
				Env: map[string]string{
					"CA_PROVIDER": "Citadel",
					"CA_ADDR":     "istio-citadel:8060",
					"VALID_TOKEN": "true",
				},
			},
		},
		Prometheus: &PrometheusConfig{
			CommonComponentConfig: CommonComponentConfig{
				EnabledField: EnabledField{
					Enabled: util.BoolPointer(false),
				},
			},
		},
	}
}

func defaultConfigVisibilitySettings(s string) []string {
	if s != "" {
		return []string{s}
	}
	return nil
}
