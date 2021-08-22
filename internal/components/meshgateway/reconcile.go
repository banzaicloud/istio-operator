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

package meshgateway

import (
	"net/http"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
)

const (
	componentName = "meshgateway"
	chartName     = "istio-meshgateway"
	releaseName   = "meshgateway"

	valuesTemplateFileName = "values.yaml.tpl"
)

var _ components.MinimalComponent = &Component{}

type Component struct {
	properties v1alpha1.MeshGatewayProperties
}

func NewChartReconciler(helmReconciler *components.HelmReconciler, properties v1alpha1.MeshGatewayProperties) components.ComponentReconciler {
	return &components.Base{
		HelmReconciler: helmReconciler,
		Component: &Component{
			properties: properties,
		},
	}
}

func (rec *Component) Name() string {
	return componentName
}

func (rec *Component) Enabled(object runtime.Object) bool {
	if mgw, ok := object.(*v1alpha1.MeshGateway); ok {
		return mgw.DeletionTimestamp.IsZero()
	}

	return true
}

func (rec *Component) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	if mgw, ok := object.(*v1alpha1.MeshGateway); ok {
		values, err := rec.values(object)
		if err != nil {
			return nil, err
		}

		return &templatereconciler.ReleaseData{
			Chart:       http.FS(assets.MeshGateway),
			Values:      values,
			Namespace:   mgw.Namespace,
			ChartName:   chartName,
			ReleaseName: releaseName,
		}, nil
	}

	return nil, errors.WrapIff(errors.NewPlain("could not prepare release data: invalid object"), "%+v", object)
}

func (rec *Component) values(object runtime.Object) (helm.Strimap, error) {
	mgw, ok := object.(*v1alpha1.MeshGateway)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to a MeshGateway"), "%+v", object)
	}

	obj := &v1alpha1.MeshGatewayWithProperties{
		MeshGateway: mgw,
		Properties:  rec.properties,
	}
	obj.SetDefaults()

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.MeshGateway, valuesTemplateFileName)
	if err != nil {
		return nil, errors.WrapIff(err, "MeshGateway cannot be converted into a map[string]interface{}")
	}

	return values, nil
}
