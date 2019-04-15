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

	apiv1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (c *Cluster) reconcileService(svc apiv1.Service) error {
	var service apiv1.Service
	err := c.ctrlRuntimeClient.Get(context.TODO(), types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}, &service)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		return err
	}

	if k8sapierrors.IsNotFound(err) {
		err = c.ctrlRuntimeClient.Create(context.TODO(), &svc)
		if err != nil {
			return err
		}
	} else {
		service.Spec.Ports = svc.Spec.Ports
		err = c.ctrlRuntimeClient.Update(context.TODO(), &service)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) reconcileEnabledServices(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	for _, enabledSvc := range remoteConfig.Spec.EnabledServices {
		svc := apiv1.Service{
			ObjectMeta: templates.ObjectMeta(enabledSvc.Name, map[string]string{}, c.istioConfig),
			Spec: apiv1.ServiceSpec{
				ClusterIP: "None",
				Ports:     enabledSvc.Ports,
			},
		}
		err := c.reconcileService(svc)
		if err != nil {
			return err
		}
	}

	return nil
}
