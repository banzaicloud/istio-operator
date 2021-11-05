/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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
	"emperror.dev/errors"
	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/v2/pkg/util"
)

const (
	resourceRevisionLabel = "resource.alpha.banzaicloud.io/revision"
)

func SetResourceRevisionLabel(obj client.Object, revision string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	obj.SetLabels(util.MergeStringMaps(labels, map[string]string{
		resourceRevisionLabel: revision,
	}))
}

func GetResourceRevisionLabel(obj client.Object) string {
	return obj.GetLabels()[resourceRevisionLabel]
}

func CheckResourceRevision(obj client.Object, revisionConstraint string) (bool, error) {
	semverConstraint, err := semver.NewConstraint(revisionConstraint)
	if err != nil {
		return false, errors.WrapIf(err, "could not create semver constraint")
	}
	currentRevision := GetResourceRevisionLabel(obj)

	if currentRevision != "" {
		if currentSemver, err := semver.NewVersion(currentRevision); err == nil && !semverConstraint.Check(currentSemver) {
			return false, nil
		}
	}

	return true, nil
}
