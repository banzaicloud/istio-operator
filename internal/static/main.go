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

package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shurcooL/vfsgen"
)

var charts = []string{
	"base",
	"istio-control/istio-discovery",
}

//go:generate go run main.go
func main() {
	crds := http.Dir(filepath.Join(getRepoRoot(), "config/crd/bases"))

	err := vfsgen.Generate(crds, vfsgen.Options{
		Filename:     filepath.Join(getRepoRoot(), "internal/static/gen/crds/generated.go"),
		PackageName:  "crds",
		VariableName: "Root",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to generate crds vfs: %+v", err))
	}

	license := http.Dir(filepath.Join(getRepoRoot(), "license"))

	err = vfsgen.Generate(license, vfsgen.Options{
		Filename:     filepath.Join(getRepoRoot(), "internal/static/gen/license/generated.go"),
		PackageName:  "license",
		VariableName: "Root",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to generate license vfs: %+v", err))
	}

	chartsPath := filepath.Join(getRepoRoot(), "deploy/charts")

	for _, dir := range charts {
		path := filepath.Join(chartsPath, dir)
		// tmpPath := filepath.Join(filepath.Join(getRepoRoot(), "charts/tmp"), dir)

		chartDir := http.Dir(path)
		staticPath := filepath.Join(getRepoRoot(), "internal/static/gen/charts", dir)
		if err := os.MkdirAll(staticPath, 0755); err != nil {
			panic(fmt.Errorf("failed to create directory for charts: %w", err))
		}

		dirParts := strings.Split(dir, "/")
		err = vfsgen.Generate(chartDir, vfsgen.Options{
			Filename:     filepath.Join(staticPath, "generated.go"),
			PackageName:  strings.ReplaceAll(dirParts[len(dirParts)-1], "-", "_"),
			VariableName: "Chart",
		})
		if err != nil {
			panic(fmt.Sprintf("failed to generate chart vfs: %+v", err))
		}
	}
}

// getRepoRoot returns the full path to the root of the repo
//nolint:dogsled
func getRepoRoot() string {
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filename)

	return filepath.Dir(path.Join(dir, ".."))
}
