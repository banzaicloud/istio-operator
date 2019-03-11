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
	rbacv1 "k8s.io/api/rbac/v1"
)

type roleMatcher struct {
	objectMatcher ObjectMatcher
}

func NewRoleMatcher(objectMatcher ObjectMatcher) *roleMatcher {
	return &roleMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two rbacv1.ClusterRole objects
func (m roleMatcher) Match(old, new *rbacv1.Role) (bool, error) {
	type Role struct {
		ObjectMeta
		Rules []rbacv1.PolicyRule `json:"rules"`
	}

	oldData, err := json.Marshal(Role{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Rules:      old.Rules,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newObject := Role{
		ObjectMeta: m.objectMatcher.GetObjectMeta(new.ObjectMeta),
		Rules:      new.Rules,
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
