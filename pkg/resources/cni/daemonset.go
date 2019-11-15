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

package cni

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) daemonSet() runtime.Object {
	labels := util.MergeStringMaps(cniLabels, labelSelector)
	hostPathType := apiv1.HostPathUnset
	return &appsv1.DaemonSet{
		ObjectMeta: templates.ObjectMeta(daemonSetName, labels, r.Config),
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: util.IntstrPointer(1),
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					HostNetwork: true,
					Tolerations: []apiv1.Toleration{
						{
							Operator: apiv1.TolerationOpExists,
							Effect:   apiv1.TaintEffectNoSchedule,
						},
						{
							Operator: apiv1.TolerationOpExists,
							Effect:   apiv1.TaintEffectNoExecute,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: apiv1.TolerationOpExists,
						},
					},
					TerminationGracePeriodSeconds: util.Int64Pointer(5),
					ServiceAccountName:            serviceAccountName,
					Containers: []apiv1.Container{
						{
							Name:            "install-cni",
							Image:           r.Config.Spec.SidecarInjector.InitCNIConfiguration.Image,
							ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "cni-bin-dir",
									MountPath: "/host/opt/cni/bin",
								},
								{
									Name:      "cni-net-dir",
									MountPath: "/host/etc/cni/net.d",
								},
							},
							Command: []string{"/install-cni.sh"},
							Env: []apiv1.EnvVar{
								{
									Name: "CNI_NETWORK_CONFIG",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "istio-cni-config",
											},
											Key: "cni_network_config",
										},
									},
								},
							},
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "cni-bin-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: r.Config.Spec.SidecarInjector.InitCNIConfiguration.BinDir,
									Type: &hostPathType,
								},
							},
						},
						{
							Name: "cni-net-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: r.Config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir,
									Type: &hostPathType,
								},
							},
						},
					},
					Affinity: r.Config.Spec.SidecarInjector.InitCNIConfiguration.Affinity,
				},
			},
		},
	}
}
