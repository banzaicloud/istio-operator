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
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	istio_crds "github.com/banzaicloud/istio-operator/pkg/manifests/istio-crds/generated"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

const (
	componentName  = "crds"
	createdByLabel = "banzaicloud.io/created-by"
	createdBy      = "istio-operator"
)

type CRDReconciler struct {
	crds   []*extensionsobj.CustomResourceDefinition
	config *rest.Config
}

func New(cfg *rest.Config, crds ...*extensionsobj.CustomResourceDefinition) (*CRDReconciler, error) {
	r := &CRDReconciler{
		crds:   crds,
		config: cfg,
	}

	return r, nil
}

func (r *CRDReconciler) LoadCRDs() error {
	dir, err := istio_crds.CRDs.Open("/")
	if err != nil {
		return err
	}

	dirFiles, err := dir.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range dirFiles {
		f, err := istio_crds.CRDs.Open(file.Name())
		if err != nil {
			return err
		}

		err = r.load(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CRDReconciler) load(f io.Reader) error {
	var b bytes.Buffer

	var yamls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return err
			}
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		yamls = append(yamls, s)
	}

	for _, yaml := range yamls {
		s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
			scheme.Scheme)

		obj, _, err := s.Decode([]byte(yaml), nil, nil)
		if err != nil {
			continue
		}

		var crd *extensionsobj.CustomResourceDefinition
		var ok bool
		if crd, ok = obj.(*extensionsobj.CustomResourceDefinition); !ok {
			continue
		}

		crd.Status = extensionsobj.CustomResourceDefinitionStatus{}
		crd.SetGroupVersionKind(schema.GroupVersionKind{})
		labels := crd.GetLabels()
		labels[createdByLabel] = createdBy
		r.crds = append(r.crds, crd)
	}

	return nil
}

func (r *CRDReconciler) Reconcile(config *istiov1beta1.Istio, log logr.Logger) error {
	log = log.WithValues("component", componentName)
	apiExtensions, err := apiextensionsclient.NewForConfig(r.config)
	if err != nil {
		return emperror.Wrap(err, "instantiating apiextensions client failed")
	}
	crdClient := apiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, obj := range r.crds {
		crd := obj.DeepCopy()
		log := log.WithValues("kind", crd.Spec.Names.Kind)
		current, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "Failed to set last applied annotation", "crd", crd)
			}
			if _, err := crdClient.Create(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created")
		} else {
			crd.ResourceVersion = current.ResourceVersion
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, crd, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", crd.Spec.Names.Kind)
			} else if patchResult.IsEmpty() {
				log.V(1).Info("CRD is in sync")
				continue
			} else {
				log.V(1).Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "Failed to set last applied annotation", "crd", crd)
			}

			if _, err := crdClient.Update(crd); err != nil {
				if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
					err := crdClient.Delete(crd.Name, &metav1.DeleteOptions{})
					if err != nil {
						return emperror.WrapWith(err, "could not delete CRD", "kind", crd.Spec.Names.Kind)
					}
					crd.ResourceVersion = ""
					if _, err := crdClient.Create(crd); err != nil {
						log.Info("resource needs to be re-created")
						return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
					}
					log.Info("CRD created")
				}

				return emperror.WrapWith(err, "updating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated")
		}
	}

	log.Info("Reconciled")

	return nil
}

func GetWatchPredicateForCRDs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetLabels()[createdByLabel] == createdBy {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetLabels()[createdByLabel] == createdBy {
				return true
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld.GetLabels()[createdByLabel] == createdBy || e.MetaNew.GetLabels()[createdByLabel] == createdBy {
				return true
			}
			return false
		},
	}
}
