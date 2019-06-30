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

package sidecarinjector

import (
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/util"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) podDisruptionBudget() runtime.Object {
	labels := util.MergeLabels(sidecarInjectorLabels, labelSelector)
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: templates.ObjectMeta(pdbName, labels, r.Config),
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: util.IntstrPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}
