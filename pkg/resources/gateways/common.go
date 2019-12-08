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

package gateways

import (
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

func (r *Reconciler) gatewayName() string {
	return r.gw.Name
}

func (r *Reconciler) serviceAccountName() string {
	return fmt.Sprintf("%s-service-account", r.gw.Name)
}

func (r *Reconciler) labels() map[string]string {
	return util.MergeStringMaps(map[string]string{
		"gateway-name": r.gatewayName(),
		"gateway-type": string(r.gw.Spec.Type),
	}, r.gw.Spec.Labels)
}

func (r *Reconciler) clusterRoleName() string {
	return fmt.Sprintf("%s-cluster-role", r.gw.Name)
}

func (r *Reconciler) clusterRoleBindingName() string {
	return fmt.Sprintf("%s-cluster-role-binding", r.gw.Name)
}

func (r *Reconciler) roleName() string {
	return fmt.Sprintf("%s-role-sds", r.gw.Name)
}

func (r *Reconciler) roleBindingName() string {
	return fmt.Sprintf("%s-role-binding-sds", r.gw.Name)
}

func (r *Reconciler) labelSelector() map[string]string {
	return r.labels()
}

func (r *Reconciler) pdbName() string {
	return r.gw.Name
}

func (r *Reconciler) hpaName() string {
	return fmt.Sprintf("%s-autoscaler", r.gw.Name)
}
