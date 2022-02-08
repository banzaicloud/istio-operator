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
package discovery_test

import (
	_ "embed"
	"fmt"
	"os"
	"testing"
	"time"

	"emperror.dev/errors"
	"emperror.dev/errors/utils/keyval"
	"github.com/go-logr/logr"
	testlogr "github.com/go-logr/logr/testing"
	"github.com/gogo/protobuf/types"
	"github.com/homeport/dyff/pkg/dyff"
	istio_mesh_v1alpha1 "istio.io/api/mesh/v1alpha1"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	assets "github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components/discovery"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

//go:embed testdata/icp-test-cr.yaml
var icpTestCR []byte

//go:embed testdata/icp-expected-values.yaml
var icpExpectedValues []byte

//go:embed testdata/icp-expected-resource-dump.yaml
var icpExpectedResourceDump []byte

//go:embed testdata/icp-passive-test-cr.yaml
var icpPassiveTestCR []byte

//go:embed testdata/icp-passive-expected-values.yaml
var icpPassiveExpectedValues []byte

//go:embed testdata/icp-passive-expected-resource-dump.yaml
var icpPassiveExpectedResourceDump []byte

func TestICPDiscoveryResourceDump(t *testing.T) {
	t.Parallel()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	reconciler := discovery.NewChartReconciler(
		templatereconciler.NewHelmReconciler(nil, nil, testlogr.TestLogger{
			T: t,
		}, fake.NewSimpleClientset().Discovery(), []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
		v1alpha1.IstioControlPlaneProperties{
			Mesh: &v1alpha1.IstioMesh{
				Spec: &v1alpha1.IstioMeshSpec{
					Config: &istio_mesh_v1alpha1.MeshConfig{
						ConnectTimeout: types.DurationProto(5 * time.Second),
					},
				},
			},
			MeshNetworks:                 getTestMeshNetworks(),
			TrustedRootCACertificatePEMs: []string{"<pem content from peer>"},
		},
		testlogr.TestLogger{
			T: t,
		},
	)

	dd, err := reconciler.GetManifest(icp)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(icpExpectedResourceDump, dd)
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

func getTestMeshNetworks() *istio_mesh_v1alpha1.MeshNetworks {
	return &istio_mesh_v1alpha1.MeshNetworks{
		Networks: map[string]*istio_mesh_v1alpha1.Network{
			"network1": {
				Endpoints: []*istio_mesh_v1alpha1.Network_NetworkEndpoints{
					{
						Ne: &istio_mesh_v1alpha1.Network_NetworkEndpoints_FromRegistry{
							FromRegistry: "demo-cluster2",
						},
					},
				},
				Gateways: []*istio_mesh_v1alpha1.Network_IstioNetworkGateway{
					{
						Gw: &istio_mesh_v1alpha1.Network_IstioNetworkGateway_Address{
							Address: "127.0.0.1",
						},
						Port:     uint32(15443),
						Locality: "us-east-1a",
					},
				},
			},
		},
	}
}

func TestICPDiscoveryValuesTemplateTransform(t *testing.T) {
	t.Parallel()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	obj := v1alpha1.IstioControlPlaneWithProperties{
		IstioControlPlane: icp,
		Properties: v1alpha1.IstioControlPlaneProperties{
			Mesh: &v1alpha1.IstioMesh{
				Spec: &v1alpha1.IstioMeshSpec{
					Config: &istio_mesh_v1alpha1.MeshConfig{
						ConnectTimeout: types.DurationProto(5 * time.Second),
					},
				},
			},
			MeshNetworks: getTestMeshNetworks(),
		},
	}

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.DiscoveryChart, "values.yaml.tpl")
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

	report, err := util.CompareYAMLs(icpExpectedValues, valuesYaml)
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

func TestPassiveICPDiscoveryResourceDump(t *testing.T) {
	t.Parallel()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpPassiveTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	reconciler := discovery.NewChartReconciler(
		templatereconciler.NewHelmReconciler(nil, nil, testlogr.TestLogger{
			T: t,
		}, fake.NewSimpleClientset().Discovery(), []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
		v1alpha1.IstioControlPlaneProperties{
			Mesh: &v1alpha1.IstioMesh{
				Spec: &v1alpha1.IstioMeshSpec{
					Config: &istio_mesh_v1alpha1.MeshConfig{
						ConnectTimeout: types.DurationProto(5 * time.Second),
					},
				},
			},
		},
		logr.DiscardLogger{},
	)

	dd, err := reconciler.GetManifest(icp)
	if err != nil {
		t.Fatal(err)
	}

	report, err := util.CompareYAMLs(icpPassiveExpectedResourceDump, dd)
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

		util.DyffReportMultilineDiffOutput(report, os.Stdout)

		t.Fatal(errors.NewPlain("generated resource dump not equals with expected"))
	}
}

func TestPassiveICPDiscoveryValuesTemplateTransform(t *testing.T) {
	t.Parallel()

	var icp *v1alpha1.IstioControlPlane
	if err := yaml.Unmarshal(icpPassiveTestCR, &icp); err != nil {
		t.Fatal(err)
	}

	obj := v1alpha1.IstioControlPlaneWithProperties{
		IstioControlPlane: icp,
		Properties: v1alpha1.IstioControlPlaneProperties{
			Mesh: &v1alpha1.IstioMesh{
				Spec: &v1alpha1.IstioMeshSpec{
					Config: &istio_mesh_v1alpha1.MeshConfig{
						ConnectTimeout: types.DurationProto(5 * time.Second),
					},
				},
			},
		},
	}

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.DiscoveryChart, "values.yaml.tpl")
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

	report, err := util.CompareYAMLs(icpPassiveExpectedValues, valuesYaml)
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

		util.DyffReportMultilineDiffOutput(report, os.Stdout)

		t.Fatal(errors.NewPlain("generated template values not equals with expected"))
	}
}
