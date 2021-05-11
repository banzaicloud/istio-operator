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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	istioclientv1alpha3 "github.com/banzaicloud/istio-client-go/pkg/networking/v1alpha3"
	istioclientv1beta1 "github.com/banzaicloud/istio-client-go/pkg/security/v1beta1"
	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components/base"
	discovery_component "github.com/banzaicloud/istio-operator/v2/internal/components/discovery"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
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

	config, err := ctrl.GetConfig()
	if err != nil {
		return ctrl.Result{}, err
	}
	var d discovery.DiscoveryInterface
	if d, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return ctrl.Result{}, err
	}

	baseReconciler := base.NewChartReconciler(
		templatereconciler.NewHelmReconciler(r.Client, r.Scheme, r.Log.WithName("base"), d, []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
	)
	_, err = baseReconciler.Reconcile(icp)
	if err != nil {
		return ctrl.Result{}, err
	}

	discoveryReconciler := discovery_component.NewChartReconciler(
		templatereconciler.NewHelmReconciler(r.Client, r.Scheme, r.Log.WithName("discovery"), d, []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
	)
	_, err = discoveryReconciler.Reconcile(icp)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *IstioControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	predicate := predicate.GenerationChangedPredicate{}
	//  Funcs{
	// 	CreateFunc: func(e event.CreateEvent) bool {
	// 		object, err := meta.Accessor(e.Object)
	// 		if err != nil {
	// 			return false
	// 		}
	// 		if _, ok := object.(*kafkabanzaicloudiov1beta1.KafkaCluster); ok {
	// 			return true
	// 		}
	// 		return false
	// 	},
	// 	UpdateFunc: func(e event.UpdateEvent) bool {
	// 		object, err := meta.Accessor(e.ObjectOld)
	// 		if err != nil {
	// 			return false
	// 		}
	// 		if _, ok := object.(*kafkabanzaicloudiov1beta1.KafkaCluster); !ok {
	// 			ownedObject := false
	// 			for _, onwerRef := range object.GetOwnerReferences() {
	// 				if onwerRef.Kind == kafka.Kind {
	// 					ownedObject = true
	// 					break
	// 				}
	// 			}

	// 			if !ownedObject {
	// 				return false
	// 			}
	// 		}

	// 		patchResult, err := patch.DefaultPatchMaker.Calculate(e.ObjectOld, e.ObjectNew, ignoreStatusFieldsExceptInternalListenerStatus())
	// 		if err != nil {
	// 			r.Log.Error(err, "could not match objects", "kind", e.ObjectOld.GetObjectKind())
	// 		} else if patchResult.IsEmpty() {
	// 			return false
	// 		}

	// 		return true
	// 	},
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha1.IstioControlPlane{}).
		Owns(&appsv1.Deployment{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&appsv1.DaemonSet{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&corev1.ConfigMap{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&corev1.Service{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&policyv1beta1.PodSecurityPolicy{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&rbacv1.Role{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&rbacv1.RoleBinding{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&rbacv1.ClusterRole{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&rbacv1.ClusterRoleBinding{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&corev1.Secret{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&istioclientv1alpha3.EnvoyFilter{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&istioclientv1beta1.PeerAuthentication{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}, ctrlBuilder.WithPredicates(predicate)).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{}, ctrlBuilder.WithPredicates(predicate)).
		Complete(r)
}
