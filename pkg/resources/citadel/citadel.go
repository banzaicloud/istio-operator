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

package citadel

import (
	"fmt"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/helm"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/helm/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	componentName  = "citadel"
	deploymentName = "istio-citadel"
)

var citadelLabels = map[string]string{
	"app": "security",
}

var labelSelector = map[string]string{
	"istio": "citadel",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface

	configuration Configuration
}

func New(configuration Configuration, client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, manifests []manifest.Manifest, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:    client,
			Config:    config,
			Manifests: manifests,
			Scheme:    scheme,
		},
		dynamic:       dc,
		configuration: configuration,
	}
}

func GetDeploymentName() string {
	return deploymentName
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	objects, err := helm.DecodeObjects(log, r.Manifests)
	if err != nil {
		return emperror.Wrap(err, "failed to decode objects from chart")
	}

	for _, o := range objects {
		fmt.Printf("***type: %T\n", o)

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

	//if !*r.configuration.CreateMeshPolicy {
	//	return nil
	//}
	//
	//var mTLSDesiredState k8sutil.DesiredState
	//if r.Config.Spec.MTLS {
	//	mTLSDesiredState = k8sutil.DesiredStatePresent
	//} else {
	//	mTLSDesiredState = k8sutil.DesiredStateAbsent
	//}
	//drs := []resources.DynamicResourceWithDesiredState{
	//	{DynamicResource: r.meshPolicy, DesiredState: k8sutil.DesiredStatePresent},
	//	{DynamicResource: r.destinationRuleDefaultMtls, DesiredState: mTLSDesiredState},
	//	{DynamicResource: r.destinationRuleApiServerMtls, DesiredState: mTLSDesiredState},
	//}
	//
	//for _, dr := range drs {
	//	o := dr.DynamicResource()
	//	err := o.Reconcile(log, r.dynamic, dr.DesiredState)
	//	if err != nil {
	//		return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
	//	}
	//}
	//
	log.Info("Reconciled")

	return nil
}
