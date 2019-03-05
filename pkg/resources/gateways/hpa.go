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

package gateways

import (
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) horizontalPodAutoscaler(gw string) runtime.Object {
	gwConfig := r.getGatewayConfig(gw)
	return &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: templates.ObjectMeta(hpaName(gw), nil, r.Config),
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: gwConfig.MaxReplicas,
			MinReplicas: &gwConfig.MinReplicas,
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       gatewayName(gw),
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: templates.TargetAvgCpuUtil80(),
		},
	}
}
