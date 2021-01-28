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
	"context"
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/caddyfile"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	multiMeshBlockChaosString = "Istio-CoreDNS"
)

// Add/Remove multi mesh domains to/from coredns configmap
func (r *Reconciler) reconcileCoreDNSConfigMap(log logr.Logger, desiredState k8sutil.DesiredState) error {
	var cm apiv1.ConfigMap

	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      "coredns",
		Namespace: "kube-system",
	}, &cm)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return emperror.Wrap(err, "could not get coredns configmap")
	}

	domains := r.Config.Spec.GetMultiMeshExpansion().GetDomains()

	corefile := cm.Data["Corefile"]
	clusterIP := ""

	if desiredState == k8sutil.DesiredStatePresent {
		var svc apiv1.Service
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      r.Config.WithRevision(serviceName),
			Namespace: r.Config.Namespace,
		}, &svc)
		if err != nil {
			return emperror.Wrap(err, "could not get Istio coreDNS service")
		}
		clusterIP = svc.Spec.ClusterIP
	}

	proxyOrForward := "proxy"
	if r.isProxyPluginDeprecated() {
		proxyOrForward = "forward"
	}

	config := caddyfile.EncodedServerBlock{}
	if desiredState == k8sutil.DesiredStatePresent {
		config.Keys = func() (d []string) {
			for _, domain := range domains {
				d = append(d, fmt.Sprintf("%s:53", domain))
			}

			return d
		}()
		config.Body = [][]interface{}{
			{"chaos", multiMeshBlockChaosString},
			{"errors"},
			{"cache", "30"},
			{proxyOrForward, ".", clusterIP},
		}
	}

	desiredCorefile, err := r.updateCorefile([]byte(corefile), config, desiredState == k8sutil.DesiredStateAbsent || len(domains) == 0)
	if err != nil {
		return emperror.Wrap(err, "could not add config to Corefile")
	}

	if desiredState == k8sutil.DesiredStateAbsent {
		domains = []string{}
	}

	err = setDomainsToAnnotation(domains, &cm)
	if err != nil {
		return err
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["Corefile"] = string(desiredCorefile)

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update coredns configmap")
	}

	return nil
}

func (r *Reconciler) updateCorefile(corefileJSON []byte, config caddyfile.EncodedServerBlock, remove bool) ([]byte, error) {
	corefileJSONData, err := caddyfile.ToJSON([]byte(corefileJSON))
	if err != nil {
		return nil, emperror.Wrap(err, "could not convert Corefile to JSON data")
	}

	var corefile caddyfile.EncodedCaddyfile
	err = json.Unmarshal(corefileJSONData, &corefile)
	if err != nil {
		return nil, emperror.Wrap(err, "could not unmarshal JSON to EncodedCaddyfile")
	}

	pos := -1
	for i, block := range corefile {
		for _, b := range block.Body {
			if len(b) == 2 && b[0] == "chaos" && b[1] == multiMeshBlockChaosString {
				pos = i
				break
			}
		}
	}

	if pos > 0 {
		if remove {
			corefile = append(corefile[:pos], corefile[pos+1:]...)
		} else {
			corefile[pos] = config
		}
	} else if !remove {
		corefile = append(corefile, config)
	}

	return util.GenerateCaddyFile(corefile)
}
