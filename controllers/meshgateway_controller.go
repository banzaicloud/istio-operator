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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components/meshgateway"
	"github.com/banzaicloud/operator-tools/pkg/helm/templatereconciler"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
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
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if mgw.Spec.IstioControlPlane == nil {
		err = errors.New("please set spec.istiocontrolplane in your meshgateway CR to be reconciled by this operator")
		logger.Error(err, "", "name", mgw.Name, "namespace", mgw.Namespace)

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

	meshGatewayReconciler := meshgateway.NewChartReconciler(
		templatereconciler.NewHelmReconciler(r.Client, r.Scheme, r.Log.WithName("meshGateway"), d, []reconciler.NativeReconcilerOpt{
			reconciler.NativeReconcilerSetControllerRef(),
		}),
	)
	_, err = meshGatewayReconciler.Reconcile(mgw)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MeshGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha1.MeshGateway{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MeshGateway",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		}).
		Owns(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Owns(&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(reconciler.SpecChangePredicate{})).
		Complete(r)
}
