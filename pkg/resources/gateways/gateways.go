package gateways

import (
	"fmt"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	resources.Reconciler
}

type GwResource func(gw string, owner *istiov1beta1.Config) runtime.Object

func rr(gw string, gwResources []GwResource) []resources.Resource {
	resources := make([]resources.Resource, 0)
	for i := range gwResources {
		i := i
		resources = append(resources, func(owner *istiov1beta1.Config) runtime.Object {
			return gwResources[i](gw, owner)
		})
	}
	return resources
}

func New(client client.Client, istio *istiov1beta1.Config) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Owner:  istio,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	var gwResources = []GwResource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.deployment,
		r.service,
		r.horizontalPodAutoscaler,
	}
	for _, res := range append(rr("ingressgateway", gwResources), rr("egressgateway", gwResources)...) {
		o := res(r.Owner)
		err := k8sutil.Reconcile(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	return nil
}

func serviceAccountName(gw string) string {
	return fmt.Sprintf("istio-%s-service-account", gw)
}

func clusterRoleName(gw string) string {
	return fmt.Sprintf("istio-%s-cluster-role", gw)
}

func clusterRoleBindingName(gw string) string {
	return fmt.Sprintf("istio-%s-cluster-role-binding", gw)
}

func gatewayName(gw string) string {
	return fmt.Sprintf("istio-%s", gw)
}

func hpaName(gw string) string {
	return fmt.Sprintf("istio-%s-autoscaler", gw)
}

func gwLabels(gw string) map[string]string {
	return map[string]string{
		"app": fmt.Sprintf("istio-%s", gw),
	}
}

func labelSelector(gw string) map[string]string {
	return util.MergeLabels(gwLabels(gw), map[string]string{
		"istio": gw,
	})
}
