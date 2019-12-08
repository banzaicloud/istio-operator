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
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/crds/generated"
)

func (c *Cluster) reconcileCRDs(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling CRDs")

	crdo, err := crds.New(c.restConfig, []*extensionsobj.CustomResourceDefinition{
		c.istiocrd(),
		c.meshgatewaycrd(),
	})
	if err != nil {
		return err
	}

	err = crdo.Reconcile(&istiov1beta1.Istio{}, c.log)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) istiocrd() *extensionsobj.CustomResourceDefinition {
	return c.getCRD("istio_v1beta1_istio.yaml")
}

func (c *Cluster) meshgatewaycrd() *extensionsobj.CustomResourceDefinition {
	return c.getCRD("istio_v1beta1_meshgateway.yaml")
}

func (c *Cluster) getCRD(name string) *extensionsobj.CustomResourceDefinition {
	var resource apiextensionsv1beta1.CustomResourceDefinition

	f, err := generated.CRDs.Open("/" + name)
	if err != nil {
		return nil
	}

	s := runtime.NewScheme()
	apiextensionsv1beta1.AddToScheme(s)

	decoder := k8syaml.NewYAMLOrJSONDecoder(f, 1024)
	out := &unstructured.Unstructured{}
	err = decoder.Decode(out)
	if err != nil {
		return nil
	}

	err = s.Convert(out, &resource, nil)
	if err != nil {
		return nil
	}

	resource.Status = apiextensionsv1beta1.CustomResourceDefinitionStatus{}

	return &resource
}
