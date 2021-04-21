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

const (
	ServiceIpAddressOverrideAnnotation = "service.banzaicloud.io/ip-address-override"
	ServiceHostnameOverrideAnnotation  = "service.banzaicloud.io/hostname-override"
)

type IngressSetupPendingError struct{}

func (e IngressSetupPendingError) Error() string {
	return "ingress gateway endpoint address is pending"
}

func GetServiceEndpointIPs(service corev1.Service) ([]string, bool, error) {
	ips := make([]string, 0)

	var hostname string

	if v, ok := service.GetAnnotations()[ServiceHostnameOverrideAnnotation]; ok && v != "" {
		hostname = v
	}

	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		if hostname == "" && service.Spec.ClusterIP != corev1.ClusterIPNone {
			ips = []string{
				service.Spec.ClusterIP,
			}
		}
	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) < 1 {
			return ips, hostname != "", IngressSetupPendingError{}
		}

		if hostname == "" {
			if service.Status.LoadBalancer.Ingress[0].IP != "" {
				ips = []string{
					service.Status.LoadBalancer.Ingress[0].IP,
				}
			} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
				hostname = service.Status.LoadBalancer.Ingress[0].Hostname
			}
		}
	}

	if v, ok := service.GetAnnotations()[ServiceIpAddressOverrideAnnotation]; ok && v != "" {
		return []string{v}, false, nil
	}

	if hostname != "" {
		hostIPs, err := getIPsForHostname(hostname)
		if err != nil {
			return ips, hostname != "", err
		}
		ips = append(ips, hostIPs...)

	}

	return ips, hostname != "", nil
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
