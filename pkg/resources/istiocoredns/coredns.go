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
	"strings"

	"github.com/caddyserver/caddy/caddyfile"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

// Add/Remove global:53 to/from coredns configmap
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

	corefile := cm.Data["Corefile"]
	desiredCorefile := corefile
	clusterIP := ""

	if desiredState == k8sutil.DesiredStatePresent {
		var svc apiv1.Service
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      serviceName,
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
	config := caddyfile.EncodedServerBlock{
		Keys: []string{"global:53"},
		Body: [][]interface{}{
			{"errors"},
			{"cache", "30"},
			{proxyOrForward, ".", clusterIP},
		},
	}

	if desiredState == k8sutil.DesiredStatePresent {
		desiredCorefile, err = r.updateCorefile([]byte(corefile), config, false)
		if err != nil {
			return emperror.Wrap(err, "could not add config for .global to Corefile")
		}
	} else if desiredState == k8sutil.DesiredStateAbsent {
		desiredCorefile, err = r.updateCorefile([]byte(corefile), config, true)
		if err != nil {
			return emperror.Wrap(err, "could not remove config for .global from Corefile")
		}
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["Corefile"] = desiredCorefile

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update coredns configmap")
	}

	return nil
}

func (r *Reconciler) updateCorefile(corefile []byte, config caddyfile.EncodedServerBlock, remove bool) (string, error) {
	corefileJSONData, err := caddyfile.ToJSON(corefile)
	if err != nil {
		return "", emperror.Wrap(err, "could not convert Corefile to JSON data")
	}

	var corefileJSON caddyfile.EncodedCaddyfile
	err = json.Unmarshal(corefileJSONData, &corefileJSON)
	if err != nil {
		return "", emperror.Wrap(err, "could not unmarshal JSON to EncodedCaddyfile")
	}

	if len(config.Keys) < 1 {
		return "", errors.New("invalid .global config")
	}

	pos := -1
	for i, block := range corefileJSON {
		if len(block.Keys) < 1 {
			continue
		}
		if block.Keys[0] == config.Keys[0] {
			pos = i
			break
		}
	}

	if remove {
		if pos > 0 {
			corefileJSON = append(corefileJSON[:pos], corefileJSON[pos+1:]...)
		}
	} else {
		if pos > 0 {
			corefileJSON[pos] = config
		} else {
			corefileJSON = append(corefileJSON, config)
		}
	}

	corefileData, err := json.Marshal(&corefileJSON)
	if err != nil {
		return "", emperror.Wrap(err, "could not marshal EncodedCaddyfile to JSON")
	}

	corefile, err = caddyfile.FromJSON(corefileData)
	if err != nil {
		return "", emperror.Wrap(err, "could not convert JSON to Caddyfile")
	}

	// convert tabs to spaces for properly display content in ConfigMap
	return strings.Replace(string(corefile), "\t", "  ", -1), nil
}
