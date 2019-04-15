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

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
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
			if _, ok := e.Meta.GetLabels()["istio"]; ok {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.MetaNew.GetLabels()["istio"]; ok {
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
	return predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if value, ok := e.Meta.GetLabels()["istio"]; ok && value == "ingressgateway" {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if value, ok := e.MetaNew.GetLabels()["istio"]; ok && value == "ingressgateway" {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}
