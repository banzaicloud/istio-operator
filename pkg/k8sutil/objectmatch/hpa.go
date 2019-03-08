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
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
)

type horizontalPodAutoscalerMatcher struct {
	objectMatcher ObjectMatcher
}

func NewHorizontalPodAutoscalerMatcher(objectMatcher ObjectMatcher) *horizontalPodAutoscalerMatcher {
	return &horizontalPodAutoscalerMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two autoscalev2beta1.HorizontalPodAutoscaler objects
func (m horizontalPodAutoscalerMatcher) Match(old, new *autoscalev2beta1.HorizontalPodAutoscaler) (bool, error) {
	type HorizontalPodAutoscaler struct {
		ObjectMeta
		Spec autoscalev2beta1.HorizontalPodAutoscalerSpec
	}

	oldData, err := json.Marshal(HorizontalPodAutoscaler{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Spec:       old.Spec,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newObject := HorizontalPodAutoscaler{
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
