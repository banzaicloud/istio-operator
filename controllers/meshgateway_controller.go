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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/components/meshgateway"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
)

const (
	hostnameSyncWaitDuration      = time.Second * 300
	pendingGatewayRequeueDuration = time.Second * 30
	meshGatewayFinalizerID        = "istio-meshgateway.servicemesh.cisco.com"
)

// MeshGatewayReconciler reconciles a MeshGateway object
type MeshGatewayReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=meshgateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=meshgateways/status,verbs=get;update;patch

func (r *MeshGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("meshgateway", req.NamespacedName)

	mgw := &servicemeshv1alpha1.MeshGateway{}
	err := r.Get(context.TODO(), req.NamespacedName, mgw)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	logger.Info("reconciling")

	icp, err := r.getRelatedIstioControlPlane(ctx, r.GetClient(), mgw, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !isIstioVersionSupported(icp.Spec.Version) {
		return ctrl.Result{}, nil
	}

	err = util.AddFinalizer(r.Client, mgw, meshGatewayFinalizerID)
	if err != nil {
		return ctrl.Result{}, err
	}

	reconciler, err := GetHelmReconciler(r, func(helmReconciler *components.HelmReconciler) components.Component {
		return meshgateway.NewChartReconciler(helmReconciler, servicemeshv1alpha1.MeshGatewayProperties{
			Revision:              fmt.Sprintf("%s.%s", icp.GetName(), icp.GetNamespace()),
			EnablePrometheusMerge: true,
			InjectionTemplate:     "gateway",
		})
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	result, err := reconciler.Reconcile(mgw)
	if err != nil {
		return result, errors.WrapIf(err, "could not reconcile mesh gateway")
	}

	if result.Requeue {
		result.RequeueAfter = 0

		return result, nil
	}

	result, err = r.setGatewayAddress(ctx, r.GetClient(), mgw, logger, result)
	if err != nil {
		return result, errors.WrapIf(err, "could not set gateway address")
	}

	err = util.RemoveFinalizer(r.Client, mgw, istioControlPlaneFinalizerID)
	if err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}

func (r *MeshGatewayReconciler) GetClient() client.Client {
	return r.Client
}

func (r *MeshGatewayReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

func (r *MeshGatewayReconciler) GetLogger() logr.Logger {
	return r.Log
}

func (r *MeshGatewayReconciler) GetName() string {
	return "meshgateway"
}

func (r *MeshGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha1.MeshGateway{}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Complete(r)
}

func (r *MeshGatewayReconciler) getRelatedIstioControlPlane(ctx context.Context, c client.Client, mgw *servicemeshv1alpha1.MeshGateway, logger logr.Logger) (*servicemeshv1alpha1.IstioControlPlane, error) {
	icp := &servicemeshv1alpha1.IstioControlPlane{}

	err := c.Get(ctx, client.ObjectKey{
		Name:      mgw.GetSpec().GetIstioControlPlane().GetName(),
		Namespace: mgw.GetSpec().GetIstioControlPlane().GetNamespace(),
	}, icp)
	if err != nil {
		updateErr := components.UpdateStatus(ctx, c, mgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), err.Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update mesh gateway state")

			return nil, errors.WithStack(err)
		}

		return nil, errors.WrapIf(err, "could not get related Istio control plane")
	}

	return icp, nil
}

func (r *MeshGatewayReconciler) getGatewayAddress(mgw *servicemeshv1alpha1.MeshGateway) ([]string, bool, error) {
	var service corev1.Service
	var ips []string
	var hasHostname bool

	err := r.Get(context.Background(), client.ObjectKey{
		Name:      mgw.GetName(),
		Namespace: mgw.GetNamespace(),
	}, &service)
	if err != nil {
		return nil, hasHostname, err
	}

	ips, hasHostname, err = k8sutil.GetServiceEndpointIPs(service)
	if err != nil {
		return nil, hasHostname, err
	}

	return ips, hasHostname, nil
}

func (r *MeshGatewayReconciler) setGatewayAddress(ctx context.Context, c client.Client, mgw *servicemeshv1alpha1.MeshGateway, logger logr.Logger, result ctrl.Result) (ctrl.Result, error) {
	var gatewayHasHostname bool
	var err error

	mgw.Status.GatewayAddress, gatewayHasHostname, err = r.getGatewayAddress(mgw)
	if err != nil {
		logger.Info(fmt.Sprintf("gateway address pending: %s", err.Error()))
		updateErr := components.UpdateStatus(ctx, c, mgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), errors.Cause(err).Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")

			return result, errors.WithStack(err)
		}

		result.RequeueAfter = pendingGatewayRequeueDuration

		return result, nil
	}

	updateErr := components.UpdateStatus(ctx, c, mgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_Available), "")
	if updateErr != nil {
		logger.Error(updateErr, "failed to update state")

		return result, errors.WithStack(err)
	}

	if gatewayHasHostname {
		logger.Info(fmt.Sprintf("gateway uses hostname, trigger reconciliation after %s", hostnameSyncWaitDuration.String()))
		result.RequeueAfter = hostnameSyncWaitDuration
	}

	return result, nil
}
