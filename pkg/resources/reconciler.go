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

package resources

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

type Reconciler struct {
	client.Client
	Config *istiov1beta1.Istio
	Scheme *runtime.Scheme
}

type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
	Cleanup(log logr.Logger) error
}

type Resource func() runtime.Object

type ResourceVariation func(t string) runtime.Object

func ResolveVariations(t string, v []ResourceVariationWithDesiredState, desiredState k8sutil.DesiredState) []ResourceWithDesiredState {
	var state k8sutil.DesiredState
	resources := make([]ResourceWithDesiredState, 0)
	for i := range v {
		i := i

		if v[i].DesiredState == k8sutil.DesiredStateAbsent || desiredState == k8sutil.DesiredStateAbsent {
			state = k8sutil.DesiredStateAbsent
		} else {
			state = k8sutil.DesiredStatePresent
		}

		resource := ResourceWithDesiredState{
			func() runtime.Object {
				return v[i].ResourceVariation(t)
			},
			state,
		}
		resources = append(resources, resource)
	}

	return resources
}

type DynamicResource func() *k8sutil.DynamicObject

type DynamicResourceWithDesiredState struct {
	DynamicResource DynamicResource
	DesiredState    k8sutil.DesiredState
}

type ResourceWithDesiredState struct {
	Resource     Resource
	DesiredState k8sutil.DesiredState
}

type ResourceVariationWithDesiredState struct {
	ResourceVariation ResourceVariation
	DesiredState      k8sutil.DesiredState
}
