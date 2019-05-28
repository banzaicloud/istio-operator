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

package istiocoredns

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

// Add/Remove global:53 to/from coredns configmap
func (r *Reconciler) reconcileCoreDNSConfigMap(log logr.Logger, desiredState k8sutil.DesiredState) error {
	var cm apiv1.ConfigMap

	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      "coredns",
		Namespace: "kube-system",
	}, &cm)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return emperror.Wrap(err, "could not get coredns configmap")
	}

	coreFile := cm.Data["CoreFile"]

	var svc apiv1.Service
	err = r.Client.Get(context.Background(), types.NamespacedName{
		Name:      serviceName,
		Namespace: r.Config.Namespace,
	}, &svc)
	if err != nil {
		return emperror.Wrap(err, "could not get Istio coreDNS service")
	}
	clusterIP := []string{svc.Spec.ClusterIP}
	global53 := fmt.Sprintf(`global:53 {
			errors
			cache 30
			proxy . %s
			}))`, clusterIP)

	if desiredState == k8sutil.DesiredStatePresent {
		if !strings.Contains(coreFile, global53) {
			coreFile += global53
		}
	} else if desiredState == k8sutil.DesiredStateAbsent {
		if strings.Contains(coreFile, global53) {
			strings.Replace(coreFile, global53, "", -1)
		}
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["CoreFile"] = coreFile

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update coredns configmap")
	}

	return nil
}
