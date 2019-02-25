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

	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var appLabels = map[string]string{
	"app": "pilot",
}

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(pilotLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
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
					Containers: []apiv1.Container{
						{
							Name:            "discovery",
							Image:           "docker.io/istio/pilot:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"discovery",
							},
							Ports: []apiv1.ContainerPort{
								{ContainerPort: 8080},
								{ContainerPort: 15010},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
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
								{Name: "PILOT_CACHE_SQUASH", Value: "5"},
								{Name: "PILOT_PUSH_THROTTLE_COUNT", Value: "100"},
								{Name: "GODEBUG", Value: "gctrace=2"},
								{Name: "PILOT_TRACE_SAMPLING", Value: "1.0"},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("500m"),
									apiv1.ResourceMemory: resource.MustParse("2048Mi"),
								},
							},
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
						},
						{
							Name:            "istio-proxy",
							Image:           "docker.io/istio/proxyv2:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{ContainerPort: 15003},
								{ContainerPort: 15005},
								{ContainerPort: 15007},
								{ContainerPort: 15011},
							},
							Args: []string{
								"proxy",
								"--serviceCluster",
								"istio-pilot",
								"--templateFile",
								"/etc/istio/proxy/envoy_pilot.yaml.tmpl",
								"--controlPlaneAuthPolicy",
								templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
							},
							Env:       templates.IstioProxyEnv(),
							Resources: templates.DefaultResources(),
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "istio-certs",
									MountPath: "/etc/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "istio-certs",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: fmt.Sprintf("istio.%s", serviceAccountName),
									Optional:   util.BoolPointer(true),
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
								},
							},
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
}
