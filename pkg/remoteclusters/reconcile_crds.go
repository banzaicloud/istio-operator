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
	"bytes"
	"time"

	"github.com/goph/emperror"
	"github.com/pkg/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/crds/generated"
)

func (c *Cluster) reconcileCRDs(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling CRDs")

	istiocrd, err := c.istiocrd()
	if err != nil {
		return emperror.Wrap(err, "could not get istio CRD")
	}

	meshgatewaycrd, err := c.meshgatewaycrd()
	if err != nil {
		return emperror.Wrap(err, "could not get meshgateway CRD")
	}

	resources := []*extensionsobj.CustomResourceDefinition{
		istiocrd,
		meshgatewaycrd,
	}

	crdo, err := crds.New(c.restConfig, resources...)
	if err != nil {
		return err
	}

	err = crdo.Reconcile(&istiov1beta1.Istio{}, c.log)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 1)
	err = c.waitForCRDs(resources)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) waitForCRDs(crds []*extensionsobj.CustomResourceDefinition) error {
	apiExtensions, err := apiextensionsclient.NewForConfig(c.restConfig)
	if err != nil {
		return errors.Wrap(err, "instantiating apiextensions client failed")
	}
	crdClient := apiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crd := range crds {
		crd, err := crdClient.Get(crd.Name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			return err
		}
		c.log.Info("wait for CRD", "name", crd.Name)
		err = c.waitForCRD(crd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) waitForCRD(crd *extensionsobj.CustomResourceDefinition) error {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1beta1.Established {
			if condition.Status == apiextensionsv1beta1.ConditionTrue {
				return nil
			}
		}
	}

	return errors.Errorf("CRD '%s' is not established yet", crd.Name)
}

func (c *Cluster) istiocrd() (*extensionsobj.CustomResourceDefinition, error) {
	return c.getCRD("istio_v1beta1_istio.yaml")
}

func (c *Cluster) meshgatewaycrd() (*extensionsobj.CustomResourceDefinition, error) {
	return c.getCRD("istio_v1beta1_meshgateway.yaml")
}

func (c *Cluster) getCRD(name string) (*extensionsobj.CustomResourceDefinition, error) {
	var resource *apiextensionsv1beta1.CustomResourceDefinition

	f, err := generated.CRDs.Open("/" + name)
	if err != nil {
		return nil, err
	}

	yaml := new(bytes.Buffer)
	_, err = yaml.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	s := runtime.NewScheme()
	apiextensionsv1beta1.AddToScheme(s)

	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, s, s)
	o, _, err := serializer.Decode(yaml.Bytes(), nil, nil)
	if err != nil {
		return nil, err
	}

	var ok bool
	if resource, ok = o.(*apiextensionsv1beta1.CustomResourceDefinition); !ok {
		return nil, errors.New("invalid resource kind")
	}

	resource.SetGroupVersionKind(schema.GroupVersionKind{})
	resource.Status = apiextensionsv1beta1.CustomResourceDefinitionStatus{}

	return resource, nil
}
