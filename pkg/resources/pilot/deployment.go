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
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

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

	if r.Config.Spec.ControlPlaneSecurityEnabled && !util.PointerToBool(r.Config.Spec.Pilot.Sidecar) {
		containerArgs = append(containerArgs, "--secureGrpcAddr", ":15011")
	} else {
		containerArgs = append(containerArgs, "--secureGrpcAddr", "")
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

	containers := []apiv1.Container{
		discoveryContainer,
	}

	args := []string{
		"proxy",
		"--serviceCluster",
		"istio-pilot",
		"--templateFile",
		"/etc/istio/proxy/envoy_pilot.yaml.tmpl",
		"--controlPlaneAuthPolicy",
		templates.ControlPlaneAuthPolicy(false, r.Config.Spec.ControlPlaneSecurityEnabled),
		"--domain",
		r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
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
			Args: args,
			Env:  templates.IstioProxyEnv(r.Config),
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

		if r.Config.Spec.Proxy.LogLevel != "" {
			proxyContainer.Args = append(proxyContainer.Args, fmt.Sprintf("--proxyLogLevel=%s", r.Config.Spec.Proxy.LogLevel))
		}
		if r.Config.Spec.Proxy.ComponentLogLevel != "" {
			proxyContainer.Args = append(proxyContainer.Args, fmt.Sprintf("--proxyComponentLogLevel=%s", r.Config.Spec.Proxy.ComponentLogLevel))
		}

		containers = append(containers, proxyContainer)
	}

	return containers
}

func (r *Reconciler) deployment() runtime.Object {
	labels := util.MergeStringMaps(pilotLabels, labelSelector)
	deployment := &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, labels, r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(k8sutil.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName,
				Namespace: r.Config.Namespace,
			}, util.PointerToInt32(r.Config.Spec.Pilot.ReplicaCount))),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.Pilot.PodAnnotations),
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
										Name: base.IstioConfigMapName,
									},
									DefaultMode: util.IntPointer(420),
								},
							},
						},
					},
					Affinity:          r.Config.Spec.Pilot.Affinity,
					NodeSelector:      r.Config.Spec.Pilot.NodeSelector,
					Tolerations:       r.Config.Spec.Pilot.Tolerations,
					PriorityClassName: r.Config.Spec.PriorityClassName,
				},
			},
		},
	}

	return deployment
}
