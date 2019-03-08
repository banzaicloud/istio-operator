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

	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type unstructuredMatcher struct {
	objectMatcher ObjectMatcher
}

func NewUnstructuredMatcher(objectMatcher ObjectMatcher) *unstructuredMatcher {
	return &unstructuredMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two unstructured.Unstructured objects
func (m unstructuredMatcher) Match(old, new *unstructured.Unstructured) (bool, error) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.GetName())
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal new object", "name", new.GetName())
	}

	matched, err := m.objectMatcher.MatchJSON(oldData, newData, new)
	if err != nil {
		return false, emperror.WrapWith(err, "could not match objects", "name", new.GetName())
	}

	return matched, nil
}
