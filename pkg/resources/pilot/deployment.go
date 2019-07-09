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

package pilot

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var appLabels = map[string]string{
	"app": "pilot",
}

func (r *Reconciler) containerArgs() []string {

	containerArgs := []string{
		"discovery",
		"--monitoringAddr=:15014",
		"--domain",
		"cluster.local",
		"--keepaliveMaxServerConnectionAge",
		"30m",
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled {
		if !util.PointerToBool(r.Config.Spec.Pilot.Sidecar) {
			containerArgs = append(containerArgs, "--secureGrpcAddr", ":15011")
		}
	} else {
		containerArgs = append(containerArgs, "--secureGrpcAddr", "")
	}
	if r.Config.Spec.WatchOneNamespace {
		containerArgs = append(containerArgs, "-a", r.Config.Namespace)
	}

	return containerArgs
}

func (r *Reconciler) containerPorts() []apiv1.ContainerPort {

	containerPorts := []apiv1.ContainerPort{
		{ContainerPort: 8080, Protocol: apiv1.ProtocolTCP},
		{ContainerPort: 15010, Protocol: apiv1.ProtocolTCP},
	}

	if !util.PointerToBool(r.Config.Spec.Pilot.Sidecar) {
		containerPorts = append(containerPorts, apiv1.ContainerPort{ContainerPort: 15011, Protocol: apiv1.ProtocolTCP})
	}

	return containerPorts
}

func (r *Reconciler) containers() []apiv1.Container {
	discoveryContainer := apiv1.Container{
		Name:            "discovery",
		Image:           r.Config.Spec.Pilot.Image,
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
			PeriodSeconds:       30,
			TimeoutSeconds:      5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
		Env: []apiv1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &apiv1.EnvVarSource{
					FieldRef: &apiv1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
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
			{Name: "PILOT_PUSH_THROTTLE", Value: "100"},
			{Name: "GODEBUG", Value: "gctrace=2"},
			{
				Name:  "PILOT_TRACE_SAMPLING",
				Value: fmt.Sprintf("%.2f", r.Config.Spec.Pilot.TraceSampling),
			},
			{Name: "PILOT_DISABLE_XDS_MARSHALING_TO_ANY", Value: "1"},
			{Name: "MESHNETWORKS_HASH", Value: r.Config.Spec.GetMeshNetworksHash()},
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Pilot.Resources,
			r.Config.Spec.DefaultResources,
		),
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "config-volume",
				MountPath: "/etc/istio/config",
			},
			{
				Name:      "istio-certs",
				MountPath: "/etc/certs",
				ReadOnly:  true,
			},
		},
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}

	if util.PointerToBool(r.Config.Spec.LocalityLB.Enabled) {
		discoveryContainer.Env = append(discoveryContainer.Env, apiv1.EnvVar{
			Name:  "PILOT_ENABLE_LOCALITY_LOAD_BALANCING",
			Value: "1",
		})
	}

	containers := []apiv1.Container{
		discoveryContainer,
	}

	if util.PointerToBool(r.Config.Spec.Pilot.Sidecar) {
		proxyContainer := apiv1.Container{
			Name:            "istio-proxy",
			Image:           r.Config.Spec.Proxy.Image,
			ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
			Ports: []apiv1.ContainerPort{
				{ContainerPort: 15003, Protocol: apiv1.ProtocolTCP},
				{ContainerPort: 15005, Protocol: apiv1.ProtocolTCP},
				{ContainerPort: 15007, Protocol: apiv1.ProtocolTCP},
				{ContainerPort: 15011, Protocol: apiv1.ProtocolTCP},
			},
			Args: []string{
				"proxy",
				"--serviceCluster",
				"istio-pilot",
				"--templateFile",
				"/etc/istio/proxy/envoy_pilot.yaml.tmpl",
				"--controlPlaneAuthPolicy",
				templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
				"--domain",
				r.Config.Namespace + ".svc.cluster.local",
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
			},
			TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
			TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
		}

		if util.PointerToBool(r.Config.Spec.SDS.Enabled) {
			proxyContainer.VolumeMounts = append(proxyContainer.VolumeMounts, apiv1.VolumeMount{
				Name:      "sds-uds-path",
				MountPath: "/var/run/sds",
				ReadOnly:  true,
			})

			if r.Config.Spec.SDS.UseTrustworthyJwt {
				proxyContainer.VolumeMounts = append(proxyContainer.VolumeMounts, apiv1.VolumeMount{
					Name:      "istio-token",
					MountPath: "/var/run/secrets/tokens",
				})
			}
		}

		containers = append(containers, proxyContainer)
	}

	return containers
}

func (r *Reconciler) deployment() runtime.Object {
	deployment := &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(pilotLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName,
				Namespace: r.Config.Namespace,
			}, r.Config.Spec.Pilot.ReplicaCount)),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeLabels(appLabels, labelSelector),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeLabels(appLabels, labelSelector),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers:         r.containers(),
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
							Name: "config-volume",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: common.IstioConfigMapName,
									},
									DefaultMode: util.IntPointer(420),
								},
							},
						},
					},
					Affinity:     r.Config.Spec.Pilot.Affinity,
					NodeSelector: r.Config.Spec.Pilot.NodeSelector,
					Tolerations:  r.Config.Spec.Pilot.Tolerations,
				},
			},
		},
	}

	if util.PointerToBool(r.Config.Spec.SDS.Enabled) {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, apiv1.Volume{
			Name: "sds-uds-path",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/run/sds",
				},
			},
		})

		if r.Config.Spec.SDS.UseTrustworthyJwt {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, apiv1.Volume{
				Name: "istio-token",
				VolumeSource: apiv1.VolumeSource{
					Projected: &apiv1.ProjectedVolumeSource{
						Sources: []apiv1.VolumeProjection{
							{
								ServiceAccountToken: &apiv1.ServiceAccountTokenProjection{
									Path:              "istio-token",
									ExpirationSeconds: util.Int64Pointer(43200),
									Audience:          "cluster.local",
								},
							},
						},
						DefaultMode: util.IntPointer(420),
					},
				},
			})
		}
	}

	return deployment
}
