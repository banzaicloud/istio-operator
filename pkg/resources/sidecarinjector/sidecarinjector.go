package sidecarinjector

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountName     = "istio-sidecar-injector-service-account"
	clusterRoleName        = "istio-sidecar-injector-cluster-role"
	clusterRoleBindingName = "istio-sidecar-injector-cluster-role-binding"
	configMapName          = "istio-sidecar-injector-config"
	webhookName            = "istio-sidecar-injector"
	deploymentName         = "istio-sidecar-injector"
	serviceName            = "istio-sidecar-injector"
)

var sidecarInjectorLabels = map[string]string{
	"app": "istio-sidecar-injector",
}

var labelSelector = map[string]string{
	"istio": "sidecar-injector",
}

type Reconciler struct {
	resources.Reconciler
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
	for _, res := range []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
		r.deployment,
		r.service,
		r.webhook,
	} {
		o := res(r.Owner)
		err := k8sutil.Reconcile(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	return nil
}
