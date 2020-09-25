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
	"context"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

const (
	managedAutoInjectionLabelKey    = "istio-operator-managed-injection"
	autoInjectionLabelKey           = "istio-injection"
	revisionedAutoInjectionLabelKey = "istio.io/rev"
)

func (r *Reconciler) reconcileAutoInjectionLabels(log logr.Logger) error {
	var autoInjectLabels = map[string]string{
		autoInjectionLabelKey:        "enabled",
		managedAutoInjectionLabelKey: "enabled",
	}

	managedNamespaces := make(map[string]bool)
	for _, ns := range r.Config.Spec.AutoInjectionNamespaces {
		managedNamespaces[ns] = true
		err := k8sutil.ReconcileNamespaceLabelsIgnoreNotFound(log, r.Client, ns, autoInjectLabels, nil, revisionedAutoInjectionLabelKey)
		if err != nil {
			log.Error(err, "failed to label namespace", "namespace", ns)
		}
	}

	var namespaces corev1.NamespaceList
	err := r.Client.List(context.Background(), &namespaces, client.MatchingLabels(map[string]string{
		managedAutoInjectionLabelKey: autoInjectLabels[managedAutoInjectionLabelKey],
	}))
	if err != nil {
		return emperror.Wrap(err, "could not list namespaces")
	}

	for _, ns := range namespaces.Items {
		if !managedNamespaces[ns.Name] {
			err := k8sutil.ReconcileNamespaceLabelsIgnoreNotFound(log, r.Client, ns.Name, nil, []string{
				autoInjectionLabelKey,
				managedAutoInjectionLabelKey,
			}, revisionedAutoInjectionLabelKey)
			if err != nil {
				log.Error(emperror.Wrap(err, "failed to label namespace"), "namespace", ns.Name)
			}
		}
	}

	return nil
}
