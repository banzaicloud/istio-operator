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

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName          = "mixer"
	serviceAccountName     = "istio-mixer-service-account"
	clusterRoleName        = "istio-mixer-cluster-role"
	clusterRoleBindingName = "istio-mixer-cluster-role-binding"
	configMapName          = "istio-statsd-prom-bridge"
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
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Config) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	rs := []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
	}
	rsv := []resources.ResourceVariation{
		r.deployment,
		r.service,
		r.horizontalPodAutoscaler,
	}
	rs = append(rs, resources.ResolveVariations("policy", rsv)...)
	rs = append(rs, resources.ResolveVariations("telemetry", rsv)...)
	for _, res := range rs {
		o := res()
		err := k8sutil.Reconcile(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.istioProxyAttributeManifest, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.kubernetesAttributeManifest, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.stdioHandler, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.accessLogLogentry, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.tcpAccessLogLogentry, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.stdioRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.stdioTcpRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.prometheusHandler, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.requestCountMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.requestDurationMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.requestSizeMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.responseSizeMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.tcpByteReceivedMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.tcpByteSentMetric, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.promHttpRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.promTcpRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.kubernetesEnvHandler, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.attributesKubernetes, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.kubeAttrRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.tcpKubeAttrRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.policyDestinationRule, DesiredState: k8sutil.CREATED},
		{DynamicResource: r.telemetryDestinationRule, DesiredState: k8sutil.CREATED},
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
