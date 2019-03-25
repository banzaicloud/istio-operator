package helm

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

func Convert(config *istiov1beta1.Istio) *IstioHelmValues {

	return &IstioHelmValues{
		Global: &GlobalConfig{
			MTLS: &MTLSConfig{
				EnabledField{
					Enabled: &config.Spec.MTLS,
				},
			},
			ControlPlaneSecurityEnabled: &config.Spec.ControlPlaneSecurityEnabled,
		},
	}
}
