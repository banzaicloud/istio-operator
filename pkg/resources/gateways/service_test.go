/*
Copyright 2020 Banzai Cloud.

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

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

func TestEnsureStatusPort_Empty(t *testing.T) {
	var ports []apiv1.ServicePort
	ports = ensureStatusPort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       istiov1beta1.PortStatusPortName,
			Protocol:   "TCP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureStatusPort_HappyPath(t *testing.T) {
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
	ports = ensureStatusPort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       istiov1beta1.PortStatusPortName,
			Protocol:   "TCP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
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
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureStatusPort_MixedProtocol(t *testing.T) {
	// when service type is LoadBalancer, protocols cannot be mixed. See ensureStatusPort() for more info
	ports := []apiv1.ServicePort{
		{
			Name:       "udp",
			Protocol:   "UDP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
		{
			Name:       "sctp",
			Protocol:   "SCTP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
	}
	ports = ensureStatusPort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "udp",
			Protocol:   "UDP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
		{
			Name:       "sctp",
			Protocol:   "SCTP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureStatusPort_PortMismatch(t *testing.T) {
	ports := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       1234,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
	}
	ports = ensureStatusPort(ports)

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       istiov1beta1.PortStatusPortName,
			Protocol:   "TCP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       1234,
			TargetPort: intstr.FromInt(istiov1beta1.PortStatusPortNumber),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureStatusPort_TargetPortMismatch(t *testing.T) {
	originalPorts := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(1234),
		},
	}
	ports := ensureStatusPort(append([]apiv1.ServicePort{}, originalPorts...))

	expectedPorts := []apiv1.ServicePort{
		{
			Name:       "foo",
			Protocol:   "TCP",
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(1234),
		},
	}
	require.Equal(t, expectedPorts, ports)
}

func TestEnsureStatusPort_PortNameTaken(t *testing.T) {
	originalPorts := []apiv1.ServicePort{
		{
			Name:       "sTaTuS-pOrT",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		},
	}
	ports := ensureStatusPort(append([]apiv1.ServicePort{}, originalPorts...))

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

func TestEnsureStatusPort_PortNameAndTargetPort(t *testing.T) {
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
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(1234),
		},
	}
	ports := ensureStatusPort(append([]apiv1.ServicePort{}, originalPorts...))

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
			Port:       istiov1beta1.PortStatusPortNumber,
			TargetPort: intstr.FromInt(1234),
		},
	}
	require.Equal(t, expectedPorts, ports)
}
