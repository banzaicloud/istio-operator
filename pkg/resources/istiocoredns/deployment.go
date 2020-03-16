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

package istiocoredns

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) coreDNSContainer() apiv1.Container {
	return apiv1.Container{
		Name:            "coredns",
		Image:           util.PointerToString(r.Config.Spec.IstioCoreDNS.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args: []string{
			"-conf",
			"/etc/coredns/Corefile",
		},
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "config-volume",
				MountPath: "/etc/coredns",
				ReadOnly:  true,
			},
		},
		Ports: []apiv1.ContainerPort{
			{
				Name:          "dns",
				ContainerPort: 53,
				Protocol:      apiv1.ProtocolUDP,
			},
			{
				Name:          "dns-tcp",
				ContainerPort: 53,
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				Name:          "metrics",
				ContainerPort: 9153,
				Protocol:      apiv1.ProtocolTCP,
			},
		},
		LivenessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.FromInt(8080),
					Scheme: apiv1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 60,
			PeriodSeconds:       5,
			FailureThreshold:    5,
			SuccessThreshold:    1,
			TimeoutSeconds:      5,
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.IstioCoreDNS.Resources,
			r.Config.Spec.DefaultResources,
		),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}

func (r *Reconciler) coreDNSPluginContainer() apiv1.Container {
	return apiv1.Container{
		Name:            "istio-coredns-plugin",
		Image:           r.Config.Spec.IstioCoreDNS.PluginImage,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Command: []string{
			"/usr/local/bin/plugin",
		},
		Ports: []apiv1.ContainerPort{
			{
				Name:          "dns-grpc",
				ContainerPort: 8053,
				Protocol:      apiv1.ProtocolTCP,
			},
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.IstioCoreDNS.Resources,
			r.Config.Spec.DefaultResources,
		),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}
}

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeStringMaps(labels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: r.Config.Spec.IstioCoreDNS.ReplicaCount,
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       util.IntstrPointer(1),
					MaxUnavailable: util.IntstrPointer(0),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeStringMaps(labels, labelSelector),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeStringMaps(labels, labelSelector),
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.IstioCoreDNS.PodAnnotations),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []apiv1.Container{
						r.coreDNSContainer(),
						r.coreDNSPluginContainer(),
					},
					DNSPolicy: apiv1.DNSDefault,
					Volumes: []apiv1.Volume{
						{
							Name: "config-volume",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  "Corefile",
											Path: "Corefile",
										},
									},
									DefaultMode: util.IntPointer(420),
								},
							},
						},
					},
					Affinity:          r.Config.Spec.IstioCoreDNS.Affinity,
					NodeSelector:      r.Config.Spec.IstioCoreDNS.NodeSelector,
					Tolerations:       r.Config.Spec.IstioCoreDNS.Tolerations,
					PriorityClassName: r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}
