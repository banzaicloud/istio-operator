package mixer

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cmLabels = map[string]string{
	"app": "istio-statsd-prom-bridge",
}

func (r *Reconciler) configMap(owner *istiov1beta1.Config) runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, util.MergeLabels(labelSelector, cmLabels), owner),
		Data: map[string]string{
			"mapping.conf": "",
		},
	}
}
