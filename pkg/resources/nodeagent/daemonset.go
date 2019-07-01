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

package nodeagent

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) daemonSet() runtime.Object {
	labels := util.MergeLabels(nodeAgentLabels, labelSelector)
	hostPathType := apiv1.HostPathUnset
	return &appsv1.DaemonSet{
		ObjectMeta: templates.ObjectMeta(daemonSetName, labels, r.Config),
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []apiv1.Container{
						{
							Name:            "nodeagent",
							Image:           r.Config.Spec.NodeAgent.Image,
							ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "sdsudspath",
									MountPath: "/var/run/sds",
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CA_PROVIDER",
									Value: "Citadel",
								},
								{
									Name:  "CA_ADDR",
									Value: "istio-citadel:8060",
								},
								{
									Name:  "VALID_TOKEN",
									Value: "true",
								},
							},
							Resources: templates.GetResourcesRequirementsOrDefault(
								r.Config.Spec.NodeAgent.Resources,
								r.Config.Spec.DefaultResources,
							),
							TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
							TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "sdsudspath",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/var/run/sds",
									Type: &hostPathType,
								},
							},
						},
					},
					Affinity:     r.Config.Spec.NodeAgent.Affinity,
					NodeSelector: r.Config.Spec.NodeAgent.NodeSelector,
					Tolerations:  r.Config.Spec.NodeAgent.Tolerations,
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
		},
	}
}

func (r *Reconciler) galleyProbe(path string) *apiv1.Probe {
	return &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{
					"/usr/local/bin/galley",
					"probe",
					fmt.Sprintf("--probe-path=%s", path),
					"--interval=10s",
				},
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
		FailureThreshold:    3,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}
}
