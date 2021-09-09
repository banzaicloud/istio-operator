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

package components

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

type (
	HelmReconciler             = templatereconciler.HelmReconciler
	NewComponentReconcilerFunc = func(helmReconciler *HelmReconciler) ComponentReconciler
)

type MinimalComponent interface {
	Name() string
	Enabled(runtime.Object) bool
	ReleaseData(runtime.Object) (*templatereconciler.ReleaseData, error)
}

type ComponentReconciler interface {
	templatereconciler.Component
	Reconcile(object runtime.Object) (reconcile.Result, error)
	GetManifest(object runtime.Object) ([]byte, error)
	GetHelmReconciler() *HelmReconciler
}

type Reconciler interface {
	GetClient() client.Client
	GetScheme() *runtime.Scheme
}

type ObjectWithStatus interface {
	client.Object
	SetStatus(status v1alpha1.ConfigState, errorMessage string)
}

type Base struct {
	HelmReconciler *HelmReconciler
	Component      MinimalComponent
}

func (rec *Base) Reconcile(object runtime.Object) (reconcile.Result, error) {
	result, err := rec.GetHelmReconciler().Reconcile(object, rec)
	if err != nil {
		return reconcile.Result{}, err
	}

	return *result, nil
}

func (rec *Base) ReleaseData(object runtime.Object) (*templatereconciler.ReleaseData, error) {
	return rec.Component.ReleaseData(object)
}

func (rec *Base) Name() string {
	return rec.Component.Name()
}

func (rec *Base) GetHelmReconciler() *HelmReconciler {
	return rec.HelmReconciler
}

func (rec *Base) GetManifest(object runtime.Object) ([]byte, error) {
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

	rbs, err := rec.GetHelmReconciler().GetResourceBuilders(parent, rec, releaseData, false)
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

func (rec *Base) IsOptional() bool {
	if c, ok := rec.Component.(interface {
		IsOptional() bool
	}); ok {
		return c.IsOptional()
	}

	return false
}

func (rec *Base) UpdateStatus(object runtime.Object, status types.ReconcileStatus, message string) error {
	if c, ok := rec.Component.(interface {
		UpdateStatus(object runtime.Object, status types.ReconcileStatus, message string) error
	}); ok {
		return c.UpdateStatus(object, status, message)
	}

	return UpdateStatus(context.Background(), rec.GetHelmReconciler().GetClient(), object, status, message)
}

func (rec *Base) Skipped(object runtime.Object) bool {
	if c, ok := rec.Component.(interface {
		Skipped(object runtime.Object) bool
	}); ok {
		return c.Skipped(object)
	}

	return false
}

func (rec *Base) Enabled(object runtime.Object) bool {
	if c, ok := rec.Component.(interface {
		Enabled(object runtime.Object) bool
	}); ok {
		return c.Enabled(object)
	}

	return true
}

func (rec *Base) PreChecks(object runtime.Object) error {
	if c, ok := rec.Component.(interface {
		PreChecks(object runtime.Object) error
	}); ok {
		return c.PreChecks(object)
	}

	return nil
}

func ConvertReconcileStatusToConfigState(status types.ReconcileStatus) v1alpha1.ConfigState {
	switch status {
	case types.ReconcileStatusReconciling:
		return v1alpha1.ConfigState_Reconciling
	case types.ReconcileStatusSucceeded, types.ReconcileStatusAvailable:
		return v1alpha1.ConfigState_Available
	case types.ReconcileStatusFailed:
		return v1alpha1.ConfigState_ReconcileFailed
	case types.ReconcileStatusUnmanaged:
		return v1alpha1.ConfigState_Unmanaged
	case types.ReconcileStatusPending, types.ReconcileStatusRemoved:
		return v1alpha1.ConfigState_Unspecified
	}

	return v1alpha1.ConfigState_Unspecified
}

func ConvertConfigStateToReconcileStatus(state v1alpha1.ConfigState) types.ReconcileStatus {
	switch state {
	case v1alpha1.ConfigState_Created:
		return types.ReconcileStatus(v1alpha1.ConfigState_Created.String())
	case v1alpha1.ConfigState_ReconcileFailed:
		return types.ReconcileStatusFailed
	case v1alpha1.ConfigState_Reconciling:
		return types.ReconcileStatusReconciling
	case v1alpha1.ConfigState_Available:
		return types.ReconcileStatusAvailable
	case v1alpha1.ConfigState_Unmanaged:
		return types.ReconcileStatusUnmanaged
	case v1alpha1.ConfigState_Unspecified:
		return types.ReconcileStatus(v1alpha1.ConfigState_Unspecified.String())
	}

	return types.ReconcileStatus(v1alpha1.ConfigState_Unspecified.String())
}

func UpdateStatus(ctx context.Context, c client.Client, object runtime.Object, status types.ReconcileStatus, message string) error {
	if imgw, ok := object.(ObjectWithStatus); ok {
		current := &unstructured.Unstructured{}
		current.SetGroupVersionKind(imgw.GetObjectKind().GroupVersionKind())
		err := c.Get(context.TODO(), client.ObjectKey{
			Namespace: imgw.GetNamespace(),
			Name:      imgw.GetName(),
		}, current)
		if err != nil {
			return errors.WrapIf(err, "could not get resource for updating status")
		}

		patch := client.MergeFrom(current)
		imgw.SetResourceVersion(current.GetResourceVersion())
		imgw.SetStatus(ConvertReconcileStatusToConfigState(status), message)

		return c.Status().Patch(context.Background(), imgw, patch)
	}

	return nil
}
