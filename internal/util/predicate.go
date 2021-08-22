/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package util

import (
	"encoding/json"
	"strings"

	"emperror.dev/errors"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

type CalculateOption = patch.CalculateOption

type ObjectChangePredicate struct {
	predicate.Funcs
	CalculateOptions []CalculateOption
}

func (p ObjectChangePredicate) Update(e event.UpdateEvent) bool {
	oldRV := e.ObjectOld.GetResourceVersion()
	e.ObjectOld.SetResourceVersion(e.ObjectNew.GetResourceVersion())
	defer e.ObjectOld.SetResourceVersion(oldRV)

	options := []CalculateOption{
		patch.IgnoreStatusFields(),
		reconciler.IgnoreManagedFields(),
	}
	options = append(options, p.CalculateOptions...)

	patchResult, err := patch.DefaultPatchMaker.Calculate(e.ObjectOld, e.ObjectNew, options...)
	if err != nil {
		return true
	} else if patchResult.IsEmpty() {
		return false
	}

	return true
}

func IgnoreMetadataAnnotations(prefixes ...string) CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteMetadataAnnotations(current, prefixes...)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete metadata annotations from modified byte sequence")
		}

		modified, err = deleteMetadataAnnotations(modified, prefixes...)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete metadata annotations from modified byte sequence")
		}

		return current, modified, nil
	}
}

func IgnoreWebhookFailurePolicy() CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteWebhookFailurePolicy(current)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete failure policy from modified byte sequence")
		}

		modified, err = deleteWebhookFailurePolicy(modified)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete failure policy from modified byte sequence")
		}

		return current, modified, nil
	}
}

func deleteWebhookFailurePolicy(obj []byte) ([]byte, error) {
	var objectMap map[string]interface{}
	err := json.Unmarshal(obj, &objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not unmarshal byte sequence")
	}

	if webhooks, ok := objectMap["webhooks"].([]interface{}); ok {
		for i, wh := range webhooks {
			if webhook, ok := wh.(map[string]interface{}); ok {
				delete(webhook, "failurePolicy")
				webhooks[i] = webhook
			}
		}
		objectMap["webhooks"] = webhooks
	}

	obj, err = json.Marshal(objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not marshal byte sequence")
	}

	return obj, nil
}

func deleteMetadataAnnotations(obj []byte, annotationPrefixes ...string) ([]byte, error) {
	var objectMap map[string]interface{}
	err := json.Unmarshal(obj, &objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not unmarshal byte sequence")
	}
	if metadata, ok := objectMap["metadata"].(map[string]interface{}); ok {
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			for k := range annotations {
				for _, prefix := range annotationPrefixes {
					if strings.HasPrefix(k, prefix) {
						delete(annotations, k)
					}
				}
			}
			metadata["annotations"] = annotations
		}
		objectMap["metadata"] = metadata
	}
	obj, err = json.Marshal(objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not marshal byte sequence")
	}

	return obj, nil
}

type ICPInjectorChangePredicate struct {
	predicate.Funcs
}

func (p ICPInjectorChangePredicate) Create(e event.CreateEvent) bool {
	return false
}

func (p ICPInjectorChangePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (p ICPInjectorChangePredicate) Update(e event.UpdateEvent) bool {
	var ok bool
	var oldICP *v1alpha1.IstioControlPlane
	var newICP *v1alpha1.IstioControlPlane

	if oldICP, ok = e.ObjectOld.(*v1alpha1.IstioControlPlane); !ok {
		return false
	}

	if newICP, ok = e.ObjectNew.(*v1alpha1.IstioControlPlane); !ok {
		return false
	}

	if oldICP.Status.GetChecksums().GetMeshConfig() != newICP.Status.GetChecksums().GetMeshConfig() {
		return true
	}

	if oldICP.Status.GetChecksums().GetSidecarInjector() != newICP.Status.GetChecksums().GetSidecarInjector() {
		return true
	}

	return false
}

func (p ICPInjectorChangePredicate) Generic(e event.GenericEvent) bool {
	return false
}
