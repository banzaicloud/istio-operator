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

package citadel

import (
	"fmt"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, util.MergeLabels(citadelLabels, labelSelector), r.Config),
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
								fmt.Sprintf("--citadel-storage-namespace=%s", r.Config.Namespace),
								fmt.Sprintf("--custom-dns-names=istio-pilot-service-account.%[1]s:istio-pilot.%[1]s,istio-ingressgateway-service-account.%[1]s:istio-ingressgateway.%[1]s", r.Config.Namespace),
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
