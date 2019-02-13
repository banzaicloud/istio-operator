package gateways

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) horizontalPodAutoscaler(gw string, owner *istiov1beta1.Config) runtime.Object {
	return &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: templates.ObjectMeta(hpaName(gw), nil, owner),
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: 5,
			MinReplicas: util.IntPointer(1),
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       gatewayName(gw),
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: templates.TargetAvgCpuUtil80(),
		},
	}
}
