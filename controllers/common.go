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
	"context"
	"sort"

	"emperror.dev/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
)

func GetHelmReconciler(r components.Reconciler, newChartReconcilerFunc components.NewChartReconcilerFunc) (components.Component, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	var d discovery.DiscoveryInterface
	if d, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return nil, err
	}

	return newChartReconcilerFunc(
		templatereconciler.NewHelmReconciler(r.GetClient(), r.GetScheme(), r.GetLogger().WithName(r.GetName()), d, []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
	), nil
}

func GetRelatedIstioCR(ctx context.Context, c client.Client, key client.ObjectKey) (*v1alpha1.IstioControlPlane, error) {
	icp := &v1alpha1.IstioControlPlane{}

	// try to get specified Istio CR
	if key.Name != "" && key.Namespace != "" {
		err := c.Get(ctx, key, icp)
		if err == nil {
			return icp, nil
		}
		if err != nil {
			return nil, errors.WrapIf(err, "could not get related Istio control plane")
		}
	}

	// get the oldest otherwise for backward compatibility
	var icps v1alpha1.IstioControlPlaneList
	err := c.List(context.TODO(), &icps)
	if err != nil {
		return nil, errors.WrapIf(err, "could not list istio control planes")
	}
	if len(icps.Items) == 0 {
		return nil, errors.New("no Istio control planes were found")
	}

	sort.Sort(v1alpha1.SortableIstioControlPlaneItems(icps.Items))

	icp = &icps.Items[0]
	icp.SetGroupVersionKind(v1alpha1.SchemeBuilder.GroupVersion.WithKind("IstioControlPlane"))

	return icp, nil
}
