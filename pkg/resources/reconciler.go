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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Config *istiov1beta1.Istio
}

type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

type Resource func() runtime.Object

type ResourceVariation func(t string) runtime.Object

func ResolveVariations(t string, v []ResourceVariation) []Resource {
	resources := make([]Resource, 0)
	for i := range v {
		i := i
		resources = append(resources, func() runtime.Object {
			return v[i](t)
		})
	}
	return resources
}

type DynamicResource func() *k8sutil.DynamicObject

type DynamicResourceWithDesiredState struct {
	DynamicResource DynamicResource
	DesiredState    k8sutil.DesiredState
}
