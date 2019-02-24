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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Cluster) reconcileConfigCrd(remoteConfig *istiov1beta1.RemoteConfig) error {
	c.log.Info("reconciling config crd")

	crdo, err := crds.New(c.restConfig, []*extensionsobj.CustomResourceDefinition{
		c.configcrd(),
	})
	if err != nil {
		return err
	}

	err = crdo.Reconcile(&istiov1beta1.Config{})
	if err != nil {
		return err
	}

	// Re-init k8s clients
	err = c.initK8SClients()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) configcrd() *extensionsobj.CustomResourceDefinition {
	return &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "configs.operator.istio.io",
			Labels: map[string]string{
				"controller-tools.k8s.io": "1.0",
			},
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   "operator.istio.io",
			Version: "v1beta1",
			Scope:   "Namespaced",
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:   "configs",
				Kind:     "Config",
				ListKind: "ConfigList",
				Singular: "config",
			},
		},
	}
}
