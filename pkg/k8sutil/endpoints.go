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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateK8sEndpoints(name string, namespace string, addresses []corev1.EndpointAddress, ports []corev1.EndpointPort) *corev1.Endpoints {
	return &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: addresses,
				Ports:     ports,
			},
		},
	}
}

func GetEndpoints(ctx context.Context, kubeClient client.Client, name string, namespace string) (*corev1.Endpoints, error) {
	endpoints := &corev1.Endpoints{}
	err := kubeClient.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, endpoints)
	if err != nil {
		return endpoints, errors.WithStackIf(err)
	}

	return endpoints, nil
}

func GetIPsForEndpoints(endpoints *corev1.Endpoints) []string {
	var endpointAddresses []string
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			endpointAddresses = append(endpointAddresses, address.IP)
		}
	}

	return endpointAddresses
}
