/*
Copyright 2020 Banzai Cloud.

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
	"encoding/json"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/goph/emperror"
	apiv1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	multiMeshDomainsAnnotationName = "istio.banzaicloud.io/multi-mesh-domains"
)

// Removed support for the proxy plugin: https://coredns.io/2019/03/03/coredns-1.4.0-release/
func (r *Reconciler) isProxyPluginDeprecated() bool {
	imageParts := strings.Split(util.PointerToString(r.Config.Spec.IstioCoreDNS.Image), ":")
	tag := imageParts[1]

	v140 := semver.New("1.4.0")
	vCoreDNSTag := semver.New(tag)

	if v140.LessThan(*vCoreDNSTag) {
		return true
	}

	return false
}

func setDomainsToAnnotation(domains []string, cm *apiv1.ConfigMap) error {
	domainsJSON, err := json.Marshal(domains)
	if err != nil {
		return emperror.Wrap(err, "could not marshal domains")
	}
	if cm.Annotations == nil {
		cm.Annotations = make(map[string]string)
	}
	if len(domains) > 0 {
		cm.Annotations[multiMeshDomainsAnnotationName] = string(domainsJSON)
	} else if _, ok := cm.Annotations[multiMeshDomainsAnnotationName]; ok {
		delete(cm.Annotations, multiMeshDomainsAnnotationName)
	}

	return nil
}

func getDomainsFromAnnotation(cm *apiv1.ConfigMap) (map[string]string, error) {
	domains := make([]string, 0)
	domainsMap := make(map[string]string)

	domainsJSON := cm.Annotations[multiMeshDomainsAnnotationName]
	if domainsJSON == "" {
		return domainsMap, nil
	}
	err := json.Unmarshal([]byte(domainsJSON), &domains)
	if err != nil {
		return nil, emperror.Wrap(err, "could not unmarshal domains")
	}

	for _, domain := range domains {
		domainsMap[domain] = domain
	}

	return domainsMap, err
}
