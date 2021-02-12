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
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) daemonSet() runtime.Object {
	labels := util.MergeStringMaps(cniLabels, labelSelector)
	hostPathType := apiv1.HostPathUnset
	return &appsv1.DaemonSet{
		ObjectMeta: templates.ObjectMetaWithRevision(daemonSetName, labels, r.Config),
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeStringMaps(labels, r.Config.RevisionLabels()),
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: util.IntstrPointer(1),
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeMultipleStringMaps(labels, r.Config.RevisionLabels(), v1beta1.DisableInjectionLabel),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
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
					ServiceAccountName:            r.Config.WithRevision(serviceAccountName),
					Containers:                    r.container(),
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
					Affinity:          r.Config.Spec.SidecarInjector.InitCNIConfiguration.Affinity,
					PriorityClassName: r.Config.Spec.PriorityClassName,
					ImagePullSecrets:  r.Config.Spec.ImagePullSecrets,
				},
			},
		},
	}
}

func (r *Reconciler) container() []apiv1.Container {
	cniConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration
	containers := []apiv1.Container{
		{
			Name:            "install-cni",
			Image:           cniConfig.Image,
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
			Command: []string{"install-cni"},
			Env: []apiv1.EnvVar{
				{
					Name: "CNI_NETWORK_CONFIG",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: r.Config.WithRevision(configMapName),
							},
							Key: "cni_network_config",
						},
					},
				},
				{
					Name:  "CNI_NET_DIR",
					Value: "/etc/cni/net.d",
				},
				{
					Name:  "CHAINED_CNI_PLUGIN",
					Value: strconv.FormatBool(util.PointerToBool(cniConfig.Chained)),
				},
				{
					Name:  "KUBECFG_FILE_NAME",
					Value: r.Config.WithRevision("ZZZ-istio-cni-kubeconfig"),
				},
				{
					Name:  "CNI_CONFIG_NAME",
					Value: r.Config.WithRevision("istio-cni"),
				},
			},
			TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
			TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
			LivenessProbe: &apiv1.Probe{
				Handler: apiv1.Handler{
					HTTPGet: &apiv1.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(8000),
						Scheme: apiv1.URISchemeHTTP,
					},
				},
				InitialDelaySeconds: 5,
			},
			ReadinessProbe: &apiv1.Probe{
				Handler: apiv1.Handler{
					HTTPGet: &apiv1.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.FromInt(8000),
						Scheme: apiv1.URISchemeHTTP,
					},
				},
			},
		},
	}

	if util.PointerToBool(cniConfig.Repair.Enabled) {
		image := cniConfig.Image
		containers = append(containers, apiv1.Container{
			Name:  "repair-cni",
			Image: image,
			Command: []string{
				"/opt/cni/bin/istio-cni-repair",
			},
			Env: []apiv1.EnvVar{
				{
					Name: "REPAIR_NODE-NAME",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							FieldPath: "spec.nodeName",
						},
					},
				},
				{
					Name:  "REPAIR_LABEL-PODS",
					Value: strconv.FormatBool(util.PointerToBool(cniConfig.Repair.LabelPods)),
				},
				{
					Name:  "REPAIR_DELETE-PODS",
					Value: strconv.FormatBool(util.PointerToBool(cniConfig.Repair.DeletePods)),
				},
				{
					Name:  "REPAIR_RUN-AS-DAEMON",
					Value: "true",
				},
				{
					Name:  "REPAIR_SIDECAR-ANNOTATION",
					Value: "sidecar.istio.io/status",
				},
				{
					Name:  "REPAIR_INIT-CONTAINER-NAME",
					Value: util.PointerToString(cniConfig.Repair.InitContainerName),
				},
				{
					Name:  "REPAIR_BROKEN-POD-LABEL-KEY",
					Value: util.PointerToString(cniConfig.Repair.BrokenPodLabelKey),
				},
				{
					Name:  "REPAIR_BROKEN-POD-LABEL-VALUE",
					Value: util.PointerToString(cniConfig.Repair.BrokenPodLabelValue),
				},
			},
		})
	}

	if util.PointerToBool(cniConfig.Taint.Enabled) {
		image := cniConfig.Image
		containers = append(containers, apiv1.Container{
			Name:  "taint-controller",
			Image: image,
			Command: []string{
				"/opt/cni/bin/istio-cni-taint",
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "TAINT_RUN-AS-DAEMON",
					Value: "true",
				},
				{
					Name:  "TAINT_CONFIGMAP-NAME",
					Value: r.Config.WithRevision(taintConfigMapName),
				},
				{
					Name:  "TAINT_CONFIGMAP-NAMESPACE",
					Value: r.Config.Namespace,
				},
			},
		})
	}

	return containers
}
