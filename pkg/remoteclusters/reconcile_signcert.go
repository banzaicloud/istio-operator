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

	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (c *Cluster) reconcileSignCert(remoteConfig *istiov1beta1.RemoteIstio) error {
	c.log.Info("reconciling sign cert")

	var secret corev1.Secret
	err := c.ctrlRuntimeClient.Get(context.TODO(), client.ObjectKey{
		Namespace: remoteConfig.Namespace,
		Name:      "cacerts",
	}, &secret)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		return err
	}

	signCert := remoteConfig.Spec.GetSignCert()
	secretData := map[string][]byte{
		"root-cert.pem":  signCert.Root,
		"ca-cert.pem":    signCert.CA,
		"ca-key.pem":     signCert.Key,
		"cert-chain.pem": signCert.Chain,
	}

	if k8sapierrors.IsNotFound(err) {
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cacerts",
				Namespace: remoteConfig.Namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: secretData,
		}
		secret.SetOwnerReferences([]metav1.OwnerReference{
			{
				Kind:               c.istioConfig.Kind,
				APIVersion:         c.istioConfig.APIVersion,
				Name:               c.istioConfig.Name,
				UID:                c.istioConfig.GetUID(),
				Controller:         util.BoolPointer(true),
				BlockOwnerDeletion: util.BoolPointer(true),
			},
		})
		err = c.ctrlRuntimeClient.Create(context.TODO(), &secret)
		if err != nil {
			return err
		}

		return nil
	}

	secret.Data = secretData
	err = c.ctrlRuntimeClient.Update(context.TODO(), &secret)
	if err != nil {
		return err
	}

	return nil
}
