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

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) deployment(t string) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName(t), labelSelector, r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName(t),
				Namespace: r.Config.Namespace,
			}, r.Config.Spec.Mixer.ReplicaCount)),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeLabels(labelSelector, util.MergeLabels(appLabel(t), mixerTypeLabel(t))),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeLabels(labelSelector, util.MergeLabels(appLabel(t), mixerTypeLabel(t))),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Volumes: []apiv1.Volume{
						{
							Name: "istio-certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName:  fmt.Sprintf("istio.%s", serviceAccountName),
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
					},
					Affinity:     r.Config.Spec.Mixer.Affinity,
					NodeSelector: r.Config.Spec.Mixer.NodeSelector,
					Tolerations:  r.Config.Spec.Mixer.Tolerations,
					Containers: []apiv1.Container{
						r.mixerContainer(t, r.Config.Namespace),
						r.istioProxyContainer(t),
					},
				},
			},
		},
	}
}

func (r *Reconciler) mixerContainer(t string, ns string) apiv1.Container {
	containerArgs := []string{
		"--address",
		"unix:///sock/mixer.socket",
		"--configDefaultNamespace",
		ns,
		"--monitoringPort",
		"15014",
	}

	if util.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		containerArgs = append(containerArgs, "--trace_zipkin_url",
			"http://"+r.Config.Spec.Tracing.Zipkin.Address+"/api/v1/spans")
	}

	if r.Config.Spec.UseMCP {
		if r.Config.Spec.ControlPlaneSecurityEnabled {
			containerArgs = append(containerArgs, "--configStoreURL", "mcps://istio-galley."+r.Config.Namespace+".svc:9901")
			if t == "telemetry" {
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

	if t == "telemetry" {
		containerArgs = append(containerArgs, "--averageLatencyThreshold", "100ms")
		containerArgs = append(containerArgs, "--loadsheddingMode", "enforce")
	}

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
	if r.Config.Spec.UseMCP {
		volumeMounts = append(volumeMounts, apiv1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	return apiv1.Container{
		Name:            t,
		Image:           r.Config.Spec.Mixer.Image,
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
		Args: containerArgs,
		Env: []apiv1.EnvVar{
			{
				Name:  "GODEBUG",
				Value: "gctrace=2",
			},
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Mixer.Resources,
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
	}
}

func (r *Reconciler) istioProxyContainer(t string) apiv1.Container {
	return apiv1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Ports: []apiv1.ContainerPort{
			{ContainerPort: 9091, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15004, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		},
		Args: []string{
			"proxy",
			"--serviceCluster",
			fmt.Sprintf("istio-%s", t),
			"--templateFile",
			fmt.Sprintf("/etc/istio/proxy/envoy_%s.yaml.tmpl", t),
			"--controlPlaneAuthPolicy",
			templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
			"--domain",
			"$(POD_NAMESPACE).svc.cluster.local",
		},
		Env: templates.IstioProxyEnv(),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Proxy.Resources,
			r.Config.Spec.DefaultResources,
		),
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "istio-certs",
				MountPath: "/etc/certs",
				ReadOnly:  true,
			},
			{
				Name:      "uds-socket",
				MountPath: "/sock",
			},
		},
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}
