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

package v1beta1

import (
	apiv1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (gw *MeshGateway) SetDefaults() {
	if gw.Spec.ReplicaCount == nil {
		gw.Spec.ReplicaCount = util.IntPointer(defaultReplicaCount)
	}
	if gw.Spec.MinReplicas == nil {
		gw.Spec.MinReplicas = util.IntPointer(defaultReplicaCount)
	}
	if gw.Spec.MaxReplicas == nil {
		gw.Spec.MaxReplicas = util.IntPointer(defaultReplicaCount)
	}
	if gw.Spec.Resources == nil {
		gw.Spec.Resources = defaultProxyResources
	}
	if gw.Spec.SDS.Enabled == nil {
		gw.Spec.SDS.Enabled = util.BoolPointer(false)
	}
	if gw.Spec.SDS.Image == "" {
		gw.Spec.SDS.Image = defaultNodeAgentImage
	}
	if gw.Spec.Type == GatewayTypeIngress && gw.Spec.ServiceType == "" {
		gw.Spec.ServiceType = defaultIngressGatewayServiceType
	}
	if gw.Spec.Type == GatewayTypeEgress && gw.Spec.ServiceType == "" {
		gw.Spec.ServiceType = defaultEgressGatewayServiceType
	}
	// always turn off SDS for egress
	if gw.Spec.Type == GatewayTypeEgress {
		gw.Spec.SDS.Enabled = util.BoolPointer(false)
	}
	if gw.Spec.RunAsRoot == nil {
		gw.Spec.RunAsRoot = util.BoolPointer(true)
	}
	if gw.Spec.SecurityContext == nil {
		if util.PointerToBool(gw.Spec.RunAsRoot) {
			gw.Spec.SecurityContext = &apiv1.SecurityContext{}
		} else {
			gw.Spec.SecurityContext = defaultSecurityContext
		}
	}

	if gw.Spec.Labels == nil {
		gw.Spec.Labels = make(map[string]string)
	}

	gw.Spec.Labels["gateway-name"] = gw.Name
	gw.Spec.Labels["gateway-type"] = string(gw.Spec.Type)
}
