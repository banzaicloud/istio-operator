package assets

import (
	"embed"
	"io/fs"
)

var (
	//go:embed crds
	crds embed.FS
	CRDs = crds

	//go:embed manifests/base
	//go:embed manifests/base/templates/_helpers.tpl
	baseChart embed.FS
	BaseChart = GetSubFS(baseChart, "manifests/base")

	//go:embed manifests/istio-control/istio-discovery
	//go:embed manifests/istio-control/istio-discovery/templates/_helpers.tpl
	discoveryChart embed.FS
	DiscoveryChart = GetSubFS(discoveryChart, "manifests/istio-control/istio-discovery")
)

func GetSubFS(fsys fs.FS, dir string) (subFS fs.FS) {
	subFS, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}

	return
}
