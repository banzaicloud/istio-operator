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
	"context"

	"emperror.dev/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterregistryv1alpha1 "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
)

func GetLocalCluster(ctx context.Context, kubeClient client.Client) (*clusterregistryv1alpha1.Cluster, error) {
	var cluster *clusterregistryv1alpha1.Cluster

	clusters := &clusterregistryv1alpha1.ClusterList{}
	err := kubeClient.List(ctx, clusters)
	if err != nil {
		return cluster, errors.WithStackIf(err)
	}

	counter := 0
	for _, c := range clusters.Items {
		c := c
		if c.Status.Type == clusterregistryv1alpha1.ClusterTypeLocal {
			counter++
			if counter > 1 {
				return cluster, errors.WithStackIf(errors.New("multiple local Cluster CR found, there should only be one"))
			}
			cluster = &c
		}
	}

	if counter == 0 {
		return cluster, errors.WithStackIf(errors.New("no local Cluster CR found, either there should be one or cluster-registry-api-enabled arg should be set to false"))
	}

	return cluster, nil
}
