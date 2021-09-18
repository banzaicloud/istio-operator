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
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/components/istiomeshgateway"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

const (
	hostnameSyncWaitDuration      = time.Second * 300
	pendingGatewayRequeueDuration = time.Second * 30
	istioMeshGatewayFinalizerID   = "istio-meshgateway.servicemesh.cisco.com"
)

// IstioMeshGatewayReconciler reconciles a IstioMeshGateway object
type IstioMeshGatewayReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiomeshgateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiomeshgateways/status,verbs=get;update;patch

func (r *IstioMeshGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("istiomeshgateway", req.NamespacedName)

	imgw := &servicemeshv1alpha1.IstioMeshGateway{}
	err := r.Get(ctx, req.NamespacedName, imgw)
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

	icp, err := r.getRelatedIstioControlPlane(ctx, r.GetClient(), imgw, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !isIstioVersionSupported(icp.Spec.Version) {
		return ctrl.Result{}, nil
	}

	err = util.AddFinalizer(r.Client, imgw, istioMeshGatewayFinalizerID)
	if err != nil {
		return ctrl.Result{}, err
	}

	enablePrometheusMerge := true
	if icp.Status.GetMeshConfig().GetEnablePrometheusMerge() != nil {
		enablePrometheusMerge = icp.Status.GetMeshConfig().GetEnablePrometheusMerge().Value
	}

	reconciler, err := NewComponentReconciler(r, func(helmReconciler *components.HelmReconciler) components.ComponentReconciler {
		return istiomeshgateway.NewChartReconciler(helmReconciler, servicemeshv1alpha1.IstioMeshGatewayProperties{
			Revision:              fmt.Sprintf("%s.%s", icp.GetName(), icp.GetNamespace()),
			EnablePrometheusMerge: utils.BoolPointer(enablePrometheusMerge),
			InjectionTemplate:     "gateway",
			InjectionChecksum:     icp.Status.GetChecksums().GetSidecarInjector(),
			MeshConfigChecksum:    icp.Status.GetChecksums().GetMeshConfig(),
			IstioControlPlane:     icp,
		})
	}, r.Log.WithName("istiomeshgateway"))
	if err != nil {
		return ctrl.Result{}, err
	}

	result, err := reconciler.Reconcile(imgw)
	if err != nil {
		return result, errors.WrapIf(err, "could not reconcile istio mesh gateway")
	}

	if result.Requeue {
		result.RequeueAfter = 0

		return result, nil
	}

	result, err = r.setGatewayAddress(ctx, r.GetClient(), imgw, logger, result)
	if err != nil {
		return result, errors.WrapIf(err, "could not set gateway address")
	}

	err = util.RemoveFinalizer(r.Client, imgw, istioMeshGatewayFinalizerID)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *IstioMeshGatewayReconciler) GetClient() client.Client {
	return r.Client
}

func (r *IstioMeshGatewayReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

func (r *IstioMeshGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)

	ctrl, err := builder.
		For(&servicemeshv1alpha1.IstioMeshGateway{}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&policyv1beta1.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policyv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{
			TypeMeta: metav1.TypeMeta{
				Kind:       "HorizontalPodAutoscaler",
				APIVersion: autoscalingv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{
			CalculateOptions: []util.CalculateOption{
				util.IgnoreMetadataAnnotations("autoscaling.alpha.kubernetes.io"),
				patch.IgnoreStatusFields(),
			},
		})).
		Build(r)
	if err != nil {
		return err
	}

	err = ctrl.Watch(&source.Kind{
		Type: &servicemeshv1alpha1.IstioControlPlane{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IstioControlPlane",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		},
	}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
		var icp *servicemeshv1alpha1.IstioControlPlane
		var ok bool
		if icp, ok = a.(*servicemeshv1alpha1.IstioControlPlane); !ok {
			return nil
		}

		imgws := &servicemeshv1alpha1.IstioMeshGatewayList{}
		err := r.Client.List(context.Background(), imgws)
		if err != nil {
			r.Log.Error(err, "could not list istiomeshgateway resources")

			return nil
		}

		resources := make([]reconcile.Request, 0)
		for _, imgw := range imgws.Items {
			if icp.GetName() == imgw.GetSpec().GetIstioControlPlane().GetName() && icp.GetNamespace() == imgw.GetSpec().GetIstioControlPlane().GetNamespace() {
				resources = append(resources, reconcile.Request{
					NamespacedName: client.ObjectKey{
						Name:      imgw.GetName(),
						Namespace: imgw.GetNamespace(),
					},
				})
			}
		}

		return resources
	}), util.ICPInjectorChangePredicate{})
	if err != nil {
		return err
	}

	return nil
}

func (r *IstioMeshGatewayReconciler) getRelatedIstioControlPlane(ctx context.Context, c client.Client, imgw *servicemeshv1alpha1.IstioMeshGateway, logger logr.Logger) (*servicemeshv1alpha1.IstioControlPlane, error) {
	icp := &servicemeshv1alpha1.IstioControlPlane{}

	err := c.Get(ctx, client.ObjectKey{
		Name:      imgw.GetSpec().GetIstioControlPlane().GetName(),
		Namespace: imgw.GetSpec().GetIstioControlPlane().GetNamespace(),
	}, icp)
	if err != nil {
		updateErr := components.UpdateStatus(ctx, c, imgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), err.Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update istio mesh gateway state")

			return nil, errors.WithStack(err)
		}

		return nil, errors.WrapIf(err, "could not get related Istio control plane")
	}

	return icp, nil
}

func (r *IstioMeshGatewayReconciler) getGatewayAddress(imgw *servicemeshv1alpha1.IstioMeshGateway) ([]string, bool, error) {
	var service corev1.Service
	var ips []string
	var hasHostname bool

	err := r.Get(context.Background(), client.ObjectKey{
		Name:      imgw.GetName(),
		Namespace: imgw.GetNamespace(),
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

func (r *IstioMeshGatewayReconciler) setGatewayAddress(ctx context.Context, c client.Client, imgw *servicemeshv1alpha1.IstioMeshGateway, logger logr.Logger, result ctrl.Result) (ctrl.Result, error) {
	var gatewayHasHostname bool
	var err error

	if !imgw.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	imgw.Status.GatewayAddress, gatewayHasHostname, err = r.getGatewayAddress(imgw)
	if err != nil {
		logger.Info(fmt.Sprintf("gateway address pending: %s", err.Error()))
		updateErr := components.UpdateStatus(ctx, c, imgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), errors.Cause(err).Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")

			return result, errors.WithStack(err)
		}

		result.RequeueAfter = pendingGatewayRequeueDuration

		return result, nil
	}

	updateErr := components.UpdateStatus(ctx, c, imgw, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_Available), "")
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
