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
	"net/http"

	"emperror.dev/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	componentName = "base"
	chartName     = "base"
	releaseName   = "istio-operator-base"

	valuesTemplateFileName = "values.yaml.tmpl"
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
	return false
}

func (rec *Reconciler) Enabled(object runtime.Object) bool {
	if controlPlane, ok := object.(*v1alpha1.IstioControlPlane); ok {
		return controlPlane.DeletionTimestamp.IsZero()
	}

	return true
}

func (rec *Reconciler) IsOptional() bool {
	return false
}

func (rec *Reconciler) PreChecks(object runtime.Object) error {
	return nil
}

func (rec *Reconciler) UpdateStatus(object runtime.Object, status types.ReconcileStatus, message string) error {
	return nil
}

func (rec *Reconciler) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	icp, ok := object.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
	}

	values, err := rec.values(object)
	if err != nil {
		return nil, err
	}

	return &templatereconciler.ReleaseData{
		Chart:       http.FS(assets.BaseChart),
		Values:      values,
		Namespace:   icp.Namespace,
		ChartName:   chartName,
		ReleaseName: releaseName,
		DesiredStateOverrides: map[reconciler.ObjectKeyWithGVK]reconciler.DesiredState{
			{
				GVK: apiextensionv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			}: reconciler.DynamicDesiredState{
				DesiredState: reconciler.StateCreated,
				BeforeCreateFunc: func(desired runtime.Object) error {
					if o, ok := desired.(client.Object); ok {
						annotations := o.GetAnnotations()
						delete(annotations, types.BanzaiCloudManagedComponent)
						delete(annotations, types.BanzaiCloudRelatedTo)
						o.SetAnnotations(annotations)
					}

					return nil
				},
			},
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
	}, nil
}

func (rec *Reconciler) Reconcile(object runtime.Object) (*reconcile.Result, error) {
	return rec.helmReconciler.Reconcile(object, rec)
}

func (rec *Reconciler) RegisterWatches(builder *controllerruntime.Builder) {}

func (rec *Reconciler) values(object runtime.Object) (helm.Strimap, error) {
	icp, ok := object.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
	}

	values, err := util.TransformStructToStriMapWithTemplate(icp, assets.BaseChart, valuesTemplateFileName)
	if err != nil {
		return nil, errors.WrapIff(err, "IstioControlPlane spec cannot be converted into a map[string]interface{}")
	}

	return values, nil
}
