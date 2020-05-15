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
	autoscalev2beta2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) horizontalPodAutoscaler() runtime.Object {
	return &autoscalev2beta2.HorizontalPodAutoscaler{
		ObjectMeta: templates.ObjectMeta(r.hpaName(), r.labelSelector(), r.gw),
		Spec: autoscalev2beta2.HorizontalPodAutoscalerSpec{
			MaxReplicas: *r.gw.Spec.MaxReplicas,
			MinReplicas: r.gw.Spec.MinReplicas,
			ScaleTargetRef: autoscalev2beta2.CrossVersionObjectReference{
				Name:       r.gatewayName(),
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: templates.TargetAvgCpuUtil80(),
		},
	}
}
