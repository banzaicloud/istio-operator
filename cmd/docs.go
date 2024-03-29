/*
Copyright 2022 Cisco Systems, Inc. and/or its affiliates.

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
	"path/filepath"

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/banzaicloud/operator-tools/pkg/docgen"
)

var logger = zap.New(zap.UseDevMode(true))

func main() {
	crds()
}

func crds() {
	lister := docgen.NewSourceLister(
		map[string]docgen.SourceDir{
			"v1alpha1": {Path: "api/v1alpha1", DestPath: "docs/crds/v1alpha1"},
		},
		logger.WithName("crdlister"))

	lister.IgnoredSources = []string{
		".*.deepcopy",
		".*.json",
		".*_test",
		".*_info",
	}

	lister.DefaultValueFromTagExtractor = func(tag string) string {
		return docgen.GetPrefixedValue(tag, `plugin:\"default:(.*)\"`)
	}

	lister.Index = docgen.NewDoc(docgen.DocItem{
		Name:     "_index",
		DestPath: "docs/crds/v1alpha1",
	}, logger.WithName("crds"))

	lister.Header = heredoc.Doc(`
		---
		title: Available CRDs
		generated_file: true
		---
		
		The following CRDs are available.  For details, click the name of the CRD.

		| Name | Description | Version |
		|---|---|---|`,
	)

	lister.Footer = heredoc.Doc(`
	`)

	lister.DocGeneratedHook = func(document *docgen.Doc) error {
		relPath, err := filepath.Rel(lister.Index.Item.DestPath, document.Item.DestPath)
		if err != nil {
			return errors.WrapIff(err, "failed to determine relpath for %s", document.Item.DestPath)
		}
		lister.Index.Append(fmt.Sprintf("| **[%s](%s/)** | %s | %s |",
			document.DisplayName,
			filepath.Join(relPath, document.Item.Name),
			document.Desc,
			document.Item.Category))

		return nil
	}

	if err := lister.Generate(); err != nil {
		panic(err)
	}
}
