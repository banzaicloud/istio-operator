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

package discovery

import (
	"net/http"

	"emperror.dev/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	assets "github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

const (
	componentName = "istio-discovery"
	chartName     = "istio-discovery"
	releaseName   = "istio-operator-discovery"

	valuesTemplateFileName = "values.yaml.tpl"
)

var _ components.MinimalComponent = &Component{}

type Component struct {
	properties v1alpha1.IstioControlPlaneProperties
}

func NewChartReconciler(helmReconciler *templatereconciler.HelmReconciler, properties v1alpha1.IstioControlPlaneProperties) components.ComponentReconciler {
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
	if controlPlane, ok := object.(*v1alpha1.IstioControlPlane); ok {
		return controlPlane.DeletionTimestamp.IsZero()
	}

	return true
}

func (rec *Component) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	icp, ok := object.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
	}

	values, err := rec.values(object)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	overlays, err := util.ConvertK8sOverlays(icp.GetSpec().GetK8SResourceOverlays())
	if err != nil {
		return nil, errors.WrapIf(err, "could not convert k8s resource overlays")
	}

	return &templatereconciler.ReleaseData{
		Chart:       http.FS(assets.DiscoveryChart),
		Values:      values,
		Namespace:   icp.Namespace,
		ChartName:   chartName,
		ReleaseName: releaseName,
		DesiredStateOverrides: map[reconciler.ObjectKeyWithGVK]reconciler.DesiredState{
			{
				GVK: admissionregistrationv1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration"),
			}: reconciler.DynamicDesiredState{
				ShouldUpdateFunc: func(current, desired runtime.Object) (bool, error) {
					options := []patch.CalculateOption{
						patch.IgnoreStatusFields(),
						reconciler.IgnoreManagedFields(),
						util.IgnoreWebhookFailurePolicy(),
					}

					patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, options...)
					if err != nil {
						return false, err
					}

					return !patchResult.IsEmpty(), nil
				},
			},
		},
		Layers: overlays,
	}, nil
}

func (rec *Component) values(object runtime.Object) (helm.Strimap, error) {
	icp, ok := object.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
	}

	obj := &v1alpha1.IstioControlPlaneWithProperties{
		IstioControlPlane: icp,
		Properties:        rec.properties,
	}

	values, err := util.TransformStructToStriMapWithTemplate(obj, assets.DiscoveryChart, valuesTemplateFileName)
	if err != nil {
		return nil, errors.WrapIff(err, "IstioControlPlane spec cannot be converted into a map[string]interface{}: %+v", icp.Spec)
	}

	return values, nil
}
