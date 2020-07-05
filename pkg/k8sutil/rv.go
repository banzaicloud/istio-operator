/*
Copyright 2020 Banzai Cloud.

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

package k8sutil

import (
	"github.com/Masterminds/semver/v3"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceRevisionLabel = "resource.alpha.banzaicloud.io/revision"
)

func SetResourceRevision(obj runtime.Object, revision string) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	labels := m.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	m.SetLabels(util.MergeStringMaps(labels, map[string]string{
		ResourceRevisionLabel: revision,
	}))

	return nil
}

func GetResourceRevision(obj runtime.Object) (string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}

	return m.GetLabels()[ResourceRevisionLabel], nil
}

func CheckResourceRevision(obj runtime.Object, revisionConstraint string) (bool, error) {
	semverConstraint, err := semver.NewConstraint(revisionConstraint)
	if err != nil {
		return false, emperror.Wrap(err, "could not create semver constraint")
	}
	currentRevision, err := GetResourceRevision(obj)
	if err != nil {
		return false, emperror.Wrap(err, "could not get current revision")
	}

	if currentRevision != "" {
		if currentSemver, err := semver.NewVersion(currentRevision); err == nil && !semverConstraint.Check(currentSemver) {
			return false, nil
		}
	}

	return true, nil
}
