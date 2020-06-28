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

package templates

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func ObjectMeta(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: objMeta.GetNamespace(),
		Labels:    labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         util.BoolPointer(true),
				BlockOwnerDeletion: util.BoolPointer(true),
			},
		},
	}
}

func ObjectMetaWithRevision(name string, labels map[string]string, config *v1beta1.Istio) metav1.ObjectMeta {
	return ObjectMeta(util.CombinedName(name, config.GetName()), util.MergeStringMaps(labels, config.RevisionLabels()), config)
}

func ObjectMetaClusterScopeWithRevision(name string, labels map[string]string, config *v1beta1.Istio) metav1.ObjectMeta {
	return ObjectMetaClusterScope(util.CombinedName(name, config.GetName(), config.GetNamespace()), util.MergeStringMaps(labels, config.RevisionLabels()), config)
}

func ObjectMetaWithAnnotations(name string, labels map[string]string, annotations map[string]string, config runtime.Object) metav1.ObjectMeta {
	o := ObjectMeta(name, labels, config)
	o.Annotations = annotations
	return o
}

func ObjectMetaClusterScope(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:   name,
		Labels: labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         util.BoolPointer(true),
				BlockOwnerDeletion: util.BoolPointer(true),
			},
		},
	}
}

func ControlPlaneAuthPolicy(istiodEnabled, controlPlaneSecurityEnabled bool) string {
	if !istiodEnabled && controlPlaneSecurityEnabled {
		return "MUTUAL_TLS"
	}
	return "NONE"
}
