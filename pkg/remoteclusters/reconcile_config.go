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

package remoteclusters

import (
	"context"
	"fmt"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) reconcileConfig(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling config")

	var istioConfig istiov1beta1.Istio
	err := c.ctrlRuntimeClient.Get(context.TODO(), types.NamespacedName{
		Name:      istio.Name,
		Namespace: remoteConfig.Namespace,
	}, &istioConfig)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		return err
	}

	caSecretName := CASecretName
	if remoteConfig.Spec.Citadel.CASecretName != "" {
		caSecretName = remoteConfig.Spec.Citadel.CASecretName
	}

	istioConfig.Spec = istio.Spec
	if len(remoteConfig.Spec.AutoInjectionNamespaces) > 0 {
		istioConfig.Spec.AutoInjectionNamespaces = remoteConfig.Spec.AutoInjectionNamespaces
	}
	istioConfig.Spec.SidecarInjector.ReplicaCount = remoteConfig.Spec.SidecarInjector.ReplicaCount
	istioConfig.Spec.Proxy.Privileged = remoteConfig.Spec.Proxy.Privileged
	istioConfig.Spec.Citadel.NodeSelector = remoteConfig.Spec.Citadel.NodeSelector
	istioConfig.Spec.Citadel.Affinity = remoteConfig.Spec.Citadel.Affinity
	istioConfig.Spec.Citadel.Tolerations = remoteConfig.Spec.Citadel.Tolerations
	istioConfig.Spec.Citadel.CASecretName = caSecretName
	istioConfig.Spec.SidecarInjector.NodeSelector = remoteConfig.Spec.SidecarInjector.NodeSelector
	istioConfig.Spec.SidecarInjector.Affinity = remoteConfig.Spec.SidecarInjector.Affinity
	istioConfig.Spec.SidecarInjector.Tolerations = remoteConfig.Spec.SidecarInjector.Tolerations
	istioConfig.Spec.SidecarInjector.InitCNIConfiguration.Affinity = remoteConfig.Spec.SidecarInjector.InitCNIConfiguration.Affinity
	istioConfig.Spec.SidecarInjector.InjectedContainerAdditionalEnvVars = remoteConfig.Spec.SidecarInjector.InjectedContainerAdditionalEnvVars

	istioConfig.Spec.SidecarInjector.Enabled = util.BoolPointer(true)
	istioConfig.Spec.Citadel.Enabled = util.BoolPointer(true)

	if util.PointerToBool(istio.Spec.Istiod.Enabled) {
		istioConfig.Spec.Citadel.SDSEnabled = util.BoolPointer(true)
		istioConfig.Spec.Citadel.ListenedNamespaces = &remoteConfig.Namespace
		if util.PointerToBool(istio.Spec.Istiod.MultiClusterSupport) {
			istioConfig.Spec.Citadel.Enabled = util.BoolPointer(false)
		} else {
			istioConfig.Spec.CAAddress = fmt.Sprintf("istio-citadel.%s.svc.%s:15012", istio.Namespace, istio.Spec.Proxy.ClusterDomain)
		}
	}

	if remoteConfig.Spec.DefaultResources != nil {
		istioConfig.Spec.DefaultResources = remoteConfig.Spec.DefaultResources
	}

	if remoteConfig.Spec.Proxy.Resources != nil {
		istioConfig.Spec.Proxy.Resources = remoteConfig.Spec.Proxy.Resources
	}

	if util.PointerToBool(istioConfig.Spec.MeshExpansion) {
		istioConfig.Spec.Gateways.MeshExpansion.Enabled = istio.Spec.Gateways.MeshExpansion.Enabled
		istioConfig.Spec.Gateways.Ingress.Enabled = istio.Spec.Gateways.Ingress.Enabled
	}

	if remoteConfig.Spec.NetworkName == "" {
		remoteConfig.Spec.NetworkName = remoteConfig.Name
	}
	istioConfig.Spec.NetworkName = remoteConfig.Spec.NetworkName
	istioConfig.Spec.ClusterName = remoteConfig.Name

	istioConfig.SetDefaults()

	if k8sapierrors.IsNotFound(err) {
		istioConfig.Name = istio.Name
		istioConfig.Namespace = remoteConfig.Namespace

		err = c.ctrlRuntimeClient.Create(context.TODO(), &istioConfig)
		if err != nil {
			return err
		}
	} else {
		err = c.ctrlRuntimeClient.Update(context.TODO(), &istioConfig)
		if err != nil {
			return err
		}
	}

	istioConfig.TypeMeta = metav1.TypeMeta{Kind: "Istio", APIVersion: istiov1beta1.SchemeGroupVersion.WithKind("Istio").GroupVersion().String()}
	c.istioConfig = &istioConfig

	return nil
}
