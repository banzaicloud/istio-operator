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

	"github.com/banzaicloud/operator-tools/pkg/utils"
)

func (c *MeshGatewayConfiguration) SetDefaults() {
	if c.ReplicaCount == nil {
		c.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if c.MinReplicas == nil {
		c.MinReplicas = utils.IntPointer(defaultReplicaCount)
	}
	if c.MaxReplicas == nil {
		c.MaxReplicas = utils.IntPointer(defaultReplicaCount)
	}
	if c.Resources == nil {
		c.Resources = defaultProxyResources
	}
	if c.SDS.Enabled == nil {
		c.SDS.Enabled = utils.BoolPointer(false)
	}
	if c.SDS.Image == "" {
		c.SDS.Image = defaultNodeAgentImage
	}
	if c.RunAsRoot == nil {
		c.RunAsRoot = utils.BoolPointer(false)
	}
	if c.SecurityContext == nil {
		if utils.PointerToBool(c.RunAsRoot) {
			c.SecurityContext = &apiv1.SecurityContext{
				RunAsUser:    utils.IntPointer64(0),
				RunAsGroup:   utils.IntPointer64(0),
				RunAsNonRoot: utils.BoolPointer(false),
			}
		} else {
			c.SecurityContext = defaultSecurityContext
		}
	}
}

func (gw *MeshGateway) SetDefaults() {
	gw.Spec.MeshGatewayConfiguration.SetDefaults()

	if gw.Spec.Type == GatewayTypeIngress && gw.Spec.ServiceType == "" {
		gw.Spec.ServiceType = defaultIngressGatewayServiceType
	}
	if gw.Spec.Type == GatewayTypeEgress && gw.Spec.ServiceType == "" {
		gw.Spec.ServiceType = defaultEgressGatewayServiceType
	}
	// always turn off SDS for egress
	if gw.Spec.Type == GatewayTypeEgress {
		gw.Spec.SDS.Enabled = utils.BoolPointer(false)
	}
	if gw.Spec.Ports == nil {
		gw.Spec.Ports = make([]ServicePort, 0)
	}

	gw.SetDefaultLabels()
}

func (gw *MeshGateway) SetDefaultLabels() {
	gw.Spec.Labels = MergeStringMaps(gw.GetDefaultLabels(), gw.Spec.Labels)
}

func (gw *MeshGateway) GetDefaultLabels() map[string]string {
	return map[string]string{
		"gateway-name": gw.Name,
		"gateway-type": string(gw.Spec.Type),
	}
}
