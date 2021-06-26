/*
Copyright 2020 Banzai Cloud.

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

package meshexpansion

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istionetworkingv1beta1 "github.com/banzaicloud/istio-client-go/pkg/networking/v1beta1"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/resources/ingressgateway"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	ResourceName             = "istio-meshexpansion-gateway"
	componentName            = "meshexpansiongateway"
	multiMeshDomainLabelName = "istio.banzaicloud.io/multi-mesh-domain"
)

var resourceLabels = map[string]string{
	"app":   "istio-meshexpansion-gateway",
	"istio": "meshexpansiongateway",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
	remote  bool
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, remote bool, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
			Scheme: scheme,
		},
		dynamic: dc,
		remote:  remote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	ports := r.Config.Spec.Gateways.MeshExpansion.Ports
	if ports == nil {
		ports = make([]istiov1beta1.ServicePort, 0)
	}

	spec := istiov1beta1.MeshGatewaySpec{
		MeshGatewayConfiguration: r.Config.Spec.Gateways.MeshExpansion.MeshGatewayConfiguration,
		Ports:                    ports,
		Type:                     istiov1beta1.GatewayTypeIngress,
	}
	spec.IstioControlPlane = &istiov1beta1.NamespacedName{
		Name:      r.Config.Name,
		Namespace: r.Config.Namespace,
	}
	spec.Labels = r.labels()
	spec.AdditionalEnvVars = k8sutil.MergeEnvVars(spec.AdditionalEnvVars, []corev1.EnvVar{
		{
			Name:  "ISTIO_META_LOCAL_ENDPOINTS_ONLY",
			Value: "true",
		},
	})
	objectMeta := templates.ObjectMetaWithRevision(ResourceName, spec.Labels, r.Config)

	var desiredState k8sutil.DesiredState

	if util.PointerToBool(r.Config.Spec.Gateways.Enabled) && util.PointerToBool(r.Config.Spec.Gateways.MeshExpansion.Enabled) {
		desiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.Gateways.MeshExpansion.CreateOnly) {
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
	object.SetDefaults()

	overlays, err := k8sutil.GetObjectModifiersForOverlays(r.Scheme, r.Config.Spec.K8SOverlays)
	if err != nil {
		return emperror.WrapWith(err, "could not get k8s overlay object modifiers")
	}

	err = k8sutil.ReconcileWithObjectModifiers(log, r.Client, object, desiredState, k8sutil.CombineObjectModifiers([]k8sutil.ObjectModifierFunc{k8sutil.GetGVKObjectModifier(r.Scheme)}, overlays))
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile resource", "resource", object.GetObjectKind().GroupVersionKind())
	}

	if r.remote {
		log.Info("Reconciled")
		return nil
	}

	meshExpansionDesiredState := k8sutil.DesiredStateAbsent
	if (desiredState != k8sutil.DesiredStateAbsent || (util.PointerToBool(r.Config.Spec.Gateways.Enabled) && util.PointerToBool(r.Config.Spec.Gateways.Ingress.Enabled))) && util.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutil.DesiredStatePresent
	}

	multimeshDesiredState := k8sutil.DesiredStateAbsent
	if (desiredState != k8sutil.DesiredStateAbsent || (util.PointerToBool(r.Config.Spec.Gateways.Enabled) && util.PointerToBool(r.Config.Spec.Gateways.Ingress.Enabled))) && util.PointerToBool(r.Config.Spec.MultiMesh) {
		multimeshDesiredState = k8sutil.DesiredStatePresent
	}

	multimeshEnvoyFilterDesiredState := k8sutil.DesiredStateAbsent
	if multimeshDesiredState == k8sutil.DesiredStatePresent && util.PointerToBool(r.Config.Spec.MultiMeshExpansion.EnvoyFilterEnabled) {
		multimeshEnvoyFilterDesiredState = k8sutil.DesiredStatePresent
	}

	selector := resourceLabels
	if desiredState == k8sutil.DesiredStateAbsent {
		selector = ingressgateway.ResourceLabels
	}

	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: func() *k8sutil.DynamicObject { return r.meshExpansionGateway(selector) }, DesiredState: meshExpansionDesiredState},
		{DynamicResource: func() *k8sutil.DynamicObject { return r.clusterAwareGateway(selector) }, DesiredState: meshExpansionDesiredState},
		{DynamicResource: func() *k8sutil.DynamicObject { return r.multimeshIngressGateway(selector) }, DesiredState: multimeshDesiredState},
		{DynamicResource: func() *k8sutil.DynamicObject { return r.multimeshEnvoyFilter(selector) }, DesiredState: multimeshEnvoyFilterDesiredState},
	}
	meshDomains := make(map[string]bool, 0)
	for _, domain := range r.Config.Spec.GetMultiMeshExpansion().GetDomains() {
		domain := domain
		drs = append(drs, resources.DynamicResourceWithDesiredState{
			DynamicResource: func() *k8sutil.DynamicObject { return r.multimeshDestinationRule(domain) }, DesiredState: multimeshDesiredState,
		})
		meshDomains[domain] = true
	}
	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.ReconcileWithObjectModifiers(log, r.dynamic, dr.DesiredState, overlays)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	// remove obsolete destination rules
	destinationrules := &istionetworkingv1beta1.DestinationRuleList{}
	err = r.Client.List(context.Background(), destinationrules, client.HasLabels{multiMeshDomainLabelName})
	if err != nil {
		return err
	}
	for _, dr := range destinationrules.Items {
		if meshDomains[dr.GetLabels()[multiMeshDomainLabelName]] {
			continue
		}
		k8sutil.Reconcile(log, r.Client, &dr, k8sutil.DesiredStateAbsent)
	}

	log.Info("Reconciled")

	return nil
}

func (r *Reconciler) labels() map[string]string {
	return util.MergeMultipleStringMaps(resourceLabels, r.Config.Spec.Gateways.MeshExpansion.MeshGatewayConfiguration.Labels, r.Config.RevisionLabels())
}
