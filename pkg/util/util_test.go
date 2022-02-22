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

package util_test

import (
	"testing"

	"gotest.tools/v3/assert"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	"github.com/banzaicloud/istio-operator/v2/pkg/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

var envoyFilter = &v1alpha3.EnvoyFilter{
	Spec: networkingv1alpha3.EnvoyFilter{
		ConfigPatches: []*networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
			{
				ApplyTo: networkingv1alpha3.EnvoyFilter_CLUSTER,
			},
		},
	},
}

func TestUpstreamPatchMaker(t *testing.T) {
	t.Parallel()

	desired := envoyFilter.DeepCopy()
	desired.Spec.ConfigPatches[0].ApplyTo = networkingv1alpha3.EnvoyFilter_HTTP_FILTER

	maker := patch.DefaultPatchMaker
	_, err := maker.Calculate(envoyFilter, desired)
	assert.Error(t, err, "Failed to generate strategic merge patch: unable to find api field in struct EnvoyFilter for the json field \"configPatches\"")
}

func TestProtoCompatiblePatchMaker(t *testing.T) {
	t.Parallel()

	desired := envoyFilter.DeepCopy()
	desired.Spec.ConfigPatches[0].ApplyTo = networkingv1alpha3.EnvoyFilter_HTTP_FILTER

	maker := util.NewProtoCompatiblePatchMaker()
	r, err := maker.Calculate(envoyFilter, desired)
	assert.NilError(t, err)

	assert.Equal(t, string(r.Patch), `{"spec":{"configPatches":[{"applyTo":"HTTP_FILTER"}]}}`)
}
