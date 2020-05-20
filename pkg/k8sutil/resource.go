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
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

func Reconcile(log logr.Logger, client runtimeClient.Client, desired runtime.Object, desiredState DesiredState) error {
	if desiredState == "" {
		desiredState = DesiredStatePresent
	}

	desiredType := reflect.TypeOf(desired)
	var current = desired.DeepCopyObject()
	var desiredCopy = desired.DeepCopyObject()
	key, err := runtimeClient.ObjectKeyFromObject(current)
	if err != nil {
		return emperror.With(err, "kind", desiredType)
	}
	log = log.WithValues("kind", desiredType, "name", key.Name)

	err = client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "kind", desiredType, "name", key.Name)
	}
	if apierrors.IsNotFound(err) {
		if desiredState == DesiredStatePresent || desiredState == DesiredStateExists {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			if err := client.Create(context.TODO(), desired); err != nil {
				return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
			}
			log.Info("resource created")
		}
	} else {
		if desiredState == DesiredStatePresent {
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", desiredType, "name", key.Name)
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

			metaAccessor := meta.NewAccessor()
			currentResourceVersion, err := metaAccessor.ResourceVersion(current)
			if err != nil {
				return err
			}

			metaAccessor.SetResourceVersion(desired, currentResourceVersion)
			prepareResourceForUpdate(current, desired)

			if err := client.Update(context.TODO(), desired); err != nil {
				if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
					log.Info("resource needs to be re-created", "error", err)
					err := client.Delete(context.TODO(), current)
					if err != nil {
						return emperror.WrapWith(err, "could not delete resource", "kind", desiredType, "name", key.Name)
					}
					log.Info("resource deleted")
					if err := client.Create(context.TODO(), desiredCopy); err != nil {
						return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
					}
					log.Info("resource created")
					return nil
				}

				return emperror.WrapWith(err, "updating resource failed", "kind", desiredType, "name", key.Name)
			}
			log.Info("resource updated")
		} else if desiredState == DesiredStateAbsent {
			if err := client.Delete(context.TODO(), current); err != nil {
				return emperror.WrapWith(err, "deleting resource failed", "kind", desiredType, "name", key.Name)
			}
			log.Info("resource deleted")
		}
	}
	return nil
}

func prepareResourceForUpdate(current, desired runtime.Object) {
	switch desired.(type) {
	case *corev1.Service:
		svc := desired.(*corev1.Service)
		svc.Spec.ClusterIP = current.(*corev1.Service).Spec.ClusterIP
	}
}

// IsObjectChanged checks whether there is an actual difference between the two objects
func IsObjectChanged(oldObj, newObj runtime.Object, ignoreStatusChange bool) (bool, error) {
	old := oldObj.DeepCopyObject()
	new := newObj.DeepCopyObject()

	metaAccessor := meta.NewAccessor()
	currentResourceVersion, err := metaAccessor.ResourceVersion(old)
	if err == nil {
		metaAccessor.SetResourceVersion(new, currentResourceVersion)
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(old, new, patch.IgnoreStatusFields())
	if err != nil {
		return true, emperror.WrapWith(err, "could not match objects", "kind", old.GetObjectKind())
	} else if patchResult.IsEmpty() {
		return false, nil
	}

	if ignoreStatusChange {
		var patch map[string]interface{}
		json.Unmarshal(patchResult.Patch, &patch)
		delete(patch, "status")
		if len(patch) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// ReconcileNamespaceLabelsIgnoreNotFound patches namespaces by adding/removing labels, returns without error if namespace is not found
func ReconcileNamespaceLabelsIgnoreNotFound(log logr.Logger, client runtimeClient.Client, namespace string, labels map[string]string, labelsToRemove []string) error {
	var ns = &corev1.Namespace{}
	err := client.Get(context.TODO(), runtimeClient.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("namespace not found, ignoring", "namespace", namespace)
			return nil
		}

		return emperror.WrapWith(err, "getting namespace failed", "namespace", namespace)
	}

	updateNeeded := false
	for dlk, dlv := range labels {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		if clv, ok := ns.Labels[dlk]; !ok || clv != dlv {
			ns.Labels[dlk] = dlv
			updateNeeded = true
		}
	}
	for _, labelKey := range labelsToRemove {
		if _, ok := ns.Labels[labelKey]; ok {
			delete(ns.Labels, labelKey)
			updateNeeded = true
		}
	}
	if updateNeeded {
		if err := client.Update(context.TODO(), ns); err != nil {
			return emperror.WrapWith(err, "updating namespace failed", "namespace", namespace)
		}
		log.Info("namespace labels reconciled", "namespace", namespace, "labels", labels)
	}

	return nil
}
