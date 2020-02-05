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
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/ingressgateway"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) SetIngressGatewayAddress(remoteIstio *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	if !util.PointerToBool(istio.Spec.Gateways.Enabled) || !util.PointerToBool(istio.Spec.Gateways.IngressConfig.Enabled) || !util.PointerToBool(istio.Spec.MeshExpansion) {
		return nil
	}

	var err error
	remoteIstio.Status.GatewayAddress, err = k8sutil.GetMeshGatewayAddress(c.ctrlRuntimeClient, client.ObjectKey{
		Name:      ingressgateway.ResourceName,
		Namespace: remoteIstio.Namespace,
	})

	return err
}
