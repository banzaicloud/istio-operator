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

package base

import (
	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/static/gen/charts/base"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	componentName = "base"
	chartName     = "base"
	releaseName   = "istio-operator-base"

	valuesTemplatePath     = "internal/components/base"
	valuesTemplateFileName = "values.tmpl"
)

var _ templatereconciler.Component = &Reconciler{}

type Reconciler struct {
	helmReconciler *templatereconciler.HelmReconciler
}

func NewChartReconciler(helmReconciler *templatereconciler.HelmReconciler) *Reconciler {
	return &Reconciler{
		helmReconciler: helmReconciler,
	}
}

func (rec *Reconciler) Name() string {
	return componentName
}

func (rec *Reconciler) Skipped(object runtime.Object) bool {
	// controlPlane, ok := object.(*v1alpha1.IstioControlPlane)
	return false
}

func (rec *Reconciler) Enabled(object runtime.Object) bool {
	// controlPlane, ok := object.(*v1alpha1.IstioControlPlane)
	return true
}

func (rec *Reconciler) PreChecks(object runtime.Object) error {
	return nil
}

func (rec *Reconciler) UpdateStatus(object runtime.Object, status types.ReconcileStatus, message string) error {
	return nil
}

func (rec *Reconciler) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	if controlPlane, ok := object.(*v1alpha1.IstioControlPlane); ok {
		values, err := rec.values(object)
		if err != nil {
			return nil, err
		}

		return &templatereconciler.ReleaseData{
			Chart:       base.Chart,
			Values:      values,
			Namespace:   controlPlane.Namespace,
			ChartName:   chartName,
			ReleaseName: releaseName,
		}, nil
	}

	return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
}

func (rec *Reconciler) Reconcile(object runtime.Object) (*reconcile.Result, error) {
	return rec.helmReconciler.Reconcile(object, rec)
}

func (rec *Reconciler) IsOptional() bool {
	return true
}

func (rec *Reconciler) RegisterWatches(builder *controllerruntime.Builder) {}

func (rec *Reconciler) values(object runtime.Object) (helm.Strimap, error) {
	icp, ok := object.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
	}

	values, err := util.TransformICPSpecToStriMapWithTemplate(
		icp.Spec,
		valuesTemplatePath,
		valuesTemplateFileName,
	)
	if err != nil {
		return nil, errors.WrapIff(errors.NewPlain("IstioControlPlane spec cannot be converted into a map[string]interface{}"), "%+v", icp.Spec)
	}

	return values, nil
}
