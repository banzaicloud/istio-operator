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
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Log              logr.Logger
	Scheme           *runtime.Scheme
	watchersInitOnce sync.Once
	builder          *ctrlBuilder.Builder
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups="policy",resources=podsecuritypolicies;poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="security.istio.io",resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="networking.istio.io",resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete
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

	r.watchersInitOnce.Do(func() {
		err = r.watchIstioCRs()
		if err != nil {
			logger.Error(err, "unable to watch Istio Custom Resources")
		}
	})

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
	r.builder = ctrl.NewControllerManagedBy(mgr)

	return r.builder.
		For(&servicemeshv1alpha1.IstioControlPlane{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IstioControlPlane",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		}).
		Owns(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DaemonSet",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&policyv1beta1.PodSecurityPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodSecurityPolicy",
				APIVersion: policyv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&policyv1beta1.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policyv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&rbacv1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRole",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{
			TypeMeta: metav1.TypeMeta{
				Kind:       "HorizontalPodAutoscaler",
				APIVersion: autoscalingv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MutatingWebhookConfiguration",
				APIVersion: admissionregistrationv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&admissionregistrationv1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ValidatingWebhookConfiguration",
				APIVersion: admissionregistrationv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Complete(r)
}

func (r *IstioControlPlaneReconciler) watchIstioCRs() error {
	return r.builder.
		Owns(&istioclientv1alpha3.EnvoyFilter{
			TypeMeta: metav1.TypeMeta{
				Kind:       "EnvoyFilter",
				APIVersion: istioclientv1alpha3.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&istioclientv1beta1.PeerAuthentication{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PeerAuthentication",
				APIVersion: istioclientv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Complete(r)
}
