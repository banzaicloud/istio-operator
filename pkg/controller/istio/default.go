package istio

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func defaultResources() apiv1.ResourceRequirements {
	return apiv1.ResourceRequirements{
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU: resource.MustParse("10m"),
		},
	}
}
