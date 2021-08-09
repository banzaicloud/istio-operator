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
	"reflect"
	"testing"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
)

//go:embed testdata/*.tmpl
var testChartValuesTemplate embed.FS

func TestTransformICPSpecToStriMapWithTemplate(t *testing.T) {
	t.Parallel()
	icp := &v1alpha1.IstioControlPlane{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &v1alpha1.IstioControlPlaneSpec{
			Istiod: &v1alpha1.IstiodConfiguration{
				Deployment: &v1alpha1.BaseKubernetesResourceConfig{
					Resources: &v1alpha1.ResourceRequirements{},
					Metadata: &v1alpha1.K8SObjectMeta{
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: map[string]string{
							"c": "d",
							"e": "f",
						},
					},
				},
			},
		},
	}
	icp.Spec.Version = "1.10"
	icp.Spec.Istiod.Deployment.Resources.Requests = make(map[string]*v1alpha1.Quantity)
	icp.Spec.Istiod.Deployment.Resources.Requests["memory"] = &v1alpha1.Quantity{
		Quantity: resource.MustParse("500m"),
	}

	values, err := util.TransformICPToStriMapWithTemplate(icp, testChartValuesTemplate, "testdata/test-chart-values.yaml.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]interface{}{
		"enableAnalysis": true,
		"istioNamespace": "default",
		"istiod": map[string]interface{}{
			"deploymentLabels": map[string]interface{}{
				"c": "d",
				"e": "f",
			},
		},
		"param1": "value",
		"param2": map[string]interface{}{
			"level2": "value",
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": "500m",
			},
		},
		"version": string("1.10"),
	}

	if !reflect.DeepEqual(values, expected) {
		t.Fatal(errors.Errorf("%#v != %#v", values, expected))
	}
}
