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
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/banzaicloud/istio-operator/pkg/util"
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
