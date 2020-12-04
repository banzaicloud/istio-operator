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

package egressgateway

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "egressgateway"
	resourceName  = "istio-egressgateway"
)

var (
	resourceLabels = map[string]string{
		"app":   "istio-egressgateway",
		"istio": "egressgateway",
	}
)

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
	remote  bool
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, remote bool) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
		remote:  remote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	spec := istiov1beta1.MeshGatewaySpec{
		MeshGatewayConfiguration: r.Config.Spec.Gateways.Egress.MeshGatewayConfiguration,
		Ports:                    r.Config.Spec.Gateways.Egress.Ports,
		Type:                     istiov1beta1.GatewayTypeEgress,
	}
	spec.IstioControlPlane = &istiov1beta1.NamespacedName{
		Name:      r.Config.Name,
		Namespace: r.Config.Namespace,
	}
	spec.Labels = r.labels()
	objectMeta := templates.ObjectMetaWithRevision(resourceName, spec.Labels, r.Config)

	var desiredState k8sutil.DesiredState

	if util.PointerToBool(r.Config.Spec.Gateways.Enabled) && util.PointerToBool(r.Config.Spec.Gateways.Egress.Enabled) {
		desiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.Gateways.Egress.CreateOnly) {
			objectMeta.OwnerReferences = nil
			desiredState = k8sutil.MeshGatewayCreateOnlyDesiredState{}
		}
	} else {
		desiredState = k8sutil.DesiredStateAbsent
	}

	object := &istiov1beta1.MeshGateway{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
	object.SetDefaultLabels()

	err := k8sutil.Reconcile(log, r.Client, object, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile resource", "resource", object.GetObjectKind().GroupVersionKind())
	}

	var multimeshEgressGatewayDesiredState k8sutil.DesiredState
	if desiredState != k8sutil.DesiredStateAbsent && util.PointerToBool(r.Config.Spec.MultiMesh) && util.PointerToBool(r.Config.Spec.Gateways.Egress.Enabled) {
		multimeshEgressGatewayDesiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.Gateways.Egress.CreateOnly) {
			multimeshEgressGatewayDesiredState = k8sutil.DesiredStateExists
		}
	} else {
		multimeshEgressGatewayDesiredState = k8sutil.DesiredStateAbsent
	}

	if r.remote {
		log.Info("Reconciled")
		return nil
	}

	var drs = []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.multimeshEgressGateway, DesiredState: multimeshEgressGatewayDesiredState},
	}
	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, dr.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}

func (r *Reconciler) labels() map[string]string {
	return util.MergeMultipleStringMaps(resourceLabels, r.Config.Spec.Gateways.Egress.MeshGatewayConfiguration.Labels, r.Config.RevisionLabels())
}
