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

// +build ignore

//go:generate go run generate.go

package main

import (
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	"github.com/shurcooL/vfsgen"
)

var CRDs http.FileSystem = http.Dir(path.Join(getRepoRoot(), "config/base/crds"))

func main() {
	err := vfsgen.Generate(CRDs, vfsgen.Options{
		Filename:     path.Join(getRepoRoot(), "pkg/crds/generated/crds.gogen.go"),
		PackageName:  "generated",
		VariableName: "CRDs",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// getRepoRoot returns the full path to the root of the repo
func getRepoRoot() string {
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filename)

	return filepath.Dir(path.Join(dir, ".."))
}
