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
					},
					Affinity: &apiv1.Affinity{},
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
	c := apiv1.Container{
		Name:            "mixer",
		Image:           r.Config.Spec.Mixer.Image,
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Ports: []apiv1.ContainerPort{
			{
				ContainerPort: 9093,
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				ContainerPort: 42422,
				Protocol:      apiv1.ProtocolTCP,
			},
		},
		Args: []string{
			"--address",
			"unix:///sock/mixer.socket",
			"--configStoreURL=k8s://",
			fmt.Sprintf("--configDefaultNamespace=%s", ns),
			"--trace_zipkin_url=http://zipkin:9411/api/v1/spans",
		},
		Env: []apiv1.EnvVar{
			{
				Name:  "GODEBUG",
				Value: "gctrace=2",
			},
		},
		Resources: templates.DefaultResources(),
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "uds-socket",
				MountPath: "/sock",
			},
		},
		LivenessProbe: &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Path:   "/version",
					Port:   intstr.FromInt(9093),
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
	if t == "policy" {
		c.Args = append(c.Args, "--numCheckCacheEntries=0")
	}
	return c
}

func (r *Reconciler) istioProxyContainer(t string) apiv1.Container {
	return apiv1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: apiv1.PullIfNotPresent,
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
		},
		Env:       templates.IstioProxyEnv(),
		Resources: templates.DefaultResources(),
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
