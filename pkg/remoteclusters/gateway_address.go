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

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/ingressgateway"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) SetIngressGatewayAddress(remoteIstio *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	if !util.PointerToBool(istio.Spec.Gateways.Enabled) || !util.PointerToBool(istio.Spec.Gateways.IngressConfig.Enabled) || !util.PointerToBool(istio.Spec.MeshExpansion) {
		return nil
	}

	c.log.Info("get ingress gateway address")

	var mgw istiov1beta1.MeshGateway
	var ips []string

	err := c.ctrlRuntimeClient.Get(context.Background(), types.NamespacedName{
		Name:      ingressgateway.ResourceName,
		Namespace: remoteIstio.Namespace,
	}, &mgw)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if mgw.Status.Status != istiov1beta1.Available {
		return errors.New("gateway is pending")
	}

	ips = mgw.Status.GatewayAddress

	remoteIstio.Status.GatewayAddress = ips

	return nil
}
