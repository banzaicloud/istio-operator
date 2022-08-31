/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
)

func DetectSupportedJWTPolicy(k8sConfig *rest.Config) (servicemeshv1alpha1.JWTPolicyType, error) {
	d, err := discovery.NewDiscoveryClientForConfig(k8sConfig)
	if err != nil {
		return servicemeshv1alpha1.JWTPolicyType_JWTPolicyType_UNSPECIFIED, err
	}

	_, s, err := d.ServerGroupsAndResources()
	if err != nil {
		return servicemeshv1alpha1.JWTPolicyType_JWTPolicyType_UNSPECIFIED, err
	}

	for _, res := range s {
		for _, api := range res.APIResources {
			if api.Name == "serviceaccounts/token" {
				return servicemeshv1alpha1.JWTPolicyType_THIRD_PARTY_JWT, nil
			}
		}
	}

	return servicemeshv1alpha1.JWTPolicyType_FIRST_PARTY_JWT, nil
}
