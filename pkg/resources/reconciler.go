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
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Owner *istiov1beta1.Config
}

type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

type Resource func(owner *istiov1beta1.Config) runtime.Object

type ResourceVariation func(t string, owner *istiov1beta1.Config) runtime.Object

func ResolveVariations(t string, v []ResourceVariation) []Resource {
	resources := make([]Resource, 0)
	for i := range v {
		i := i
		resources = append(resources, func(owner *istiov1beta1.Config) runtime.Object {
			return v[i](t, owner)
		})
	}
	return resources
}

type DynamicResource func(owner *istiov1beta1.Config) *k8sutil.DynamicObject
