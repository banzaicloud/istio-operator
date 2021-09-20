/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPodsForService(ctx context.Context, kubeClient client.Client, serviceName string, serviceNamespace string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}

	service, err := GetService(ctx, kubeClient, serviceName, serviceNamespace)
	if err != nil {
		return pods, errors.WithStackIf(err)
	}

	ls, err := labels.Parse(labels.Set(service.Spec.Selector).String())
	if err != nil {
		return pods, errors.WithStackIf(err)
	}

	err = kubeClient.List(ctx, pods, client.MatchingLabelsSelector{
		Selector: ls,
	})
	if err != nil {
		return pods, errors.WithStackIf(err)
	}

	return pods, nil
}

func GetPodIPsForPodList(pods *corev1.PodList) []string {
	var podAddresses []string

	for _, pod := range pods.Items {
		ready := 0
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				ready++
			}
		}
		if len(pod.Spec.Containers) == ready {
			podAddresses = append(podAddresses, pod.Status.PodIP)
		}
	}

	return podAddresses
}
