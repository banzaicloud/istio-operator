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
	corev1 "k8s.io/api/core/v1"
)

type configMapMatcher struct {
	objectMatcher ObjectMatcher
}

func NewConfigMapMatcher(objectMatcher ObjectMatcher) *configMapMatcher {
	return &configMapMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two corev1.ConfigMap objects
func (m configMapMatcher) Match(old, new *corev1.ConfigMap) (bool, error) {
	type ConfigMap struct {
		ObjectMeta
		Data       map[string]string `json:"data"`
		BinaryData map[string][]byte `json:"binaryData"`
	}

	oldData, err := json.Marshal(ConfigMap{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Data:       old.Data,
		BinaryData: old.BinaryData,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newConfigMap := ConfigMap{
		ObjectMeta: m.objectMatcher.GetObjectMeta(new.ObjectMeta),
		Data:       new.Data,
		BinaryData: new.BinaryData,
	}
	newData, err := json.Marshal(newConfigMap)
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal new object", "name", new.Name)
	}

	matched, err := m.objectMatcher.MatchJSON(oldData, newData, newConfigMap)
	if err != nil {
		return false, emperror.WrapWith(err, "could not match objects", "name", new.Name)
	}

	return matched, nil
}
