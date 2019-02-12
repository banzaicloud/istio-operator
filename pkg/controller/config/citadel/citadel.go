package citadel

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountName     = "istio-citadel-service-account"
	clusterRoleName        = "istio-citadel-cluster-role"
	clusterRoleBindingName = "istio-citadel-cluster-role-binding"
	deploymentName         = "istio-citadel"
	serviceName            = "istio-citadel"
)

var citadelLabels = map[string]string{
	"app": "security",
}

var labelSelector = map[string]string{
	"istio": "citadel",
}

type Reconciler struct {
	config.Reconciler
}

func New(client client.Client, istio *istiov1beta1.Config) *Reconciler {
	return &Reconciler{
		Reconciler: config.Reconciler{
			Client: client,
			Owner:  istio,
			Resources: []config.Resource{
				serviceAccount,
				clusterRole,
				clusterRoleBinding,
				deployment,
				service,
			},
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	for _, res := range r.Resources {
		o := res(r.Owner)
		err := k8sutil.ReconcileResource(log, r.Client, o)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}
	return nil
}
