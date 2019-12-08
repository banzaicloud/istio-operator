package v1beta1

import "github.com/banzaicloud/istio-operator/pkg/util"

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
}
