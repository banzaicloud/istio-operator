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

	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
)

type serviceMatcher struct {
	objectMatcher ObjectMatcher
}

func NewServiceMatcher(objectMatcher ObjectMatcher) *serviceMatcher {
	return &serviceMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two corev1.Service objects
func (m serviceMatcher) Match(old, new *corev1.Service) (bool, error) {
	type Service struct {
		ObjectMeta
		Spec corev1.ServiceSpec
	}

	if !reflect.DeepEqual(m.objectMatcher.GetObjectMeta(old.ObjectMeta), m.objectMatcher.GetObjectMeta(new.ObjectMeta)) {
		return false, nil
	}

	oldData, err := json.Marshal(Service{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Spec:       old.Spec,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newObject := Service{
		ObjectMeta: m.objectMatcher.GetObjectMeta(new.ObjectMeta),
		Spec:       new.Spec,
	}
	newData, err := json.Marshal(newObject)
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal new object", "name", new.Name)
	}

	matched, err := m.objectMatcher.MatchJSON(oldData, newData, newObject)
	if err != nil {
		return false, emperror.WrapWith(err, "could not match objects", "name", new.Name)
	}

	return matched, nil
}
