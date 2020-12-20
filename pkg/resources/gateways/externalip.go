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
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

func (r *Reconciler) GetGatewayAddress() ([]string, bool, error) {
	var service corev1.Service
	var ips []string
	var hasHostname bool

	err := r.Get(context.Background(), types.NamespacedName{
		Name:      r.gatewayName(),
		Namespace: r.gw.Namespace,
	}, &service)
	if err != nil {
		return nil, hasHostname, err
	}

	ips, hasHostname, err = k8sutil.GetServiceEndpointIPs(service)
	if err != nil {
		return nil, hasHostname, err
	}

	return ips, hasHostname, nil
}
