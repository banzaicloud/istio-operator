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

package controllers

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

func NewComponentReconciler(r components.Reconciler, newComponentFunc components.NewComponentReconcilerFunc, logger logr.Logger) (components.ComponentReconciler, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	var d discovery.DiscoveryInterface
	if d, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return nil, err
	}

	return newComponentFunc(
		templatereconciler.NewHelmReconcilerWith(
			r.GetClient(),
			r.GetScheme(),
			logger,
			d,
			templatereconciler.WithNativeReconcilerOptions(
				reconciler.NativeReconcilerSetControllerRef(),
			),
			templatereconciler.WithGenericReconcilerOptions(
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateErrorMessageIgnored(),
			),
			templatereconciler.ManageNamespace(false),
		),
	), nil
}
