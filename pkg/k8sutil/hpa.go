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

package k8sutil

import (
	"context"

	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetHPAReplicaCountOrDefault get desired replica count from HPA if exists, returns the given default otherwise
func GetHPAReplicaCountOrDefault(client client.Client, name types.NamespacedName, defaultReplicaCount int32) int32 {
	var hpa autoscalev2beta1.HorizontalPodAutoscaler
	err := client.Get(context.Background(), name, &hpa)
	if err != nil {
		return defaultReplicaCount
	}

	return hpa.Status.DesiredReplicas
}
