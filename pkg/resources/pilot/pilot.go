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

package pilot

import (
	"fmt"
	"github.com/banzaicloud/istio-operator/pkg/helm"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	componentName = "pilot"
)

type Reconciler struct {
	resources.Reconciler
}

func New(client client.Client, config *istiov1beta1.Istio, manifests []manifest.Manifest, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:    client,
			Config:    config,
			Manifests: manifests,
			Scheme:    scheme,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	objects, err := helm.DecodeObjects(log, r.Manifests)
	if err != nil {
		return emperror.Wrap(err, "failed to decode objects from chart")
	}

	for _, o := range objects {
		ro := o.(runtime.Object)
		err := controllerutil.SetControllerReference(r.Config, o, r.Scheme)
		if err != nil {
			return emperror.WrapWith(err, "failed to set controller reference", "resource", ro.GetObjectKind().GroupVersionKind())
		}
		err = k8sutil.Reconcile(log, r.Client, ro)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", ro.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
