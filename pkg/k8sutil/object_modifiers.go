/*
Copyright 2021 Banzai Cloud.

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
	"sort"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/goph/emperror"

	"github.com/banzaicloud/operator-tools/pkg/resources"
)

type ObjectModifierFunc = resources.ObjectModifierFunc

func RunObjectModifiers(o runtime.Object, objectModifiers []ObjectModifierFunc) (runtime.Object, error) {
	var err error

	keys := make([]int, 0, len(objectModifiers))
	for k := range objectModifiers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, m := range keys {
		o, err = objectModifiers[m](o)
		if err != nil {
			return nil, emperror.Wrap(err, "could not modify object")
		}
	}

	return o, nil
}

func CombineObjectModifiers(modifiers ...[]ObjectModifierFunc) []ObjectModifierFunc {
	oms := make([]ObjectModifierFunc, 0)

	for _, m := range modifiers {
		oms = append(oms, m...)
	}

	return oms
}

func GetObjectModifiersForOverlays(scheme *runtime.Scheme, overlays []resources.K8SResourceOverlay) ([]ObjectModifierFunc, error) {
	parser := resources.NewObjectParser(scheme)

	oms := []resources.ObjectModifierFunc{}
	for _, overlay := range overlays {
		om, err := resources.PatchYAMLModifier(overlay, parser)
		if err != nil {
			return nil, err
		}
		oms = append(oms, om)
	}

	return oms, nil
}

func GetGVKObjectModifier(scheme *runtime.Scheme) ObjectModifierFunc {
	return func(o runtime.Object) (runtime.Object, error) {
		if o.GetObjectKind().GroupVersionKind().Group == "" {
			gvks, _, err := scheme.ObjectKinds(o)
			if err != nil {
				return nil, err
			}
			if len(gvks) > 0 {
				o.GetObjectKind().SetGroupVersionKind(gvks[0])
			}
		}
		return o, nil
	}
}
