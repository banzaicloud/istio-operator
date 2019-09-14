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

package mixer

import (
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
)

const (
	policyComponentName    = "policy"
	telemetryComponentName = "telemetry"
	serviceAccountName     = "istio-mixer-service-account"
	clusterRoleName        = "istio-mixer-cluster-role"
	clusterRoleBindingName = "istio-mixer-cluster-role-binding"
)

var mixerLabels = map[string]string{
	"app": "mixer",
}

var labelSelector = map[string]string{
	"istio": "mixer",
}

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface

	component         string
	k8sResourceConfig istiov1beta1.K8sResourceConfiguration
	telemetryConfig   istiov1beta1.TelemetryConfiguration
	policyConfig      istiov1beta1.PolicyConfiguration
}

func NewPolicyReconciler(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,

		component:         policyComponentName,
		k8sResourceConfig: config.Spec.Policy.K8sResourceConfiguration,
		policyConfig:      config.Spec.Policy,
	}
}

func NewTelemetryReconciler(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,

		component:         telemetryComponentName,
		k8sResourceConfig: config.Spec.Telemetry.K8sResourceConfiguration,
		telemetryConfig:   config.Spec.Telemetry,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", r.component)

	log.Info("Reconciling")

	var mixerDesiredState k8sutil.DesiredState
	var pdbDesiredState k8sutil.DesiredState
	if (r.component == policyComponentName && util.PointerToBool(r.policyConfig.Enabled)) ||
		(r.component == telemetryComponentName && util.PointerToBool(r.telemetryConfig.Enabled)) {
		mixerDesiredState = k8sutil.DesiredStatePresent
		if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
			pdbDesiredState = k8sutil.DesiredStatePresent
		} else {
			pdbDesiredState = k8sutil.DesiredStateAbsent
		}
	} else {
		mixerDesiredState = k8sutil.DesiredStateAbsent
	}

	rs := []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount},
		{Resource: r.clusterRole},
		{Resource: r.clusterRoleBinding},
	}
	rsv := []resources.ResourceVariationWithDesiredState{
		{ResourceVariation: r.deployment},
		{ResourceVariation: r.service},
		{ResourceVariation: r.horizontalPodAutoscaler},
		{ResourceVariation: r.podDisruptionBudget, DesiredState: pdbDesiredState},
	}

	rs = append(rs, resources.ResolveVariations(r.component, rsv, mixerDesiredState)...)
	for _, res := range rs {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.istioProxyAttributeManifest},
		{DynamicResource: r.kubernetesAttributeManifest},
		{DynamicResource: r.kubernetesEnvHandler},
		{DynamicResource: r.attributesKubernetes},
		{DynamicResource: r.kubeAttrRule},
		{DynamicResource: r.tcpKubeAttrRule},
	}

	if r.component == telemetryComponentName {
		if util.PointerToBool(r.Config.Spec.Mixer.StdioAdapterEnabled) {
			drs = append(drs, []resources.DynamicResourceWithDesiredState{
				// stdio adapter
				{DynamicResource: r.stdioHandler},
				{DynamicResource: r.accessLogLogentry},
				{DynamicResource: r.tcpAccessLogLogentry},
				{DynamicResource: r.stdioRule},
				{DynamicResource: r.stdioTcpRule},
			}...)
		}
		drs = append(drs, []resources.DynamicResourceWithDesiredState{
			// prometheus adapter
			{DynamicResource: r.prometheusHandler},
			{DynamicResource: r.requestCountMetric},
			{DynamicResource: r.requestDurationMetric},
			{DynamicResource: r.requestSizeMetric},
			{DynamicResource: r.responseSizeMetric},
			{DynamicResource: r.tcpByteReceivedMetric},
			{DynamicResource: r.tcpByteSentMetric},
			{DynamicResource: r.tcpConnectionsOpenedMetric},
			{DynamicResource: r.tcpConnectionsClosedMetric},
			{DynamicResource: r.promHttpRule},
			{DynamicResource: r.promTcpRule},
			{DynamicResource: r.promTcpConnectionOpenRule},
			{DynamicResource: r.promTcpConnectionClosedRule},

			// telemetry
			{DynamicResource: r.telemetryDestinationRule},
		}...)
	}

	if r.component == policyComponentName {
		drs = append(drs, resources.DynamicResourceWithDesiredState{
			DynamicResource: r.policyDestinationRule,
		})
	}

	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, mixerDesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}

func deploymentName(t string) string {
	return fmt.Sprintf("istio-%s", t)
}

func serviceName(t string) string {
	return fmt.Sprintf("istio-%s", t)
}

func hpaName(t string) string {
	return fmt.Sprintf("istio-%s-autoscaler", t)
}

func appLabel(t string) map[string]string {
	return map[string]string{
		"app": t,
	}
}

func mixerTypeLabel(t string) map[string]string {
	return map[string]string{
		"istio-mixer-type": t,
	}
}
