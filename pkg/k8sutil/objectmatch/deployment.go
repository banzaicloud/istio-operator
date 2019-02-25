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
	appsv1 "k8s.io/api/apps/v1"
)

type DeploymentMatcher struct{}

// Match compares two appsv1.Deployment objects
func (m DeploymentMatcher) Match(old, new *appsv1.Deployment) (bool, error) {
	type Deployment struct {
		ObjectMeta
		Spec appsv1.DeploymentSpec
	}

	delete(old.ObjectMeta.Annotations, "deployment.kubernetes.io/revision")

	oldData, err := json.Marshal(Deployment{
		ObjectMeta: getObjectMeta(old.ObjectMeta),
		Spec:       old.Spec,
	})
	if err != nil {
		return false, emperror.Wrap(err, "could not marshal object")
	}
	newObject := Deployment{
		ObjectMeta: getObjectMeta(new.ObjectMeta),
		Spec:       new.Spec,
	}
	newData, err := json.Marshal(newObject)
	if err != nil {
		return false, emperror.Wrap(err, "could not marshal object")
	}

	matched, err := match(oldData, newData, newObject)
	if err != nil {
		return false, emperror.Wrap(err, "could not match objects")
	}

	return matched, nil
}
