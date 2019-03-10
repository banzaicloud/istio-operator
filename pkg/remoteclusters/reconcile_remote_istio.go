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
	"github.com/banzaicloud/istio-operator/pkg/resources/citadel"
	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/nodeagent"
	"github.com/banzaicloud/istio-operator/pkg/resources/sidecarinjector"
)

func (c *Cluster) reconcileComponents(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling components")

	reconcilers := []resources.ComponentReconciler{
		common.New(c.ctrlRuntimeClient, c.istioConfig, true),
		citadel.New(citadel.Configuration{
			DeployMeshPolicy: false,
		}, c.ctrlRuntimeClient, c.dynamicClient, c.istioConfig),
		sidecarinjector.New(c.ctrlRuntimeClient, c.istioConfig),
	}

	if c.istioConfig.Spec.NodeAgent.Enabled {
		reconcilers = append(reconcilers, nodeagent.New(c.ctrlRuntimeClient, c.istioConfig))
	}

	for _, rec := range reconcilers {
		err := rec.Reconcile(c.log)
		if err != nil {
			return err
		}
	}

	return nil
}
