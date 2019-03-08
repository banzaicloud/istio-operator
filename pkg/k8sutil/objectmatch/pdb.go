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
	policyv1beta1 "k8s.io/api/policy/v1beta1"
)

type podDisruptionBudgetMatcher struct {
	objectMatcher ObjectMatcher
}

func NewPodDisruptionBudgetMatcher(objectMatcher ObjectMatcher) *podDisruptionBudgetMatcher {
	return &podDisruptionBudgetMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two autoscalev2beta1.HorizontalPodAutoscaler objects
func (m podDisruptionBudgetMatcher) Match(old, new *policyv1beta1.PodDisruptionBudget) (bool, error) {
	type PodDisruptionBudgetMatcher struct {
		ObjectMeta
		Spec policyv1beta1.PodDisruptionBudgetSpec
	}

	oldData, err := json.Marshal(PodDisruptionBudgetMatcher{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Spec:       old.Spec,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newObject := PodDisruptionBudgetMatcher{
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
