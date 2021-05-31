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

package k8sutil

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/go-logr/logr"
)

func GetWatchPredicateForIstio() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*istiov1beta1.Istio)
			new := e.ObjectNew.(*istiov1beta1.Istio)
			if !reflect.DeepEqual(old.Spec, new.Spec) ||
				old.GetDeletionTimestamp() != new.GetDeletionTimestamp() ||
				old.GetGeneration() != new.GetGeneration() {
				return true
			}
			return false
		},
	}
}

func GetWatchPredicateForRemoteIstioAvailability() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			new := e.ObjectNew.(*istiov1beta1.RemoteIstio)
			if new.Status.Status == istiov1beta1.Available {
				return true
			}
			return false
		},
	}
}

func GetWatchPredicateForRemoteIstio() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*istiov1beta1.RemoteIstio)
			new := e.ObjectNew.(*istiov1beta1.RemoteIstio)
			if !reflect.DeepEqual(old.Spec, new.Spec) ||
				old.GetDeletionTimestamp() != new.GetDeletionTimestamp() ||
				old.GetGeneration() != new.GetGeneration() {
				return true
			}
			return false
		},
	}
}

func GetWatchPredicateForIstioServicePods() predicate.Funcs {
	return predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if _, ok := e.Object.GetLabels()["istio"]; ok {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.ObjectNew.GetLabels()["istio"]; ok {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func GetWatchPredicateForIstioService(name string) predicate.Funcs {
	return predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if value, ok := e.Object.GetLabels()["istio"]; ok && value == name {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if value, ok := e.ObjectNew.GetLabels()["istio"]; ok && value == name {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func GetWatchPredicateForIstioIngressGateway() predicate.Funcs {
	return GetWatchPredicateForIstioService("ingressgateway")
}

func GetWatchPredicateForMeshGateway() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.ObjectOld.(*istiov1beta1.MeshGateway); !ok {
				return false
			}
			old := e.ObjectOld.(*istiov1beta1.MeshGateway)
			new := e.ObjectNew.(*istiov1beta1.MeshGateway)
			if !reflect.DeepEqual(old.Spec, new.Spec) ||
				old.GetDeletionTimestamp() != new.GetDeletionTimestamp() ||
				old.GetGeneration() != new.GetGeneration() ||
				!reflect.DeepEqual(old.Status.GatewayAddress, new.Status.GatewayAddress) {
				return true
			}
			return false
		},
	}
}

func GetWatchPredicateForOwnedResources(owner runtime.Object, isController bool, scheme *runtime.Scheme, logger logr.Logger) predicate.Funcs {
	ownerMatcher := NewOwnerReferenceMatcher(owner, isController, scheme)
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// If a new namespace is created, we need to reconcile to mutate the injection labels
			if _, ok := e.Object.(*corev1.Namespace); ok {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// We don't want to run reconcile if a namespace is deleted
			if _, ok := e.Object.(*corev1.Namespace); ok {
				return false
			}
			related, object, err := ownerMatcher.Match(e.Object)
			if err != nil {
				logger.Error(err, "could not determine relation", "kind", e.Object.GetObjectKind())
			}
			if related {
				logger.Info("related object deleted", "trigger", object.GetName())
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// If a namespace is updated, we need to reconcile to mutate the injection labels
			if _, ok := e.ObjectNew.(*corev1.Namespace); ok {
				return true
			}
			related, object, err := ownerMatcher.Match(e.ObjectNew)
			if err != nil {
				logger.Error(err, "could not determine relation", "kind", e.ObjectNew.GetObjectKind())
			}
			if related {
				changed, err := IsObjectChanged(e.ObjectOld, e.ObjectNew, true)
				if err != nil {
					logger.Error(err, "could not check whether object is changed", "kind", e.ObjectNew.GetObjectKind())
				}
				if !changed {
					return false
				}

				logger.Info("related object changed", "trigger", object.GetName())
			}
			return true
		},
	}
}
