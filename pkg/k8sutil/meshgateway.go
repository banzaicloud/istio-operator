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

package k8sutil

import (
	"context"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

func GetMeshGatewayAddress(client client.Client, key client.ObjectKey) ([]string, error) {
	var mgw istiov1beta1.MeshGateway

	ips := make([]string, 0)

	err := client.Get(context.TODO(), key, &mgw)
	if err != nil && !k8serrors.IsNotFound(err) {
		return ips, err
	}

	if mgw.Status.Status != istiov1beta1.Available {
		return ips, errors.New("gateway is pending")
	}

	ips = mgw.Status.GatewayAddress

	return ips, nil
}
