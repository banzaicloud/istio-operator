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
	return ReconcileWithObjectModifiers(log, client, desired, desiredState, nil)
}

func ReconcileWithObjectModifiers(log logr.Logger, client runtimeClient.Client, desired runtime.Object, desiredState DesiredState, objectModifiers []ObjectModifierFunc) error {
	if desiredState == nil {
		desiredState = DesiredStatePresent
	}

	desired, err := RunObjectModifiers(desired, objectModifiers)
	if err != nil {
		return err
	}

	desiredType := reflect.TypeOf(desired)
	current := desired.DeepCopyObject().(runtimeClient.Object)
	desiredCopy := desired.DeepCopyObject().(runtimeClient.Object)
	key := runtimeClient.ObjectKeyFromObject(current)
	log = log.WithValues("kind", desiredType, "name", key.Name)

	err = client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "kind", desiredType, "name", key.Name)
	}
	if apierrors.IsNotFound(err) {
		if desiredState != DesiredStateAbsent {
			should, err := desiredState.ShouldCreate(current)
			if err != nil {
				return emperror.WrapWith(err, "could not execute ShouldCreate func")
			}
			if !should {
				log.V(1).Info("resource should not be created")
				return nil
			}
			if err := desiredState.BeforeCreate(desired); err != nil {
				return emperror.WrapWith(err, "could not execute BeforeCreate func")
			}
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			if err := client.Create(context.TODO(), desired.(runtimeClient.Object)); err != nil {
				return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
			}
			if err := desiredState.AfterCreate(desired); err != nil {
				return emperror.WrapWith(err, "could not execute AfterCreate func")
			}
			log.Info("resource created")
		}
	} else {
		if desiredState != DesiredStateAbsent && desiredState != DesiredStateExists {
			should, err := desiredState.ShouldUpdate(current, desiredCopy)
			if err != nil {
				return emperror.WrapWith(err, "could not execute ShouldUpdate func")
			}
			if !should {
				log.V(1).Info("resource should not be updated")
				return nil
			}
			if err := desiredState.BeforeUpdate(current, desiredCopy); err != nil {
				return emperror.WrapWith(err, "could not execute BeforeUpdate func")
			}

			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", desiredType, "name", key.Name)
			} else if patchResult.IsEmpty() {
				log.V(1).Info("resource is in sync")
				if err := desiredState.AfterUpdate(current, desiredCopy, true); err != nil {
					return emperror.WrapWith(err, "could not execute AfterUpdate func")
				}
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

			if err := client.Update(context.TODO(), desired.(runtimeClient.Object)); err != nil {
				if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
					should, err := desiredState.ShouldRecreate(current, desiredCopy)
					if err != nil {
						return emperror.WrapWith(err, "could not execute ShoudReCreate func")
					}
					if !should {
						log.V(1).Info("resource should not be re-created")
						return nil
					}
					log.Info("resource needs to be re-created", "error", err)
					if err := desiredState.BeforeRecreate(current, desiredCopy); err != nil {
						return emperror.WrapWith(err, "could not execute BeforeRecreate func")
					}
					if err := client.Delete(context.TODO(), current); err != nil {
						return emperror.WrapWith(err, "could not delete resource", "kind", desiredType, "name", key.Name)
					}
					log.Info("resource deleted")
					if err := client.Create(context.TODO(), desiredCopy); err != nil {
						return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
					}
					log.Info("resource created")
					if err := desiredState.AfterRecreate(current, desiredCopy); err != nil {
						return emperror.WrapWith(err, "could not execute AfterRecreate func")
					}
					return nil
				}

				return emperror.WrapWith(err, "updating resource failed", "kind", desiredType, "name", key.Name)
			}
			if err := desiredState.AfterUpdate(current, desiredCopy, false); err != nil {
				return emperror.WrapWith(err, "could not execute AfterUpdate func")
			}
			log.Info("resource updated")
		} else if desiredState == DesiredStateAbsent {
			should, err := desiredState.ShouldDelete(current)
			if err != nil {
				return emperror.WrapWith(err, "could not execute ShouldDelete func")
			}
			if !should {
				log.V(1).Info("resource should not be deleted")
				return nil
			}
			if err := desiredState.BeforeDelete(current); err != nil {
				return emperror.WrapWith(err, "could not execute BeforeDelete func")
			}
			if err := client.Delete(context.TODO(), current); err != nil {
				return emperror.WrapWith(err, "deleting resource failed", "kind", desiredType, "name", key.Name)
			}
			if err := desiredState.AfterDelete(current); err != nil {
				return emperror.WrapWith(err, "could not execute AfterDelete func")
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
	case *corev1.ServiceAccount:
		sa := desired.(*corev1.ServiceAccount)
		sa.Secrets = current.(*corev1.ServiceAccount).Secrets
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
func ReconcileNamespaceLabelsIgnoreNotFound(log logr.Logger, client runtimeClient.Client, namespace string, labels map[string]string, labelsToRemove []string, customLabelsToIgnoreReconcile ...string) error {
	ns := &corev1.Namespace{}
	err := client.Get(context.TODO(), runtimeClient.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("namespace not found, ignoring", "namespace", namespace)
			return nil
		}

		return emperror.WrapWith(err, "getting namespace failed", "namespace", namespace)
	}

	for _, customLabel := range customLabelsToIgnoreReconcile {
		if _, ok := ns.Labels[customLabel]; ok {
			log.V(1).Info("namespace has a custom label, ignoring namespace", "namespace", namespace, "customLabel", customLabel)
			return nil
		}
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
