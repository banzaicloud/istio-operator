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
	"github.com/goph/emperror"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type OwnerReferenceMatcher struct {
	ownerType          runtime.Object
	ownerTypeGroupKind schema.GroupKind
	isController       bool
	scheme             *runtime.Scheme
}

// NewOwnerReferenceMatcher initializes a new owner reference matcher
func NewOwnerReferenceMatcher(ownerType runtime.Object, ctrl bool, scheme *runtime.Scheme) *OwnerReferenceMatcher {
	m := &OwnerReferenceMatcher{
		ownerType:    ownerType,
		isController: ctrl,
		scheme:       scheme,
	}

	m.setOwnerTypeGroupKind()

	return m
}

// Match matches if an object is owned by the initialised owner
func (e *OwnerReferenceMatcher) Match(object runtime.Object) (bool, metav1.Object, error) {
	o, err := meta.Accessor(object)
	if err != nil {
		return false, o, emperror.WrapWith(err, "could not access object meta", "kind", object.GetObjectKind())
	}

	for _, owner := range e.getOwnersReferences(o) {
		groupVersion, err := schema.ParseGroupVersion(owner.APIVersion)
		if err != nil {
			return false, o, emperror.WrapWith(err, "could not parse api version", "apiVersion", owner.APIVersion)
		}

		if owner.Kind == e.ownerTypeGroupKind.Kind && groupVersion.Group == e.ownerTypeGroupKind.Group {
			return true, o, nil
		}
	}

	return false, o, nil
}

func (e *OwnerReferenceMatcher) getOwnersReferences(object metav1.Object) []metav1.OwnerReference {
	if object == nil {
		return nil
	}
	if !e.isController {
		return object.GetOwnerReferences()
	}
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		return []metav1.OwnerReference{*ownerRef}
	}
	return nil
}

func (e *OwnerReferenceMatcher) setOwnerTypeGroupKind() error {
	kinds, _, err := e.scheme.ObjectKinds(e.ownerType)
	if err != nil || len(kinds) < 1 {
		return emperror.WrapWith(err, "could not get object kinds", "ownerType", e.ownerType)
	}

	e.ownerTypeGroupKind = schema.GroupKind{Group: kinds[0].Group, Kind: kinds[0].Kind}
	return nil
}
