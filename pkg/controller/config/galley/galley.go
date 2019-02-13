package galley

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountName     = "istio-galley-service-account"
	clusterRoleName        = "istio-galley-cluster-role"
	clusterRoleBindingName = "istio-galley-admin-role-binding"
	configMapName          = "istio-galley-configuration"
	webhookName            = "istio-galley"
	deploymentName         = "istio-galley"
	serviceName            = "istio-galley"
)

var galleyLabels = map[string]string{
	"app": "istio-galley",
}

var labelSelector = map[string]string{
	"istio": "galley",
}

type Reconciler struct {
	config.Reconciler
}

func New(client client.Client, istio *istiov1beta1.Config) *Reconciler {
	return &Reconciler{
		Reconciler: config.Reconciler{
			Client: client,
			Owner:  istio,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	for _, res := range []config.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
		r.deployment,
		r.service,
	} {
		o := res(r.Owner)
		err := k8sutil.ReconcileResource(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	// TODO: wait for deployment to be available?
	return nil
}
