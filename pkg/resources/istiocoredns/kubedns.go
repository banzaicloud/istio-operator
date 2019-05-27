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
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

// Add/Remove .global to/from kube-dns configmap
func (r *Reconciler) reconcileKubeDNSConfigMap(log logr.Logger, desiredState k8sutil.DesiredState) error {
	var cm apiv1.ConfigMap

	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      "kube-dns",
		Namespace: "kube-system",
	}, &cm)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return emperror.Wrap(err, "could not get kube-dns configmap")
	}

	stubDomains := make(map[string][]string, 0)
	if cm.Data["stubDomains"] != "" {
		err = json.Unmarshal([]byte(cm.Data["stubDomains"]), &stubDomains)
		if err != nil {
			return emperror.Wrap(err, "could not unmarshal stubDomains")
		}
	}

	if desiredState == k8sutil.DesiredStatePresent {
		var svc apiv1.Service
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      serviceName,
			Namespace: r.Config.Namespace,
		}, &svc)
		if err != nil {
			return emperror.Wrap(err, "could not get Istio coreDNS service")
		}
		stubDomains["global"] = []string{svc.Spec.ClusterIP}
	} else if desiredState == k8sutil.DesiredStateAbsent {
		_, ok := stubDomains["global"]
		if ok {
			delete(stubDomains, "global")
		}
	}

	stubDomainsData, err := json.Marshal(&stubDomains)
	if err != nil {
		return emperror.Wrap(err, "could not marshal updated stub domains")
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["stubDomains"] = string(stubDomainsData)

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update kube-dns configmap")
	}

	return nil
}
