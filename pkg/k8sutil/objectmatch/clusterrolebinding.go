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

type clusterRoleBindingMatcher struct {
	objectMatcher ObjectMatcher
}

func NewClusterRoleBindingMatcher(objectMatcher ObjectMatcher) *clusterRoleBindingMatcher {
	return &clusterRoleBindingMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two rbacv1.ClusterRoleBinding objects
func (m clusterRoleBindingMatcher) Match(old, new *rbacv1.ClusterRoleBinding) (bool, error) {
	type ClusterRoleBinding struct {
		ObjectMeta
		Subjects []rbacv1.Subject `json:"subjects,omitempty"`
		RoleRef  rbacv1.RoleRef   `json:"roleRef"`
	}

	oldData, err := json.Marshal(ClusterRoleBinding{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Subjects:   old.Subjects,
		RoleRef:    old.RoleRef,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}
	newObject := ClusterRoleBinding{
		ObjectMeta: m.objectMatcher.GetObjectMeta(new.ObjectMeta),
		Subjects:   new.Subjects,
		RoleRef:    new.RoleRef,
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
