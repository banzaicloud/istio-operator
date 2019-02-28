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

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) deployment(gw string) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(gatewayName(gw), labelSelector(gw), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(r.Config.Spec.Gateways.ReplicaCount),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector(gw),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labelSelector(gw),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName(gw),
					Containers: []apiv1.Container{
						{
							Name:            "istio-proxy",
							Image:           r.Config.Spec.Proxy.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"proxy",
								"router",
								"-v", "2",
								"--discoveryRefreshDelay", "1s",
								"--drainDuration", "45s",
								"--parentShutdownDuration", "1m0s",
								"--connectTimeout", "10s",
								"--serviceCluster", fmt.Sprintf("istio-%s", gw),
								"--zipkinAddress", fmt.Sprintf("zipkin.%s:9411", r.Config.Namespace),
								"--proxyAdminPort", "15000",
								"--controlPlaneAuthPolicy", templates.ControlPlaneAuthPolicy(r.Config.Spec.ControlPlaneSecurityEnabled),
								"--discoveryAddress", fmt.Sprintf("istio-pilot.%s:8080", r.Config.Namespace),
							},
							Ports: r.ports(gw),
							Env: append(templates.IstioProxyEnv(), apiv1.EnvVar{
								Name: "ISTIO_META_POD_NAME",
								ValueFrom: &apiv1.EnvVarSource{
									FieldRef: &apiv1.ObjectFieldSelector{

										FieldPath:  "metadata.name",
										APIVersion: "v1",
									},
								},
							}),
							Resources: templates.DefaultResources(),
							VolumeMounts: []apiv1.VolumeMount{
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
							},
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					Volumes: []apiv1.Volume{
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
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
}

func (r *Reconciler) ports(gw string) []apiv1.ContainerPort {
	switch gw {
	case "ingressgateway":
		return []apiv1.ContainerPort{
			{ContainerPort: 80, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 443, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 31400, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15011, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 8060, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 853, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15030, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15031, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	case "egressgateway":
		return []apiv1.ContainerPort{
			{ContainerPort: 80, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 443, Protocol: apiv1.ProtocolTCP},
			{ContainerPort: 15090, Protocol: apiv1.ProtocolTCP, Name: "http-envoy-prom"},
		}
	}
	return nil
}
