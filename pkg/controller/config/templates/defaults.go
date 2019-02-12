package templates

import (
	"github.com/banzaicloud/istio-operator/pkg/util"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultDeployAnnotations() map[string]string {
	return map[string]string{
		"sidecar.istio.io/inject":                    "false",
		"scheduler.alpha.kubernetes.io/critical-pod": "",
	}
}

func DefaultResources() apiv1.ResourceRequirements {
	return apiv1.ResourceRequirements{
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU: resource.MustParse("10m"),
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

func IstioProxyEnv() []apiv1.EnvVar {
	return []apiv1.EnvVar{
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
}
