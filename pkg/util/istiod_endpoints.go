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

package util

import (
	"context"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
)

func GetIstiodEndpointAddresses(ctx context.Context, kubeClient client.Client, icpName string, namespace string) ([]corev1.EndpointAddress, error) {
	var gatewayAddresses []corev1.EndpointAddress

	picpList := &servicemeshv1alpha1.PeerIstioControlPlaneList{}
	err := kubeClient.List(ctx, picpList, client.InNamespace(namespace))
	if err != nil {
		return gatewayAddresses, errors.WithStackIf(err)
	}

	for _, picp := range picpList.Items {
		if picp.Status.IstioControlPlaneName == icpName {
			// TODO: mgw ip in different network, istiod pod ip in case of flat network
			for _, address := range picp.Status.GatewayAddress {
				gatewayAddresses = append(gatewayAddresses,
					corev1.EndpointAddress{
						IP: address,
					})
			}
		}
	}

	return gatewayAddresses, nil
}

func GetIstiodEndpointPorts(ctx context.Context, kubeClient client.Client, name string, namespace string) ([]corev1.EndpointPort, error) {
	istiodPorts := []corev1.EndpointPort{}

	service := &corev1.Service{}
	err := kubeClient.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil {
		return istiodPorts, errors.WithStackIf(err)
	}

	for _, port := range service.Spec.Ports {
		istiodPorts = append(istiodPorts, corev1.EndpointPort{
			Name:        port.Name,
			Port:        port.Port,
			Protocol:    port.Protocol,
			AppProtocol: port.AppProtocol,
		})
	}

	return istiodPorts, nil
}
