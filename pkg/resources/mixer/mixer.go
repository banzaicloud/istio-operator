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
		r.destinationRule,
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
	drs := []resources.DynamicResource{
		r.istioProxyAttributeManifest,
		r.kubernetesAttributeManifest,
		r.stdioHandler,
		r.accessLogLogentry,
		r.tcpAccessLogLogentry,
		r.stdioRule,
		r.stdioTcpRule,
		r.prometheusHandler,
		r.requestCountMetric,
		r.requestDurationMetric,
		r.requestSizeMetric,
		r.responseSizeMetric,
		r.tcpByteReceivedMetric,
		r.tcpByteSentMetric,
		r.promHttpRule,
		r.promTcpRule,
		r.kubernetesEnvHandler,
		r.attributesKubernetes,
		r.kubeAttrRule,
		r.tcpKubeAttrRule,
	}
	for _, dr := range drs {
		o := dr()
		err := o.Reconcile(log, r.dynamic)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}
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

func destinationRuleName(t string) string {
	return fmt.Sprintf("istio-%s", t)
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
