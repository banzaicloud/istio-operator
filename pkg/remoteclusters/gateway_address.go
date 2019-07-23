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

package remoteclusters

import (
	"context"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) getIngressGatewayAddress(remoteIstio *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	if !util.PointerToBool(istio.Spec.MeshExpansion) {
		return nil
	}

	c.log.Info("get ingress gateway address")

	var service corev1.Service
	ips := make([]string, 0)

	err := c.ctrlRuntimeClient.Get(context.Background(), types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: remoteIstio.Namespace,
	}, &service)
	if err != nil {
		return err
	}

	if len(service.Status.LoadBalancer.Ingress) < 1 {
		return IngressSetupPendingError{}
	}

	if service.Status.LoadBalancer.Ingress[0].IP != "" {
		ips = []string{
			service.Status.LoadBalancer.Ingress[0].IP,
		}
	} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
		hostIPs, err := net.LookupIP(service.Status.LoadBalancer.Ingress[0].Hostname)
		if err != nil {
			return err
		}
		for _, ip := range hostIPs {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	remoteIstio.Status.GatewayAddress = ips

	return nil
}
