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

package k8sutil

import (
	"reflect"

	patch "github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

type DynamicObject struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Spec      map[string]interface{}
	Gvr       schema.GroupVersionResource
	Kind      string
	Owner     *istiov1beta1.Istio
}

func (d *DynamicObject) Reconcile(log logr.Logger, client dynamic.Interface, desiredState DesiredState) error {
	if desiredState == nil {
		desiredState = DesiredStatePresent
	}
	desired := d.unstructured()
	desiredType := reflect.TypeOf(desired)
	log = log.WithValues("type", reflect.TypeOf(d), "name", d.Name)
	current, err := client.Resource(d.Gvr).Namespace(d.Namespace).Get(d.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "name", d.Name, "kind", desiredType)
	}
	if apierrors.IsNotFound(err) {
		if desiredState == DesiredStatePresent || desiredState == DesiredStateExists {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Create(desired, metav1.CreateOptions{}); err != nil {
				return emperror.WrapWith(err, "creating resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource created", "kind", d.Gvr.Resource)
		}
	} else {
		if desiredState == DesiredStatePresent {
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", desiredType)
			} else if patchResult.IsEmpty() {
				log.V(1).Info("resource is in sync")
				return nil
			} else {
				log.V(1).Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}
			// Need to set this before resourceversion is set, as it would constantly change otherwise
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			desired.SetResourceVersion(current.GetResourceVersion())
			if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Update(desired, metav1.UpdateOptions{}); err != nil {
				return emperror.WrapWith(err, "updating resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource updated", "kind", d.Gvr.Resource)
		} else if desiredState == DesiredStateAbsent {
			if err := client.Resource(d.Gvr).Namespace(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{}); err != nil {
				return emperror.WrapWith(err, "deleting resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource deleted", "kind", d.Gvr.Resource)
		}
	}
	return nil
}

func (d *DynamicObject) unstructured() *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": d.Spec,
		},
	}
	u.SetName(d.Name)
	if len(d.Namespace) > 0 {
		u.SetNamespace(d.Namespace)
	}
	if d.Labels != nil {
		u.SetLabels(d.Labels)
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   d.Gvr.Group,
		Version: d.Gvr.Version,
		Kind:    d.Kind,
	})
	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         d.Owner.APIVersion,
			Kind:               d.Owner.Kind,
			Name:               d.Owner.Name,
			UID:                d.Owner.UID,
			Controller:         util.BoolPointer(true),
			BlockOwnerDeletion: util.BoolPointer(true),
		},
	})
	return u
}
