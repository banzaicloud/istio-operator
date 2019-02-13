package mixer

import (
	"fmt"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) destinationRule(t string, owner *istiov1beta1.Config) runtime.Object {
	return &networkingv1alpha3.DestinationRule{
		ObjectMeta: templates.ObjectMeta(destinationRuleName(t), nil, owner),
		Spec: v1alpha3.DestinationRule{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", t, owner.Namespace),
			TrafficPolicy: &v1alpha3.TrafficPolicy{
				ConnectionPool: &v1alpha3.ConnectionPoolSettings{
					Http: &v1alpha3.ConnectionPoolSettings_HTTPSettings{
						Http2MaxRequests:         10000,
						MaxRequestsPerConnection: 10000,
					},
				},
			},
		},
	}
}
