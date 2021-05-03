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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
)

// IstioControlPlaneReconciler reconciles a IstioControlPlane object
type IstioControlPlaneReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes/status,verbs=get;update;patch

func (r *IstioControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("istiocontrolplane", req.NamespacedName)

	icp := &servicemeshv1alpha1.IstioControlPlane{}
	err := r.Get(context.TODO(), req.NamespacedName, icp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if icp.Spec.Version == "" {
		err = errors.New("please set spec.version in your istiocontrolplane CR to be reconciled by this operator")
		logger.Error(err, "", "name", icp.Name, "namespace", icp.Namespace)

		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	if !isIstioVersionSupported(icp.Spec.Version) {
		err = errors.New("intended Istio version is unsupported by this version of the operator")
		logger.Error(err, "", "version", icp.Spec.Version)

		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	return ctrl.Result{}, nil
}

func (r *IstioControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha1.IstioControlPlane{}).
		Complete(r)
}
