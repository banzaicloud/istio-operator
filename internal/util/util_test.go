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

package util_test

import (
	"embed"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
)

//go:embed testdata/test_istiocontrolplane.yaml
var icpFile []byte

//go:embed testdata/test_values.yaml.tmpl
var valuesFS embed.FS

//go:embed testdata/expected_values.yaml
var expectedValuesFile []byte

func TestTransformICPSpecToStriMapWithTemplate(t *testing.T) {
	t.Parallel()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpFile, &icp); err != nil {
		t.Fatal(err)
	}

	values, err := util.TransformStructToStriMapWithTemplate(icp, valuesFS, "testdata/test_values.yaml.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	var expectedValues map[string]interface{}
	if err := yaml.Unmarshal(expectedValuesFile, &expectedValues); err != nil {
		t.Fatal(err)
	}

	if diff := pretty.Compare(values, expectedValues); diff != "" {
		t.Errorf("diff: (-got +want)\n%s", diff)
	}
}
