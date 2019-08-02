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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) getIngressGatewayAddress(remoteIstio *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	if !util.PointerToBool(istio.Spec.MeshExpansion) {
		return nil
	}

	c.log.Info("get ingress gateway address")

	var service corev1.Service
	var ips []string

	err := c.ctrlRuntimeClient.Get(context.Background(), types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: remoteIstio.Namespace,
	}, &service)
	if err != nil {
		return err
	}

	ips, err = k8sutil.GetServiceEndpointIPs(service)
	if err != nil {
		return err
	}

	remoteIstio.Status.GatewayAddress = ips

	return nil
}
