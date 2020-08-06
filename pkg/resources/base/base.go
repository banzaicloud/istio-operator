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

package base

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName                 = "common"
	istiodName                    = "istiod"
	istiodPilotName               = "istiod-pilot"
	istioReaderName               = "istio-reader"
	istioReaderServiceAccountName = "istio-reader-service-account"
	istiodServiceAccountName      = "istiod-service-account"
	IstioConfigMapName            = "istio"
)

var istiodLabel = map[string]string{
	"app": istiodName,
}

var pilotLabel = map[string]string{
	"app": "pilot",
}

var istioReaderLabel = map[string]string{
	"app": istioReaderName,
}

type Reconciler struct {
	resources.Reconciler
	remote bool
}

func New(client client.Client, config *istiov1beta1.Istio, isRemote bool) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		remote: isRemote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	for _, res := range []resources.Resource{
		r.serviceAccountReader,
		r.serviceAccount,
		r.role,
		r.clusterRole,
		r.clusterRoleReader,
		r.clusterRoleBindingReader,
		r.clusterRoleBinding,
		r.roleBinding,
		r.configMap,
	} {
		o := res()
		err := k8sutil.Reconcile(log, r.Client, o, k8sutil.DesiredStatePresent)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
