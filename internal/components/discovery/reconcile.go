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
	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	istio_discovery "github.com/banzaicloud/istio-operator/v2/internal/static/gen/charts/istio-control/istio-discovery"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	ComponentName = "istio-discovery"
	ChartName     = "istio-discovery"
	ReleaseName   = "istio-discovery"
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
	return ComponentName
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
			Chart:       istio_discovery.Chart,
			Values:      values,
			Namespace:   controlPlane.Namespace,
			ChartName:   ChartName,
			ReleaseName: ReleaseName,
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
	if _, ok := object.(*v1alpha1.IstioControlPlane); ok {
		return helm.Strimap{
			"global": helm.Strimap{
				"jwtPolicy": "first-party-jwt",
			},
		}, nil
	}

	return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to an IstioControlPlane"), "%+v", object)
}
