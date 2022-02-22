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
	"fmt"
	"net/http"

	"emperror.dev/errors"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/assets"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/logger"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

const (
	componentName = "base"
	chartName     = "base"
	releaseName   = "istio-operator-base"

	managedByValue = "istio-operator"

	valuesTemplateFileName = "values.yaml.tpl"
)

var _ components.MinimalComponent = &Component{}

type Component struct {
	logger                logger.Logger
	supportedIstioVersion string
}

func NewComponentReconciler(helmReconciler *templatereconciler.HelmReconciler, logger logger.Logger, supportedIstioVersion string) components.ComponentReconciler {
	return &components.Base{
		HelmReconciler: helmReconciler,
		Component: &Component{
			logger:                logger,
			supportedIstioVersion: supportedIstioVersion,
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
		return nil, err
	}

	prepareCRDFunc := func(obj client.Object) {
		annotations := obj.GetAnnotations()
		delete(annotations, types.BanzaiCloudManagedComponent)
		delete(annotations, types.BanzaiCloudRelatedTo)
		obj.SetAnnotations(annotations)
		k8sutil.SetResourceRevisionLabel(obj, rec.supportedIstioVersion)
		k8sutil.SetManagedByLabel(obj, managedByValue)
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
				DesiredState: reconciler.StatePresent,
				BeforeCreateFunc: func(desired runtime.Object) error {
					if o, ok := desired.(client.Object); ok {
						prepareCRDFunc(o)
					}

					return nil
				},
				BeforeUpdateFunc: func(current, desired runtime.Object) error {
					if o, ok := desired.(client.Object); ok {
						prepareCRDFunc(o)
					}

					return nil
				},
				ShouldUpdateFunc: func(current, desired runtime.Object) (bool, error) {
					var obj client.Object
					var ok bool
					if obj, ok = current.(client.Object); !ok {
						return false, errors.New("invalid object")
					}

					if k8sutil.GetManagedByLabel(obj) != managedByValue {
						rec.logger.V(1).Info("current crd is not owned by us, skip update", "name", obj.GetName())

						return false, nil
					}

					ok, err := k8sutil.CheckResourceRevision(obj, fmt.Sprintf("<=%s", rec.supportedIstioVersion))
					if err != nil {
						return false, errors.WithStackIf(err)
					}

					if !ok {
						rec.logger.V(1).Info("current crd is newer, skip update", "name", obj.GetName())
					}

					return ok, nil
				},
			},
		},
	}, nil
}

func (rec *Component) values(object runtime.Object) (helm.Strimap, error) {
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
