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

package reconcile

import (
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/helm"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

type Reconciler struct {
	client.Client
	Component string
	Config    *istiov1beta1.Istio
	Manifests []manifest.Manifest
	Scheme    *runtime.Scheme
}

func New(client client.Client, component string, config *istiov1beta1.Istio, manifests []manifest.Manifest, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Client:    client,
		Component: component,
		Config:    config,
		Manifests: manifests,
		Scheme:    scheme,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger, currentResources []*unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	log = log.WithValues("component", r.Component)
	log.Info("Reconciling")

	objects, err := helm.DecodeObjects(log, r.Manifests)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to decode objects from chart")
	}

	var reconciledResources []*unstructured.Unstructured

	for _, o := range objects {
		ro := o.(runtime.Object)
		gvk, err := apiutil.GVKForObject(ro, r.Scheme)
		if err != nil {
			return nil, err
		}
		err = controllerutil.SetControllerReference(r.Config, o, r.Scheme)
		if err != nil {
			return nil, emperror.WrapWith(err, "failed to set controller reference", "resource", gvk)
		}
		err = k8sutil.Reconcile(log, r.Client, ro, k8sutil.DesiredStatePresent)
		if err != nil {
			return nil, emperror.WrapWith(err, "failed to reconcile resource", "resource", gvk)
		}
		u, err := r.unstructuredMeta(ro)
		if err != nil {
			return nil, emperror.WrapWith(err, "failed to convert resource metadata", "resource", gvk)
		}
		reconciledResources = append(reconciledResources, u)
	}
	// delete currently managed resources that are no longer needed
	for _, cr := range currentResources {
		var found bool
		for _, mr := range reconciledResources {
			if reflect.DeepEqual(mr, cr) {
				found = true
				break
			}
		}
		if !found {
			gvk, err := apiutil.GVKForObject(cr, r.Scheme)
			if err != nil {
				return nil, err
			}
			err = k8sutil.Reconcile(log, r.Client, cr, k8sutil.DesiredStateAbsent)
			if err != nil {
				return nil, emperror.WrapWith(err, "failed to reconcile resource", "resource", gvk)
			}
		}
	}
	log.Info("Reconciled")

	return reconciledResources, nil
}

// unstructuredMeta returns an unstructured object containing gvk and name/namespace for an object to store in status
func (r *Reconciler) unstructuredMeta(ro runtime.Object) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	err := r.Scheme.Convert(ro, u, nil)
	if err != nil {
		return nil, err
	}
	um := &unstructured.Unstructured{}
	um.SetGroupVersionKind(u.GroupVersionKind())
	um.SetName(u.GetName())
	um.SetNamespace(u.GetNamespace())
	return um, nil
}
