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
package istiomeshgateway_test

import (
	_ "embed"
	"fmt"
	"os"
	"testing"

	"emperror.dev/errors"
	"emperror.dev/errors/utils/keyval"
	testlogr "github.com/go-logr/logr/testing"
	"github.com/homeport/dyff/pkg/dyff"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components/istiomeshgateway"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

//go:embed testdata/imgw-test-cr.yaml
var imgwTestCR []byte

//go:embed testdata/icp-test-cr.yaml
var icpTestCR []byte

//go:embed testdata/imgw-expected-values.yaml
var imgwExpectedValues []byte

//go:embed testdata/imgw-expected-resource-dump.yaml
var imgwExpectedResourceDump []byte

func TestIMGWResourceDump(t *testing.T) {
	t.Parallel()

	var imgw *v1alpha1.IstioMeshGateway
	if err := yaml.Unmarshal(imgwTestCR, &imgw); err != nil {
		t.Fatal(err)
	}

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	reconciler := istiomeshgateway.NewChartReconciler(
		templatereconciler.NewHelmReconciler(nil, nil, testlogr.TestLogger{
			T: t,
		}, fake.NewSimpleClientset().Discovery(), []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
		v1alpha1.IstioMeshGatewayProperties{
			Revision:                "cp-v110x.istio-system",
			EnablePrometheusMerge:   utils.BoolPointer(true),
			InjectionTemplate:       "gateway",
			InjectionChecksum:       "08fdba0c89f9bbd6624201d98758746d1bddc78e9004b00259f33b20b7f9efba",
			MeshConfigChecksum:      "319ffd3f807ef4516499c6ad68279a1cd07778f5847e65f9aef908eceb1693e3",
			IstioControlPlane:       icp,
			GenerateExternalService: true,
		},
		testlogr.TestLogger{
			T: t,
		},
	)

	dd, err := reconciler.GetManifest(imgw)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(imgwExpectedResourceDump, dd)
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

func TestIMGWTemplateTransform(t *testing.T) {
	t.Parallel()

	var imgw *v1alpha1.IstioMeshGateway
	if err := yaml.Unmarshal(imgwTestCR, &imgw); err != nil {
		t.Fatal(err)
	}

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	obj := &v1alpha1.IstioMeshGatewayWithProperties{
		IstioMeshGateway: imgw,
		Properties: v1alpha1.IstioMeshGatewayProperties{
			Revision:                "cp-revision-1",
			EnablePrometheusMerge:   utils.BoolPointer(false),
			InjectionTemplate:       "gateway",
			InjectionChecksum:       "08fdba0c89f9bbd6624201d98758746d1bddc78e9004b00259f33b20b7f9efba",
			MeshConfigChecksum:      "319ffd3f807ef4516499c6ad68279a1cd07778f5847e65f9aef908eceb1693e3",
			IstioControlPlane:       icp,
			GenerateExternalService: true,
		},
	}
	obj.SetDefaults()

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.IstioMeshGateway, "values.yaml.tpl")
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

	report, err := util.CompareYAMLs(imgwExpectedValues, valuesYaml)
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
