/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package assets

import (
	"embed"
	"io/fs"
)

var (
	//go:embed manifests/base
	//go:embed manifests/base/templates/_helpers.tpl
	baseChart embed.FS
	BaseChart = GetSubFS(baseChart, "manifests/base")

	//go:embed manifests/istio-discovery
	//go:embed manifests/istio-discovery/templates/_helpers.tpl
	discoveryChart embed.FS
	DiscoveryChart = GetSubFS(discoveryChart, "manifests/istio-discovery")

	//go:embed manifests/istio-cni
	//go:embed manifests/istio-cni/templates/_helpers.tpl
	cniChart embed.FS
	CNIChart = GetSubFS(cniChart, "manifests/istio-cni")

	//go:embed manifests/istio-meshexpansion
	//go:embed manifests/istio-meshexpansion/templates/_helpers.tpl
	meshExpansionChart embed.FS
	MeshExpansionChart = GetSubFS(meshExpansionChart, "manifests/istio-meshexpansion")

	//go:embed manifests/istio-meshgateway
	//go:embed manifests/istio-meshgateway/templates/_helpers.tpl
	istioMeshGateway embed.FS
	IstioMeshGateway = GetSubFS(istioMeshGateway, "manifests/istio-meshgateway")

	//go:embed manifests/istio-sidecar-injector
	//go:embed manifests/istio-sidecar-injector/templates/_helpers.tpl
	istioSidecarInjector embed.FS
	IstioSidecarInjector = GetSubFS(istioSidecarInjector, "manifests/istio-sidecar-injector")

	//go:embed manifests/resource-sync-rule
	//go:embed manifests/resource-sync-rule/templates/_helpers.tpl
	resourceSyncRule embed.FS
	ResourceSyncRule = GetSubFS(resourceSyncRule, "manifests/resource-sync-rule")

	//go:embed manifests/ztunnel
	//go:embed manifests/ztunnel/templates/_helpers.tpl
	ztunnelChart embed.FS
	ZtunnelChart = GetSubFS(ztunnelChart, "manifests/ztunnel")
)

func GetSubFS(fsys fs.FS, dir string) (subFS fs.FS) {
	subFS, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}

	return
}
