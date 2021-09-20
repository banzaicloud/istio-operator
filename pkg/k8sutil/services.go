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
	"bytes"
	"context"
	"net"
	"sort"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceIPAddressOverrideAnnotation = "service.banzaicloud.io/ip-address-override"
	serviceHostnameOverrideAnnotation  = "service.banzaicloud.io/hostname-override"
)

type IngressSetupPendingError struct{}

func (e IngressSetupPendingError) Error() string {
	return "ingress gateway endpoint address is pending"
}

func GetServiceEndpointIPs(service corev1.Service) ([]string, bool, error) {
	ips := make([]string, 0)

	// check whether the load balancer was assigned
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) < 1 {
		return nil, false, IngressSetupPendingError{}
	}

	// ip address is overridden by annotation
	if overriddenIPAddress, ok := service.GetAnnotations()[serviceIPAddressOverrideAnnotation]; ok && overriddenIPAddress != "" {
		return []string{overriddenIPAddress}, false, nil
	}

	// hostname is overridden by annotation
	if overriddenHostname, ok := service.GetAnnotations()[serviceHostnameOverrideAnnotation]; ok && overriddenHostname != "" {
		hostIPs, err := getIPsForHostname(overriddenHostname)
		if err != nil {
			return nil, true, err
		}
		ips = append(ips, hostIPs...)

		return ips, true, nil
	}

	switch service.Spec.Type { // nolint:exhaustive
	case corev1.ServiceTypeClusterIP:
		if service.Spec.ClusterIP != corev1.ClusterIPNone {
			ips = []string{
				service.Spec.ClusterIP,
			}
		}
	case corev1.ServiceTypeLoadBalancer:
		if service.Status.LoadBalancer.Ingress[0].IP != "" {
			ips = []string{
				service.Status.LoadBalancer.Ingress[0].IP,
			}
		} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
			hostIPs, err := getIPsForHostname(service.Status.LoadBalancer.Ingress[0].Hostname)
			if err != nil {
				return nil, true, err
			}
			ips = append(ips, hostIPs...)

			return ips, true, nil
		}
	}

	return ips, false, nil
}

func getIPsForHostname(hostname string) ([]string, error) {
	ips := make([]string, 0)

	hostIPs, err := net.LookupIP(hostname)
	if err != nil {
		return ips, err
	}
	sort.Slice(hostIPs, func(i, j int) bool {
		return bytes.Compare(hostIPs[i], hostIPs[j]) < 0
	})
	for _, ip := range hostIPs {
		if ip.To4() != nil {
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

func GetService(ctx context.Context, kubeClient client.Client, serviceName string, serviceNamespace string) (*corev1.Service, error) {
	service := &corev1.Service{}
	err := kubeClient.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: serviceNamespace,
	}, service)
	if err != nil {
		return service, errors.WithStackIf(err)
	}

	return service, nil
}
