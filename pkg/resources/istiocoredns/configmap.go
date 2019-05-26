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

package istiocoredns

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
)

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, labels, r.Config),
		Data: map[string]string{
			"Corefile": `.:53 {
    errors
    health
    proxy global 127.0.0.1:8053 {
        protocol grpc insecure
    }
    prometheus :9153
    proxy . /etc/resolv.conf
    cache 30
    reload
}
`,
		},
	}
}
