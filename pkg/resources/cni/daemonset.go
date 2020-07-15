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
	"fmt"
	"strconv"
	"strings"

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
					Labels:      util.MergeStringMaps(labels, r.Config.RevisionLabels()),
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
			Command: []string{"/install-cni.sh"},
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
		},
	}

	if util.PointerToBool(cniConfig.Repair.Enabled) {
		image := cniConfig.Image
		if !strings.Contains(cniConfig.Image, "/") {
			image = fmt.Sprintf("%s/%s:%s", r.repairHub(), r.repairImage(), r.repairTag())
		}
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

	return containers
}

func (r *Reconciler) repairHub() string {
	repairConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration.Repair
	if util.PointerToString(repairConfig.Hub) == "" {
		return "docker.io/istio"
	}

	return util.PointerToString(repairConfig.Hub)
}

func (r *Reconciler) repairImage() string {
	cniConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration
	if cniConfig.Image == "" {
		return "install-cni"
	}

	return cniConfig.Image
}

func (r *Reconciler) repairTag() string {
	repairConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration.Repair
	if util.PointerToString(repairConfig.Tag) == "" {
		return "1.6.7"
	}

	return util.PointerToString(repairConfig.Tag)
}
