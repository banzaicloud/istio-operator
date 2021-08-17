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

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/types"
)

type (
	HelmReconciler         = templatereconciler.HelmReconciler
	NewChartReconcilerFunc = func(helmReconciler *HelmReconciler) Component
)

type Component interface {
	templatereconciler.Component
	Reconcile(object runtime.Object) (reconcile.Result, error)
	GetManifest(object runtime.Object) ([]byte, error)
}

type Reconciler interface {
	GetClient() client.Client
	GetScheme() *runtime.Scheme
	GetLogger() logr.Logger
	GetName() string
}

type ObjectWithStatus interface {
	client.Object
	SetStatus(status v1alpha1.ConfigState, errorMessage string)
	GetStatus() interface{}
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
	if mgw, ok := object.(ObjectWithStatus); ok {
		current := &unstructured.Unstructured{}
		current.SetGroupVersionKind(mgw.GetObjectKind().GroupVersionKind())
		err := c.Get(context.TODO(), client.ObjectKey{
			Namespace: mgw.GetNamespace(),
			Name:      mgw.GetName(),
		}, current)
		if err != nil {
			return errors.WrapIf(err, "could not get resource for updating status")
		}

		patch := client.MergeFrom(current)
		mgw.SetResourceVersion(current.GetResourceVersion())
		mgw.SetStatus(ConvertReconcileStatusToConfigState(status), message)

		return c.Status().Patch(context.Background(), mgw, patch)
	}

	return nil
}
