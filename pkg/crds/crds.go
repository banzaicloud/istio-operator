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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	istio_crds "github.com/banzaicloud/istio-operator/pkg/manifests/istio-crds/generated"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	componentName     = "crds"
	createdByLabel    = "banzaicloud.io/created-by"
	managedBy         = "istio-operator"
	eventRecorderName = "istio-crd-controller"
)

type CRDReconciler struct {
	crds     []runtime.Object
	config   *rest.Config
	revision string
	recorder record.EventRecorder
	client   client.Client
}

func New(mgr manager.Manager, revision string, crds ...runtime.Object) (*CRDReconciler, error) {
	r := &CRDReconciler{
		crds:     crds,
		config:   mgr.GetConfig(),
		revision: revision,
		recorder: mgr.GetEventRecorderFor(eventRecorderName),
		client:   mgr.GetClient(),
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

		if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
			crd.Status = apiextensionsv1.CustomResourceDefinitionStatus{}
			crd.SetGroupVersionKind(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
			crd.SetLabels(util.MergeStringMaps(crd.GetLabels(), map[string]string{
				createdByLabel:       managedBy,
				types.ManagedByLabel: managedBy,
			}))
			r.crds = append(r.crds, crd)
			continue
		}

		if crd, ok := obj.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
			crd.Status = apiextensionsv1beta1.CustomResourceDefinitionStatus{}
			crd.SetGroupVersionKind(apiextensionsv1beta1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
			crd.SetLabels(util.MergeStringMaps(crd.GetLabels(), map[string]string{
				createdByLabel:       managedBy,
				types.ManagedByLabel: managedBy,
			}))
			r.crds = append(r.crds, crd)
		}
	}

	return nil
}

func (r *CRDReconciler) Reconcile(config *istiov1beta1.Istio, log logr.Logger) error {
	log = log.WithValues("component", componentName)

	for _, obj := range r.crds {
		var name, kind string
		if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
			name = crd.Name
			kind = crd.Spec.Names.Kind
		} else if crd, ok := obj.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
			name = crd.Name
			kind = crd.Spec.Names.Kind
		} else {
			log.Error(errors.New("invalid GVK"), "cannot reconcile CRD", "gvk", obj.GetObjectKind().GroupVersionKind())
			continue
		}

		crd := obj.DeepCopyObject().(client.Object)
		current := obj.DeepCopyObject().(client.Object)
		err := k8sutil.SetResourceRevision(crd, r.revision)
		if err != nil {
			return emperror.Wrap(err, "could not set resource revision")
		}
		log := log.WithValues("kind", kind)
		err = r.client.Get(context.Background(), client.ObjectKey{
			Name: name,
		}, current)
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", kind)
		}
		if apierrors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "failed to set last applied annotation", "crd", crd)
			}
			if err := r.client.Create(context.Background(), crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", kind)
			}
			log.Info("CRD created")
		} else {
			managedByUs := false
			if current.GetLabels()[createdByLabel] == managedBy || current.GetLabels()[types.ManagedByLabel] == managedBy {
				managedByUs = true
			}

			if !managedByUs {
				log.V(1).Info("current crd is not managed by us, skip update", "name", current.GetName())
				continue
			}

			if ok, err := k8sutil.CheckResourceRevision(current, fmt.Sprintf("<=%s", r.revision)); !ok {
				if err != nil {
					log.Error(err, "could not check resource revision")
				} else {
					log.V(1).Info("CRD is too new for us")
				}
				continue
			}

			crd.SetResourceVersion(current.GetResourceVersion())

			patchResult, err := patch.DefaultPatchMaker.Calculate(current, crd, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", kind)
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
				log.Error(err, "failed to set last applied annotation", "crd", crd)
			}

			if err := r.client.Update(context.Background(), crd); err != nil {
				errorMessage := "updating CRD failed, consider updating the CRD manually if needed"
				r.recorder.Eventf(
					config,
					corev1.EventTypeWarning,
					"IstioCRDUpdateFailure",
					errorMessage,
					"kind",
					kind,
				)
				return emperror.WrapWith(err, errorMessage, "kind", kind)
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
			if e.Object.GetLabels()[createdByLabel] == managedBy || e.Object.GetLabels()[types.ManagedByLabel] == managedBy {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld.GetLabels()[createdByLabel] == managedBy || e.ObjectNew.GetLabels()[createdByLabel] == managedBy {
				return true
			}
			if e.ObjectOld.GetLabels()[types.ManagedByLabel] == managedBy || e.ObjectNew.GetLabels()[types.ManagedByLabel] == managedBy {
				return true
			}
			return false
		},
	}
}
