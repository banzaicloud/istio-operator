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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/pkg/resources/templates"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/mholt/caddy/caddyfile"
)

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMetaWithRevision(configMapName, labels, r.Config),
		Data:       r.data(),
	}
}

func (r *Reconciler) data() map[string]string {
	var data map[string]string
	proxyOrGrpc := "proxy"
	if r.isProxyPluginDeprecated() {
		proxyOrGrpc = "grpc"
	}

	// base config
	config := caddyfile.EncodedCaddyfile([]caddyfile.EncodedServerBlock{
		{
			Keys: []string{".:53"},
			Body: [][]interface{}{
				{"errors"},
				{"health"},
				{proxyOrGrpc, ".", "/etc/resolv.conf"},
				{"prometheus", ":9153"},
				{"cache", "30"},
				{"reload"},
			},
		},
	})

	// get config for specific domains
	for _, domain := range r.Config.Spec.GetMultiClusterDomains() {
		config = append(config, r.getCoreDNSConfigBlockForDomain(domain))
	}

	cf, err := util.GenerateCaddyFile(config)
	if err != nil {
		return data
	}

	return map[string]string{
		"Corefile": string(cf),
	}
}

func (r *Reconciler) getCoreDNSConfigBlockForDomain(domain string) caddyfile.EncodedServerBlock {
	proxyOrGrpc := "proxy"
	if r.isProxyPluginDeprecated() {
		proxyOrGrpc = "grpc"
	}

	return caddyfile.EncodedServerBlock{
		Keys: []string{fmt.Sprintf("%s:53", domain)},
		Body: [][]interface{}{
			{"errors"},
			{proxyOrGrpc, domain, "127.0.0.1:8053"},
			{"cache", "30"},
		},
	}
}
