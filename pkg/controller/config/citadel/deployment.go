package citadel

import (
	"fmt"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/controller/config/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func deployment(owner *istiov1beta1.Config) runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(citadelLabels, labelSelector), owner),
		Spec: appsv1.DeploymentSpec{
			Replicas: util.IntPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labelSelector,
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []apiv1.Container{
						{
							Name:            "citadel",
							Image:           "docker.io/istio/citadel:1.0.5",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"--append-dns-names=true",
								"--grpc-port=8060",
								"--grpc-hostname=citadel",
								fmt.Sprintf("--citadel-storage-namespace=%s", owner.Namespace),
								fmt.Sprintf("--custom-dns-names=istio-pilot-service-account.%[1]s:istio-pilot.%[1]s,istio-ingressgateway-service-account.%[1]s:istio-ingressgateway.%[1]s", istio.Namespace),
								"--self-signed-ca=true",
							},
							Resources: templates.DefaultResources(),
						},
					},
					Affinity: &apiv1.Affinity{},
				},
			},
		},
	}
}
