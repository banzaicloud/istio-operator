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

package gateways

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestEnsureHealthProbePort_Empty(t *testing.T) {
	var ports []apiv1.ServicePort
	ports = ensureHealthProbePort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "status-port",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_HappyPath(t *testing.T) {
	ports := []apiv1.ServicePort{
		{
			Name:       "http",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
		{
			Name:       "https",
			Protocol:   "TCP",
			Port:       443,
			TargetPort: intstr.FromInt(4430),
		},
	}
	ports = ensureHealthProbePort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "http",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
		{
			Name:       "https",
			Protocol:   "TCP",
			Port:       443,
			TargetPort: intstr.FromInt(4430),
		},
		{
			Name:       "status-port",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_MixedProtocol(t *testing.T) {
	// when service type is LoadBalancer, protocols cannot be mixed. See ensureHealthProbePort() for more info
	ports := []apiv1.ServicePort{
		{
			Name:       "udp",
			Protocol:   "UDP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
		{
			Name:       "sctp",
			Protocol:   "SCTP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	ports = ensureHealthProbePort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "udp",
			Protocol:   "UDP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
		{
			Name:       "sctp",
			Protocol:   "SCTP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_PortMismatch(t *testing.T) {
	ports := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       1234,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	ports = ensureHealthProbePort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       1234,
			TargetPort: intstr.FromInt(healthProbePort),
		},
		{
			Name:       "status-port",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(healthProbePort),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_TargetPortMismatch(t *testing.T) {
	originalPorts := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(1234),
		},
	}
	ports := ensureHealthProbePort(append([]apiv1.ServicePort{}, originalPorts...))

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(1234),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_PortNameTaken(t *testing.T) {
	originalPorts := []apiv1.ServicePort{
		{
			Name:       "sTaTuS-pOrT",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
	}
	ports := ensureHealthProbePort(append([]apiv1.ServicePort{}, originalPorts...))

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "sTaTuS-pOrT",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureHealthProbePort_PortNameAndTargetPort(t *testing.T) {
	originalPorts := []apiv1.ServicePort{
		{
			Name:       "sTaTuS-pOrT",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(1234),
		},
	}
	ports := ensureHealthProbePort(append([]apiv1.ServicePort{}, originalPorts...))

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "sTaTuS-pOrT",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       healthProbePort,
			TargetPort: intstr.FromInt(1234),
		},
	}
	require.Equal(t, expectedPorts, ports)
}
