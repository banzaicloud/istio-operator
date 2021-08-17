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
	"context"
	"fmt"
	"net/http"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	componentName = "meshgateway"
	chartName     = "istio-meshgateway"
	releaseName   = "meshgateway"

	valuesTemplateFileName = "values.yaml.tpl"
)

var _ components.Component = &Reconciler{}

type Reconciler struct {
	helmReconciler *templatereconciler.HelmReconciler
}

func NewChartReconciler(helmReconciler *templatereconciler.HelmReconciler) components.Component {
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
	if mgw, ok := object.(*v1alpha1.MeshGateway); ok {
		return mgw.DeletionTimestamp.IsZero()
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
	return components.UpdateStatus(context.Background(), rec.helmReconciler.GetClient(), object, status, message)
}

func (rec *Reconciler) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
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

func (rec *Reconciler) GetManifest(object runtime.Object) ([]byte, error) {
	content := []byte{}

	releaseData, err := rec.ReleaseData(object)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get release data")
	}

	var parent reconciler.ResourceOwner
	var ok bool
	if parent, ok = object.(reconciler.ResourceOwner); !ok {
		return nil, errors.New("cannot convert object to ResourceOwner interface")
	}

	rbs, err := rec.helmReconciler.GetResourceBuilders(parent, rec, releaseData, false)
	if err != nil {
		return nil, err
	}

	for _, rb := range rbs {
		obj, _, err := rb()
		if err != nil {
			return nil, err
		}

		y, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		content = append(content, []byte(fmt.Sprintf("---\n%s\n", string(y)))...)
	}

	return content, err
}

func (rec *Reconciler) Reconcile(object runtime.Object) (reconcile.Result, error) {
	result, err := rec.helmReconciler.Reconcile(object, rec)
	if err != nil {
		return reconcile.Result{}, err
	}

	return *result, nil
}

func (rec *Reconciler) RegisterWatches(builder *controllerruntime.Builder) {}

func (rec *Reconciler) values(object runtime.Object) (helm.Strimap, error) {
	mgw, ok := object.(*v1alpha1.MeshGateway)
	if !ok {
		return nil, errors.WrapIff(errors.NewPlain("object cannot be converted to a MeshGateway"), "%+v", object)
	}

	values, err := util.TransformICPToStriMapWithTemplate(&v1alpha1.MeshGatewayWithProperties{
		MeshGateway: mgw,
		Properties: v1alpha1.MeshGatewayProperties{
			Revision:              "cp-v110x.istio-system",
			EnablePrometheusMerge: true,
			InjectionTemplate:     "gateway",
		},
	}, assets.MeshGateway, valuesTemplateFileName)
	if err != nil {
		return nil, errors.WrapIff(err, "MeshGateway cannot be converted into a map[string]interface{}")
	}

	return values, nil
}
