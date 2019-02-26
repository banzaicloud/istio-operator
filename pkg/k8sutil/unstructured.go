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

	"github.com/banzaicloud/istio-operator/pkg/k8sutil/objectmatch"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
)

type DynamicObject struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Spec      map[string]interface{}
	Gvr       schema.GroupVersionResource
	Kind      string
	Owner     *istiov1beta1.Config
}

func (d *DynamicObject) Reconcile(log logr.Logger, client dynamic.Interface) error {
	desired := d.unstructured()
	desiredType := reflect.TypeOf(desired)
	log = log.WithValues("type", reflect.TypeOf(d), "name", d.Name)
	current, err := client.Resource(d.Gvr).Namespace(d.Namespace).Get(d.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "name", d.Name, "kind", desiredType)
	}
	if apierrors.IsNotFound(err) {
		if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Create(desired, metav1.CreateOptions{}); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "name", d.Name, "kind", desiredType)
		}
		log.Info("resource created", "name", d.Name, "kind", d.Gvr.Resource)
	}
	if err == nil {
		objectsEquals, err := objectmatch.Match(current, desired)
		if err != nil {
			log.Error(err, "could not match objects", "kind", desiredType, "name", d.Name)
		} else if objectsEquals {
			log.V(1).Info("resource is in sync")
			return nil
		}

		desired.SetResourceVersion(current.GetResourceVersion())
		if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Update(desired, metav1.UpdateOptions{}); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "name", d.Name, "kind", desiredType)
		}
		log.Info("resource updated", "name", d.Name, "kind", d.Gvr.Resource)
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
			APIVersion: d.Owner.APIVersion,
			Kind:       d.Owner.Kind,
			Name:       d.Owner.Name,
			UID:        d.Owner.UID,
		},
	})
	return u
}
