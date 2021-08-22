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
package meshgateway_test

import (
	_ "embed"
	"fmt"
	"os"
	"testing"

	"emperror.dev/errors"
	"emperror.dev/errors/utils/keyval"
	"github.com/homeport/dyff/pkg/dyff"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components/meshgateway"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

//go:embed testdata/mgw-test-cr.yaml
var mgwTestCR []byte

//go:embed testdata/mgw-expected-values.yaml
var mgwExpectedValues []byte

//go:embed testdata/mgw-expected-resource-dump.yaml
var mgwExpectedResourceDump []byte

func TestMGWResourceDump(t *testing.T) {
	t.Parallel()

	var mgw *v1alpha1.MeshGateway
	if err := yaml.Unmarshal(mgwTestCR, &mgw); err != nil {
		t.Fatal(err)
	}

	reconciler := meshgateway.NewChartReconciler(
		templatereconciler.NewHelmReconciler(nil, nil, nil, fake.NewSimpleClientset().Discovery(), []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
		v1alpha1.MeshGatewayProperties{
			Revision:              "cp-v110x.istio-system",
			EnablePrometheusMerge: true,
			InjectionTemplate:     "gateway",
			InjectionChecksum:     "08fdba0c89f9bbd6624201d98758746d1bddc78e9004b00259f33b20b7f9efba",
			MeshConfigChecksum:    "319ffd3f807ef4516499c6ad68279a1cd07778f5847e65f9aef908eceb1693e3",
		},
	)

	dd, err := reconciler.GetManifest(mgw)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(mgwExpectedResourceDump, dd)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Diffs) > 0 {
		if err := (&dyff.HumanReport{
			Report:       report,
			OmitHeader:   false,
			NoTableStyle: true,
		}).WriteReport(os.Stdout); err != nil {
			t.Fatal(err)
		}

		t.Fatal(errors.NewPlain("generated resource dump not equals with expected"))
	}
}

func TestMGWTemplateTransform(t *testing.T) {
	t.Parallel()

	var mgw *v1alpha1.MeshGateway
	if err := yaml.Unmarshal(mgwTestCR, &mgw); err != nil {
		t.Fatal(err)
	}

	obj := &v1alpha1.MeshGatewayWithProperties{
		MeshGateway: mgw,
		Properties: v1alpha1.MeshGatewayProperties{
			Revision:              "cp-revision-1",
			EnablePrometheusMerge: false,
			InjectionTemplate:     "gateway",
			InjectionChecksum:     "08fdba0c89f9bbd6624201d98758746d1bddc78e9004b00259f33b20b7f9efba",
			MeshConfigChecksum:    "319ffd3f807ef4516499c6ad68279a1cd07778f5847e65f9aef908eceb1693e3",
		},
	}
	obj.SetDefaults()

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.MeshGateway, "values.yaml.tpl")
	if err != nil {
		kv := keyval.ToMap(errors.GetDetails(err))
		if t, ok := kv["template"]; ok {
			fmt.Printf("%s\n", t.(string))
		}
		t.Fatal(err)
	}

	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(mgwExpectedValues, valuesYaml)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Diffs) > 0 {
		if err := (&dyff.HumanReport{
			Report:       report,
			OmitHeader:   false,
			NoTableStyle: true,
		}).WriteReport(os.Stdout); err != nil {
			t.Fatal(err)
		}

		t.Fatal(errors.NewPlain("generated template values not equals with expected"))
	}
}
