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

package controllers

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
)

func setDynamicDefaults(icp *v1alpha1.IstioControlPlane, k8sConfig *rest.Config, logger logr.Logger) {
	if icp.Spec.JwtPolicy == v1alpha1.JWTPolicyType_UNSPECIFIED {
		// try to detect supported jwt policy
		supportedJWTPolicy, err := k8sutil.DetectSupportedJWTPolicy(k8sConfig)
		if err != nil {
			logger.Error(err, "could not detect supported jwt policy")
		} else {
			icp.Spec.JwtPolicy = supportedJWTPolicy
			logger.Info("supported jwt policy", "policy", icp.Spec.JwtPolicy)
		}
	}
}
