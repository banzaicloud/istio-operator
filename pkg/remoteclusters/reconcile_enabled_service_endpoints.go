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

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	apiv1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Cluster) reconcileServiceEndpoints(endp apiv1.Endpoints) error {
	var endpoints apiv1.Endpoints
	err := c.ctrlRuntimeClient.Get(context.TODO(), types.NamespacedName{
		Name:      endp.Name,
		Namespace: endp.Namespace,
	}, &endpoints)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		return err
	}

	if k8sapierrors.IsNotFound(err) {
		err = c.ctrlRuntimeClient.Create(context.TODO(), &endp)
		if err != nil {
			return err
		}
	} else {
		endpoints.Subsets = endp.Subsets
		err = c.ctrlRuntimeClient.Update(context.TODO(), &endpoints)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) ReconcileEnabledServiceEndpoints(remoteConfig *istiov1beta1.RemoteIstio) error {
	for _, enabledSvc := range remoteConfig.Spec.EnabledServices {
		addresses := make([]apiv1.EndpointAddress, 0)
		for _, ip := range enabledSvc.IPs {
			if ip == "" {
				continue
			}
			addresses = append(addresses, apiv1.EndpointAddress{
				IP: ip,
			})
		}

		if len(addresses) == 0 {
			continue
		}

		endp := apiv1.Endpoints{
			ObjectMeta: templates.ObjectMeta(enabledSvc.Name, map[string]string{}, c.istioConfig),
			Subsets: []apiv1.EndpointSubset{
				{
					Addresses: addresses,
					Ports: []apiv1.EndpointPort{
						{
							Port: 65000,
						},
					},
				},
			},
		}

		err := c.reconcileServiceEndpoints(endp)
		if err != nil {
			return err
		}
	}

	return nil
}
