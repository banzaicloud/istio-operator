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
	"context"

	"emperror.dev/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
	"github.com/banzaicloud/operator-tools/pkg/logger"
)

func setDynamicDefaults(ctx context.Context, kubeClient client.Client, icp *v1alpha1.IstioControlPlane, k8sConfig *rest.Config, logger logger.Logger, clusterRegistryAPIEnabled bool) error {
	if icp.Spec.JwtPolicy == v1alpha1.JWTPolicyType_JWTPolicyType_UNSPECIFIED {
		// try to detect supported jwt policy
		supportedJWTPolicy, err := k8sutil.DetectSupportedJWTPolicy(k8sConfig)
		if err != nil {
			logger.Error(err, "could not detect supported jwt policy")
		} else {
			icp.Spec.JwtPolicy = supportedJWTPolicy
			logger.V(1).Info("supported jwt policy", "policy", icp.Spec.JwtPolicy)
		}
	}

	if icp.Spec.ClusterID == "" {
		icp.Spec.ClusterID = "Kubernetes"
		if clusterRegistryAPIEnabled {
			cluster, err := k8sutil.GetLocalCluster(ctx, kubeClient)
			if err != nil {
				return errors.WithStackIf(err)
			}

			icp.Spec.ClusterID = cluster.GetName()
		}
	}

	return nil
}
