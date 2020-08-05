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
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) daemonSet() runtime.Object {
	labels := util.MergeMultipleStringMaps(nodeAgentLabels, labelSelector, r.Config.RevisionLabels())
	hostPathType := apiv1.HostPathUnset
	om := templates.ObjectMetaWithRevision(daemonSetName, labels, r.Config)
	om.Annotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	return &appsv1.DaemonSet{
		ObjectMeta: om,
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
							Image:           util.PointerToString(r.Config.Spec.NodeAgent.Image),
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
								{
									Name:  "TRUST_DOMAIN",
									Value: r.Config.Spec.TrustDomain,
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath:  "metadata.namespace",
											APIVersion: "v1",
										},
									},
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
					Affinity:          r.Config.Spec.NodeAgent.Affinity,
					NodeSelector:      r.Config.Spec.NodeAgent.NodeSelector,
					Tolerations:       r.Config.Spec.NodeAgent.Tolerations,
					PriorityClassName: r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}
