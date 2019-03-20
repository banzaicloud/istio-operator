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

package crds

import (
	"bytes"
	"os"
	"path"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/releaseutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil/objectmatch"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var log = logf.Log.WithName("crds")

const (
	componentName = "crds"
)

type CrdOperator struct {
	crds   []*extensionsobj.CustomResourceDefinition
	config *rest.Config
}

func New(cfg *rest.Config, crds []*extensionsobj.CustomResourceDefinition) (*CrdOperator, error) {
	return &CrdOperator{
		crds:   crds,
		config: cfg,
	}, nil
}

func DecodeCRDs(chartPath string) ([]*extensionsobj.CustomResourceDefinition, error) {
	log.Info("ensuring CRDs have been installed")
	crdPath := path.Join(chartPath, "istio-init/files")
	crdDir, err := os.Stat(crdPath)
	if !crdDir.IsDir() {
		return nil, errors.Errorf("Cannot locate any CRD files in %s", crdPath)
	}
	var crds []*extensionsobj.CustomResourceDefinition
	err = filepath.Walk(crdPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		c, err := processCRDFile(path)
		if err != nil {
			return err
		}
		crds = append(crds, c...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	log.Info("finished processing CRDs", "number of CRDs", len(crds))
	return crds, nil
}

func processCRDFile(fileName string) ([]*extensionsobj.CustomResourceDefinition, error) {
	allErrors := []error{}
	buf := &bytes.Buffer{}
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}
	var crds []*extensionsobj.CustomResourceDefinition
	for _, raw := range releaseutil.SplitManifests(string(buf.Bytes())) {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, gvk, err := decode([]byte(raw), nil, nil)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		switch obj.(type) {
		case *extensionsobj.CustomResourceDefinition:
			crd := obj.(*extensionsobj.CustomResourceDefinition)
			log.Info("found custom resource definition", "group", crd.Spec.Group, "version", crd.Spec.Version, "name", crd.Name)
			crds = append(crds, crd)
		default:
			log.V(1).Info("decoded object is not a custom resource definition", "groupVersionKind", gvk)
		}
	}
	log.Info("finished processing file", "file", fileName, "number of CRDs", len(crds))
	err = utilerrors.NewAggregate(allErrors)
	if err != nil {
		return nil, err
	}
	return crds, nil
}

func (r *CrdOperator) Reconcile(config *istiov1beta1.Istio, log logr.Logger) error {
	log = log.WithValues("component", componentName)
	apiExtensions, err := apiextensionsclient.NewForConfig(r.config)
	if err != nil {
		return emperror.Wrap(err, "instantiating apiextensions client failed")
	}
	crdClient := apiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range r.crds {
		log := log.WithValues("kind", crd.Spec.Names.Kind)
		current, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if config.Name != "" {
				crd.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
					{
						Kind:               config.Kind,
						APIVersion:         config.APIVersion,
						Name:               config.Name,
						UID:                config.GetUID(),
						Controller:         util.BoolPointer(true),
						BlockOwnerDeletion: util.BoolPointer(true),
					},
				}
			}
			if _, err := crdClient.Create(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created")
		}
		if err == nil {
			objectsEquals, err := objectmatch.New(log).Match(current, crd)
			if err != nil {
				log.Error(err, "could not match objects", "kind", crd.Spec.Names.Kind)
			} else if objectsEquals {
				log.V(1).Info("CRD is in sync")
				continue
			}
			crd.ResourceVersion = current.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return emperror.WrapWith(err, "updating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated")
		}
	}

	log.Info("Reconciled")

	return nil
}
