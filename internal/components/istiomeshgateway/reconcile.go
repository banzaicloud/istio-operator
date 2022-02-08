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

package istiomeshgateway

import (
	"net/http"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

const (
	componentName = "istio-meshgateway"
	chartName     = "istio-meshgateway"
	releaseName   = "istio-meshgateway"

	valuesTemplateFileName = "values.yaml.tpl"
)

var _ components.MinimalComponent = &Component{}

type Component struct {
	properties v1alpha1.IstioMeshGatewayProperties
	logger     logr.Logger
}

func NewChartReconciler(helmReconciler *components.HelmReconciler, properties v1alpha1.IstioMeshGatewayProperties, logger logr.Logger) components.ComponentReconciler {
	return &components.Base{
		HelmReconciler: helmReconciler,
		Component: &Component{
			properties: properties,
			logger:     logger,
		},
	}
}

func (rec *Component) Name() string {
	return componentName
}

func (rec *Component) Enabled(object runtime.Object) bool {
	if imgw, ok := object.(*v1alpha1.IstioMeshGateway); ok {
		return imgw.DeletionTimestamp.IsZero()
	}

	return true
}

func (rec *Component) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	if imgw, ok := object.(*v1alpha1.IstioMeshGateway); ok {
		values, err := rec.values(object)
		if err != nil {
			return nil, err
		}

		overlays, err := util.ConvertK8sOverlays(imgw.GetSpec().GetK8SResourceOverlays())
		if err != nil {
			return nil, errors.WrapIf(err, "could not convert k8s resource overlays")
		}

		return &templatereconciler.ReleaseData{
			Chart:       http.FS(assets.IstioMeshGateway),
			Values:      values,
			Namespace:   imgw.Namespace,
			ChartName:   chartName,
			ReleaseName: releaseName,
			Layers:      overlays,
			DesiredStateOverrides: map[reconciler.ObjectKeyWithGVK]reconciler.DesiredState{
				{
					GVK: policyv1beta1.SchemeGroupVersion.WithKind("PodDisruptionBudget"),
				}: reconciler.DynamicDesiredState{
					ShouldUpdateFunc: func(current, desired runtime.Object) (bool, error) {
						options := []patch.CalculateOption{
							patch.IgnoreStatusFields(),
							reconciler.IgnoreManagedFields(),
							patch.IgnorePDBSelector(),
						}

						patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, options...)
						if err != nil {
							rec.logger.Error(err, "could not calculate patch result")

							return false, err
						}

						return !patchResult.IsEmpty(), nil
					},
				},
			},
		}, nil
	}

	return nil, errors.WrapIff(errors.NewPlain("could not prepare release data: invalid object"), "%+v", object)
}

func (rec *Component) values(object runtime.Object) (helm.Strimap, error) {
	imgw, ok := object.(*v1alpha1.IstioMeshGateway)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to a IstioMeshGateway"), "%+v", object)
	}

	obj := &v1alpha1.IstioMeshGatewayWithProperties{
		IstioMeshGateway: imgw,
		Properties:       rec.properties,
	}
	obj.SetDefaults()

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.IstioMeshGateway, valuesTemplateFileName)
	if err != nil {
		return nil, errors.WrapIff(err, "IstioMeshGateway cannot be converted into a map[string]interface{}")
	}

	return values, nil
}
