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

package citadel

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) containerArgs() []string {
	var containerArgs []string

	if util.PointerToBool(r.Config.Spec.SDS.Enabled) {
		containerArgs = append(containerArgs, "--sds-enabled=true")
	}

	containerArgs = append(containerArgs,
		"--append-dns-names=true",
		"--grpc-port=8060",
		fmt.Sprintf("--citadel-storage-namespace=%s", r.Config.Namespace),
		fmt.Sprintf("--custom-dns-names=istio-pilot-service-account.%[1]s:istio-pilot.%[1]s", r.Config.Namespace),
		"--monitoring-port=15014",
		fmt.Sprintf("--trust-domain=%s", r.Config.Spec.TrustDomain),
	)

	if r.Config.Spec.Citadel.CASecretName == "" {
		containerArgs = append(containerArgs, "--self-signed-ca=true")
	} else {
		containerArgs = append(containerArgs,
			"--self-signed-ca=false",
			"--signing-cert=/etc/cacerts/ca-cert.pem",
			"--signing-key=/etc/cacerts/ca-key.pem",
			"--root-cert=/etc/cacerts/root-cert.pem",
			"--cert-chain=/etc/cacerts/cert-chain.pem",
		)
	}

	if util.PointerToBool(r.Config.Spec.Citadel.HealthCheck) {
		containerArgs = append(containerArgs,
			"--liveness-probe-path=/tmp/ca.liveness",
			"--liveness-probe-interval=60s",
			"--probe-check-interval=15s",
		)
	}

	if r.Config.Spec.Citadel.WorkloadCertTTL != "" {
		containerArgs = append(containerArgs,
			"--workload-cert-ttl",
			r.Config.Spec.Citadel.WorkloadCertTTL,
		)
	}

	if r.Config.Spec.Citadel.MaxWorkloadCertTTL != "" {
		containerArgs = append(containerArgs,
			"--max-workload-cert-ttl",
			r.Config.Spec.Citadel.MaxWorkloadCertTTL,
		)
	}

	if len(r.Config.Spec.Citadel.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.Citadel.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) containerEnvs() []apiv1.EnvVar {
	envs := []apiv1.EnvVar{
		{
			Name:  "CITADEL_ENABLE_NAMESPACES_BY_DEFAULT",
			Value: strconv.FormatBool(util.PointerToBool(r.Config.Spec.Citadel.EnableNamespacesByDefault)),
		},
	}

	envs = k8sutil.MergeEnvVars(envs, r.Config.Spec.Citadel.AdditionalEnvVars)

	return envs
}

func (r *Reconciler) deployment() runtime.Object {

	var citadelContainer = apiv1.Container{
		Name:            "citadel",
		Image:           util.PointerToString(r.Config.Spec.Citadel.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args:            r.containerArgs(),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Citadel.Resources,
			r.Config.Spec.DefaultResources,
		),
		Env:                      r.containerEnvs(),
		TerminationMessagePath:   apiv1.TerminationMessagePathDefault,
		TerminationMessagePolicy: apiv1.TerminationMessageReadFile,
	}

	if util.PointerToBool(r.Config.Spec.Citadel.HealthCheck) {
		citadelContainer.LivenessProbe = &apiv1.Probe{
			Handler: apiv1.Handler{
				Exec: &apiv1.ExecAction{
					Command: []string{
						"/usr/local/bin/istio_ca",
						"probe",
						"--probe-path=/tmp/ca.liveness",
						"--interval=125s",
					},
				},
			},
			InitialDelaySeconds: 60,
			PeriodSeconds:       60,
			FailureThreshold:    30,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		}
	}

	if r.Config.Spec.Citadel.CASecretName != "" {
		citadelContainer.VolumeMounts = []apiv1.VolumeMount{
			{
				Name:      "cacerts",
				MountPath: "/etc/cacerts",
				ReadOnly:  true,
			},
		}
	}

	var podSpec = apiv1.PodSpec{
		ServiceAccountName:            serviceAccountName,
		DNSPolicy:                     apiv1.DNSClusterFirst,
		RestartPolicy:                 apiv1.RestartPolicyAlways,
		TerminationGracePeriodSeconds: util.Int64Pointer(int64(30)),
		SecurityContext:               &apiv1.PodSecurityContext{},
		SchedulerName:                 "default-scheduler",
		Containers: []apiv1.Container{
			citadelContainer,
		},
		Affinity:          r.Config.Spec.Citadel.Affinity,
		NodeSelector:      r.Config.Spec.Citadel.NodeSelector,
		Tolerations:       r.Config.Spec.Citadel.Tolerations,
		PriorityClassName: r.Config.Spec.PriorityClassName,
	}

	var optional = false
	if r.Config.Spec.Citadel.CASecretName != "" {
		podSpec.Volumes = []apiv1.Volume{
			{
				Name: "cacerts",
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName:  r.Config.Spec.Citadel.CASecretName,
						Optional:    &optional,
						DefaultMode: util.IntPointer(420),
					},
				},
			},
		}
	}

	var deployment = &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeStringMaps(citadelLabels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.MergeStringMaps(citadelLabels, labelSelector),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      util.MergeStringMaps(citadelLabels, labelSelector),
					Annotations: util.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.Citadel.PodAnnotations),
				},
				Spec: podSpec,
			},
		},
	}

	return deployment
}
