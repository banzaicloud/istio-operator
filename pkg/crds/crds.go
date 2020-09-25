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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	istio_crds "github.com/banzaicloud/istio-operator/pkg/manifests/istio-crds/generated"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

const (
	componentName     = "crds"
	createdByLabel    = "banzaicloud.io/created-by"
	createdBy         = "istio-operator"
	eventRecorderName = "istio-crd-controller"
)

type CRDReconciler struct {
	crds     []*apiextensionsv1.CustomResourceDefinition
	config   *rest.Config
	revision string
	recorder record.EventRecorder
}

func New(mgr manager.Manager, revision string, crds ...*apiextensionsv1.CustomResourceDefinition) (*CRDReconciler, error) {
	r := &CRDReconciler{
		crds:     crds,
		config:   mgr.GetConfig(),
		revision: revision,
		recorder: mgr.GetEventRecorderFor(eventRecorderName),
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

		var crd *apiextensionsv1.CustomResourceDefinition
		var ok bool
		if crd, ok = obj.(*apiextensionsv1.CustomResourceDefinition); !ok {
			continue
		}

		crd.Status = apiextensionsv1.CustomResourceDefinitionStatus{}
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
	crdClient := apiExtensions.ApiextensionsV1().CustomResourceDefinitions()
	for _, obj := range r.crds {
		crd := obj.DeepCopy()
		err = k8sutil.SetResourceRevision(crd, r.revision)
		if err != nil {
			return emperror.Wrap(err, "could not set resource revision")
		}
		log := log.WithValues("kind", crd.Spec.Names.Kind)
		current, err := crdClient.Get(context.Background(), crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "Failed to set last applied annotation", "crd", crd)
			}
			if _, err := crdClient.Create(context.Background(), crd, metav1.CreateOptions{}); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created")
		} else {
			if ok, err := k8sutil.CheckResourceRevision(current, fmt.Sprintf("<=%s", r.revision)); !ok {
				if err != nil {
					log.Error(err, "could not check resource revision")
				} else {
					log.V(1).Info("CRD is too new for us")
				}
				continue
			}
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

			if _, err := crdClient.Update(context.Background(), crd, metav1.UpdateOptions{}); err != nil {
				errorMessage := "updating CRD failed, consider updating the CRD manually if needed"
				r.recorder.Eventf(
					config,
					"Warning",
					"IstioCRDUpdateFailure",
					errorMessage,
					"kind",
					crd.Spec.Names.Kind,
				)
				return emperror.WrapWith(err, errorMessage, "kind", crd.Spec.Names.Kind)
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
