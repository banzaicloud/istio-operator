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
	"bytes"
	"net"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

type IngressSetupPendingError struct{}

func (e IngressSetupPendingError) Error() string {
	return "ingress gateway endpoint address is pending"
}

func GetServiceEndpointIPs(service corev1.Service) ([]string, error) {
	ips := make([]string, 0)

	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		if service.Spec.ClusterIP != corev1.ClusterIPNone {
			ips = []string{
				service.Spec.ClusterIP,
			}
		}
	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) < 1 {
			return ips, IngressSetupPendingError{}
		}

		if service.Status.LoadBalancer.Ingress[0].IP != "" {
			ips = []string{
				service.Status.LoadBalancer.Ingress[0].IP,
			}
		} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
			hostIPs, err := net.LookupIP(service.Status.LoadBalancer.Ingress[0].Hostname)
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
		}
	}

	return ips, nil
}
