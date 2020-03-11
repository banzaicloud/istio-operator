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

package templates

import (
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

func DefaultDeployAnnotations() map[string]string {
	return map[string]string{
		"sidecar.istio.io/inject":                    "false",
		"scheduler.alpha.kubernetes.io/critical-pod": "",
	}
}

func GetResourcesRequirementsOrDefault(requirements *apiv1.ResourceRequirements, defaults *apiv1.ResourceRequirements) apiv1.ResourceRequirements {
	if requirements != nil {
		return *requirements
	}

	return *defaults
}

func DefaultRollingUpdateStrategy() appsv1.DeploymentStrategy {
	return appsv1.DeploymentStrategy{
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       util.IntstrPointer(1),
			MaxUnavailable: util.IntstrPointer(0),
		},
	}
}

func TargetAvgCpuUtil80() []autoscalev2beta1.MetricSpec {
	return []autoscalev2beta1.MetricSpec{
		{
			Type: autoscalev2beta1.ResourceMetricSourceType,
			Resource: &autoscalev2beta1.ResourceMetricSource{
				Name:                     apiv1.ResourceCPU,
				TargetAverageUtilization: util.IntPointer(80),
			},
		},
	}
}

func IstioProxyEnv(config *v1beta1.Istio) []apiv1.EnvVar {
	envs := []apiv1.EnvVar{
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
		{
			Name: "INSTANCE_IP",
			ValueFrom: &apiv1.EnvVarSource{
				FieldRef: &apiv1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}

	envs = append(envs, apiv1.EnvVar{
		Name:  "SDS_ENABLED",
		Value: strconv.FormatBool(util.PointerToBool(config.Spec.SDS.Enabled)),
	})

	if util.PointerToBool(config.Spec.Istiod.Enabled) {
		envs = append(envs, []apiv1.EnvVar{
			{
				Name:  "JWT_POLICY",
				Value: string(config.Spec.JWTPolicy),
			},
			{
				Name:  "PILOT_CERT_PROVIDER",
				Value: string(config.Spec.Pilot.CertProvider),
			},
		}...)
	}

	return envs
}
