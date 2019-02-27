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

package objectmatch

import (
	"encoding/json"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func Match(old, new interface{}) (bool, error) {
	if reflect.TypeOf(old) != reflect.TypeOf(new) {
		return false, emperror.With(errors.New("old and new object types mismatch"), "oldType", reflect.TypeOf(old), "newType", reflect.TypeOf(new))
	}

	switch old.(type) {
	default:
		return false, nil
	case *unstructured.Unstructured:
		oldObject := old.(*unstructured.Unstructured)
		newObject := new.(*unstructured.Unstructured)

		m := UnstructuredMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *corev1.ServiceAccount:
		oldObject := old.(*corev1.ServiceAccount)
		newObject := new.(*corev1.ServiceAccount)

		m := ServiceAccountMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *rbacv1.ClusterRole:
		oldObject := old.(*rbacv1.ClusterRole)
		newObject := new.(*rbacv1.ClusterRole)

		m := ClusterRoleMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *rbacv1.ClusterRoleBinding:
		oldObject := old.(*rbacv1.ClusterRoleBinding)
		newObject := new.(*rbacv1.ClusterRoleBinding)

		m := ClusterRoleBindingMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *appsv1.Deployment:
		oldObject := old.(*appsv1.Deployment)
		newObject := new.(*appsv1.Deployment)

		m := DeploymentMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *corev1.Service:
		oldObject := old.(*corev1.Service)
		newObject := new.(*corev1.Service)

		m := ServiceMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *corev1.ConfigMap:
		oldObject := old.(*corev1.ConfigMap)
		newObject := new.(*corev1.ConfigMap)

		m := ConfigMapMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *extensionsobj.CustomResourceDefinition:
		oldObject := old.(*extensionsobj.CustomResourceDefinition)
		newObject := new.(*extensionsobj.CustomResourceDefinition)

		m := CRDMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *autoscalev2beta1.HorizontalPodAutoscaler:
		oldObject := old.(*autoscalev2beta1.HorizontalPodAutoscaler)
		newObject := new.(*autoscalev2beta1.HorizontalPodAutoscaler)

		m := HorizontalPodAutoscalerMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	case *admissionv1beta1.MutatingWebhookConfiguration:
		oldObject := old.(*admissionv1beta1.MutatingWebhookConfiguration)
		newObject := new.(*admissionv1beta1.MutatingWebhookConfiguration)

		m := MutatingWebhookConfigurationMatcher{}
		ok, err := m.Match(oldObject, newObject)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return ok, nil
	}
}

type ObjectMeta struct {
	Labels          map[string]string       `json:"labels,omitempty"`
	Annotations     map[string]string       `json:"annotations,omitempty"`
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences,omitempty"`
}

func getObjectMeta(objectMeta metav1.ObjectMeta) ObjectMeta {
	return ObjectMeta{
		Labels:          objectMeta.Labels,
		Annotations:     objectMeta.Annotations,
		OwnerReferences: objectMeta.OwnerReferences,
	}
}

func match(old, new []byte, obj interface{}) (bool, error) {
	var patch []byte
	var err error

	_, unstructed := obj.(*unstructured.Unstructured)
	if unstructed {
		patch, err = jsonpatch.CreateMergePatch(old, new)
		if err != nil {
			return false, emperror.Wrap(err, "could not create json merge patch")
		}
	} else {
		patch, err = strategicpatch.CreateTwoWayMergePatch(old, new, obj)
		if err != nil {
			return false, emperror.Wrap(err, "could not create two way merge patch")
		}
	}

	patch, _, err = deleteNullInJsonPatch(patch)
	if err != nil {
		return false, emperror.Wrap(err, "could not remove nil values from json merge patch")
	}

	if string(patch) == "{}" {
		return true, nil
	}

	return false, nil
}

func deleteNullInJsonPatch(patch []byte) ([]byte, map[string]interface{}, error) {
	var patchMap map[string]interface{}

	err := json.Unmarshal(patch, &patchMap)
	if err != nil {
		return nil, nil, emperror.Wrap(err, "could not unmarshal json patch")
	}

	filteredMap, err := deleteNullInObj(patchMap)
	if err != nil {
		return nil, nil, emperror.Wrap(err, "could not delete null values from patch map")
	}

	o, err := json.Marshal(filteredMap)
	if err != nil {
		return nil, nil, emperror.Wrap(err, "could not marshal filtered patch map")
	}

	return o, filteredMap, err
}

func deleteNullInObj(m map[string]interface{}) (map[string]interface{}, error) {
	var err error
	filteredMap := make(map[string]interface{})

	for key, val := range m {
		if val == nil {
			continue
		}

		switch typedVal := val.(type) {
		default:
			return nil, errors.Errorf("unknown type: %v", reflect.TypeOf(typedVal))
		case []interface{}, string, float64, bool, int64, nil:
			filteredMap[key] = val
		case map[string]interface{}:
			if len(typedVal) == 0 {
				filteredMap[key] = typedVal
				continue
			}

			var filteredSubMap map[string]interface{}
			filteredSubMap, err = deleteNullInObj(typedVal)
			if err != nil {
				return nil, emperror.Wrap(err, "could not delete null values from filtered sub map")
			}

			if len(filteredSubMap) != 0 {
				filteredMap[key] = filteredSubMap
			}
		}
	}
	return filteredMap, nil
}
