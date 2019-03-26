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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"k8s.io/helm/pkg/manifest"
)

func (c *Cluster) reconcileComponents(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling components")

	reconcilers := []*resources.Reconciler{
		resources.New(c.ctrlRuntimeClient, c.istioConfig, []manifest.Manifest{}, nil),
		resources.New(c.ctrlRuntimeClient, c.istioConfig, []manifest.Manifest{}, nil),
		resources.New(c.ctrlRuntimeClient, c.istioConfig, []manifest.Manifest{}, nil),
		//TODO
	}

	for _, rec := range reconcilers {
		err := rec.Reconcile(c.log)
		if err != nil {
			return err
		}
	}

	return nil
}
