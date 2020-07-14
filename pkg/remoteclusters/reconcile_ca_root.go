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

	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	caRootConfigMapName = "istio-ca-root-cert"
	certKeyInConfigMap  = "root-cert.pem"
)

func (c *Cluster) reconcileCARootToNamespaces(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	desiredState := k8sutil.DesiredStatePresent
	if !util.PointerToBool(istio.Spec.Istiod.Enabled) || istio.Spec.Pilot.CertProvider != istiov1beta1.PilotCertProviderTypeIstiod {
		desiredState = k8sutil.DesiredStateAbsent
	}

	var namespaces corev1.NamespaceList

	err := c.ctrlRuntimeClient.List(context.Background(), &client.ListOptions{}, &namespaces)
	if err != nil {
		return err
	}

	signCert := remoteConfig.Spec.GetSignCert()
	configMapData := map[string]string{
		certKeyInConfigMap: string(signCert.Root),
	}

	for _, ns := range namespaces.Items {
		err = c.reconcileCARootInNamespace(istio.WithName(caRootConfigMapName), ns.Name, configMapData, desiredState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) reconcileCARootInNamespace(name, namespace string, configMapData map[string]string, desiredState k8sutil.DesiredState) error {
	c.log.Info("reconciling ca root configmap", "namespace", namespace)

	configmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: configMapData,
	}
	ownerRef, err := k8sutil.SetOwnerReferenceToObject(&configmap, c.istioConfig)
	if err != nil {
		return err
	}
	configmap.SetOwnerReferences(ownerRef)

	err = k8sutil.Reconcile(c.log, c.ctrlRuntimeClient, &configmap, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile resource", "resource", configmap.GetObjectKind().GroupVersionKind())
	}

	c.log.Info("ca root configmap reconciled", "namespace", namespace)

	return nil
}
