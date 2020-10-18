/*
Copyright 2020 Banzai Cloud.

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

package remoteclusters

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) reconcileNamespaceInjectionLabels(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	localNamespaces := make(map[string]bool, 0)
	namespaces := &corev1.NamespaceList{}
	// local -> remote
	err := c.cl.List(context.Background(), namespaces, client.MatchingLabels(istio.RevisionLabels()))
	if err != nil {
		return err
	}

	for _, ns := range namespaces.Items {
		err = c.reconcileNamespaceInjectionLabel(ns.Name, istio.LegacyInjectionLabels(), istio.RevisionLabels())
		if err != nil {
			return err
		}
		localNamespaces[ns.Name] = true
	}

	remoteNamespaces := &corev1.NamespaceList{}
	// remote -> local
	c.ctrlRuntimeClient.List(context.Background(), remoteNamespaces, client.MatchingLabels(istio.RevisionLabels()))
	if err != nil {
		return err
	}
	for _, ns := range remoteNamespaces.Items {
		if localNamespaces[ns.Name] {
			continue
		}
		ns.Labels = util.ReduceMapByMap(ns.Labels, istio.RevisionLabels())
		err = c.ctrlRuntimeClient.Update(context.Background(), &ns)
		if err != nil {
			return err
		}
		c.log.V(1).Info("injection label removed", "namespace", ns.Name, "labels", istio.RevisionLabels())
	}

	return nil
}

func (c *Cluster) reconcileNamespaceInjectionLabel(name string, legacyInjectionLabels, matchingLabels map[string]string) error {
	ns := &corev1.Namespace{}

	// get namespace
	err := c.ctrlRuntimeClient.Get(context.Background(), client.ObjectKey{
		Name: name,
	}, ns)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// check if the labels already set
	if labels.SelectorFromValidatedSet(matchingLabels).Matches(labels.Set(ns.GetLabels())) && !labels.SelectorFromValidatedSet(legacyInjectionLabels).Matches(labels.Set(ns.GetLabels())) {
		c.log.V(1).Info("injection label already matches", "namespace", ns.Name, "labels", matchingLabels)
		return nil
	}

	// update labels
	ns.Labels = util.ReduceMapByMap(util.MergeStringMaps(ns.Labels, matchingLabels), legacyInjectionLabels)
	err = c.ctrlRuntimeClient.Update(context.Background(), ns)
	if err != nil {
		return err
	}

	c.log.V(1).Info("injection label updated", "namespace", ns.Name, "labels", matchingLabels)

	return nil
}
