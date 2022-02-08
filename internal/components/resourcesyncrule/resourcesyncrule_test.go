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
package resourcesyncrule_test

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
	"github.com/banzaicloud/istio-operator/v2/internal/components/resourcesyncrule"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

//go:embed testdata/icp-active-test-cr.yaml
var icpActiveTestCR []byte

//go:embed testdata/icp-passive-test-cr.yaml
var icpPassiveTestCR []byte

//go:embed testdata/rsr-expected-active-values.yaml
var rsrExpectedActiveValues []byte

//go:embed testdata/rsr-expected-passive-values.yaml
var rsrExpectedPassiveValues []byte

//go:embed testdata/rsr-expected-active-resource-dump.yaml
var rsrExpectedActiveResourceDump []byte

//go:embed testdata/rsr-expected-passive-resource-dump.yaml
var rsrExpectedPassiveResourceDump []byte

func TestResourceSyncRuleResourceDump(t *testing.T) {
	t.Parallel()

	testResourceSyncRuleResourceDump(t, icpActiveTestCR, rsrExpectedActiveResourceDump)
	testResourceSyncRuleResourceDump(t, icpPassiveTestCR, rsrExpectedPassiveResourceDump)
}

func testResourceSyncRuleResourceDump(t *testing.T, icpTestCR, rsrExpectedResourceDump []byte) {
	t.Helper()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	reconciler := resourcesyncrule.NewChartReconciler(
		templatereconciler.NewHelmReconciler(nil, nil, testlogr.TestLogger{
			T: t,
		}, fake.NewSimpleClientset().Discovery(), []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
		true,
	)

	dd, err := reconciler.GetManifest(icp)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(rsrExpectedResourceDump, dd)
	if err != nil {
		t.Log(string(dd))
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

		if err := util.DyffReportMultilineDiffOutput(report, os.Stdout); err != nil {
			t.Fatal(err)
		}

		t.Fatal(errors.NewPlain("generated resource dump not equals with expected"))
	}
}

func TestResourceSyncRuleValuesTemplateTransform(t *testing.T) {
	t.Parallel()

	testResourceSyncRuleValuesTemplateTransform(t, icpActiveTestCR, rsrExpectedActiveValues)
	testResourceSyncRuleValuesTemplateTransform(t, icpPassiveTestCR, rsrExpectedPassiveValues)
}

func testResourceSyncRuleValuesTemplateTransform(t *testing.T, icpTestCR, rsrExpectedValues []byte) {
	t.Helper()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	values, err := util.TransformStructToStriMapWithTemplate(icp, assets.ResourceSyncRule, "values.yaml.tpl")
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

	report, err := util.CompareYAMLs(rsrExpectedValues, valuesYaml)
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
