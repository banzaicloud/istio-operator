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

	"github.com/banzaicloud/istio-operator/v2/static/util"
)

var charts = []string{}

//go:generate go run main.go
func main() {
	chartsPath := filepath.Join(getRepoRoot(), "deploy/charts")

	for _, dir := range charts {
		actualChartPath := filepath.Join(chartsPath, dir)

		chartDir := util.ZeroModTimeFileSystem{
			Source: http.Dir(actualChartPath),
		}
		staticPath := filepath.Join(getRepoRoot(), "static/gen/charts", dir)
		if err := os.MkdirAll(staticPath, 0755); err != nil {
			panic(fmt.Errorf("failed to create directory for charts: %w", err))
		}

		dirParts := strings.Split(dir, "/")
		err := vfsgen.Generate(chartDir, vfsgen.Options{
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

	return filepath.Dir(path.Join(dir, "."))
}
