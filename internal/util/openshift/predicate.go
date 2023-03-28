/*
Copyright 2023 Cisco Systems, Inc. and/or its affiliates.

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

package openshift

import (
	"context"
	"encoding/json"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/cisco-open/cluster-registry-controller/pkg/clustermeta"
)

const openshiftTokenSecretAnnotationKey = "openshift.io/token-secret.name"

func IgnoreOpenshiftImagePullSecrets(client client.Client) util.CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		clusterMetadata, err := clustermeta.GetClusterMetadata(context.Background(), client)
		if err != nil {
			return []byte{}, []byte{}, err
		}

		if clusterMetadata.Distribution == clustermeta.OPENSHIFT {
			current, err = removeSecretsFromServiceAccount(client, current)
			if err != nil {
				return []byte{}, []byte{}, errors.WrapIf(err, "could not delete imagePullSecrets from service account")
			}

			modified, err = removeSecretsFromServiceAccount(client, modified)
			if err != nil {
				return []byte{}, []byte{}, errors.WrapIf(err, "could not delete imagePullSecrets from service account")
			}
		}

		return current, modified, nil
	}
}

func removeSecretsFromServiceAccount(client client.Client, obj []byte) ([]byte, error) {
	var objectMap map[string]interface{}
	err := json.Unmarshal(obj, &objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not unmarshal byte sequence")
	}

	namespace := "default"
	if metadata, ok := objectMap["metadata"].(map[string]interface{}); ok {
		value, mapHasKey := metadata["namespace"]
		ns, ok := value.(string)
		if mapHasKey && ok {
			namespace = ns
		}
	}

	if imagePullSecrets, ok := objectMap["imagePullSecrets"].([]interface{}); ok {
		objectMap["imagePullSecrets"], err = removeSecretsFromField(client, imagePullSecrets, namespace)
		if err != nil {
			return []byte{}, err
		}
	}

	if serviceAccountSecrets, ok := objectMap["secrets"].([]interface{}); ok {
		objectMap["secrets"], err = removeSecretsFromField(client, serviceAccountSecrets, namespace)
		if err != nil {
			return []byte{}, err
		}
	}

	obj, err = json.Marshal(objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not marshal byte sequence")
	}

	return obj, nil
}

func removeSecretsFromField(client client.Client, secretsField []interface{}, namespace string) ([]interface{}, error) {
	for i, s := range secretsField {
		if secret, ok := s.(map[string]interface{}); ok {
			value, mapHasKey := secret["name"]
			secretName, ok := value.(string)
			secret := &corev1.Secret{}
			if mapHasKey && ok {
				err := client.Get(context.Background(), types.NamespacedName{
					Name:      secretName,
					Namespace: namespace,
				}, secret)
				if err != nil {
					return secretsField, err
				}
				if _, ok := secret.Annotations[openshiftTokenSecretAnnotationKey]; ok {
					secretsField = append(secretsField[:i], secretsField[i+1:]...)
				}
			}
		}
	}
	return secretsField, nil
}
