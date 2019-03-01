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

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
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

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio) *Reconciler {
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
		{DynamicResource: r.istioProxyAttributeManifest},
		{DynamicResource: r.kubernetesAttributeManifest},
		{DynamicResource: r.stdioHandler},
		{DynamicResource: r.accessLogLogentry},
		{DynamicResource: r.tcpAccessLogLogentry},
		{DynamicResource: r.stdioRule},
		{DynamicResource: r.stdioTcpRule},
		{DynamicResource: r.prometheusHandler},
		{DynamicResource: r.requestCountMetric},
		{DynamicResource: r.requestDurationMetric},
		{DynamicResource: r.requestSizeMetric},
		{DynamicResource: r.responseSizeMetric},
		{DynamicResource: r.tcpByteReceivedMetric},
		{DynamicResource: r.tcpByteSentMetric},
		{DynamicResource: r.promHttpRule},
		{DynamicResource: r.promTcpRule},
		{DynamicResource: r.kubernetesEnvHandler},
		{DynamicResource: r.attributesKubernetes},
		{DynamicResource: r.kubeAttrRule},
		{DynamicResource: r.tcpKubeAttrRule},
		{DynamicResource: r.policyDestinationRule},
		{DynamicResource: r.telemetryDestinationRule},
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
